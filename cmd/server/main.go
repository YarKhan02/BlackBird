package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	apihttp "github.com/YarKhan02/BlackBird/internal/api/http"
	"github.com/YarKhan02/BlackBird/internal/config"
	"github.com/YarKhan02/BlackBird/internal/domain/role"
	"github.com/YarKhan02/BlackBird/internal/domain/token"
	"github.com/YarKhan02/BlackBird/internal/domain/user"
	"github.com/YarKhan02/BlackBird/internal/infrastructure/crypto"
	"github.com/YarKhan02/BlackBird/internal/infrastructure/postgre"
	"github.com/YarKhan02/BlackBird/internal/infrastructure/redis"
	"github.com/YarKhan02/BlackBird/internal/observability/logging"
	"github.com/YarKhan02/BlackBird/internal/observability/sentryobs"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const serviceName = "blackbird"

func main() {
	cfg, err := config.Load()
	if err != nil {
		// Logger isn't built yet; use a minimal stderr logger for this fatal.
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).
			Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	logger, logCloser := logging.New(logging.Options{
		Level:    cfg.LogLevel,
		FilePath: cfg.LogFile,
		Service:  serviceName,
		Env:      cfg.Env,
	})
	defer logCloser.Close()
	slog.SetDefault(logger)

	sentryEnabled, sentryFlush := sentryobs.Init(sentryobs.Options{
		DSN:              cfg.SentryDSN,
		Environment:      cfg.SentryEnvironment,
		Release:          serviceName,
		TracesSampleRate: cfg.SentryTracesSampleRate,
	}, logger)
	defer sentryFlush()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		logger.Error("database ping failed", slog.Any("error", err))
		os.Exit(1)
	}

	blocklist, err := redis.NewBlocklist(cfg.RedisURL)
	if err != nil {
		logger.Error("failed to connect to redis", slog.Any("error", err))
		os.Exit(1)
	}
	defer blocklist.Close()

	key, err := crypto.LoadRSAPrivateKey(cfg.RSAPrivateKeyPath)
	if err != nil {
		logger.Error("failed to load RSA key", slog.Any("error", err))
		os.Exit(1)
	}

	userRepo := postgre.NewUserRepository(db)
	roleRepo := postgre.NewRoleRepository(db)
	tokenRepo := postgre.NewTokenRepository(db)

	roleSvc := role.NewService(roleRepo)
	userSvc := user.NewService(userRepo, roleSvc)
	tokenSvc := token.NewService(key, tokenRepo, cfg.JWTIssuer, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

	srv := apihttp.NewServer(cfg, logger, sentryEnabled, userSvc, tokenSvc, roleSvc, blocklist)

	logger.Info("server starting", slog.String("addr", cfg.Addr))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server error", slog.Any("error", err))
		os.Exit(1)
	}
}
