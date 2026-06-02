// Package sentryobs wires the Sentry SDK for error tracking. Everything is a
// no-op when no DSN is configured, so local/dev runs are unaffected.
package sentryobs

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	chimid "github.com/go-chi/chi/v5/middleware"
)

// Options configures Sentry initialization.
type Options struct {
	DSN              string
	Environment      string
	Release          string
	TracesSampleRate float64
}

// Init initializes the global Sentry client. When DSN is empty it returns
// (false, no-op flush) and the application runs without error tracking.
func Init(opts Options, logger *slog.Logger) (enabled bool, flush func()) {
	noop := func() {}

	if opts.DSN == "" {
		logger.Info("sentry disabled (no SENTRY_DSN set)")
		return false, noop
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              opts.DSN,
		Environment:      opts.Environment,
		Release:          opts.Release,
		TracesSampleRate: opts.TracesSampleRate,
		EnableTracing:    opts.TracesSampleRate > 0,
	})
	if err != nil {
		logger.Error("sentry init failed; continuing without error tracking", slog.Any("error", err))
		return false, noop
	}

	logger.Info("sentry initialized", slog.String("environment", opts.Environment))
	return true, func() { sentry.Flush(2 * time.Second) }
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

// Middleware returns an HTTP middleware that:
//   - installs a per-request Sentry hub and recovers panics (re-panicking so
//     the outer chi Recoverer still produces a 500 response), and
//   - reports server-side errors (5xx responses) as Sentry events enriched
//     with request context.
func Middleware() func(http.Handler) http.Handler {
	sh := sentryhttp.New(sentryhttp.Options{Repanic: true})

	return func(next http.Handler) http.Handler {
		captureErrors := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: w}
			next.ServeHTTP(rec, r)

			if rec.status < http.StatusInternalServerError {
				return
			}

			hub := sentry.GetHubFromContext(r.Context())
			if hub == nil {
				return
			}
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetLevel(sentry.LevelError)
				scope.SetTag("http.method", r.Method)
				scope.SetTag("http.path", r.URL.Path)
				scope.SetTag("http.status", strconv.Itoa(rec.status))
				if rid := chimid.GetReqID(r.Context()); rid != "" {
					scope.SetTag("request_id", rid)
				}
				hub.CaptureMessage("HTTP " + strconv.Itoa(rec.status) + " " + r.Method + " " + r.URL.Path)
			})
		})

		// sentryhttp must wrap the error-capture handler so the hub is present
		// in the request context when we inspect the status code.
		return sh.Handle(captureErrors)
	}
}
