package users

import (
	"net/http"

	"github.com/immad-qureshi/auction-system/internal/ratelimiter"
)

func (api *UsersAPI) RegisterRoutes(mux *http.ServeMux, strict *ratelimiter.RateLimiter) {
	mux.Handle("POST /register", strict.Middleware(http.HandlerFunc(api.Register)))
	mux.Handle("POST /login", strict.Middleware(http.HandlerFunc(api.Login)))
}
