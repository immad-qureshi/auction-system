package auctions

import (
	"net/http"

	"github.com/immad-qureshi/auction-system/internal/ratelimiter"
)

func (api *AuctionAPI) RegisterRoutes(mux *http.ServeMux, strict *ratelimiter.RateLimiter, loose *ratelimiter.RateLimiter) {
	mux.Handle("POST /auction", strict.Middleware(api.Authentication.AuthMiddleWare(api.CreateAuction)))
	mux.Handle("GET /auctions", loose.Middleware(http.HandlerFunc(api.GetAllAuctions)))
	mux.Handle("GET /auctions/{id}", loose.Middleware(http.HandlerFunc(api.GetAuction)))
	mux.Handle("POST /auctions/{id}/bid", loose.Middleware(api.Authentication.AuthMiddleWare(api.PlaceBid)))
}
