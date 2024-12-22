package gateway

import (
	"context"
	"go.uber.org/zap"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/lib/auth"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/erix"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/transport"

	"dev.gaijin.team/go/golib/stacktrace"
	"github.com/google/uuid"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type tokenKey struct{}

var (
	ErrInvalidAuthorizationHeader = erix.NewStatus("invalid authorization header", erix.CodeUnauthorized)
	ErrInvalidToken               = erix.NewStatus("invalid token", erix.CodeUnauthorized)
	ErrArtistTokenRequired        = erix.NewStatus("artist token required", erix.CodeForbidden)
)

func ContextWithLogger(log logger.Logger, next gateway.HandlerFunc) gateway.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		var traceID string

		traceID = r.Header.Get(transport.TraceIdKey)
		if traceID == "" {
			traceID = uuid.NewString()

			r.Header.Set(transport.TraceIdKey, traceID)
		}

		ctx := context.WithValue(r.Context(), logger.LoggerKey, log)
		r = r.WithContext(ctx)

		next(w, r, pathParams)
	}
}

func RecoveryMw(next gateway.HandlerFunc) gateway.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		log := logger.GetLoggerFromCtx(r.Context())

		defer func() { //nolint:contextcheck
			recov := recover()
			if recov == nil {
				return
			}

			stack := stacktrace.Capture(0, math.MaxInt)

			frames := wrapFramesFromStack(stack)

			log.Error(
				r.Context(), "panic was recovered",
				zap.Any("panic", recov),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
				zap.Array("stacktrace", frames))

			errText := "internal error (trace_id " + logger.TraceIDFromContext(r.Context()) + ")"
			jsonErr(w, erix.NewStatus(errText, erix.CodeInternalServerError))
		}()

		next(w, r, pathParams)
	}
}

type HandlerErrFunc func(http.ResponseWriter, *http.Request, map[string]string) error

func LoggingMw(next HandlerErrFunc) gateway.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		ctx := r.Context()

		log := logger.GetLoggerFromCtx(ctx)

		start := time.Now()
		handlerErr := next(w, r, pathParams)
		elapsed := time.Since(start)

		if handlerErr != nil {
			log.Error(
				ctx, "incoming request",
				zap.Duration("elapsed", elapsed),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
				zap.Error(handlerErr),
			)

			jsonErr(w, handlerErr)
		}
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

func TokenFromCtx(ctx context.Context) (auth.Token, bool) {
	token, ok := ctx.Value(tokenKey{}).(auth.Token)
	return token, ok
}

func CtxWithToken(ctx context.Context, token auth.Token) context.Context {
	return context.WithValue(ctx, tokenKey{}, token)
}
