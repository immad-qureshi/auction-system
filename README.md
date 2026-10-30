
# Auction System

Auction System is a backend service written in Go. Its REST API lets users register, log in, create auctions and place bids on auctions created by other users.

When a user places a bid, the auction's row is locked inside a database transaction. A second bid on the same auction has to wait until the first one finishes, then checks against the updated highest bid, so two bids at the same moment can't overwrite each other. Bids on different auctions don't wait for each other because only that one auction's row is locked.

It's written in Go and uses PostgreSQL as it's database. The HTTP server uses Go's standard net/http package and the pgx driver connects it to Postgre. Passwords are hashed with bcrypt and users authenticate with JWTs.
## Features

- **Accounts:** users register with a unique username and email. Passwords are hashed with bcrypt and never stored in plain text.
- **Authentication:** logging in returns a JWT that expires after 24 hours. Creating an auction and placing a bid require it in the `Authorization: Bearer <token>` header.
- **Auctions:** authenticated users create auctions with a title, description, starting price, and start and end times. A new auction starts in the `pending` state.
- **Bidding rules:** a bid must be higher than the current highest bid (or the starting price, if there are no bids yet). It's accepted only while the auction is running, and sellers can't bid on their own auctions.
- **Safe concurrent bidding:** each bid runs in a database transaction that locks the auction's row. A second bid on the same auction waits for the first to finish, so two simultaneous bids can't overwrite each other. Bids on other auctions aren't affected.
- **Automatic auction lifecycle:** a background worker checks the database every second. It moves auctions from `pending` to `active` when their start time arrives, and to `finished` when their end time passes, recording the highest bidder as the winner if there were any bids.
- **Input validation:** request fields are checked for length and format, unknown JSON fields are rejected, and database queries are parameterized.
- **Rate limiting:** limits are applied per IP address in two tiers. `/register`, `/login` and `POST /auction` allow 5 requests per minute. The read endpoints and bidding allow 200 requests per minute. Requests over the limit get a `429` response.
- **Graceful shutdown:** on `Ctrl+C` or `SIGTERM` the server stops accepting new requests, waits up to 10 seconds for requests in progress, and stops the background workers before closing the database connection.
- **Database migrations:** the schema is defined in versioned SQL files and applied with [golang-migrate](https://github.com/golang-migrate/migrate).
## Tech Stack

- **Language:** Go
- **HTTP server:** standard library `net/http`, with its method-based routing (`POST /login`)
- **Database:** PostgreSQL 17, accessed through `database/sql` with the [pgx](https://github.com/jackc/pgx) driver
- **Migrations:** [golang-migrate](https://github.com/golang-migrate/migrate)
- **Authentication:** JWT with [golang-jwt](https://github.com/golang-jwt/jwt), passwords hashed with bcrypt (`golang.org/x/crypto`)
- **Configuration:** environment variables, loaded from a `.env` file with [godotenv](https://github.com/joho/godotenv)
- **Logging:** `log/slog` from the standard library
- **Local development:** Docker Compose runs the PostgreSQL database

## API Reference

All requests and responses use JSON. Responses have the form `{"message": "...", "data": ...}`, and `data` is left out when there's nothing to return. Endpoints marked **Yes** under Auth need the header `Authorization: Bearer <token>`, using the token from `/login`. Rate limits apply per IP address.

| Method | Endpoint             | Auth | Rate limit | Description              |
| ------ | -------------------- | ---- | ---------- | ------------------------ |
| POST   | `/register`          | No   | 5/min      | Create a user account    |
| POST   | `/login`             | No   | 5/min      | Log in and get a JWT     |
| POST   | `/auction`           | Yes  | 5/min      | Create an auction        |
| GET    | `/auctions`          | No   | 200/min    | List all auctions        |
| GET    | `/auctions/{id}`     | No   | 200/min    | Get a single auction     |
| POST   | `/auctions/{id}/bid` | Yes  | 200/min    | Place a bid on an auction |

### Request bodies

`POST /register`

```json
{ "username": "alice", "email": "alice@example.com", "password": "password123" }
```

Username: 3 to 50 characters (letters, numbers, underscore). Password: 8 to 72 bytes.

`POST /login`

```json
{ "email": "alice@example.com", "password": "password123" }
```

`POST /auction`

```json
{
  "title": "Vintage camera",
  "description": "Working 35mm film camera in good condition",
  "start_time": "2026-11-01T10:00:00Z",
  "end_time": "2026-11-02T10:00:00Z",
  "starting_price": 100
}
```

Times use the RFC 3339 format. The start can't be in the past, and the end must be after the start. The title is 3 to 150 characters, the description 10 to 5000, and the starting price a whole number greater than 0.

`POST /auctions/{id}/bid`

```json
{ "amount": 150 }
```

The amount must be higher than the current highest bid, or higher than the starting price if there are no bids yet.

### Status codes

| Code | Meaning                                                                                       |
| ---- | --------------------------------------------------------------------------------------------- |
| 200  | Success                                                                                       |
| 201  | Created (user or auction)                                                                     |
| 400  | Invalid JSON, invalid auction id, or failed validation                                        |
| 401  | Wrong email or password, or a missing, invalid or expired token                               |
| 403  | The seller tried to bid on their own auction                                                  |
| 404  | Auction not found                                                                             |
| 409  | Username or email already taken, or a bid on an auction that hasn't started, has finished, or is too low |
| 429  | Rate limit exceeded                                                                           |
| 500  | Unexpected server error                                                                       |
## Project Structure

```
.
├── cmd/api                  # application entry point
├── internal/
│   ├── auctions/         # auctions and bidding
│   ├── auth/             # JWT and authentication middleware
│   ├── ratelimiter/      # per-IP rate limiting
│   ├── response/         # JSON response helper
│   └── users/            # registration and login
├── migrations/           # versioned SQL schema files
├── docker-compose.yml    # PostgreSQL for local development
└── .env.example          # example configuration
```

- `cmd/api/main.go` loads the configuration, connects to the database, starts the background workers and the HTTP server, and handles graceful shutdown.
- `internal/auctions` holds the auction handlers, the bidding logic, the database queries, and the background worker that activates and finishes auctions.
- `internal/auth` generates and verifies JWTs and provides the middleware that protects routes.
- `internal/ratelimiter` limits requests per IP address and removes expired entries in the background.
- `internal/response` writes the JSON responses shared by all handlers.
- `internal/users` handles registration, login, and password hashing.
- `migrations/` holds the numbered `up` and `down` SQL files that create and remove the schema.

Each feature package follows the same layout: `handlers.go` for HTTP, `store.go` for database queries, `models.go` for types, `validation.go` for input checks, `errors.go` for error values, and `routes.go` for route registration. Handlers use a store interface, so the database code can be replaced without changing them.
## Getting Started

Tested on Linux.

### Prerequisites

- [Go](https://go.dev/dl/) (version 1.26.5+)
- [Docker](https://docs.docker.com/get-docker/) with Docker Compose
- [golang-migrate](https://github.com/golang-migrate/migrate), the tool that runs the database migrations:

  ```bash
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```

  Make sure `~/go/bin` is on your `PATH`.

### Setup

1. Clone the repository:

   ```bash
   git clone https://github.com/immad-qureshi/auction-system.git
   cd auction-system
   ```

2. Create your config file from the example:

   ```bash
   cp .env.example .env
   ```

3. Open `.env` and edit the values:

   - `JWT_SECRET` must be at least 32 bytes. Generate one with `openssl rand -base64 32`.
   - Use only letters and numbers in `POSTGRES_PASSWORD`. It is also part of the `DATABASE_URL`, and characters such as `@`, `:`, `/`, `#` and `?` break the URL.
   - The user, password and database name in `DATABASE_URL` must match the three `POSTGRES_` values.

4. Start the database:

   ```bash
   docker compose up -d
   ```

   Run `docker compose ps` to check that the `db` container is running.

5. Load the variables from `.env` into your terminal session and run the migrations:

   ```bash
   set -a; source .env; set +a
   migrate -path migrations -database "$DATABASE_URL" up
   ```

   You should see `1/u init`.

6. Start the server. From the project root, run:

   ```bash
   go run ./cmd/api
   ```

   You should see `connected successfully to db`. The server listens on the address in `HTTP_SERVER` (`:8080` by default).

### Try it out

Register a user:

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "email": "alice@example.com", "password": "password123"}'
```

The response should look like this:

```json
{"message":"user created","data":{"user_id":1,"username":"alice","email":"alice@example.com"}}
```

Log in to get a token:

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "password": "password123"}'
```

Send the token in the `Authorization: Bearer <token>` header on the endpoints that require authentication (see the endpoints table above).

Note: `/register`, `/login` and `POST /auction` allow 5 requests per minute per IP address. After that you will get a `429 Too Many Requests` response until the window passes.

### Stopping

Stop the server with `Ctrl+C`. It finishes the requests in progress before shutting down.

To stop the database:

```bash
docker compose down        # stops the container and keeps your data
docker compose down -v     # stops the container and deletes your data
```

## Roadmap

- [x] Database migrations
- [x] Docker Compose setup for PostgreSQL
- [ ] Unit tests for validation, authentication, and the rate limiter
- [ ] Integration tests for bidding, including simultaneous bids on one auction
- [ ] Dockerfile for the API, so the whole app runs with `docker compose up`
- [ ] GitHub Actions pipeline that builds and tests on every push
- [ ] Database indexes for the auction worker's queries
- [ ] Pagination for `GET /auctions`
- [ ] Deployment to AWS