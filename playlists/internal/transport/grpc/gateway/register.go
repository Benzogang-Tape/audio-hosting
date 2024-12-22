package gateway

import (
	"context"
	"dev.gaijin.team/go/golib/e"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/lib/auth"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service/covers"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"net/http"
)

type CoversService interface {
	GetRawCover(ctx context.Context, playlistID string) (covers.GetRawCoverOutput, error)
	UploadRawCover(ctx context.Context, input covers.UploadRawCoverInput) (covers.UploadRawCoverOutput, error)
}

type CoversHandlers struct {
	service CoversService
}

func RegisterHandlers(
	mux *gateway.ServeMux,
	log logger.Logger,
	coversService CoversService,
	pub string,
) error {
	tokenParser, err := auth.NewParser(pub)
	if err != nil {
		return e.NewFrom("creating token parser", err)
	}

	h := CoversHandlers{
		service: coversService,
	}

	mws := func(hand HandlerErrFunc) gateway.HandlerFunc {
		return ContextWithLogger(log,
			RecoveryMw(
				LoggingMw(hand)))
	}
	authMws := func(hand HandlerErrFunc) gateway.HandlerFunc {
		return mws(AuthMw(true, tokenParser, hand))
	}

	err = mux.HandlePath(
		http.MethodPost,
		"/playlists/api/v1/playlist/{playlist_id}/cover",
		authMws(h.UploadRawCoverHandler()),
	)
	if err != nil {
		return e.NewFrom("register post playlist/cover", err)
	}

	err = mux.HandlePath(
		http.MethodGet,
		"/playlists/api/v1/playlist/{playlist_id}/cover/{cover_id}",
		mws(h.GetRawCoverHandler()),
	)
	if err != nil {
		return e.NewFrom("register get playlist/cover", err)
	}

	return nil
}
