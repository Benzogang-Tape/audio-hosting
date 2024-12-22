package playlists_test

import (
	"context"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/client/songs"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"
	"github.com/brianvoe/gofakeit/v7"
	"testing"

	playlistsmocks "github.com/Benzogang-Tape/audio-hosting/playlists/internal/mocks/playlists"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service/playlists"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CreatePlaylistSuite struct {
	suite.Suite

	pr *playlistsmocks.PlaylistsRepo
	sr *playlistsmocks.SongsRepo

	s     *playlists.ServicePlaylists
	ctx   context.Context
	input playlists.CreatePlaylistInput
}

func (s *CreatePlaylistSuite) SetupTest() {
	s.pr = playlistsmocks.NewPlaylistsRepo(s.T())
	s.sr = playlistsmocks.NewSongsRepo(s.T())

	s.s = playlists.New(s.pr, s.sr)

	log := logger.New("test", "prod")

	s.ctx = context.WithValue(context.Background(), logger.LoggerKey, log)
	s.input = validCreatePlaylistInput()
}

func (s *CreatePlaylistSuite) TestHappyPath() {
	s.pr.EXPECT().SavePlaylist(mock.Anything, mock.Anything).Return(nil).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]songs.Song{}, nil).Maybe()

	output, err := s.s.CreatePlaylist(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *CreatePlaylistSuite) TestDbError() {
	s.pr.EXPECT().SavePlaylist(mock.Anything, mock.Anything).Return(gofakeit.ErrorDatabase()).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]songs.Song{}, nil).Maybe()

	output, err := s.s.CreatePlaylist(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func (s *CreatePlaylistSuite) TestSongsServiceError() {
	s.pr.EXPECT().SavePlaylist(mock.Anything, mock.Anything).Return(nil).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return(nil, gofakeit.ErrorDatabase()).Maybe()

	output, err := s.s.CreatePlaylist(s.ctx, s.input)
	s.Error(err)
	s.NotEmpty(output)
}

func TestCreatePlaylist(t *testing.T) {
	suite.Run(t, new(CreatePlaylistSuite))
}

func validCreatePlaylistInput() playlists.CreatePlaylistInput {
	return playlists.CreatePlaylistInput{
		Title:    gofakeit.BookTitle(),
		TrackIDs: []uuid.UUID{uuid.New()},
		AuthorID: uuid.New(),
		CoverURL: gofakeit.URL(),
		IsAlbum:  false,
	}
}
