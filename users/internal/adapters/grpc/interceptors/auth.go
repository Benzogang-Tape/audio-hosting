package interceptors

import (
	"context"
	"errors"
	"strings"

	"github.com/Benzogang-Tape/audio-hosting/users/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AuthInterceptor struct {
	parser       auth.Parser
	publicRoutes []string
}

func NewAuthInterceptor(parser auth.Parser, publicRoutes []string) *AuthInterceptor {
	return &AuthInterceptor{
		parser:       parser,
		publicRoutes: publicRoutes,
	}
}

const (
	AuthorizationMetadataName = "authorization"
)

type ClaimsCtx struct{}

var (
	ErrInvalidMetadata = errors.New("invalid metadata")
	ErrUnauthorized    = errors.New("unauthorized")
)

func (ai *AuthInterceptor) JWT(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	for _, route := range ai.publicRoutes {
		if route == info.FullMethod {
			return handler(ctx, req)
		}
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrInvalidMetadata
	}

	header := md.Get(AuthorizationMetadataName)
	if len(header) == 0 {
		return nil, ErrInvalidMetadata
	}

	headerParts := strings.Split(header[0], " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, ErrInvalidMetadata
	}

	claims, err := ai.parser.Parse(headerParts[1])
	if err != nil {
		return nil, ErrUnauthorized
	}

	return handler(context.WithValue(ctx, ClaimsCtx{}, claims), req)
}
