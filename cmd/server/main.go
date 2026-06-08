package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	apihttp "github.com/YarKhan02/BlackBird/internal/api/http"
	"github.com/YarKhan02/BlackBird/internal/config"
	"github.com/YarKhan02/BlackBird/internal/domain/app"
	"github.com/YarKhan02/BlackBird/internal/domain/role"
	"github.com/YarKhan02/BlackBird/internal/domain/token"
	"github.com/YarKhan02/BlackBird/internal/domain/user"
	"github.com/YarKhan02/BlackBird/internal/infrastructure/crypto"
	"github.com/YarKhan02/BlackBird/internal/infrastructure/postgre"
	"github.com/YarKhan02/BlackBird/internal/infrastructure/redis"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}

	blocklist, err := redis.NewBlocklist(cfg.RedisURL)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer blocklist.Close()

	key, err := crypto.LoadRSAPrivateKey(cfg.RSAPrivateKeyPath)
	if err != nil {
		log.Fatalf("failed to load RSA key: %v", err)
	}

	appRepo := postgre.NewAppRepository(db)
	userRepo := postgre.NewUserRepository(db)
	roleRepo := postgre.NewRoleRepository(db)
	tokenRepo := postgre.NewTokenRepository(db)

	appSvc := app.NewService(appRepo)
	roleSvc := role.NewService(roleRepo)
	userSvc := user.NewService(userRepo, roleSvc)
	tokenSvc := token.NewService(key, tokenRepo, cfg.JWTIssuer, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

	srv := apihttp.NewServer(cfg, appSvc, userSvc, tokenSvc, roleSvc, blocklist)

	log.Printf("listening on %s", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}
