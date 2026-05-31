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
```

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

## Notes

- Refresh cookies are marked `Secure`. For local HTTP testing, use HTTPS or adjust cookie settings.
- Rate limiting is in-memory per instance; for multi-instance deployments, use a shared limiter (e.g. Redis).
- Seed `global_roles` with roles like `admin` and `user` before assigning them.

## License

MIT
