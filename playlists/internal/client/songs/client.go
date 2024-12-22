package songs

import (
	"context"
	"strconv"
	"time"

	"dev.gaijin.team/go/golib/e"
	"github.com/AlekSi/pointer"
	songs "github.com/Benzogang-Tape/audio-hosting/playlists/api/protogen/clients/songs"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/transport/grpc/uniceptors"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/transport"
	retryer "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	client songs.SongsServiceClient
	conn   *grpc.ClientConn
}

type Config struct {
	Host    string        `env:"SONGS_HOST" env-default:"localhost" yaml:"host"`
	Port    int           `env:"SONGS_PORT" env-default:"50052" yaml:"port"`
	Retries uint          `env:"SONGS_RETRIES" env-default:"5" yaml:"retries"`
	Timeout time.Duration `env:"SONGS_TIMEOUT" env-default:"5s" yaml:"timeout"`
}

func New(cfg Config) (*Client, error) {
	retryOpts := []retryer.CallOption{
		retryer.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		retryer.WithMax(cfg.Retries),
		retryer.WithPerRetryTimeout(cfg.Timeout),
	}

	conn, err := grpc.NewClient(
		cfg.Host+":"+strconv.Itoa(cfg.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			retryer.UnaryClientInterceptor(retryOpts...),
		),
	)
	if err != nil {
		return nil, e.NewFrom("creating songs client", err)
	}

	return &Client{
		client: songs.NewSongsServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *Client) Close() error {
	err := c.conn.Close()
	if err != nil {
		return e.NewFrom("closings songs client conn", err)
	}

	return nil
}

func (c *Client) GetSong(ctx context.Context, id string) (Song, error) {
	token, _ := uniceptors.RawTokenFromCtx(ctx)

	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		transport.TraceIdKey: logger.TraceIDFromContext(ctx),
		transport.AuthKey:    token,
	}))

	params := &songs.GetSongRequest{
		Id: id,
	}

	resp, err := c.client.GetSong(ctx, params)

	if err != nil || resp.GetSong() == nil {
		return Song{}, e.NewFrom("getting song", err)
	}

	return convSong(resp.GetSong()), nil
}

func (c *Client) GetSongs(ctx context.Context, ids []string) ([]Song, error) {
	token, _ := uniceptors.RawTokenFromCtx(ctx)

	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		transport.TraceIdKey: logger.TraceIDFromContext(ctx),
		transport.AuthKey:    token,
	}))

	params := &songs.GetSongsRequest{ //nolint:exhaustruct
		Page:     pointer.To(int32(1)),
		PageSize: pointer.To(int32(len(ids))), //nolint:gosec
		Ids:      ids,
	}

	resp, err := c.client.GetSongs(ctx, params)
	if err != nil || len(resp.GetSongs()) == 0 {
		return nil, e.NewFrom("getting songs", err)
	}

	tracks := make([]Song, 0, len(resp.GetSongs()))
	for _, song := range resp.GetSongs() {
		tracks = append(tracks, convSong(song))
	}

	ordTracks := c.getOrderedSongs(ctx, ids, tracks)

	return ordTracks, nil
}

func (*Client) getOrderedSongs(ctx context.Context, ids []string, tracks []Song) []Song {
	var orderedTracks []Song

	log := logger.GetLoggerFromCtx(ctx)

	for _, id := range ids {
		for _, track := range tracks {
			if track.ID == id {
				orderedTracks = append(orderedTracks, track)
				break
			}
		}
	}

	log.Debug(
		ctx, "ordered tracks",
		zap.String("layout", "client/songs"),
		zap.Int("count ids", len(ids)),
		zap.Int("count tracks", len(orderedTracks)),
		zap.Int("loss", len(ids)-len(orderedTracks)))

	return orderedTracks
}

func (c *Client) ReleaseSongs(ctx context.Context, ids []string) error {
	token, ok := uniceptors.RawTokenFromCtx(ctx)
	if !ok {
		return e.New("token not found in ctx")
	}

	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		transport.TraceIdKey: logger.TraceIDFromContext(ctx),
		transport.AuthKey:    token,
	}))

	params := &songs.ReleaseSongsRequest{
		Ids:    ids,
		Notify: false,
	}

	_, err := c.client.ReleaseSongs(ctx, params)
	if err != nil {
		return e.NewFrom("releasing songs", err)
	}

	return nil
}
