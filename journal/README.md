# Database Migration

## Development

```sh
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Then run the following command to apply the migrations:

```sh
./bin/db-migrate.sh
```

## Production

```sh
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/golang-migrate/migrate/v4/database/pgx
go get -u github.com/golang-migrate/migrate/v4/source/file
```

# Chi Router

This is setting up a Chi router in Go and attaching a chain of middleware. Every incoming HTTP request will pass through these middleware in the order they are registered.

```
r := chi.NewRouter()
```

Creates a new router instance.

---

```
r.Use(chimid.RequestID)
```

Adds a unique request ID to every request.

Example:

```
X-Request-ID: a1b2c3d4
```

Useful for tracing logs across services.

---

```
r.Use(chimid.RealIP)
```

Extracts the real client IP from headers such as:

```
X-Forwarded-For
X-Real-IP
```

Important when your application is behind a reverse proxy, load balancer, or API Gateway.

Without it:

```
RemoteAddr = Load Balancer IP
```

With it:

```
RemoteAddr = Actual User IP
```

---

```
r.Use(chimid.Recoverer)
```

Protects your server from crashing due to panics.

Without Recoverer:

```
panic("something broke")
```

could terminate the request and potentially affect the server.

With Recoverer:

- Panic is caught
- Stack trace is logged
- HTTP 500 returned
- Server keeps running

```
r.Use(apimiddleware.Logger)
```

---

Custom logging middleware.

Typically logs:

```
GET /users 200 15ms
POST /login 401 8ms
```

May also log:

- Request ID
- IP Address
- User Agent
- Response time

depending on implementation.

```
r.Use(apimiddleware.RateLimit(
    cfg.RateLimitRequests,
    cfg.RateLimitWindow,
))
```

---

Custom rate-limiting middleware.

Example configuration:

```
RateLimitRequests = 100
RateLimitWindow   = time.Minute
```

Meaning:

```
100 requests per minute per IP
```

If exceeded:

```
HTTP/1.1 429 Too Many Requests
```

# SQL Structure

This is safe from SQL injection as written.

The SQL is static and embedded at build time via //go:embed, so it is not constructed from user input.

The query uses positional parameters ($1…$7) and QueryRowContext binds arguments separately, so user input cannot change the SQL structure.

# Authentication Flow

![image](auth_integration_flow.svg)