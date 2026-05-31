package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/YarKhan02/BlackBird/internal/domain/token"
	"github.com/YarKhan02/BlackBird/internal/infrastructure/redis"
)

type contextKey string

const claimsKey contextKey = "auth_claims"

func Auth(tokenSvc *token.Service, blocklist *redis.Blocklist) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			claims, err := tokenSvc.ValidateAccessToken(parts[1])
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			if blocklist != nil && claims.ID != "" {
				blocked, err := blocklist.Contains(r.Context(), claims.ID)
				if err != nil || blocked {
					http.Error(w, "invalid token", http.StatusUnauthorized)
					return
				}
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireGlobalRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context())
			if claims == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			for _, roleName := range claims.GlobalRoles {
				if roleName == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, "forbidden", http.StatusForbidden)
		})
	}
}

func GetClaims(ctx context.Context) *token.Claims {
	claims, _ := ctx.Value(claimsKey).(*token.Claims)
	return claims
}
