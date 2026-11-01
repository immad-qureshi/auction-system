package auctions

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestValidateAuction(t *testing.T) {
	now := time.Now()

	validRequest := CreateAuctionRequest{
		Title:         "Aston Martin Valkyrie",
		Description:   "A beautiful 2024 Aston Martin Valkyrie with twin turbo V8 engine",
		StartTime:     now.Add(time.Hour),
		EndTime:       now.Add(24 * time.Hour),
		StartingPrice: 1000,
	}

	tests := []struct {
		name    string
		modify  func(*CreateAuctionRequest)
		wantErr error
	}{
		{
			name:    "valid auction",
			wantErr: nil,
		},
		{
			name: "title too short",
			modify: func(req *CreateAuctionRequest) {
				req.Title = "AB"
			},
			wantErr: ErrTitleTooShort,
		},
		{
			name: "first valid title length",
			modify: func(req *CreateAuctionRequest) {
				req.Title = "ABC"
			},
			wantErr: nil,
		},
		{
			name: "title too long",
			modify: func(req *CreateAuctionRequest) {
				req.Title = strings.Repeat("A", 151)
			},
			wantErr: ErrTitleTooLong,
		},
		{
			name: "last valid title length",
			modify: func(req *CreateAuctionRequest) {
				req.Title = strings.Repeat("A", 150)
			},
			wantErr: nil,
		},
		{
			name: "title contains control character",
			modify: func(req *CreateAuctionRequest) {
				req.Title = "Car\nModel"
			},
			wantErr: ErrInvalidTitle,
		},
		{
			name: "100 non ASCII characters title",
			modify: func(req *CreateAuctionRequest) {
				req.Title = strings.Repeat("é", 100)
			},
			wantErr: nil,
		},
		{
			name: "description too short",
			modify: func(req *CreateAuctionRequest) {
				req.Description = "Too short"
			},
			wantErr: ErrDescriptionTooShort,
		},
		{
			name: "first valid description length",
			modify: func(req *CreateAuctionRequest) {
				req.Description = strings.Repeat("A", 10)
			},
			wantErr: nil,
		},
		{
			name: "description too long",
			modify: func(req *CreateAuctionRequest) {
				req.Description = strings.Repeat("B", 5001)
			},
			wantErr: ErrDescriptionTooLong,
		},
		{
			name: "last valid description length",
			modify: func(req *CreateAuctionRequest) {
				req.Description = strings.Repeat("B", 5000)
			},
		},
		{
			name: "description contains null byte",
			modify: func(req *CreateAuctionRequest) {
				req.Description = "Description\x00invalid"
			},
			wantErr: ErrInvalidDescription,
		},
		{
			name: "start time is zero",
			modify: func(req *CreateAuctionRequest) {
				req.StartTime = time.Time{}
			},
			wantErr: ErrInvalidStartTime,
		},
		{
			name: "start time is in the past",
			modify: func(req *CreateAuctionRequest) {
				req.StartTime = now.Add(-2 * time.Minute)
			},
			wantErr: ErrInvalidStartTime,
		},
		{
			name: "start time is exactly 1 minute ago",
			modify: func(req *CreateAuctionRequest) {
				req.StartTime = now.Add(-time.Minute)
			},
			wantErr: nil,
		},
		{
			name: "end time is zero",
			modify: func(req *CreateAuctionRequest) {
				req.EndTime = time.Time{}
			},
			wantErr: ErrInvalidEndTime,
		},
		{
			name: "end time is not after start time",
			modify: func(req *CreateAuctionRequest) {
				req.EndTime = req.StartTime.Add(-1 * time.Minute)
			},
			wantErr: ErrInvalidEndTime,
		},
		{
			name: "end time equals start time",
			modify: func(req *CreateAuctionRequest) {
				req.EndTime = req.StartTime
			},
			wantErr: ErrInvalidEndTime,
		},
		{
			name: "starting price is zero",
			modify: func(req *CreateAuctionRequest) {
				req.StartingPrice = 0
			},
			wantErr: ErrInvalidStartingPrice,
		},
		{
			name: "starting price is negative",
			modify: func(req *CreateAuctionRequest) {
				req.StartingPrice = -2
			},
			wantErr: ErrInvalidStartingPrice,
		},
	}

	for _, tt := range tests {
		req := validRequest

		if tt.modify != nil {
			tt.modify(&req)
		}

		t.Run(tt.name, func(t *testing.T) {
			got := ValidateAuction(req, now)

			if !errors.Is(got, tt.wantErr) {
				t.Errorf("ValidateAuction() error = %v, want %v", got, tt.wantErr)
			}
		})
	}
}

func TestValidateBidAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  int64
		wantErr error
	}{
		{
			name:    "valid bid",
			amount:  100,
			wantErr: nil,
		},
		{
			name:    "zero bid",
			amount:  0,
			wantErr: ErrInvalidBid,
		},
		{
			name:    "negative bid",
			amount:  -5,
			wantErr: ErrInvalidBid,
		},
		{
			name:    "minimum valid bid",
			amount:  1,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateBidAmount(tt.amount)

			if got != tt.wantErr {
				t.Errorf("got %v, want %v", got, tt.wantErr)
			}
		})
	}
}
