# Accounts Service

A small Go service providing CRUD over accounts, structured with a clean internal layout and ready for GitOps/Istio deployment.

## Layout
- `cmd/accounts/main.go` – wiring (config, logger, db, routes)
- `internal/config` – env config via cleanenv
- `internal/account` – model, repo, service, handler
- `internal/app` – HTTP routes + middleware
- `migrations` – SQL to create/drop accounts table
- `Dockerfile` – multi-stage build
- `docker-compose.yml` – local dev with Postgres

## Running locally
```bash
cd services/accounts
docker-compose up --build
```

## Configuration
- `ENV` (dev|prod)
- `LOG_LEVEL` (debug|info|warn|error)
- `HTTP_ADDR` (default `:8080`)
- `DATABASE_URL` (e.g., `postgres://user:pass@host:5432/db?sslmode=disable`)

## API
- `POST /accounts` `{name, currency}` → 201 with account
- `GET /accounts/{id}` → 200/404
- `PATCH /accounts/{id}` `{name}` → 200
- `DELETE /accounts/{id}` → 200
- `GET /health` → 200
