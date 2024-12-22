package playlists_test

import (
	"context"
	client "github.com/Benzogang-Tape/audio-hosting/playlists/internal/client/songs"
	playlistsmocks "github.com/Benzogang-Tape/audio-hosting/playlists/internal/mocks/playlists"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service/playlists"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/pgconv"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"slices"
	"testing"
	"time"
)

type LikeDislikeSuit struct {
	suite.Suite

	pr *playlistsmocks.PlaylistsRepo
	sr *playlistsmocks.SongsRepo

	input playlists.LikeDislikePlaylistInput
	s     *playlists.ServicePlaylists
	ctx   context.Context
}

func (s *LikeDislikeSuit) SetupTest() {
	s.pr = playlistsmocks.NewPlaylistsRepo(s.T())
	s.sr = playlistsmocks.NewSongsRepo(s.T())

	s.s = playlists.New(s.pr, s.sr)

	log := logger.New("test", "prod")

	s.ctx = context.WithValue(context.Background(), logger.LoggerKey, log)
	s.input = playlists.LikeDislikePlaylistInput{
		PlaylistID: gofakeit.UUID(),
		UserID:     gofakeit.UUID(),
	}
}

func (s *LikeDislikeSuit) TestHappyPathForLike() {
	s.pr.EXPECT().LikePlaylist(mock.Anything, mock.Anything).Return(uuid.New(), nil).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]client.Song{}, nil).Maybe()

	err := s.s.LikePlaylist(s.ctx, s.input)
	s.NoError(err)
}

func (s *LikeDislikeSuit) TestErrorForLike() {
	s.pr.EXPECT().LikePlaylist(mock.Anything, mock.Anything).Return(uuid.Nil, gofakeit.ErrorDatabase()).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]client.Song{}, nil).Maybe()

	err := s.s.LikePlaylist(s.ctx, s.input)
	s.Error(err)
}

func (s *LikeDislikeSuit) TestLikeUnavailablePlaylist() {
	s.pr.EXPECT().LikePlaylist(mock.Anything, mock.Anything).Return(uuid.Nil, pgx.ErrNoRows).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]client.Song{}, nil).Maybe()

	err := s.s.LikePlaylist(s.ctx, s.input)
	s.Error(err)
}

func (s *LikeDislikeSuit) TestHappyPathForDislike() {
	s.pr.EXPECT().DislikePlaylist(mock.Anything, mock.Anything).Return(nil).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]client.Song{}, nil).Maybe()

	err := s.s.DislikePlaylist(s.ctx, s.input)
	s.NoError(err)
}

func TestLikeDislikePlaylist(t *testing.T) {
	suite.Run(t, new(LikeDislikeSuit))
}

type CopyPlaylistSuit struct {
	suite.Suite

	pr *playlistsmocks.PlaylistsRepo
	sr *playlistsmocks.SongsRepo

	input playlists.CopyPlaylistInput
	s     *playlists.ServicePlaylists
	ctx   context.Context
}

func (s *CopyPlaylistSuit) SetupTest() {
	s.pr = playlistsmocks.NewPlaylistsRepo(s.T())
	s.sr = playlistsmocks.NewSongsRepo(s.T())

	s.s = playlists.New(s.pr, s.sr)

	log := logger.New("test", "prod")

	s.ctx = context.WithValue(context.Background(), logger.LoggerKey, log)
	s.input = playlists.CopyPlaylistInput{
		PlaylistID: gofakeit.UUID(),
		UserID:     gofakeit.UUID(),
	}
}

func (s *CopyPlaylistSuit) TestHappyPath() {
	s.pr.EXPECT().CopyPlaylist(mock.Anything, mock.Anything).Return(uuid.New(), nil).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]client.Song{}, nil).Maybe()

	id, err := s.s.CopyPlaylist(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(id)
}

func (s *CopyPlaylistSuit) TestCopyOwnPlaylist() {
	// Why pgx.ErrNoRows: because it won't insert any row
	s.pr.EXPECT().CopyPlaylist(mock.Anything, mock.Anything).Return(uuid.Nil, pgx.ErrNoRows).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]client.Song{}, nil).Maybe()

	id, err := s.s.CopyPlaylist(s.ctx, s.input)
	s.Error(err)
	s.Empty(id)
}

func (s *CopyPlaylistSuit) TestCopyPrivatePlaylist() {
	s.pr.EXPECT().CopyPlaylist(mock.Anything, mock.Anything).Return(uuid.Nil, pgx.ErrNoRows).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]client.Song{}, nil).Maybe()

	id, err := s.s.CopyPlaylist(s.ctx, s.input)
	s.Error(err)
	s.Empty(id)
}

func (s *CopyPlaylistSuit) TestDbError() {
	s.pr.EXPECT().CopyPlaylist(mock.Anything, mock.Anything).Return(uuid.Nil, gofakeit.ErrorDatabase()).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]client.Song{}, nil).Maybe()

	id, err := s.s.CopyPlaylist(s.ctx, s.input)
	s.Error(err)
	s.Empty(id)
}

func TestCopyPlaylist(t *testing.T) {
	suite.Run(t, new(CopyPlaylistSuit))
}

type GetMyPlaylistsSuit struct {
	suite.Suite

	pr *playlistsmocks.PlaylistsRepo
	sr *playlistsmocks.SongsRepo

	s     *playlists.ServicePlaylists
	ctx   context.Context
	input string
}

func (s *GetMyPlaylistsSuit) SetupTest() {
	s.pr = playlistsmocks.NewPlaylistsRepo(s.T())
	s.sr = playlistsmocks.NewSongsRepo(s.T())

	s.s = playlists.New(s.pr, s.sr)

	log := logger.New("test", "prod")

	s.ctx = context.WithValue(context.Background(), logger.LoggerKey, log)
	s.input = gofakeit.UUID()
}

func (s *GetMyPlaylistsSuit) TestHappyPath() {
	s.pr.EXPECT().UserPlaylists(mock.Anything, mock.Anything).Return(validUserPlaylistRows(s.input), nil).Once()

	output, err := s.s.GetMyPlaylists(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *GetMyPlaylistsSuit) TestDbError() {
	s.pr.EXPECT().UserPlaylists(mock.Anything, mock.Anything).Return(nil, gofakeit.ErrorDatabase()).Once()

	output, err := s.s.GetMyPlaylists(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func (s *GetMyPlaylistsSuit) TestNoRows() {
	s.pr.EXPECT().UserPlaylists(mock.Anything, mock.Anything).Return(nil, pgx.ErrNoRows).Once()

	output, err := s.s.GetMyPlaylists(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func TestGetMyPlaylists(t *testing.T) {
	suite.Run(t, new(GetMyPlaylistsSuit))
}

func validUserPlaylistRows(id string) []postgres.UserPlaylistsRow {
	var rows []postgres.UserPlaylistsRow

	for i := 0; i < 10; i++ {
		rows = append(rows, postgres.UserPlaylistsRow{
			ID:        uuid.New(),
			Title:     gofakeit.BookTitle(),
			AuthorID:  uuid.MustParse(id),
			CoverUrl:  pgconv.Text(gofakeit.URL()),
			CreatedAt: gofakeit.Date(),
			IsAlbum:   false,
			IsPublic:  true,
		})
	}

	return rows
}

type GetMyCollectionSuit struct {
	suite.Suite

	pr *playlistsmocks.PlaylistsRepo
	sr *playlistsmocks.SongsRepo

	s     *playlists.ServicePlaylists
	ctx   context.Context
	input string
}

func (s *GetMyCollectionSuit) SetupTest() {
	s.pr = playlistsmocks.NewPlaylistsRepo(s.T())
	s.sr = playlistsmocks.NewSongsRepo(s.T())

	s.s = playlists.New(s.pr, s.sr)

	log := logger.New("test", "prod")

	s.ctx = context.WithValue(context.Background(), logger.LoggerKey, log)
	s.input = gofakeit.UUID()
}

func (s *GetMyCollectionSuit) TestHappyPath() {
	s.pr.EXPECT().MyCollection(mock.Anything, mock.Anything).Return(validMyCollectionRows(), nil).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]client.Song{}, nil).Once()

	output, err := s.s.GetMyCollection(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *GetMyCollectionSuit) TestDbError() {
	s.pr.EXPECT().MyCollection(mock.Anything, mock.Anything).Return(nil, gofakeit.ErrorDatabase()).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]client.Song{}, nil).Maybe()

	output, err := s.s.GetMyCollection(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func (s *GetMyCollectionSuit) TestNoRows() {
	s.pr.EXPECT().MyCollection(mock.Anything, mock.Anything).Return(nil, pgx.ErrNoRows).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]client.Song{}, nil).Maybe()

	output, err := s.s.GetMyCollection(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *GetMyCollectionSuit) TestSongsServiceError() {
	s.pr.EXPECT().MyCollection(mock.Anything, mock.Anything).Return(validMyCollectionRows(), nil).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return(nil, gofakeit.ErrorDatabase()).Once()

	output, err := s.s.GetMyCollection(s.ctx, s.input)
	s.Error(err)
	s.NotEmpty(output)
}

func TestGetMyCollection(t *testing.T) {
	suite.Run(t, new(GetMyCollectionSuit))
}

func validMyCollectionRows() []postgres.MyCollectionRow {
	var rows []postgres.MyCollectionRow

	for i := 0; i < 10; i++ {
		rows = append(rows, postgres.MyCollectionRow{
			TrackID: uuid.UUID{},
			LikedAt: time.Time{},
		})
	}

	slices.SortFunc(rows, func(a, b postgres.MyCollectionRow) int {
		if a.LikedAt.Before(b.LikedAt) {
			return 1
		} else if a.LikedAt.After(b.LikedAt) {
			return -1
		}

		return 0
	})

	return rows
}

type LikeDislikeSuite struct {
	suite.Suite

	pr *playlistsmocks.PlaylistsRepo
	sr *playlistsmocks.SongsRepo

	s     *playlists.ServicePlaylists
	ctx   context.Context
	input playlists.LikeDislikeTrackInput
}

func (s *LikeDislikeSuite) SetupTest() {
	s.pr = playlistsmocks.NewPlaylistsRepo(s.T())
	s.sr = playlistsmocks.NewSongsRepo(s.T())

	s.s = playlists.New(s.pr, s.sr)

	log := logger.New("test", "prod")

	s.ctx = context.WithValue(context.Background(), logger.LoggerKey, log)
	s.input = playlists.LikeDislikeTrackInput{
		TrackID: uuid.NewString(),
		UserID:  uuid.NewString(),
	}
}

func (s *LikeDislikeSuite) TestHappyPathForLike() {
	s.sr.EXPECT().GetSong(mock.Anything, mock.Anything).Return(client.Song{}, nil).Once()
	s.pr.EXPECT().LikeTrack(mock.Anything, mock.Anything).Return(uuid.MustParse(s.input.TrackID), nil).Once()

	err := s.s.LikeTrack(s.ctx, s.input)
	s.NoError(err)
}

func (s *LikeDislikeSuite) TestDbErrorForLike() {
	s.sr.EXPECT().GetSong(mock.Anything, mock.Anything).Return(client.Song{}, nil).Once()
	s.pr.EXPECT().LikeTrack(mock.Anything, mock.Anything).Return(uuid.Nil, gofakeit.ErrorDatabase()).Once()

	err := s.s.LikeTrack(s.ctx, s.input)
	s.Error(err)
}

func (s *LikeDislikeSuite) TestSongDoesNotExistForLike() {
	s.sr.EXPECT().GetSong(mock.Anything, mock.Anything).Return(client.Song{}, status.Error(codes.NotFound, "test")).Once()

	err := s.s.LikeTrack(s.ctx, s.input)
	s.Error(err)
}

func (s *LikeDislikeSuite) TestSongsServiceUnavailableForLike() {
	s.sr.EXPECT().GetSong(mock.Anything, mock.Anything).Return(client.Song{}, status.Error(codes.Unavailable, "test")).Once()

	err := s.s.LikeTrack(s.ctx, s.input)
	s.Error(err)
}

func (s *LikeDislikeSuite) TestHappyPathForDislike() {
	s.pr.EXPECT().DislikeTrack(mock.Anything, mock.Anything).Return(nil).Once()

	err := s.s.DislikeTrack(s.ctx, s.input)
	s.NoError(err)
}

func (s *LikeDislikeSuite) TestDbErrorForDislike() {
	s.pr.EXPECT().DislikeTrack(mock.Anything, mock.Anything).Return(gofakeit.ErrorDatabase()).Once()

	err := s.s.DislikeTrack(s.ctx, s.input)
	s.Error(err)
}

func TestLikeDislikeTrack(t *testing.T) {
	suite.Run(t, new(LikeDislikeSuite))
}
