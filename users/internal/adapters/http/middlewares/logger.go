package middlewares

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/users/pkg/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func LoggerInterceptor(next runtime.HandlerFunc) runtime.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		start := time.Now()

		log := logger.GetLoggerFromCtx(r.Context())

		log.Info(
			"get request",
			slog.String("method", r.Method),
		)

		next(w, r, pathParams)

		defer func() {
			log.Info(
				"request completed",
				slog.String("method", r.Method),
				slog.String("duration", time.Since(start).String()),
			)
		}()
	}
}

func LoggerToCtxInterceptor(log *slog.Logger) func(next runtime.HandlerFunc) runtime.HandlerFunc {
	return func(next runtime.HandlerFunc) runtime.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			ctx := logger.SetLoggerToCtx(r.Context(), log)

			next(w, r.WithContext(ctx), pathParams)
		}
	}
}
