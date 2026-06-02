// Package logging configures the application-wide structured logger.
package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Options controls how the logger is built.
type Options struct {
	// Level is one of debug, info, warn, error (case-insensitive).
	Level string
	// FilePath, when non-empty, enables JSON logs written to a rotating file
	// in addition to stdout. The parent directory is created if needed.
	FilePath string
	// Service is added as a constant attribute on every record so log
	// shippers (Loki/Promtail) can label the stream.
	Service string
	// Env is added as a constant attribute (development/production/...).
	Env string
}

// New builds a JSON slog.Logger writing to stdout and, optionally, a rotating
// file. The returned closer flushes/closes the file sink; it is a no-op when
// no file sink is configured.
func New(opts Options) (*slog.Logger, io.Closer) {
	writers := []io.Writer{os.Stdout}

	var fileSink *lumberjack.Logger
	if opts.FilePath != "" {
		if dir := filepath.Dir(opts.FilePath); dir != "" {
			_ = os.MkdirAll(dir, 0o755)
		}
		fileSink = &lumberjack.Logger{
			Filename:   opts.FilePath,
			MaxSize:    50, // megabytes
			MaxBackups: 5,
			MaxAge:     28, // days
			Compress:   true,
		}
		writers = append(writers, fileSink)
	}

	handler := slog.NewJSONHandler(io.MultiWriter(writers...), &slog.HandlerOptions{
		Level: parseLevel(opts.Level),
	})

	logger := slog.New(handler)

	attrs := []any{}
	if opts.Service != "" {
		attrs = append(attrs, slog.String("service", opts.Service))
	}
	if opts.Env != "" {
		attrs = append(attrs, slog.String("env", opts.Env))
	}
	if len(attrs) > 0 {
		logger = logger.With(attrs...)
	}

	return logger, closerFor(fileSink)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type noopCloser struct{}

func (noopCloser) Close() error { return nil }

func closerFor(sink *lumberjack.Logger) io.Closer {
	if sink == nil {
		return noopCloser{}
	}
	return sink
}
