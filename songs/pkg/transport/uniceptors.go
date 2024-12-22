package transport

import (
	"context"
	"math"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/erix"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/logger"

	"dev.gaijin.team/go/golib/stacktrace"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	TraceIdKey    = "X-Trace-Id"
	TraceIdLogKey = "trace_id"
	AuthKey       = "Authorization"
)

func ContextWithLogger[T any, T2 any](base zerolog.Logger) Uniceptor[T, T2] {
	return func(next GrpcHandler[T, T2]) GrpcHandler[T, T2] {
		return func(ctx context.Context, req T) (resp T2, err error) {
			var traceId string

			md, ok := metadata.FromIncomingContext(ctx)
			if ok && len(md.Get(TraceIdKey)) > 0 {
				traceId = md.Get(TraceIdKey)[0]
			} else {
				traceId = uuid.NewString()
			}

			log := base.With().
				Str(TraceIdLogKey, traceId).
				Logger()

			ctx = logger.WithLoggerAndTraceId(ctx, log, traceId)

			return next(ctx, req)
		}
	}
}

func Logging[T any, T2 any](method string) Uniceptor[T, T2] {
	return func(next GrpcHandler[T, T2]) GrpcHandler[T, T2] {
		return func(ctx context.Context, req T) (resp T2, err error) {
			log := logger.FromContext(ctx)

			start := time.Now()
			res, handlerErr := next(ctx, req)
			elapsed := time.Since(start)

			var logLevel zerolog.Level = zerolog.InfoLevel

			if handlerErr != nil {
				logLevel = zerolog.ErrorLevel

				code := erix.GrpcCode(handlerErr)
				err = status.Error(code, erix.LastReason(handlerErr))
			}

			log.WithLevel(logLevel).
				Err(handlerErr).
				Dur("elapsed", elapsed).
				Str("method", method).
				Msg("incoming request")

			return res, err //nolint:wrapcheck
		}
	}
}

func Recovery[T any, T2 any](method string) Uniceptor[T, T2] {
	return func(next GrpcHandler[T, T2]) GrpcHandler[T, T2] {
		return func(ctx context.Context, req T) (resp T2, err error) {
			log := logger.FromContext(ctx)

			defer func() {
				recov := recover()
				if recov == nil {
					return
				}

				stack := stacktrace.Capture(0, math.MaxInt)

				marshaledStack := logger.ArrayFunc(func(a *zerolog.Array) {
					for _, frame := range stack.FramesIter() {
						a.Object(logger.ObjectFunc(func(e *zerolog.Event) {
							e.Str("func", frame.Function)
							e.Str("file", frame.File)
							e.Int("line", frame.Line)
						}))
					}
				})

				log.Error().
					Any("panic", recov).
					Str("method", method).
					Array("stacktrace", marshaledStack).
					Msg("panic was recovered")

				err = status.Errorf(
					codes.Internal, "internal error (trace_id %s)",
					logger.TraceIdFromContext(ctx))
			}()

			return next(ctx, req)
		}
	}
}

type ValidatorAll interface {
	ValidateAll() error
}

func Validation[T ValidatorAll, T2 any](enabled bool) Uniceptor[T, T2] {
	return func(next GrpcHandler[T, T2]) GrpcHandler[T, T2] {
		return func(ctx context.Context, req T) (resp T2, err error) {
			if !enabled {
				return next(ctx, req)
			}

			err = req.ValidateAll()

			if err != nil {
				return *new(T2), status.Error(codes.InvalidArgument, err.Error())
			}

			return next(ctx, req)
		}
	}
}
