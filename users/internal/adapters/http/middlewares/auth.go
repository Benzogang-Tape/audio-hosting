package middlewares

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Benzogang-Tape/audio-hosting/users/pkg/auth"
	"github.com/Benzogang-Tape/audio-hosting/users/pkg/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type AuthMiddleware struct {
	parser       auth.Parser
	publicRoutes []string
}

func NewAuthMiddleware(parser auth.Parser, publicRoutes []string) *AuthMiddleware {
	return &AuthMiddleware{
		parser:       parser,
		publicRoutes: publicRoutes,
	}
}

const (
	authorizationHeader = "Authorization"
)

type ClaimsCtx struct{}

var ErrInvalidAuthHeader = errors.New("invalid auth header")

func (am *AuthMiddleware) JWT() func(next runtime.HandlerFunc) runtime.HandlerFunc {
	return func(next runtime.HandlerFunc) runtime.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			for _, route := range am.publicRoutes {
				if route == r.URL.Path {
					next(w, r, pathParams)
					return
				}
			}

			header, err := parseAuthHeaders(r)
			if err != nil {
				if errors.Is(err, ErrInvalidAuthHeader) {
					logger.GetLoggerFromCtx(r.Context()).
						Error("invalid auth header", slog.String("error", err.Error()))

					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}

				logger.GetLoggerFromCtx(r.Context()).
					Error("internal error", slog.String("error", err.Error()))

				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			claims, err := am.parser.Parse(header[1])
			if err != nil {
				logger.GetLoggerFromCtx(r.Context()).
					Error("internal error", slog.String("error", err.Error()))

				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, ClaimsCtx{}, claims)
			r = r.WithContext(ctx)

			next(w, r, pathParams)
		}
	}
}

func parseAuthHeaders(r *http.Request) ([]string, error) {
	authHeader := r.Header.Get(authorizationHeader)
	if authHeader == "" {
		return nil, ErrInvalidAuthHeader
	}

	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, ErrInvalidAuthHeader
	}

	if len(headerParts[1]) == 0 {
		return nil, ErrInvalidAuthHeader
	}

	return headerParts, nil
}
