package users

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalidUsername  = errors.New("invalid username")
	ErrUsernameTooShort = errors.New("username can't be less than 3 characters")
	ErrUsernameTooLong  = errors.New("username can't be more than 50 characters")
	ErrInvalidEmail     = errors.New("invalid email")
	ErrEmailTooLong     = errors.New("email can't be more than 255 characters")
	ErrPasswordBlank    = errors.New("password can't be blank")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrPasswordTooShort = errors.New("password can't be less than 8 characters")
	ErrPasswordTooLong  = errors.New("password too long")
)

var usernameRegex = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

func ValidateRegistration(req RegisterRequest) error {

	if utf8.RuneCountInString(req.Username) < 3 {
		return ErrUsernameTooShort
	}

	if utf8.RuneCountInString(req.Username) > 50 {
		return ErrUsernameTooLong
	}

	if !usernameRegex.MatchString(req.Username) {
		return ErrInvalidUsername
	}

	if utf8.RuneCountInString(req.Email) > 255 {
		return ErrEmailTooLong
	}

	addr, err := mail.ParseAddress(req.Email)
	if err != nil || addr.Address != req.Email {
		return ErrInvalidEmail
	}

	if strings.TrimSpace(req.Password) == "" {
		return ErrInvalidPassword
	}

	if utf8.RuneCountInString(req.Password) < 8 {
		return ErrPasswordTooShort
	}

	//Used len here because bcrypt truncates password after 72 bytes not characters
	if len(req.Password) > 72 {
		return ErrPasswordTooLong
	}

	return nil
}

func ValidateLogin(req LoginRequest) error {
	if req.Email == "" {
		return ErrInvalidEmail
	}

	if req.Password == "" {
		return ErrInvalidPassword
	}

	if len(req.Password) > 72 {
		return ErrInvalidPassword
	}

	return nil
}
