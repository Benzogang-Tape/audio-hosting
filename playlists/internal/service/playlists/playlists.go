package playlists

import (
	"context"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/client/songs"

	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage/postgres"

	"github.com/google/uuid"
)

type ServicePlaylists struct {
	repo  PlaylistsRepo
	songs SongsRepo
}

type PlaylistsRepo interface {
	// crud
	SavePlaylist(ctx context.Context, arg postgres.SavePlaylistParams) error
	Playlist(ctx context.Context, id uuid.UUID) (postgres.PlaylistRow, error)
	DeletePlaylists(ctx context.Context, arg postgres.DeletePlaylistsParams) error
	PatchPlaylist(ctx context.Context, arg postgres.PatchPlaylistParams) (postgres.Playlist, error)
	UpdatePlaylist(ctx context.Context, arg postgres.UpdatePlaylistParams) (postgres.Playlist, error)
	PublicPlaylists(ctx context.Context, arg postgres.PublicPlaylistsParams) ([]postgres.PublicPlaylistsRow, error)

	// user's actions
	LikePlaylist(ctx context.Context, arg postgres.LikePlaylistParams) (uuid.UUID, error)
	DislikePlaylist(ctx context.Context, arg postgres.DislikePlaylistParams) error
	CopyPlaylist(ctx context.Context, arg postgres.CopyPlaylistParams) (uuid.UUID, error)
	UserPlaylists(ctx context.Context, userID uuid.UUID) ([]postgres.UserPlaylistsRow, error)
	MyCollection(ctx context.Context, userID uuid.UUID) ([]postgres.MyCollectionRow, error)
	DislikeTrack(ctx context.Context, arg postgres.DislikeTrackParams) error
	LikeTrack(ctx context.Context, arg postgres.LikeTrackParams) (uuid.UUID, error)

	// Transactions
	Begin(ctx context.Context) (PlaylistsRepo, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type SongsRepo interface {
	GetSong(ctx context.Context, id string) (songs.Song, error)
	GetSongs(ctx context.Context, id []string) ([]songs.Song, error)
	ReleaseSongs(ctx context.Context, ids []string) error
}

func New(repo PlaylistsRepo, songsRepo SongsRepo) *ServicePlaylists {
	return &ServicePlaylists{
		repo:  repo,
		songs: songsRepo,
	}
}
