package interceptors

import (
	"context"
	"log/slog"

	"github.com/Benzogang-Tape/audio-hosting/users/pkg/logger"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	requestIDMetadataName = "X-Trace-Id"
)

func RequestIDInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	var requestID string

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		requestID = uuid.New().String()
	} else {
		requestIDHeader := md.Get(requestIDMetadataName)
		if len(requestIDHeader) == 0 {
			requestID = uuid.New().String()
		} else if requestIDHeader[0] == "" {
			requestID = uuid.New().String()
		} else {
			requestID = requestIDHeader[0]
		}
	}

	log := logger.GetLoggerFromCtx(ctx).With(
		slog.String("trace_id", requestID),
	)

	ctx = logger.SetLoggerToCtx(ctx, log)

	return handler(ctx, req)
}
