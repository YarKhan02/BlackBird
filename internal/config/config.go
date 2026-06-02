package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr              string
	DatabaseURL       string
	RedisURL          string
	RSAPrivateKeyPath string
	RateLimitRequests int
	RateLimitWindow   time.Duration
	Env               string
	JWTIssuer         string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	LogLevel          string
	LogFile           string

	SentryDSN              string
	SentryEnvironment      string
	SentryTracesSampleRate float64
}

func Load() (*Config, error) {
	// Load .env if present to simplify local development.
	_ = godotenv.Load()

	cfg := &Config{
		Addr:              getEnv("ADDR", ":8080"),
		DatabaseURL:       mustEnv("DATABASE_URL"),
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379"),
		RSAPrivateKeyPath: getEnv("RSA_PRIVATE_KEY_PATH", "./keys/private.pem"),
		Env:               getEnv("ENV", "development"),
		JWTIssuer:         getEnv("JWT_ISSUER", "auth.shoukan-labs.com"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		LogFile:           getEnv("LOG_FILE", "./logs/app.log"),
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

	cfg.SentryDSN = getEnv("SENTRY_DSN", "")
	cfg.SentryEnvironment = getEnv("SENTRY_ENVIRONMENT", cfg.Env)

	sampleRate, err := strconv.ParseFloat(getEnv("SENTRY_TRACES_SAMPLE_RATE", "0.0"), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid SENTRY_TRACES_SAMPLE_RATE: %w", err)
	}
	cfg.SentryTracesSampleRate = sampleRate

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
