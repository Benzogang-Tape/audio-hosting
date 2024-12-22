package logger

import (
	"context"
	"go.uber.org/zap"
)

const (
	LoggerKey     = "logger"
	ServiceName   = "service"
	TraceIdKey    = "X-Trace-Id"
	TraceIdLogKey = "traceId"
)

type Logger interface {
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
	Debug(ctx context.Context, msg string, fields ...zap.Field)
}

type logger struct {
	serviceName string
	logger      *zap.Logger
}

func (l logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(ServiceName, l.serviceName))

	traceId := TraceIDFromContext(ctx)
	if traceId != "" {
		fields = append(fields, zap.String(TraceIdLogKey, traceId))
	}

	l.logger.Info(msg, fields...)
}

func (l logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(ServiceName, l.serviceName))

	traceId := TraceIDFromContext(ctx)
	if traceId != "" {
		fields = append(fields, zap.String(TraceIdLogKey, traceId))
	}

	l.logger.Error(msg, fields...)
}

func (l logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(ServiceName, l.serviceName))

	traceId := TraceIDFromContext(ctx)
	if traceId != "" {
		fields = append(fields, zap.String(TraceIdLogKey, traceId))
	}

	l.logger.Debug(msg, fields...)
}

func New(serviceName string, env string) Logger {
	var zapLogger *zap.Logger

	if env == "dev" {
		zapLogger, _ = zap.NewDevelopment()
	} else {
		zapLogger, _ = zap.NewProduction()
	}
	defer zapLogger.Sync()

	return &logger{
		serviceName: serviceName,
		logger:      zapLogger,
	}
}

func GetLoggerFromCtx(ctx context.Context) Logger {
	return ctx.Value(LoggerKey).(Logger)
}

func TraceIDFromContext(ctx context.Context) string {
	value, ok := ctx.Value(TraceIdLogKey).(string)
	if !ok {
		return ""
	}

	return value
}
