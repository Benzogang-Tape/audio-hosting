package playlists_test

import (
	"context"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/client/songs"
	playlistsmocks "github.com/Benzogang-Tape/audio-hosting/playlists/internal/mocks/playlists"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service/playlists"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type DeletePlaylistSuit struct {
	suite.Suite

	pr *playlistsmocks.PlaylistsRepo
	sr *playlistsmocks.SongsRepo

	s     *playlists.ServicePlaylists
	ctx   context.Context
	input playlists.DeletePlaylistInput
}

func (s *DeletePlaylistSuit) SetupTest() {
	s.pr = playlistsmocks.NewPlaylistsRepo(s.T())
	s.sr = playlistsmocks.NewSongsRepo(s.T())

	s.s = playlists.New(s.pr, s.sr)

	log := logger.New("test", "prod")

	s.ctx = context.WithValue(context.Background(), logger.LoggerKey, log)
	s.input = playlists.DeletePlaylistInput{
		PlaylistIDs: []string{uuid.NewString(), uuid.NewString()},
		UserID:      uuid.NewString(),
	}
}

func (s *DeletePlaylistSuit) TestHappyPath() {
	s.pr.EXPECT().DeletePlaylists(mock.Anything, mock.Anything).Return(nil).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]songs.Song{}, nil).Maybe()

	err := s.s.DeletePlaylist(s.ctx, s.input)
	s.NoError(err)
}

func (s *DeletePlaylistSuit) TestDbError() {
	s.pr.EXPECT().DeletePlaylists(mock.Anything, mock.Anything).Return(gofakeit.ErrorDatabase()).Once()
	s.sr.EXPECT().GetSongs(mock.Anything, mock.Anything).Return([]songs.Song{}, nil).Maybe()

	err := s.s.DeletePlaylist(s.ctx, s.input)
	s.Error(err)
}

func TestDelete(t *testing.T) {
	suite.Run(t, new(DeletePlaylistSuit))
}
