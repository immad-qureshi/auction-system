package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type PostgresUserStore struct {
	DB *sql.DB
}

var dummyHash = mustHash("dummy-password")

func mustHash(s string) []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	return hash
}

func (p *PostgresUserStore) Create(ctx context.Context, req RegisterRequest) (RegisterResponse, error) {
	var user RegisterResponse

	password_hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	if err != nil {
		return RegisterResponse{}, fmt.Errorf("password hash generation: %w", err)
	}

	err = p.DB.QueryRowContext(ctx, "INSERT INTO users (username,email,password_hash,role) VALUES ($1,$2,$3,$4) RETURNING id,username,email", req.Username, req.Email, password_hash, "User").Scan(&user.UserID, &user.Username, &user.Email)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_username_key":
				return RegisterResponse{}, ErrUsernameTaken
			case "users_email_key":
				return RegisterResponse{}, ErrEmailTaken
			}
		}
		return RegisterResponse{}, fmt.Errorf("insert user: %w", err)
	}

	return user, nil
}

func (p *PostgresUserStore) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	var resp LoginResponse
	var passwordHash string

	err := p.DB.QueryRowContext(ctx, "SELECT id,username,email,password_hash,role FROM users WHERE email=$1", req.Email).Scan(&resp.User.ID, &resp.User.Username, &resp.User.Email, &passwordHash, &resp.User.Role)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password))
			return LoginResponse{}, ErrInvalidEmailOrPassword
		}
		return LoginResponse{}, fmt.Errorf("query user by email: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return LoginResponse{}, ErrInvalidEmailOrPassword
		}
		return LoginResponse{}, fmt.Errorf("compare password hash: %w", err)
	}

	return resp, nil
}
