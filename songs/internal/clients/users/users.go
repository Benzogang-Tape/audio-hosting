package users

import (
	"context"
	"slices"

	"github.com/Benzogang-Tape/audio-hosting/songs/api/protogen/api/clients/users"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/config"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/repoerrs"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/transport"

	"dev.gaijin.team/go/golib/e"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	c    users.UsersServiceClient
	conn *grpc.ClientConn
}

type Config struct {
	Target string
}

func New() (*Client, error) {
	conf := config.Get()

	return NewWithConfig(Config{
		Target: conf.Connections.UsersService.Target,
	})
}

func NewWithConfig(conf Config) (*Client, error) {
	conn, err := grpc.NewClient(conf.Target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, e.NewFrom("new users client", err)
	}

	return &Client{
		c:    users.NewUsersServiceClient(conn),
		conn: conn,
	}, nil
}

func (c *Client) Close() error {
	err := c.conn.Close()
	if err != nil {
		return e.NewFrom("close users client", err)
	}

	return nil
}

func (c *Client) ArtistsByIds(ctx context.Context, ids []uuid.UUID) ([]Artist, error) {
	log := logger.FromContext(ctx)

	log.Debug().Array("ids", logger.Stringers[uuid.UUID](ids)).Msg("getting artists by ids")

	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		transport.TraceIdKey: logger.TraceIdFromContext(ctx),
	}))

	filters := make([]*users.Filter, len(ids))
	for i := range ids {
		filters[i] = &users.Filter{
			Field:    "id",
			Operator: "eq",
			Value:    ids[i].String(),
		}
	}

	resp, err := c.c.GetArtists(ctx, &users.GetArtistsRequest{ //nolint:exhaustruct
		Filter: filters,
		Pagination: &users.PaginationRequest{
			Offset: 0,
			Limit:  int64(len(ids)),
		},
	})
	if err != nil {
		return nil, e.NewFrom("get artists in client", err)
	}

	if len(resp.GetArtists()) == 0 {
		return nil, repoerrs.ErrEmptyResult
	}

	result := make([]Artist, 0, len(resp.GetArtists()))

	// Linear search is better here than binary one
	// because we have a small number of artists.
	for _, id := range ids {
		idx := slices.IndexFunc(resp.GetArtists(), func(artist *users.Artist) bool {
			return artist.GetId() == id.String()
		})
		if idx == -1 {
			continue
		}

		artist := resp.GetArtists()[idx]

		result = append(result, Artist{
			Id:        uuid.MustParse(artist.GetId()),
			Name:      artist.GetName(),
			Label:     artist.GetLabel(),
			AvatarUrl: artist.GetAvatarUrl(),
		})
	}

	return result, nil
}

func (c *Client) ArtistsMatchingName(ctx context.Context, name string) ([]Artist, error) {
	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		transport.TraceIdKey: logger.TraceIdFromContext(ctx),
	}))

	const limitArtists = 100

	resp, err := c.c.GetArtists(ctx, &users.GetArtistsRequest{ //nolint:exhaustruct
		Filter: []*users.Filter{
			{
				Field:    "name",
				Operator: "like",
				Value:    name,
			},
		},
		Pagination: &users.PaginationRequest{
			Offset: int64(0),
			Limit:  int64(limitArtists),
		},
	})
	if len(resp.GetArtists()) == 0 {
		return nil, repoerrs.ErrEmptyResult
	}

	if err != nil {
		return nil, e.NewFrom("get artists by name in client", err)
	}

	result := make([]Artist, len(resp.GetArtists()))
	for i, artist := range resp.GetArtists() {
		result[i] = Artist{
			Id:        uuid.MustParse(artist.GetId()),
			Name:      artist.GetName(),
			Label:     artist.GetLabel(),
			AvatarUrl: artist.GetAvatarUrl(),
		}
	}

	return result, nil
}
