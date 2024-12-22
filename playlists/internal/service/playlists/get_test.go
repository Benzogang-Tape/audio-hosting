package playlists_test

import (
	"context"
	"github.com/AlekSi/pointer"
	client "github.com/Benzogang-Tape/audio-hosting/playlists/internal/client/songs"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/lib/auth"
	playlistsmocks "github.com/Benzogang-Tape/audio-hosting/playlists/internal/mocks/playlists"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/models"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service/playlists"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/transport/grpc/uniceptors"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/pgconv"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type GetPlaylistSuite struct {
	suite.Suite

	pr *playlistsmocks.PlaylistsRepo
	sr *playlistsmocks.SongsRepo

	s     *playlists.ServicePlaylists
	ctx   context.Context
	input string
}

// TODO: do test when songsService gives error (after implementing client)
func (s *GetPlaylistSuite) SetupTest() {
	s.pr = playlistsmocks.NewPlaylistsRepo(s.T())
	s.sr = playlistsmocks.NewSongsRepo(s.T())

	s.s = playlists.New(s.pr, s.sr)

	log := logger.New("test", "prod")

	s.ctx = context.WithValue(context.Background(), logger.LoggerKey, log)
	s.input = validGetPlaylistInput()
}

func (s *GetPlaylistSuite) TestHappyPath() {
	pl := validPostgresRow(s.input)
	pl.Playlist.IsPublic = true

	s.pr.EXPECT().Playlist(mock.Anything, mock.Anything).Return(pl, nil).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]client.Song{}, nil).Maybe()

	output, err := s.s.GetPlaylist(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *GetPlaylistSuite) TestNotFound() {
	s.pr.EXPECT().Playlist(mock.Anything, mock.Anything).Return(postgres.PlaylistRow{}, pgx.ErrNoRows).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]client.Song{}, nil).Maybe()

	output, err := s.s.GetPlaylist(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func (s *GetPlaylistSuite) TestError() {
	s.pr.EXPECT().Playlist(mock.Anything, mock.Anything).Return(postgres.PlaylistRow{}, gofakeit.ErrorDatabase()).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]client.Song{}, nil).Maybe()

	output, err := s.s.GetPlaylist(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func (s *GetPlaylistSuite) TestGetPrivatePlaylistWithoutToken() {
	pl := validPostgresRow(s.input)
	pl.Playlist.IsPublic = false

	if pl.Playlist.AuthorID.String() == s.input {
		s.input = gofakeit.UUID()
	}

	s.pr.EXPECT().Playlist(mock.Anything, mock.Anything).Return(pl, nil).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]client.Song{}, nil).Maybe()

	output, err := s.s.GetPlaylist(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

// Try to get private playlist without permission (author_id != user_id)
func (s *GetPlaylistSuite) TestGetPrivatePlaylistWithToken() {
	out := validPostgresRow(s.input)
	out.Playlist.IsPublic = false

	if out.Playlist.AuthorID.String() == s.input {
		s.input = gofakeit.UUID()
	}

	s.ctx = uniceptors.CtxWithToken(s.ctx, auth.Token{
		Subject:  uuid.MustParse(s.input),
		IsArtist: true,
		Exp:      time.Now().Add(time.Minute).Unix(),
	})

	s.pr.EXPECT().Playlist(mock.Anything, mock.Anything).Return(out, nil).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]client.Song{}, nil).Maybe()

	output, err := s.s.GetPlaylist(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func (s *GetPlaylistSuite) TestSongsServiceError() {
	out := validPostgresRow(s.input)
	out.Playlist.IsPublic = true

	s.pr.EXPECT().Playlist(mock.Anything, mock.Anything).Return(out, nil).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return(nil, gofakeit.ErrorDatabase()).Maybe()

	output, err := s.s.GetPlaylist(s.ctx, s.input)
	s.Error(err)
	s.NotEmpty(output)
}

func TestGetPlaylist(t *testing.T) {
	suite.Run(t, new(GetPlaylistSuite))
}

func validGetPlaylistInput() string {
	return gofakeit.UUID()
}

func validPlaylist(id string) models.Playlist {
	return models.Playlist{
		Metadata: models.PlaylistMetadata{
			ID:             id,
			Title:          gofakeit.BookTitle(),
			AuthorID:       gofakeit.UUID(),
			CoverURL:       gofakeit.URL(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			ReleasedAt:     time.Now(),
			IsAlbum:        gofakeit.Bool(),
			IsMyCollection: false,
			IsPublic:       gofakeit.Bool(),
		},
		Songs: []client.Song{},
	}
}

func validPostgresRow(id string) postgres.PlaylistRow {
	pl := validPlaylist(id)

	return postgres.PlaylistRow{
		Playlist: postgres.Playlist{
			ID:         uuid.MustParse(pl.Metadata.ID),
			Title:      pl.Metadata.Title,
			AuthorID:   uuid.MustParse(pl.Metadata.AuthorID),
			TrackIds:   []uuid.UUID{uuid.New(), uuid.New()},
			CoverUrl:   pgconv.Text(pl.Metadata.CoverURL),
			CreatedAt:  pl.Metadata.CreatedAt,
			UpdatedAt:  pgconv.Timestamptz(pl.Metadata.UpdatedAt),
			ReleasedAt: pgconv.Timestamptz(pl.Metadata.ReleasedAt),
			IsAlbum:    pl.Metadata.IsAlbum,
			IsPublic:   pl.Metadata.IsPublic,
		},
	}
}

type GetPlaylistsSuite struct {
	suite.Suite

	pr *playlistsmocks.PlaylistsRepo

	s     *playlists.ServicePlaylists
	ctx   context.Context
	input playlists.GetPlaylistsInput
}

func (s *GetPlaylistsSuite) SetupTest() {
	s.pr = playlistsmocks.NewPlaylistsRepo(s.T())

	s.s = playlists.New(s.pr, nil)

	log := logger.New("test", "prod")
	s.ctx = context.WithValue(context.Background(), logger.LoggerKey, log)
}

// byArtistID
func (s *GetPlaylistsSuite) TestByArtistIDHappyPath() {
	s.input = validInput(true, false, false)

	s.pr.EXPECT().PublicPlaylists(mock.Anything, mock.Anything).Return(validRows(5), nil).Once()

	output, err := s.s.GetPlaylists(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *GetPlaylistsSuite) TestByArtistIDError() {
	s.input = validInput(true, false, false)

	s.pr.EXPECT().PublicPlaylists(mock.Anything, mock.Anything).Return(nil, gofakeit.ErrorDatabase()).Once()

	output, err := s.s.GetPlaylists(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func (s *GetPlaylistsSuite) TestByArtistIDNoRows() {
	s.input = validInput(true, false, false)

	s.pr.EXPECT().PublicPlaylists(mock.Anything, mock.Anything).Return([]postgres.PublicPlaylistsRow{}, pgx.ErrNoRows).Once()

	output, err := s.s.GetPlaylists(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

// byMatchingTitle
func (s *GetPlaylistsSuite) TestByMatchingTitleHappyPath() {
	s.input = validInput(false, true, false)

	s.pr.EXPECT().PublicPlaylists(mock.Anything, mock.Anything).Return(validRows(5), nil).Once()

	output, err := s.s.GetPlaylists(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *GetPlaylistsSuite) TestByMatchingTitleError() {
	s.input = validInput(false, true, false)

	s.pr.EXPECT().PublicPlaylists(mock.Anything, mock.Anything).Return(nil, gofakeit.ErrorDatabase()).Once()

	output, err := s.s.GetPlaylists(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func (s *GetPlaylistsSuite) TestByMatchingTitleNoRows() {
	s.input = validInput(false, true, false)

	s.pr.EXPECT().PublicPlaylists(mock.Anything, mock.Anything).Return([]postgres.PublicPlaylistsRow{}, pgx.ErrNoRows).Once()

	output, err := s.s.GetPlaylists(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

// byIDs
func (s *GetPlaylistsSuite) TestByIDsHappyPath() {
	s.input = validInput(false, false, true)

	s.pr.EXPECT().PublicPlaylists(mock.Anything, mock.Anything).Return(validRows(5), nil).Once()

	output, err := s.s.GetPlaylists(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *GetPlaylistsSuite) TestByIDsError() {
	s.input = validInput(false, false, true)

	s.pr.EXPECT().PublicPlaylists(mock.Anything, mock.Anything).Return(nil, gofakeit.ErrorDatabase()).Once()

	output, err := s.s.GetPlaylists(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func (s *GetPlaylistsSuite) TestByIDsNoRows() {
	s.input = validInput(false, false, true)

	s.pr.EXPECT().PublicPlaylists(mock.Anything, mock.Anything).Return([]postgres.PublicPlaylistsRow{}, pgx.ErrNoRows).Once()

	output, err := s.s.GetPlaylists(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *GetPlaylistsSuite) TestEmptyFilters() {
	s.input = validInput(false, false, false)

	s.pr.EXPECT().PublicPlaylists(mock.Anything, mock.Anything).Return(validRows(3), nil).Maybe()

	output, err := s.s.GetPlaylists(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func (s *GetPlaylistsSuite) TestMultipleFilters() {
	s.input = validInput(true, true, true)

	s.pr.EXPECT().PublicPlaylists(mock.Anything, mock.Anything).Return(validRows(3), nil).Maybe()

	output, err := s.s.GetPlaylists(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func TestGetPlaylists(t *testing.T) {
	suite.Run(t, new(GetPlaylistsSuite))
}

func validPublicPlaylistsRow() postgres.PublicPlaylistsRow {
	return postgres.PublicPlaylistsRow{
		Playlist: validPostgresPlaylist(uuid.NewString()),
	}
}

func validRows(n int) []postgres.PublicPlaylistsRow {
	rows := make([]postgres.PublicPlaylistsRow, 0, n)

	for i := 0; i < n; i++ {
		rows = append(rows, validPublicPlaylistsRow())
	}

	return rows
}

func validInput(byArtistID, byMatchingTitle, byIDs bool) playlists.GetPlaylistsInput {
	var params playlists.GetPlaylistsInput

	params.Limit = 1000
	params.Page = 1

	if byArtistID {
		params.ArtistID = pointer.To(uuid.NewString())
	}

	if byMatchingTitle {
		params.MatchTitle = pointer.To(gofakeit.BookTitle())
	}

	if byIDs {
		params.PlaylistIDs = []string{uuid.NewString(), uuid.NewString()}
	}

	return params
}
