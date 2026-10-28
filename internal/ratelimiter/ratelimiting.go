package ratelimiter

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/immad-qureshi/auction-system/internal/response"
)

type client struct {
	requests    int
	windowStart time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*client
	limit   int
	window  time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*client),
		limit:   limit,
		window:  window,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		rl.mu.Lock()

		c, exists := rl.clients[ip]
		now := time.Now()

		if !exists || now.Sub(c.windowStart) >= rl.window {
			c = &client{
				requests:    0,
				windowStart: now,
			}

			rl.clients[ip] = c
		}

		if c.requests >= rl.limit {
			rl.mu.Unlock()
			w.Header().Set("Retry-After", "60")
			response.JSON(w, http.StatusTooManyRequests, response.Body{
				Message: "too many requests, try again later",
			})
			return
		}

		c.requests++

		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) RateLimiterCleanup(ctx context.Context, wg *sync.WaitGroup) {
	ticker := time.NewTicker(3 * time.Minute)
	defer wg.Done()
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("rl cleanup worker stopped")
			return
		case <-ticker.C:
			rl.mu.Lock()
			count := 0
			now := time.Now()
			for ip, c := range rl.clients {
				if now.Sub(c.windowStart) >= rl.window {
					delete(rl.clients, ip)
					count++
				}
			}
			if count > 0 {
				slog.Info("ip entries removed from rl map", "count", count)
			}
			rl.mu.Unlock()
		}
	}
}
