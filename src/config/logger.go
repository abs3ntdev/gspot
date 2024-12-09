package config

import (
	"context"
	"log/slog"
)

type loggerContextKeyType struct{}

var loggerContextKey = loggerContextKeyType{}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, logger)
}

func GetLogger(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerContextKey).(*slog.Logger)
	if !ok {
		return slog.Default()
	}
	return logger
}
