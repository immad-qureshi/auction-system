package auctions

import "time"

type AuctionState string

const (
	AuctionPending  AuctionState = "pending"
	AuctionActive   AuctionState = "active"
	AuctionFinished AuctionState = "finished"
)

type Auction struct {
	ID            int64
	SellerID      int64
	Title         string
	Description   string
	StartingPrice int64
	CurrentBid    *int64
	StartTime     time.Time
	EndTime       time.Time
	State         AuctionState
	CreatedAt     time.Time
	WinnerID      *int64
	WinningBid    *int64
	FinishedAt    *time.Time
}

type CreateAuctionRequest struct {
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	StartingPrice int64     `json:"starting_price"`
}

type AuctionResponse struct {
	ID            int64        `json:"id"`
	SellerID      int64        `json:"seller_id"`
	Title         string       `json:"title"`
	Description   string       `json:"description"`
	StartingPrice int64        `json:"starting_price"`
	CurrentBid    *int64       `json:"current_bid"`
	StartTime     time.Time    `json:"start_time"`
	EndTime       time.Time    `json:"end_time"`
	State         AuctionState `json:"state"`
	CreatedAt     time.Time    `json:"created_at"`
	WinnerID      *int64       `json:"winner_id"`
	WinningBid    *int64       `json:"winning_bid"`
	FinishedAt    *time.Time   `json:"finished_at"`
}

type Bid struct {
	ID        int64
	BidderID  int64
	AuctionID int64
	Amount    int64
}

type PlaceBidRequest struct {
	Amount int64 `json:"amount"`
}
