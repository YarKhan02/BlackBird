package sentryobs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	chimid "github.com/go-chi/chi/v5/middleware"
)

// TestMiddlewareCaptures verifies that both 5xx responses and panics produce
// Sentry envelopes, using a fake ingest endpoint so no real DSN is needed.
func TestMiddlewareCaptures(t *testing.T) {
	var received int64

	ingest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/envelope/") {
			atomic.AddInt64(&received, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ingest.Close()

	host := strings.TrimPrefix(ingest.URL, "http://")
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:       "http://publickey@" + host + "/1",
		Transport: sentry.NewHTTPSyncTransport(),
	}); err != nil {
		t.Fatalf("sentry init: %v", err)
	}

	r := chi.NewRouter()
	r.Use(chimid.RequestID)
	r.Use(chimid.Recoverer)
	r.Use(Middleware())
	r.Get("/boom", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	r.Get("/panic", func(http.ResponseWriter, *http.Request) {
		panic("kaboom")
	})
	r.Get("/ok", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(r)
	defer srv.Close()

	get := func(path string) int {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		defer resp.Body.Close()
		return resp.StatusCode
	}

	if code := get("/ok"); code != http.StatusOK {
		t.Fatalf("/ok expected 200, got %d", code)
	}
	if code := get("/boom"); code != http.StatusInternalServerError {
		t.Fatalf("/boom expected 500, got %d", code)
	}
	if code := get("/panic"); code != http.StatusInternalServerError {
		t.Fatalf("/panic expected 500 (recovered), got %d", code)
	}

	sentry.Flush(2_000_000_000) // 2s in nanoseconds

	// Expect at least 2 envelopes: one for the 500 and one for the panic.
	// The successful /ok request must not produce an event.
	if got := atomic.LoadInt64(&received); got < 2 {
		t.Fatalf("expected >=2 Sentry envelopes, got %d", got)
	}
}
