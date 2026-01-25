# Transfers Service

Simple money transfer API that records transfers between accounts and enforces idempotency with Redis.

## Endpoints
- `GET /health` – readiness probe
- `POST /transfers` – create a transfer
- `GET /transfers` – list transfers (optional `accountId`, `limit`)
- `GET /transfers/{id}` – fetch a transfer

### Create transfer payload
```json
{
  "from": "<account-uuid>",
  "to": "<account-uuid>",
  "amount_cents": 5000,
  "idempotency_key": "optional-key"
}
```

## Running locally

### With Go
```bash
export DATABASE_URL="postgres://transfers:transfers@localhost:5433/transfers?sslmode=disable"
export REDIS_ADDR="localhost:6379"
export ACCOUNTS_URL="http://localhost:8080" # accounts service must be reachable

createdb transfers || true
psql "postgres://localhost:5433/postgres" -c "create user transfers with password 'transfers'" || true
psql "postgres://localhost:5433/postgres" -c "grant all privileges on database transfers to transfers" || true

GO111MODULE=on go run ./cmd/transfers
```

### With Docker Compose
```bash
docker compose up --build
```
The compose file starts Postgres, Redis, and the transfers service. Set `ACCOUNTS_URL` to a reachable accounts endpoint (defaults to `http://accounts:8080`).

## Configuration
- `ENV` (default `dev`)
- `LOG_LEVEL` (default `debug`)
- `HTTP_ADDR` (default `:8081`)
- `DATABASE_URL` (required)
- `REDIS_ADDR` (default `redis-master.db.svc.cluster.local:6379`)
- `REDIS_PASSWORD` (optional)
- `ACCOUNTS_URL` (default `http://accounts.apps.svc.cluster.local:8080`)

