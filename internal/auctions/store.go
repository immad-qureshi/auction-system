package auctions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

type PostgresAuctionStore struct {
	DB *sql.DB
}

func (p *PostgresAuctionStore) Create(ctx context.Context, sellerID int64, req CreateAuctionRequest) (AuctionResponse, error) {
	var auction AuctionResponse

	err := p.DB.QueryRowContext(ctx, "INSERT INTO auctions (seller_id,title,description,starting_price,start_time,end_time) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id,seller_id,title,description,starting_price,current_bid,start_time,end_time,state", sellerID, req.Title, req.Description, req.StartingPrice, req.StartTime, req.EndTime).Scan(&auction.ID, &auction.SellerID, &auction.Title, &auction.Description, &auction.StartingPrice, &auction.CurrentBid, &auction.StartTime, &auction.EndTime, &auction.State)

	if err != nil {
		return AuctionResponse{}, fmt.Errorf("auction creation in db: %w", err)
	}

	return auction, nil
}

func (p *PostgresAuctionStore) GetAll(ctx context.Context) ([]AuctionResponse, error) {
	var auctions []AuctionResponse

	rows, err := p.DB.QueryContext(ctx, "SELECT id,seller_id,title,description,starting_price,current_bid,start_time,end_time,state,created_at,winner_id,winning_bid,finished_at FROM auctions ORDER BY start_time ASC")

	if err != nil {
		return nil, fmt.Errorf("fetching all auctions from db: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var auction AuctionResponse

		err := rows.Scan(&auction.ID, &auction.SellerID, &auction.Title, &auction.Description, &auction.StartingPrice, &auction.CurrentBid, &auction.StartTime, &auction.EndTime, &auction.State, &auction.CreatedAt, &auction.WinnerID, &auction.WinningBid, &auction.FinishedAt)

		if err != nil {
			return nil, fmt.Errorf("rows scanning: %w", err)
		}

		auctions = append(auctions, auction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return auctions, nil
}

func (p *PostgresAuctionStore) Get(ctx context.Context, auctionID int) (AuctionResponse, error) {
	var auction AuctionResponse

	err := p.DB.QueryRowContext(ctx, "SELECT id,seller_id,title,description,starting_price,current_bid,start_time,end_time,state,created_at,winner_id,winning_bid,finished_at FROM auctions WHERE id=$1", auctionID).Scan(&auction.ID, &auction.SellerID, &auction.Title, &auction.Description, &auction.StartingPrice, &auction.CurrentBid, &auction.StartTime, &auction.EndTime, &auction.State, &auction.CreatedAt, &auction.WinnerID, &auction.WinningBid, &auction.FinishedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AuctionResponse{}, ErrAuctionNotFound
		}
		return AuctionResponse{}, fmt.Errorf("auction search: %w", err)
	}

	return auction, nil
}

func (p *PostgresAuctionStore) PlaceBid(ctx context.Context, amount int64, auctionID int64, bidderID int64) (Bid, error) {
	var bid Bid
	var auction Auction

	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return Bid{}, fmt.Errorf("transaction: %w", err)
	}

	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, "SELECT id,seller_id,title,description,starting_price,current_bid,start_time,end_time,state FROM auctions WHERE id=$1 FOR UPDATE", auctionID).Scan(&auction.ID, &auction.SellerID, &auction.Title, &auction.Description, &auction.StartingPrice, &auction.CurrentBid, &auction.StartTime, &auction.EndTime, &auction.State)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Bid{}, ErrAuctionNotFound
		}
		return Bid{}, fmt.Errorf("auction search: %w", err)
	}

	now := time.Now()

	if auction.StartTime.After(now) {
		return Bid{}, ErrAuctionNotStarted
	}

	//EndTime.After(now) is true only while the auction is still running (EndTime > now).
	//Negating it rejects when EndTime <= now, which includes the exact instant.

	if !auction.EndTime.After(now) {
		return Bid{}, ErrAuctionFinished
	}

	if auction.SellerID == bidderID {
		return Bid{}, ErrSellerCannotBid
	}

	if auction.CurrentBid == nil && amount <= auction.StartingPrice {
		return Bid{}, ErrLowBid
	}

	if auction.CurrentBid != nil && amount <= *auction.CurrentBid {
		return Bid{}, ErrLowBid
	}

	err = tx.QueryRowContext(ctx, "INSERT INTO bids (bidder_id,auction_id,amount) VALUES ($1,$2,$3) RETURNING id,bidder_id,auction_id,amount", bidderID, auctionID, amount).Scan(&bid.ID, &bid.BidderID, &bid.AuctionID, &bid.Amount)

	if err != nil {
		return Bid{}, fmt.Errorf("bid insert: %w", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE auctions SET current_bid=$1 WHERE id=$2", amount, auctionID)

	if err != nil {
		return Bid{}, fmt.Errorf("auction update: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return Bid{}, fmt.Errorf("transaction commit: %w", err)
	}

	return bid, nil
}

func (p *PostgresAuctionStore) FinalizeExpiredAuctions(ctx context.Context) error {
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer tx.Rollback()

	query := `WITH expired AS(
	SELECT id
	FROM auctions
	WHERE end_time <= NOW()
	AND STATE IN ('active','pending')
	ORDER BY end_time
	FOR UPDATE SKIP LOCKED
	)
	UPDATE auctions AS a
	SET winner_id = b.bidder_id,
	winning_bid = b.amount,
	state = 'finished',
	finished_at = NOW()
	FROM expired AS e
	LEFT JOIN LATERAL (
	SELECT bidder_id,amount
	FROM bids
	WHERE auction_id = e.id
	ORDER BY amount DESC, id DESC
	LIMIT 1
	) AS b on TRUE
	WHERE a.id = e.id;
	`

	result, err := tx.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("finalize expired auctions: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get finalized count: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit finalization: %w", err)
	}

	if count > 0 {
		slog.Info("auction finalization completed", "count", count)
	}
	return nil
}

func (p *PostgresAuctionStore) ActivatePendingAuctions(ctx context.Context) error {
	result, err := p.DB.ExecContext(ctx, `UPDATE auctions SET state = 'active' WHERE state = 'pending' AND start_time <= NOW() AND end_time > NOW()`)

	if err != nil {
		return fmt.Errorf("activate pending auctions: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get activated count: %w", err)
	}

	if count > 0 {
		slog.Info("auctions state changed to active", "count", count)
	}
	return nil
}
