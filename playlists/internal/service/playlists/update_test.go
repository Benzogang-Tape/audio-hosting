package playlists_test

import (
	"context"
	"github.com/AlekSi/pointer"
	playlistsmocks "github.com/Benzogang-Tape/audio-hosting/playlists/internal/mocks/playlists"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service/playlists"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type UpdatePlaylistSuite struct {
	suite.Suite

	pr *playlistsmocks.PlaylistsRepo
	sr *playlistsmocks.SongsRepo

	s     *playlists.ServicePlaylists
	ctx   context.Context
	input playlists.UpdatePlaylistInput
}

func (s *UpdatePlaylistSuite) SetupTest() {
	s.pr = playlistsmocks.NewPlaylistsRepo(s.T())
	s.sr = playlistsmocks.NewSongsRepo(s.T())

	s.s = playlists.New(s.pr, s.sr)

	log := logger.New("test", "prod")

	s.ctx = context.WithValue(context.Background(), logger.LoggerKey, log)
}

func (s *UpdatePlaylistSuite) TestHappyPath() {
	s.input = validUpdatePlaylistInput()

	s.pr.EXPECT().PatchPlaylist(mock.Anything, mock.Anything).Return(validPostgresPlaylist(s.input.PlaylistID), nil).Once()

	output, err := s.s.UpdatePlaylist(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *UpdatePlaylistSuite) TestNoRows() {
	s.input = validUpdatePlaylistInput()

	s.pr.EXPECT().PatchPlaylist(mock.Anything, mock.Anything).Return(postgres.Playlist{}, pgx.ErrNoRows).Once()

	output, err := s.s.UpdatePlaylist(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func (s *UpdatePlaylistSuite) TestDbError() {
	s.input = validUpdatePlaylistInput()

	s.pr.EXPECT().PatchPlaylist(mock.Anything, mock.Anything).Return(postgres.Playlist{}, gofakeit.ErrorDatabase()).Once()

	output, err := s.s.UpdatePlaylist(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func TestUpdatePlaylist(t *testing.T) {
	suite.Run(t, new(UpdatePlaylistSuite))
}

func validUpdatePlaylistInput() playlists.UpdatePlaylistInput {
	return playlists.UpdatePlaylistInput{
		PlaylistID: uuid.NewString(),
		UserID:     uuid.NewString(),
		Title:      pointer.To(gofakeit.BookTitle()),
		CoverURL:   pointer.To(gofakeit.URL()),
		IsPublic:   pointer.To(true),
	}
}

func validPostgresPlaylist(id string) postgres.Playlist {
	return postgres.Playlist{
		ID:         uuid.MustParse(id),
		Title:      gofakeit.BookTitle(),
		AuthorID:   uuid.UUID{},
		CoverUrl:   pgtype.Text{},
		TrackIds:   []uuid.UUID{},
		CreatedAt:  time.Now(),
		UpdatedAt:  pgtype.Timestamptz{},
		ReleasedAt: pgtype.Timestamptz{},
		IsAlbum:    false,
		IsPublic:   true,
	}
}

type ReleaseAlbumSuite struct {
	suite.Suite

	pr *playlistsmocks.PlaylistsRepo
	sr *playlistsmocks.SongsRepo

	s     *playlists.ServicePlaylists
	ctx   context.Context
	input playlists.ReleaseAlbumInput
}

func (s *ReleaseAlbumSuite) SetupTest() {
	s.pr = playlistsmocks.NewPlaylistsRepo(s.T())
	s.sr = playlistsmocks.NewSongsRepo(s.T())

	s.s = playlists.New(s.pr, s.sr)

	log := logger.New("test", "prod")

	s.ctx = context.WithValue(context.Background(), logger.LoggerKey, log)

	s.input = playlists.ReleaseAlbumInput{
		AlbumID:               gofakeit.UUID(),
		UserID:                gofakeit.UUID(),
		SuppressNotifications: gofakeit.Bool(),
	}
}

func (s *ReleaseAlbumSuite) TestHappyPath() {
	s.pr.EXPECT().Begin(mock.Anything).Return(s.pr, nil).Once()
	s.pr.EXPECT().PatchPlaylist(mock.Anything, mock.Anything).Return(validPostgresPlaylist(s.input.AlbumID), nil).Once()
	s.sr.EXPECT().ReleaseSongs(mock.Anything, mock.Anything).Return(nil).Once()
	s.pr.EXPECT().Commit(mock.Anything).Return(nil).Once()
	s.pr.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	err := s.s.ReleaseAlbum(s.ctx, s.input)
	s.NoError(err)
}

func (s *ReleaseAlbumSuite) TestNoRows() {
	s.pr.EXPECT().Begin(mock.Anything).Return(s.pr, nil).Once()
	s.pr.EXPECT().PatchPlaylist(mock.Anything, mock.Anything).Return(postgres.Playlist{}, pgx.ErrNoRows).Once()
	s.sr.EXPECT().ReleaseSongs(mock.Anything, mock.Anything).Return(nil).Maybe()
	s.pr.EXPECT().Commit(mock.Anything).Return(nil).Maybe()
	s.pr.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	err := s.s.ReleaseAlbum(s.ctx, s.input)
	s.Error(err)
}

func (s *ReleaseAlbumSuite) TestDbError() {
	s.pr.EXPECT().Begin(mock.Anything).Return(s.pr, nil).Once()
	s.pr.EXPECT().PatchPlaylist(mock.Anything, mock.Anything).Return(postgres.Playlist{}, gofakeit.ErrorDatabase()).Once()
	s.sr.EXPECT().ReleaseSongs(mock.Anything, mock.Anything).Return(nil).Maybe()
	s.pr.EXPECT().Commit(mock.Anything).Return(nil).Maybe()
	s.pr.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	err := s.s.ReleaseAlbum(s.ctx, s.input)
	s.Error(err)
}

func (s *ReleaseAlbumSuite) TestBeginError() {
	s.pr.EXPECT().Begin(mock.Anything).Return(s.pr, gofakeit.ErrorDatabase()).Once()
	s.pr.EXPECT().PatchPlaylist(mock.Anything, mock.Anything).Return(postgres.Playlist{}, gofakeit.ErrorDatabase()).Maybe()
	s.sr.EXPECT().ReleaseSongs(mock.Anything, mock.Anything).Return(nil).Maybe()
	s.pr.EXPECT().Commit(mock.Anything).Return(nil).Maybe()
	s.pr.EXPECT().Rollback(mock.Anything).Return(nil).Maybe()

	err := s.s.ReleaseAlbum(s.ctx, s.input)
	s.Error(err)
}

func (s *ReleaseAlbumSuite) TestReleaseSongsError() {
	s.pr.EXPECT().Begin(mock.Anything).Return(s.pr, nil).Once()
	s.pr.EXPECT().PatchPlaylist(mock.Anything, mock.Anything).Return(validPostgresPlaylist(s.input.AlbumID), nil).Maybe()
	s.sr.EXPECT().ReleaseSongs(mock.Anything, mock.Anything).Return(gofakeit.Error()).Maybe()
	s.pr.EXPECT().Commit(mock.Anything).Return(nil).Maybe()
	s.pr.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	err := s.s.ReleaseAlbum(s.ctx, s.input)
	s.Error(err)
}

func (s *ReleaseAlbumSuite) TestCommitError() {
	s.pr.EXPECT().Begin(mock.Anything).Return(s.pr, nil).Once()
	s.pr.EXPECT().PatchPlaylist(mock.Anything, mock.Anything).Return(validPostgresPlaylist(s.input.AlbumID), nil).Once()
	s.sr.EXPECT().ReleaseSongs(mock.Anything, mock.Anything).Return(nil).Once()
	s.pr.EXPECT().Commit(mock.Anything).Return(gofakeit.Error()).Once()
	s.pr.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	err := s.s.ReleaseAlbum(s.ctx, s.input)
	s.Error(err)
}

func TestReleaseAlbum(t *testing.T) {
	suite.Run(t, new(ReleaseAlbumSuite))
}
