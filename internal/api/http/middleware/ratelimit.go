package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	clients map[string]*clientWindow
}

type clientWindow struct {
	start time.Time
	count int
}

func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	rl := &rateLimiter{
		limit:   limit,
		window:  window,
		clients: make(map[string]*clientWindow),
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r.RemoteAddr)
			if !rl.allow(ip) {
				w.Header().Set("Retry-After", window.String())
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cw, ok := rl.clients[ip]
	if !ok || now.Sub(cw.start) >= rl.window {
		rl.clients[ip] = &clientWindow{start: now, count: 1}
		return true
	}

	if cw.count >= rl.limit {
		return false
	}

	cw.count++
	return true
}

func clientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}
