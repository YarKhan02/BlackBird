package middleware

import (
	"context"
	"net/http"
	"time"
    "sync"
)

type OriginFetcher func(ctx context.Context) ([]string, error)

func DynamicCORS(staticOrigins []string, fetcher OriginFetcher) func(http.Handler) http.Handler {
    staticSet := make(map[string]struct{}, len(staticOrigins))
    for _, o := range staticOrigins {
        staticSet[o] = struct {}{}
    }

    var (
        mu          sync.RWMutex
        cached      map[string]struct{}
        cacheAt     time.Time
        cacheTTL =  60 * time.Second
    )

    isAllowed := func(ctx context.Context, origin string) bool {
        if _, ok := staticSet[origin]; ok {
            return true
        }

        mu.RLock()
        if cached != nil && time.Since(cacheAt) < cacheTTL {
            _, ok := cached[origin]
            mu.RUnlock()
            return ok
        }
        mu.RUnlock()

        mu.Lock()
        defer mu.Unlock()

        if cached == nil || time.Since(cacheAt) >= cacheTTL {
            dbOrigins, err := fetcher(ctx)
            if err == nil {
                fresh := make(map[string]struct{}, len(dbOrigins))
                for _, o := range dbOrigins {
                    fresh[o] = struct {}{}
                }
                cached = fresh
                cacheAt = time.Now()
            }
        }

        if cached != nil {
            _, ok := cached[origin]
            return ok
        }

        return false
    }

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")

            if origin != "" && isAllowed(r.Context(), origin) {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                w.Header().Set("Access-Control-Allow-Credentials", "true")  // required for cookies
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
                w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Client-ID")
                w.Header().Set("Vary", "Origin")
            }

            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}