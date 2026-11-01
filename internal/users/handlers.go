package users

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/immad-qureshi/auction-system/internal/auth"
	"github.com/immad-qureshi/auction-system/internal/response"
)

type UsersAPI struct {
	Store          UserStore
	Authentication *auth.Auth
}

type UserStore interface {
	Create(ctx context.Context, req RegisterRequest) (RegisterResponse, error)
	Login(ctx context.Context, req LoginRequest) (LoginResponse, error)
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func (api *UsersAPI) handleErrors(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidEmailOrPassword):
		response.JSON(w, http.StatusUnauthorized, response.Body{
			Message: ErrInvalidEmailOrPassword.Error(),
		})
	case errors.Is(err, ErrEmailTaken):
		response.JSON(w, http.StatusConflict, response.Body{
			Message: ErrEmailTaken.Error(),
		})
	case errors.Is(err, ErrUsernameTaken):
		response.JSON(w, http.StatusConflict, response.Body{
			Message: ErrUsernameTaken.Error(),
		})
	default:
		slog.Error("request failed", "err", err)
		response.JSON(w, http.StatusInternalServerError, response.Body{
			Message: "something went wrong,try again later",
		})
	}
}

func (api *UsersAPI) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, response.Body{
			Message: "invalid json",
		})
		return
	}

	req.Username = NormalizeUsername(req.Username)
	req.Email = NormalizeEmail(req.Email)

	err := ValidateRegistration(req)

	if err != nil {
		response.JSON(w, http.StatusBadRequest, response.Body{
			Message: err.Error(),
		})
		return
	}

	user, err := api.Store.Create(r.Context(), req)

	if err != nil {
		api.handleErrors(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, response.Body{
		Message: "user created",
		Data:    user,
	})

}

func (api *UsersAPI) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, response.Body{
			Message: "invalid json",
		})
		return
	}

	req.Email = NormalizeEmail(req.Email)

	err := ValidateLogin(req)
	if err != nil {
		response.JSON(w, http.StatusUnauthorized, response.Body{
			Message: err.Error(),
		})
		return
	}

	user, err := api.Store.Login(r.Context(), req)
	if err != nil {
		api.handleErrors(w, err)
		return
	}

	jwtToken, err := api.Authentication.GenerateJWT(user.User.ID, user.User.Username)
	if err != nil {
		slog.Error("login token generation", "err", err)
		response.JSON(w, http.StatusInternalServerError, response.Body{
			Message: "something went wrong, try again later",
		})
		return
	}

	user.Token = jwtToken

	response.JSON(w, http.StatusOK, response.Body{
		Message: "login successful",
		Data:    user,
	})
}
