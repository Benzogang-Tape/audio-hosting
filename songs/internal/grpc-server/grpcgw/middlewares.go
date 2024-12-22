package grpcgw

import (
	"context"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/audio/auth"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/erix"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/transport"

	"dev.gaijin.team/go/golib/stacktrace"
	"github.com/google/uuid"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
)

type tokenKey struct{}

var (
	ErrInvalidAuthorizationHeader = erix.NewStatus("invalid authorization header", erix.CodeUnauthorized)
	ErrInvalidToken               = erix.NewStatus("invalid token", erix.CodeUnauthorized)
	ErrArtistTokenRequired        = erix.NewStatus("artist token required", erix.CodeForbidden)
)

func ContextWithLogger(base zerolog.Logger, next gateway.HandlerFunc) gateway.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		var traceId string

		traceId = r.Header.Get(transport.TraceIdKey)
		if traceId == "" {
			traceId = uuid.NewString()
		}

		log := base.With().
			Str(transport.TraceIdLogKey, traceId).
			Logger()

		ctx := logger.WithLoggerAndTraceId(r.Context(), log, traceId)
		r = r.WithContext(ctx)

		next(w, r, pathParams)
	}
}

func RecoveryMw(next gateway.HandlerFunc) gateway.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		log := logger.FromContext(r.Context())

		defer func() { //nolint:contextcheck
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
				Str("path", r.URL.Path).
				Str("method", r.Method).
				Array("stacktrace", marshaledStack).
				Msg("panic was recovered")

			errText := "internal error (trace_id " + logger.TraceIdFromContext(r.Context()) + ")"
			jsonErr(w, erix.NewStatus(errText, erix.CodeInternalServerError))
		}()

		next(w, r, pathParams)
	}
}

type HandlerErrFunc func(http.ResponseWriter, *http.Request, map[string]string) error

func LoggingMw(next HandlerErrFunc) gateway.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		log := logger.FromContext(r.Context())

		start := time.Now()
		handlerErr := next(w, r, pathParams)
		elapsed := time.Since(start)

		var logLevel zerolog.Level = zerolog.InfoLevel

		if handlerErr != nil {
			if erix.HttpCode(handlerErr) >= http.StatusInternalServerError {
				logLevel = zerolog.ErrorLevel
			}

			jsonErr(w, handlerErr)
		}

		log.WithLevel(logLevel).
			Err(handlerErr).
			Dur("elapsed", elapsed).
			Str("path", r.URL.Path).
			Str("method", r.Method).
			Msg("incoming request")
	}
}

func AuthMw(mustArtist bool, parser auth.Parser, next HandlerErrFunc) HandlerErrFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) error {
		authParts := strings.Split(r.Header.Get("Authorization"), " ")

		if len(authParts) != 2 || authParts[0] != "Bearer" {
			return ErrInvalidAuthorizationHeader
		}

		token, err := parser.Parse(authParts[1])
		if err != nil {
			return ErrInvalidToken.Wrap(err)
		}

		if mustArtist && !token.IsArtist {
			return ErrArtistTokenRequired
		}

		r = r.WithContext(context.WithValue(r.Context(), tokenKey{}, token))

		return next(w, r, pathParams)
	}
}

func TokenFromCtx(ctx context.Context) auth.Token {
	tkn, ok := ctx.Value(tokenKey{}).(auth.Token)
	if !ok {
		panic("token not found in context")
	}

	return tkn
}
