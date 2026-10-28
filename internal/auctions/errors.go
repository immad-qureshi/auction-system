package auctions

import "errors"

var (
	ErrAuctionNotFound   = errors.New("auction not found")
	ErrAuctionNotStarted = errors.New("auction has not started yet")
	ErrAuctionFinished   = errors.New("auction is finished")
	ErrSellerCannotBid   = errors.New("seller can't bid on his own auction")
	ErrLowBid            = errors.New("bid can't be less than current bid or starting price")
)
