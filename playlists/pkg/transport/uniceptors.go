package transport

import (
	"context"
	"math"
	"runtime"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/erix"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"

	"dev.gaijin.team/go/golib/stacktrace"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	TraceIdKey    = "X-Trace-Id"
	TraceIdLogKey = "traceId"
	AuthKey       = "Authorization"
)

func ContextWithLogger[T any, T2 any](log logger.Logger) Uniceptor[T, T2] {
	return func(next GrpcHandler[T, T2]) GrpcHandler[T, T2] {
		return func(ctx context.Context, req T) (resp T2, err error) {
			var traceId string

			md, ok := metadata.FromIncomingContext(ctx)
			if ok && len(md.Get(TraceIdKey)) > 0 {
				traceId = md.Get(TraceIdKey)[0]
			} else {
				traceId = uuid.NewString()
				ctx = metadata.AppendToOutgoingContext(ctx, TraceIdKey, traceId)
			}

			ctx = context.WithValue(ctx, logger.LoggerKey, log)
			ctx = context.WithValue(ctx, TraceIdLogKey, traceId)

			return next(ctx, req)
		}
	}
}

func Logging[T any, T2 any](method string) Uniceptor[T, T2] {
	return func(next GrpcHandler[T, T2]) GrpcHandler[T, T2] {
		return func(ctx context.Context, req T) (resp T2, err error) {
			log := logger.GetLoggerFromCtx(ctx)

			log.Debug(ctx, "trace_id", zap.String(TraceIdLogKey, logger.TraceIDFromContext(ctx)))

			start := time.Now()
			res, handlerErr := next(ctx, req)
			elapsed := time.Since(start)

			if handlerErr != nil {
				code := erix.GrpcCode(handlerErr)
				err = status.Error(code, erix.LastReason(handlerErr))

				log.Error(ctx,
					"incoming request",
					zap.Duration("elapsed", elapsed),
					zap.String("method", method),
					zap.Error(handlerErr),
				)
			}

			return res, err //nolint:wrapcheck
		}
	}
}

func Recovery[T any, T2 any](method string) Uniceptor[T, T2] {
	return func(next GrpcHandler[T, T2]) GrpcHandler[T, T2] {
		return func(ctx context.Context, req T) (resp T2, err error) {
			log := logger.GetLoggerFromCtx(ctx)

			defer func() {
				recov := recover()
				if recov == nil {
					return
				}

				stack := stacktrace.Capture(0, math.MaxInt)

				frames := wrapFramesFromStack(stack)

				log.Error(ctx,
					"panic was recovered",
					zap.Any("panic", recov),
					zap.String("method", method),
					zap.Array("stacktrace", frames))

				err = status.Errorf(
					codes.Internal, "internal error (trace_id %s)",
					logger.TraceIDFromContext(ctx))
			}()

			return next(ctx, req)
		}
	}
}

// Wrapper for stacktrace frames to use in zap.Array.
type wrappedFrame struct {
	runtime.Frame
}

func (w wrappedFrame) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("file", w.File)
	enc.AddInt("line", w.Line)
	enc.AddString("function", w.Function)

	return nil
}

type wrappedFrames struct {
	frames []wrappedFrame
}

func wrapFramesFromStack(stack *stacktrace.Stack) wrappedFrames {
	var frames []wrappedFrame

	for _, frame := range stack.FramesIter() {
		frames = append(frames, wrappedFrame{Frame: frame})
	}

	return wrappedFrames{frames: frames}
}

func (w wrappedFrames) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, frame := range w.frames {
		err := enc.AppendObject(frame)
		if err != nil {
			return err
		}
	}

	return nil
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
