# BlackBird

BlackBird is a lightweight authentication service written in Go. It provides JWT-based access tokens, refresh token rotation, and role-based authorization for your applications. It uses Postgres for persistence, Redis for token blocklisting, and exposes a JWKS endpoint so your apps can verify tokens locally.

## Features

- Email/password registration and login
- RS256 JWT access tokens with roles embedded in claims
- Refresh token rotation with hashed storage
- JWKS endpoint for public key distribution
- Admin role management and user banning
- Basic rate limiting and request logging

## How it works

1. Users register and login to receive a short-lived access token and a refresh token (stored as an HttpOnly cookie).
2. Your app includes the access token in the `Authorization: Bearer <token>` header for API requests.
3. When the access token expires, call `/auth/refresh` to rotate the refresh token and get a new access token.
4. Your app verifies access tokens using the JWKS endpoint.

## Quick start

### Prerequisites

- Go 1.25+
- Postgres
- Redis

### 1) Create the database

Update the connection string in `bin/db-create.sh` if needed, then run:

```sh
./bin/db-create.sh
```

### 2) Run migrations

Install migrate and apply migrations:

```sh
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

./bin/db-migrate.sh
```

### 3) Generate an RSA key

Create a private key for signing tokens:

```sh
mkdir -p keys
openssl genrsa -out keys/private.pem 2048
```

### 4) Configure environment variables

Minimal example:

```sh
export DATABASE_URL="postgresql://user:pass@localhost:5432/blackbird"
```

Recommended full set:

```sh
export ADDR=":8080"
export DATABASE_URL="postgresql://user:pass@localhost:5432/blackbird"
export REDIS_URL="redis://localhost:6379"
export RSA_PRIVATE_KEY_PATH="./keys/private.pem"
export RATE_LIMIT_REQUESTS="100"
export RATE_LIMIT_WINDOW="60s"
export ENV="development"
export JWT_ISSUER="auth.example.com"
export ACCESS_TOKEN_TTL="15m"
export REFRESH_TOKEN_TTL="720h"

# Observability
export LOG_LEVEL="info"
export LOG_FILE="./logs/app.log"
export SENTRY_DSN=""                  # empty disables Sentry (no-op)
export SENTRY_ENVIRONMENT="development"
export SENTRY_TRACES_SAMPLE_RATE="0.0"
```

See `.env.example` for a copy-paste template.

### 5) Run the server

```sh
go run ./cmd/server
```

## API overview

Public endpoints:

- `POST /auth/register` - create a new user
- `POST /auth/login` - login and get access token
- `POST /auth/refresh` - rotate refresh token and get new access token
- `GET /auth/jwks` - public keys for token verification
- `GET /healthz` - health check

Authenticated endpoints:

- `GET /users/me` - current user profile
- `POST /users/me/password` - change password

Admin-only endpoints (requires global role `admin`):

- `GET /users/{id}`
- `POST /users/{id}/ban`
- `POST /users/{id}/unban`
- `GET /users/{id}/roles`
- `POST /users/{id}/roles/global`
- `DELETE /users/{id}/roles/global/{role}`
- `POST /users/{id}/roles/app`
- `DELETE /users/{id}/roles/app/{appID}/{role}`
- `GET /roles/global`

## Using BlackBird in your app

1. Call `POST /auth/login` with email and password to get an access token.
2. Store the refresh token cookie (HttpOnly) returned by the server.
3. Attach the access token to your app requests:

```http
Authorization: Bearer <access_token>
```

4. Verify tokens in your app using the JWKS endpoint:

- Fetch keys from `GET /auth/jwks`
- Verify `RS256` signatures
- Validate `iss` (issuer), `exp` (expiration), and `sub` (user ID)
- Use `global_roles` and `apps` claims for authorization

## Observability

BlackBird ships logs, metrics, and error tracking out of the box.

| Signal  | How                                                                 | Where to view                         |
| ------- | ------------------------------------------------------------------- | ------------------------------------- |
| Logs    | Structured JSON (`slog`) to stdout + rotating `./logs/app.log`      | Grafana Explore (via Loki/Promtail)   |
| Metrics | Prometheus client at `GET /metrics` (request count/latency/in-flight) | Prometheus + Grafana dashboard        |
| Errors  | Sentry SDK (panics + 5xx responses), enriched with `request_id`     | sentry.io project (when DSN set)      |

### Run the monitoring stack

The app runs locally (`go run ./cmd/server`); the backends run in Docker:

```sh
docker compose -f monitoring/docker-compose.yml up -d
```

This starts:

- **Prometheus** on `http://localhost:9090` (scrapes `host.docker.internal:8080/metrics`)
- **Loki** on `http://localhost:3100` (log storage)
- **Promtail** (tails `./logs/app.log` and ships to Loki)
- **Grafana** on `http://localhost:3000` (login `admin` / `admin`)

Grafana is pre-provisioned with the Prometheus + Loki datasources and a
"BlackBird - HTTP Overview" dashboard (request rate, error rate, latency
percentiles, in-flight, and a live Loki log panel).

### Metrics exposed

- `blackbird_http_requests_total{method,route,status}`
- `blackbird_http_request_duration_seconds{method,route}` (histogram)
- `blackbird_http_requests_in_flight`

The `route` label uses the chi route pattern (e.g. `/users/{id}`) to keep
cardinality bounded.

### Sentry

Set `SENTRY_DSN` to a sentry.io DSN to enable error tracking. When the DSN is
empty, Sentry is fully disabled (no-op) and the app runs normally. Panics are
captured and re-raised so the standard recovery still returns a 500; server-side
errors (5xx) are reported as events tagged with the request ID, method, path,
and status.

## Notes

- Refresh cookies are marked `Secure`. For local HTTP testing, use HTTPS or adjust cookie settings.
- The `monitoring/` stack assumes Docker Desktop (`host.docker.internal` resolves to the host). On plain Linux, the compose file maps `host.docker.internal` to the host gateway.
- Rate limiting is in-memory per instance; for multi-instance deployments, use a shared limiter (e.g. Redis).
- Seed `global_roles` with roles like `admin` and `user` before assigning them.

## License

MIT
