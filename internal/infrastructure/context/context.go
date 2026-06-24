package context

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey string

const (
	requestIDKey 	ctxKey = "request_id"
	loggerKey		ctxKey = "logger"
)

func GetLogger(ctx context.Context) *zap.Logger {
	l, ok := ctx.Value(loggerKey).(*zap.Logger)
	if !ok {
		return zap.NewNop()
	}
	return l
}

func GetRequestIDKey(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}