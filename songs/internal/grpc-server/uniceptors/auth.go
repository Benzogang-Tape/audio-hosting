package uniceptors

import (
	"context"
	"strings"

	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/audio/auth"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/erix"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/transport"

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
				return *new(T2), ErrInsufficientRights
			}

			authValue := md.Get(transport.AuthKey)
			if len(authValue) == 0 {
				return *new(T2), ErrInsufficientRights
			}

			authToken := strings.Split(authValue[0], " ")
			if len(authToken) != authorizationHeaderPartsCount {
				return *new(T2), ErrInsufficientRights
			}

			token, err := parser.Parse(authToken[1])
			if err != nil {
				return *new(T2), ErrInsufficientRights.Wrap(err)
			}

			if mustArtist && !token.IsArtist {
				return *new(T2), ErrNotArtist
			}

			ctx = context.WithValue(ctx, tokenKey{}, token)

			return next(ctx, req)

		}
	}
}

func TokenFromCtx(ctx context.Context) auth.Token {
	token, ok := ctx.Value(tokenKey{}).(auth.Token)
	if !ok {
		panic("token not found in context")
	}

	return token
}
