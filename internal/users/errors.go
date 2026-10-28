package users

import "errors"

var (
	ErrInvalidEmailOrPassword = errors.New("invalid email or password")
	ErrEmailTaken             = errors.New("email already taken")
	ErrUsernameTaken          = errors.New("username already taken")
)
