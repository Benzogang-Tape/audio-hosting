package logger

import (
	"context"

	"github.com/rs/zerolog"
)

type loggerKeyType struct{}

var loggerKey = loggerKeyType{}

type traceIdKeyType struct{}

var traceIdKey = traceIdKeyType{}

func WithLoggerAndTraceId(ctx context.Context, logger zerolog.Logger, traceId string) context.Context {
	return WithLogger(WithTraceId(ctx, traceId), logger)
}

func WithLogger(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func WithTraceId(ctx context.Context, traceId string) context.Context {
	return context.WithValue(ctx, traceIdKey, traceId)
}

func FromContext(ctx context.Context) zerolog.Logger {
	if logger, ok := ctx.Value(loggerKey).(zerolog.Logger); ok {
		return logger
	}

	return zerolog.Nop()
}

func TraceIdFromContext(ctx context.Context) string {
	if traceId, ok := ctx.Value(traceIdKey).(string); ok {
		return traceId
	}

	return ""
}
