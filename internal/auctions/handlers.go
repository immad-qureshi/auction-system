package auctions

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/immad-qureshi/auction-system/internal/auth"
	"github.com/immad-qureshi/auction-system/internal/response"
)

type AuctionAPI struct {
	Store          AuctionStore
	Authentication *auth.Auth
}

type AuctionStore interface {
	Create(ctx context.Context, sellerID int64, req CreateAuctionRequest) (AuctionResponse, error)
	GetAll(ctx context.Context) ([]AuctionResponse, error)
	Get(ctx context.Context, auctionID int) (AuctionResponse, error)
	PlaceBid(ctx context.Context, amount int64, auctionID int64, bidderID int64) (Bid, error)
	FinalizeExpiredAuctions(ctx context.Context) error
	ActivatePendingAuctions(ctx context.Context) error
}

func NormalizeTitle(title string) string {
	return strings.TrimSpace(title)
}

func NormalizeDescription(description string) string {
	return strings.TrimSpace(description)
}

func (api *AuctionAPI) RunAuctionWorker(ctx context.Context, wg *sync.WaitGroup) {
	ticker := time.NewTicker(time.Second)
	defer wg.Done()

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("auction worker stopped")
			return
		case <-ticker.C:
			if err := api.Store.ActivatePendingAuctions(ctx); err != nil {
				slog.Error("failed to activate auctions", "err", err)
			}

			if err := api.Store.FinalizeExpiredAuctions(ctx); err != nil {
				slog.Error("failed to finalize auctions", "err", err)
			}
		}
	}
}

func (api *AuctionAPI) CreateAuction(w http.ResponseWriter, r *http.Request) {
	var req CreateAuctionRequest

	sellerID, ok := auth.GetUserIDFromContext(r)

	if !ok {
		response.JSON(w, http.StatusUnauthorized, response.Body{
			Message: "user id not found",
		})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		log.Println("CreateAuction : json decoder err : ", err)
		response.JSON(w, http.StatusBadRequest, response.Body{
			Message: "invalid json",
		})
		return
	}

	req.Title = NormalizeTitle(req.Title)
	req.Description = NormalizeDescription(req.Description)

	if err := ValidateAuction(req, time.Now()); err != nil {
		response.JSON(w, http.StatusBadRequest, response.Body{
			Message: err.Error(),
		})
		return
	}

	auction, err := api.Store.Create(r.Context(), sellerID, req)

	if err != nil {
		response.JSON(w, http.StatusInternalServerError, response.Body{
			Message: "something went wrong, try again later",
		})
		return
	}

	response.JSON(w, http.StatusCreated, response.Body{
		Message: "auction successfully created",
		Data:    auction,
	})

}

func (api *AuctionAPI) GetAllAuctions(w http.ResponseWriter, r *http.Request) {

	auctions, err := api.Store.GetAll(r.Context())

	if err != nil {
		response.JSON(w, http.StatusInternalServerError, response.Body{
			Message: "something went wrong, try again later",
		})
		return
	}

	response.JSON(w, http.StatusOK, response.Body{
		Message: "fetched all auctions successfully",
		Data:    auctions,
	})
}

func (api *AuctionAPI) GetAuction(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	auctionID, err := strconv.Atoi(idString)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, response.Body{
			Message: "invalid auction id",
		})
		return
	}

	auction, err := api.Store.Get(r.Context(), auctionID)

	if err != nil {
		if errors.Is(err, ErrAuctionNotFound) {
			response.JSON(w, http.StatusNotFound, response.Body{
				Message: ErrAuctionNotFound.Error(),
			})
			return
		}
		slog.Error("auction search by id failed", "err", err)
		response.JSON(w, http.StatusInternalServerError, response.Body{
			Message: "something went wrong, try again later",
		})
		return
	}

	response.JSON(w, http.StatusOK, response.Body{
		Message: "auction found",
		Data:    auction,
	})
}

func (api *AuctionAPI) PlaceBid(w http.ResponseWriter, r *http.Request) {
	var req PlaceBidRequest

	bidderID, ok := auth.GetUserIDFromContext(r)
	auctionIdString := r.PathValue("id")

	auctionID, err := strconv.Atoi(auctionIdString)

	if err != nil {
		response.JSON(w, http.StatusBadRequest, response.Body{
			Message: "invalid auction id",
		})
		return
	}

	if !ok {
		response.JSON(w, http.StatusUnauthorized, response.Body{
			Message: "user id not found",
		})
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, response.Body{
			Message: "invalid json",
		})
		return
	}

	if err = ValidateBidAmount(req.Amount); err != nil {
		response.JSON(w, http.StatusBadRequest, response.Body{
			Message: ErrInvalidBid.Error(),
		})
		return
	}

	auction, err := api.Store.PlaceBid(r.Context(), req.Amount, int64(auctionID), bidderID)

	if err != nil {
		switch {
		case errors.Is(err, ErrAuctionNotFound):
			response.JSON(w, http.StatusNotFound, response.Body{
				Message: ErrAuctionNotFound.Error(),
			})
		case errors.Is(err, ErrAuctionNotStarted):
			response.JSON(w, http.StatusConflict, response.Body{
				Message: ErrAuctionNotStarted.Error(),
			})
		case errors.Is(err, ErrAuctionFinished):
			response.JSON(w, http.StatusConflict, response.Body{
				Message: ErrAuctionFinished.Error(),
			})
		case errors.Is(err, ErrLowBid):
			response.JSON(w, http.StatusConflict, response.Body{
				Message: ErrLowBid.Error(),
			})
		case errors.Is(err, ErrSellerCannotBid):
			response.JSON(w, http.StatusForbidden, response.Body{
				Message: ErrSellerCannotBid.Error(),
			})
		default:
			slog.Error("bid request failed", "err", err)
			response.JSON(w, http.StatusInternalServerError, response.Body{
				Message: "something went wrong, try again later",
			})
		}
		return
	}

	response.JSON(w, http.StatusOK, response.Body{
		Message: "bid placed",
		Data:    auction,
	})
}
