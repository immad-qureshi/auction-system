package users

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateRegistration(t *testing.T) {

	validRequest := RegisterRequest{
		Username: "IamMike",
		Email:    "mike@example.com",
		Password: "mike123456",
	}

	tests := []struct {
		name    string
		modify  func(*RegisterRequest)
		wantErr error
	}{
		{
			name:    "valid request",
			wantErr: nil,
		},
		{
			name: "username below minimum length",
			modify: func(req *RegisterRequest) {
				req.Username = "Jo"
			},
			wantErr: ErrUsernameTooShort,
		},
		{
			name: "valid username at minimum length",
			modify: func(req *RegisterRequest) {
				req.Username = "Joe"
			},
			wantErr: nil,
		},
		{
			name: "username exceeds maximum length",
			modify: func(req *RegisterRequest) {
				req.Username = strings.Repeat("H", 51)
			},
			wantErr: ErrUsernameTooLong,
		},
		{
			name: "valid username at maximum length",
			modify: func(req *RegisterRequest) {
				req.Username = strings.Repeat("H", 50)
			},
			wantErr: nil,
		},
		{
			name: "invalid username",
			modify: func(req *RegisterRequest) {
				req.Username = "Joey@-Mike"
			},
			wantErr: ErrInvalidUsername,
		},
		{
			name: "email exceeds maximum length",
			modify: func(req *RegisterRequest) {
				req.Email = strings.Repeat("a", 246) + "@gmail.com"
			},
			wantErr: ErrEmailTooLong,
		},
		{
			name: "valid email at maximum length",
			modify: func(req *RegisterRequest) {
				req.Email = strings.Repeat("a", 245) + "@gmail.com"
			},
			wantErr: nil,
		},
		{
			name: "invalid email format",
			modify: func(req *RegisterRequest) {
				req.Email = "joey@@gmail.com"
			},
			wantErr: ErrInvalidEmail,
		},
		{
			name: "email contains display name",
			modify: func(req *RegisterRequest) {
				req.Email = "John Doe <john@gmail.com>"
			},
			wantErr: ErrInvalidEmail,
		},
		{
			name: "empty password",
			modify: func(req *RegisterRequest) {
				req.Password = ""
			},
			wantErr: ErrInvalidPassword,
		},
		{
			name: "whitespace only password",
			modify: func(req *RegisterRequest) {
				req.Password = "    "
			},
			wantErr: ErrInvalidPassword,
		},
		{
			name: "password less than minimum length",
			modify: func(req *RegisterRequest) {
				req.Password = "1234567"
			},
			wantErr: ErrPasswordTooShort,
		},
		{
			name: "valid password at minimum length",
			modify: func(req *RegisterRequest) {
				req.Password = "12345678"
			},
			wantErr: nil,
		},
		{
			name: "valid password at maximum length",
			modify: func(req *RegisterRequest) {
				req.Password = strings.Repeat("A", 72)
			},
			wantErr: nil,
		},
		{
			name: "password exceeds maximum length",
			modify: func(req *RegisterRequest) {
				req.Password = strings.Repeat("A", 73)
			},
			wantErr: ErrPasswordTooLong,
		},
	}

	for _, tt := range tests {

		req := validRequest

		t.Run(tt.name, func(t *testing.T) {

			if tt.modify != nil {
				tt.modify(&req)
			}

			got := ValidateRegistration(req)

			if !errors.Is(got, tt.wantErr) {
				t.Errorf("ValidateRegistration() = %v,want %v", got, tt.wantErr)
			}
		})
	}
}

func TestValidateLogin(t *testing.T) {
	validRequest := LoginRequest{
		Email:    "alex@gmail.com",
		Password: "alex12345",
	}

	tests := []struct {
		name    string
		modify  func(*LoginRequest)
		wantErr error
	}{
		{
			name:    "valid request",
			wantErr: nil,
		},
		{
			name: "empty email",
			modify: func(req *LoginRequest) {
				req.Email = ""
			},
			wantErr: ErrInvalidEmailOrPassword,
		},
		{
			name: "valid email at maximum length",
			modify: func(req *LoginRequest) {
				req.Email = strings.Repeat("b", 245) + "@gmail.com"
			},
			wantErr: nil,
		},
		{
			name: "email exceeds maximum length",
			modify: func(req *LoginRequest) {
				req.Email = strings.Repeat("b", 246) + "@gmail.com"
			},
			wantErr: ErrInvalidEmailOrPassword,
		},
		{
			name: "invalid email format",
			modify: func(req *LoginRequest) {
				req.Email = "joey@@gmail.com"
			},
			wantErr: ErrInvalidEmailOrPassword,
		},
		{
			name: "email contains display name",
			modify: func(req *LoginRequest) {
				req.Email = "John Doe <john@gmail.com>"
			},
			wantErr: ErrInvalidEmailOrPassword,
		},
		{
			name: "empty password",
			modify: func(req *LoginRequest) {
				req.Password = ""
			},
			wantErr: ErrInvalidEmailOrPassword,
		},
		{
			name: "whitespace only password",
			modify: func(req *LoginRequest) {
				req.Password = strings.Repeat(" ", 8)
			},
			wantErr: ErrInvalidEmailOrPassword,
		},
		{
			name: "password less than minimum length",
			modify: func(req *LoginRequest) {
				req.Password = "1234567"
			},
			wantErr: ErrInvalidEmailOrPassword,
		},
		{
			name: "valid password at minimum length",
			modify: func(req *LoginRequest) {
				req.Password = "12345678"
			},
			wantErr: nil,
		},
		{
			name: "password exceeds maximum length",
			modify: func(req *LoginRequest) {
				req.Password = strings.Repeat("y", 73)
			},
			wantErr: ErrInvalidEmailOrPassword,
		},
		{
			name: "valid password at maximum length",
			modify: func(req *LoginRequest) {
				req.Password = strings.Repeat("u", 72)
			},
			wantErr: nil,
		},
		{
			name: "password contains 36 characters but is 72 bytes",
			modify: func(req *LoginRequest) {
				req.Password = strings.Repeat("é", 36)
			},
			wantErr: nil,
		},
		{
			name: "password contains 37 characters but is 74 bytes",
			modify: func(req *LoginRequest) {
				req.Password = strings.Repeat("é", 37)
			},
			wantErr: ErrInvalidEmailOrPassword,
		},
	}

	for _, tt := range tests {

		req := validRequest
		t.Run(tt.name, func(t *testing.T) {
			if tt.modify != nil {
				tt.modify(&req)
			}

			got := ValidateLogin(req)

			if !errors.Is(got, tt.wantErr) {
				t.Errorf("ValidateLogin() = %v, want %v", got, tt.wantErr)
			}
		})
	}
}
