# gotemporal

Production-minded demo of an order lifecycle service built with Go, Temporal, Postgres, and Ent. The service exposes a small HTTP API, orchestrates order state transitions as a Temporal workflow, and persists state in Postgres.

## Architecture at a glance
- **API**: Go + Chi (`POST /orders`, `GET /orders/{id}`), JSON responses.
- **Workflow**: Temporal workflow drives order creation → payment pending → payment confirmed → shipping → completed (or failed).
- **Persistence**: Postgres with Ent schema for strongly-typed access.
- **Workers**: Temporal worker registered in the app process (single binary).
- **Docs**: Redoc served from `/docs`, OpenAPI at `/openapi.yaml`.

## Prerequisites
- Docker and Docker Compose
- Open ports: `3000` (API/Redoc), `7233` (Temporal), `8080` (Temporal UI), `5432` (Postgres)

## Quick start (docker-compose)
```bash
docker compose up -d
```

After startup:
- API & Redoc: http://localhost:3000/docs
- Temporal UI: http://localhost:8080
- Temporal Frontend: localhost:7233
- Postgres: localhost:5432 (user/pass/db `temporal/temporal/app`)

Stop the stack:
```bash
docker compose down
```

## Configuration
Environment variables (see defaults in `internal/config/config.go`):
- `APP_HTTP_ADDR` – HTTP listen address (default `:3000`)
- `DATABASE_URL` – Postgres DSN (default `postgres://temporal:temporal@localhost:5432/app?sslmode=disable`)
- `TEMPORAL_HOSTPORT` – Temporal Frontend address (default `localhost:7233`)
- `TEMPORAL_NAMESPACE` – Temporal namespace (default `default`)
- `TASK_QUEUE` – Temporal task queue (default `orders`)
- `PAYMENT_DELAY_SECONDS` – Simulated payment delay (default `3`)
- `SHIPPING_DELAY_SECONDS` – Simulated shipping delay (default `3`)

## API
- `POST /orders` – create an order
  - body: `{"amount": 100, "customer_id": "c-123"}`
- `GET /orders/{id}` – fetch order by ID

### Sample calls
```bash
curl -X POST http://localhost:3000/orders \
  -H 'Content-Type: application/json' \
  -d '{"amount":150,"customer_id":"cust-1"}'

curl http://localhost:3000/orders/<order-id>
```

## Local development
```bash
go test ./...
go run ./cmd/app
```

Ent schema lives in `ent/schema`. If you change it, regenerate with:
```bash
go generate ./ent
```

## Notes
- Redoc is served from the running app; ensure the `app` container is healthy.
- Temporal worker runs inside the app process; if Temporal is not yet healthy, restart the `app` container after Temporal is up.

