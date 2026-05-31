package http

import (
	"net/http"
	"time"

	"github.com/YarKhan02/BlackBird/internal/api/http/handler"
	apimiddleware "github.com/YarKhan02/BlackBird/internal/api/http/middleware"
	"github.com/YarKhan02/BlackBird/internal/config"
	"github.com/YarKhan02/BlackBird/internal/domain/role"
	"github.com/YarKhan02/BlackBird/internal/domain/token"
	"github.com/YarKhan02/BlackBird/internal/domain/user"
	"github.com/YarKhan02/BlackBird/internal/infrastructure/redis"
	"github.com/go-chi/chi/v5"
	chimid "github.com/go-chi/chi/v5/middleware"
)

func NewServer(
	cfg *config.Config,
	userSvc *user.Service,
	tokenSvc *token.Service,
	roleSvc *role.Service,
	blocklist *redis.Blocklist,
) *http.Server {
	r := chi.NewRouter()
	r.Use(chimid.RequestID)
	r.Use(chimid.RealIP)
	r.Use(chimid.Recoverer)
	r.Use(apimiddleware.Logger)
	r.Use(apimiddleware.RateLimit(cfg.RateLimitRequests, cfg.RateLimitWindow))

	authHandler := handler.NewAuthHandler(userSvc, tokenSvc)
	userHandler := handler.NewUserHandler(userSvc)
	roleHandler := handler.NewRoleHandler(roleSvc)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Get("/jwks", authHandler.JWKS)
	})

	r.Route("/users", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(apimiddleware.Auth(tokenSvc, blocklist))
			r.Get("/me", userHandler.Me)
			r.Post("/me/password", userHandler.ChangePassword)
		})

		r.Group(func(r chi.Router) {
			r.Use(apimiddleware.Auth(tokenSvc, blocklist))
			r.Use(apimiddleware.RequireGlobalRole("admin"))
			r.Get("/{id}", userHandler.GetByID)
			r.Post("/{id}/ban", userHandler.Ban)
			r.Post("/{id}/unban", userHandler.Unban)

			r.Get("/{id}/roles", roleHandler.GetUserRoles)
			r.Post("/{id}/roles/global", roleHandler.AddGlobalRole)
			r.Delete("/{id}/roles/global/{role}", roleHandler.RemoveGlobalRole)
			r.Post("/{id}/roles/app", roleHandler.AddAppRole)
			r.Delete("/{id}/roles/app/{appID}/{role}", roleHandler.RemoveAppRole)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(apimiddleware.Auth(tokenSvc, blocklist))
		r.Use(apimiddleware.RequireGlobalRole("admin"))
		r.Get("/roles/global", roleHandler.ListGlobal)
	})

	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
