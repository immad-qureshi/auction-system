CREATE TABLE users(
 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 username VARCHAR(50) NOT NULL UNIQUE,
 email VARCHAR(255) NOT NULL UNIQUE,
 password_hash TEXT NOT NULL,
 role TEXT NOT NULL
);

CREATE TABLE auctions(
 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 seller_id BIGINT NOT NULL REFERENCES users(id),
 title VARCHAR(150) NOT NULL,
 description VARCHAR(5000) NOT NULL,
 starting_price BIGINT NOT NULL CHECK(
  starting_price > 0
 ),
 current_bid BIGINT CHECK(
    current_bid IS NULL OR current_bid >= starting_price
 ),
 start_time TIMESTAMPTZ NOT NULL,
 end_time TIMESTAMPTZ NOT NULL,
 state VARCHAR(20) NOT NULL DEFAULT 'pending'
   CHECK (state IN ('active','pending','finished','cancelled')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  winner_id BIGINT REFERENCES users(id),
  winning_bid BIGINT,
  finished_at TIMESTAMPTZ,
  CHECK(end_time > start_time)
);

CREATE TABLE bids(
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    bidder_id BIGINT NOT NULL REFERENCES users(id),
    auction_id BIGINT NOT NULL REFERENCES auctions(id),
    amount BIGINT NOT NULL CHECK(amount > 0)
);