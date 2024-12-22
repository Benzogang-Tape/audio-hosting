package uniceptors

import (
	"context"
	"strings"

	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/lib/auth"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/erix"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/transport"

	"google.golang.org/grpc/metadata"
)

type TokenParser interface {
	Parse(token string) (auth.Token, error)
}

var (
	ErrNotArtist          = erix.NewStatus("insufficient rights: not artist", erix.CodeForbidden)
	ErrInsufficientRights = erix.NewStatus("insufficient rights", erix.CodeForbidden)
)

type tokenKey struct{}

func Auth[T, T2 any](mustArtist bool, parser TokenParser) transport.Uniceptor[T, T2] {
	return func(next transport.GrpcHandler[T, T2]) transport.GrpcHandler[T, T2] {
		return func(ctx context.Context, req T) (T2, error) {
			const authorizationHeaderPartsCount = 2

			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return *new(T2), ErrInsufficientRights.WithField("md", md)
			}

			authValue := md.Get(transport.AuthKey)
			if len(authValue) == 0 {
				return *new(T2), ErrInsufficientRights.WithField("authValue", authValue)
			}

			ctx = context.WithValue(ctx, transport.AuthKey, authValue[0])

			authToken := strings.Split(authValue[0], " ")
			if len(authToken) != authorizationHeaderPartsCount {
				return *new(T2), ErrInsufficientRights.WithField("authToken", authToken)
			}

			token, err := parser.Parse(authToken[1])
			if err != nil {
				return *new(T2), ErrInsufficientRights.Wrap(err)
			}

			if mustArtist && !token.IsArtist {
				return *new(T2), ErrNotArtist
			}

			ctx = context.WithValue(ctx, tokenKey{}, token)
			ctx = context.WithValue(ctx, transport.AuthKey, authToken[1])

			return next(ctx, req)
		}
	}
}

func ParseToken[T, T2 any](parser TokenParser) transport.Uniceptor[T, T2] {
	return func(next transport.GrpcHandler[T, T2]) transport.GrpcHandler[T, T2] {
		return func(ctx context.Context, req T) (T2, error) {
			const authorizationHeaderPartsCount = 2

			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return next(ctx, req)
			}

			authValue := md.Get(transport.AuthKey)
			if len(authValue) == 0 {
				return next(ctx, req)
			}

			authToken := strings.Split(authValue[0], " ")
			if len(authToken) != authorizationHeaderPartsCount {
				return next(ctx, req)
			}

			token, err := parser.Parse(authToken[1])
			if err != nil {
				return next(ctx, req)
			}

			ctx = context.WithValue(ctx, tokenKey{}, token)
			ctx = context.WithValue(ctx, transport.AuthKey, authToken[1])

			return next(ctx, req)
		}
	}
}

func TokenFromCtx(ctx context.Context) (auth.Token, bool) {
	token, ok := ctx.Value(tokenKey{}).(auth.Token)
	return token, ok
}

func CtxWithToken(ctx context.Context, token auth.Token) context.Context {
	return context.WithValue(ctx, tokenKey{}, token)
}

func RawTokenFromCtx(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(transport.AuthKey).(string)
	if !ok {
		return "", false
	}
	return "Bearer " + token, ok
}
