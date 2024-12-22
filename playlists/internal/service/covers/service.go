package covers

import (
	"context"
	"fmt"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage/minio"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage/postgres"
	"github.com/google/uuid"
	"io"
)

type ServiceCovers struct {
	objRepo ObjectRepository
	repo    Repository

	coverURLTmpl string
}

type ObjectRepository interface {
	GetCoverObject(
		ctx context.Context,
		id string,
	) (io.Reader, error)

	PutCoverObject(
		ctx context.Context,
		image minio.CoverObject,
	) error
}

type Repository interface {
	Playlist(ctx context.Context, id uuid.UUID) (postgres.PlaylistRow, error)
	PatchPlaylist(ctx context.Context, arg postgres.PatchPlaylistParams) (postgres.Playlist, error)
	BeginCovers(ctx context.Context) (Repository, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Config struct {
	HostUsesTLS bool
	Host        string
}

func New(repo Repository, objRepo ObjectRepository, cfg Config) *ServiceCovers {
	schema := "http"
	if cfg.HostUsesTLS {
		schema = "https"
	}

	return &ServiceCovers{
		repo:         repo,
		objRepo:      objRepo,
		coverURLTmpl: fmt.Sprintf("%s://%s/playlists/api/v1/playlist/{playlist_id}/cover/", schema, cfg.Host),
	}
}
