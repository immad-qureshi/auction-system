package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/immad-qureshi/auction-system/internal/auctions"
	"github.com/immad-qureshi/auction-system/internal/auth"
	"github.com/immad-qureshi/auction-system/internal/ratelimiter"
	"github.com/immad-qureshi/auction-system/internal/users"
	"github.com/joho/godotenv"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func dbConnection() (*sql.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("missing dbURL")
	}
	db, err := sql.Open("pgx", dbURL)

	if err != nil {
		return nil, fmt.Errorf("db open: %w", err)
	}

	err = db.Ping()

	if err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}

	slog.Info("connected successfully to db")
	return db, nil
}

func main() {
	if err := run(); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func run() error {

	var wg sync.WaitGroup

	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf(".env file load: %w", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		return fmt.Errorf("jwtSecret can't be less than 32 bytes")
	}

	authentication := auth.Auth{
		JWTSecret: []byte(jwtSecret),
	}

	db, err := dbConnection()
	if err != nil {
		return fmt.Errorf("db connection: %w", err)
	}
	defer db.Close()
	defer wg.Wait()

	mux := http.NewServeMux()

	ctx, stopCtx := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopCtx()

	httpAddr := os.Getenv("HTTP_SERVER")
	if httpAddr == "" {
		httpAddr = ":8080"
	}

	srv := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	userStore := &users.PostgresUserStore{DB: db}
	userAPI := &users.UsersAPI{Store: userStore, Authentication: &authentication}

	auctionStore := &auctions.PostgresAuctionStore{DB: db}
	auctionAPI := &auctions.AuctionAPI{Store: auctionStore, Authentication: &authentication}

	wg.Add(1)
	go auctionAPI.RunAuctionWorker(ctx, &wg)

	strict := ratelimiter.NewRateLimiter(5, time.Minute)
	wg.Add(1)
	go strict.RateLimiterCleanup(ctx, &wg)
	loose := ratelimiter.NewRateLimiter(200, time.Minute)
	wg.Add(1)
	go loose.RateLimiterCleanup(ctx, &wg)

	userAPI.RegisterRoutes(mux, strict)
	auctionAPI.RegisterRoutes(mux, strict, loose)

	errCh := make(chan error, 1)

	go func(errCh chan error) {
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		errCh <- err
	}(errCh)

	select {
	case <-ctx.Done():
		stopCtx()
	case err := <-errCh:
		return fmt.Errorf("server start: %w", err)
	}

	slog.Info("starting graceful shutdown")
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = srv.Shutdown(timeoutCtx)
	if err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	return nil
}
