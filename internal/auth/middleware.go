package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/immad-qureshi/auction-system/internal/response"
)

func (a *Auth) AuthMiddleWare(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			response.JSON(w, http.StatusUnauthorized, response.Body{
				Message: "missing authorization header",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := a.ParseToken(tokenString)
		if err != nil {
			response.JSON(w, http.StatusUnauthorized, response.Body{
				Message: "invalid or expired token",
			})
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)

		next(w, r.WithContext(ctx))

	}
}
