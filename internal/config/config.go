package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr              string
	DatabaseURL       string
	RedisURL          string
	RSAPrivateKeyPath string
	RSAPrivateKeyPEM  string
	MigrationsPath    string
	RateLimitRequests int
	RateLimitWindow   time.Duration
	Env               string
	JWTIssuer         string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	AllowedOrigins    []string
}

func Load() (*Config, error) {
	// Load .env if present to simplify local development.
	_ = godotenv.Load()

	originsRaw := getEnv("ALLOWED_ORIGINS", "http://localhost:3000")
	origins := strings.Split(originsRaw, ",")
	for i, o := range origins {
		origins[i] = strings.TrimSpace(o)
	}

	cfg := &Config{
		Addr:              normalizeAddr(getEnv("ADDR", getEnv("PORT", "8080"))),
		DatabaseURL:       mustEnv("DATABASE_URL"),
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379"),
		RSAPrivateKeyPath: getEnv("RSA_PRIVATE_KEY_PATH", "./keys/private.pem"),
		RSAPrivateKeyPEM:  getEnv("RSA_PRIVATE_KEY_PEM", ""),
		MigrationsPath:    getEnv("MIGRATIONS_PATH", "file://./migrations"),
		Env:               getEnv("ENV", "development"),
		JWTIssuer:         getEnv("JWT_ISSUER", "auth.shoukan-labs.com"),
		AllowedOrigins:    origins,
	}

	limitStr := getEnv("RATE_LIMIT_REQUESTS", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_REQUESTS: %w", err)
	}
	cfg.RateLimitRequests = limit

	window, err := time.ParseDuration(getEnv("RATE_LIMIT_WINDOW", "60s"))
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_WINDOW: %w", err)
	}
	cfg.RateLimitWindow = window

	accessTTL, err := time.ParseDuration(getEnv("ACCESS_TOKEN_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid ACCESS_TOKEN_TTL: %w", err)
	}
	cfg.AccessTokenTTL = accessTTL

	refreshTTL, err := time.ParseDuration(getEnv("REFRESH_TOKEN_TTL", "720h")) // 30 days
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TOKEN_TTL: %w", err)
	}
	cfg.RefreshTokenTTL = refreshTTL

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}

func normalizeAddr(raw string) string {
	addr := strings.TrimSpace(raw)
	if addr == "" {
		return ":8080"
	}

	if strings.HasPrefix(addr, ":") {
		return addr
	}

	if _, err := strconv.Atoi(addr); err == nil {
		return ":" + addr
	}

	if _, _, err := net.SplitHostPort(addr); err == nil {
		return addr
	}

	return addr
}
