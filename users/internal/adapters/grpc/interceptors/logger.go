package interceptors

import (
	"context"
	"log/slog"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/users/pkg/logger"
	"google.golang.org/grpc"
)

func LoggerInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	start := time.Now()

	log := logger.GetLoggerFromCtx(ctx)

	log.Info(
		"get request",
		slog.String("method", info.FullMethod),
	)

	resp, err = handler(ctx, req)

	defer func() {
		log.Info(
			"request completed",
			slog.String("method", info.FullMethod),
			slog.String("duration", time.Since(start).String()),
		)
	}()

	return resp, err
}

func LoggerToCtxInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx = logger.SetLoggerToCtx(ctx, log)

		return handler(ctx, req)
	}
}
