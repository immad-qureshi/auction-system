package auctions

import (
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var (
	ErrInvalidTitle         = errors.New("invalid title")
	ErrTitleTooShort        = errors.New("title can't be less than 3 characters")
	ErrTitleTooLong         = errors.New("title can't be more than 150 characters")
	ErrInvalidDescription   = errors.New("invalid description")
	ErrDescriptionTooShort  = errors.New("description can't be less than 10 characters")
	ErrDescriptionTooLong   = errors.New("description can't be more than 5000 characters")
	ErrInvalidStartTime     = errors.New("start time is required and can't be in the past")
	ErrInvalidEndTime       = errors.New("end time is required and must be after start time")
	ErrInvalidStartingPrice = errors.New("starting price is required and must be greater than 0")
	ErrInvalidBid           = errors.New("invalid bid")
)

const startTimeGrace = time.Minute

func ValidateAuction(req CreateAuctionRequest, now time.Time) error {

	titleLen := utf8.RuneCountInString(req.Title)

	if titleLen < 3 {
		return ErrTitleTooShort
	}

	if titleLen > 150 {
		return ErrTitleTooLong
	}

	if strings.IndexFunc(req.Title, unicode.IsControl) >= 0 {
		return ErrInvalidTitle
	}

	descriptionLen := utf8.RuneCountInString(req.Description)

	if descriptionLen < 10 {
		return ErrDescriptionTooShort
	}

	if descriptionLen > 5000 {
		return ErrDescriptionTooLong
	}

	if strings.ContainsRune(req.Description, 0) {
		return ErrInvalidDescription
	}

	if req.StartTime.IsZero() || req.StartTime.Before(now.Add(-startTimeGrace)) {
		return ErrInvalidStartTime
	}

	if req.EndTime.IsZero() || !req.EndTime.After(req.StartTime) {
		return ErrInvalidEndTime
	}

	if req.StartingPrice <= 0 {
		return ErrInvalidStartingPrice
	}

	return nil
}

func ValidateBidAmount(amount int64) error {
	if amount <= 0 {
		return ErrInvalidBid
	}

	return nil
}
