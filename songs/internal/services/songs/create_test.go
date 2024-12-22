package songs_test

import (
	"context"
	"testing"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/clients/users"
	songsmocks "github.com/Benzogang-Tape/audio-hosting/songs/internal/mocks/songs"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/services/songs"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/repoerrs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CreateSongSuite struct {
	suite.Suite

	sm *songsmocks.SongRepo
	su *songsmocks.UserRepo

	s     *songs.Service
	ctx   context.Context
	input songs.CreateSongInput
}

func (s *CreateSongSuite) SetupTest() {
	s.sm = songsmocks.NewSongRepo(s.T())
	s.su = songsmocks.NewUserRepo(s.T())

	s.s = songs.NewWithConfig(songs.Config{
		Dependencies: songs.Dependencies{
			SongRepo: s.sm,
			UserRepo: s.su,
		},
	})

	s.ctx = context.Background()
	s.input = validCreateSongInput()
}

func (s *CreateSongSuite) TestHappyPath() {
	s.sm.EXPECT().SaveSong(mock.Anything, mock.Anything).Return(nil).Once()
	s.su.EXPECT().ArtistsByIds(mock.Anything, mock.Anything).Return(validArtists(2), nil).Once()

	output, err := s.s.CreateSong(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *CreateSongSuite) TestSongRepoError() {
	s.sm.EXPECT().SaveSong(mock.Anything, mock.Anything).Return(gofakeit.ErrorDatabase()).Once()

	output, err := s.s.CreateSong(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func (s *CreateSongSuite) TestSongRepo_UniqueError() {
	s.sm.EXPECT().SaveSong(mock.Anything, mock.Anything).Return(repoerrs.ErrUnique).Once()

	output, err := s.s.CreateSong(s.ctx, s.input)
	s.ErrorIs(err, songs.ErrSongExists)
	s.Empty(output)
}

// This case may happen when users service is not available
func (s *CreateSongSuite) TestUserRepoError() {
	s.sm.EXPECT().SaveSong(mock.Anything, mock.Anything).Return(nil).Once()
	s.su.EXPECT().ArtistsByIds(mock.Anything, mock.Anything).Return(nil, gofakeit.ErrorDatabase()).Once()

	output, err := s.s.CreateSong(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func TestCreateSong(t *testing.T) {
	suite.Run(t, new(CreateSongSuite))
}

func validCreateSongInput() songs.CreateSongInput {
	return songs.CreateSongInput{
		Name:        gofakeit.Name(),
		SingerId:    uuid.New(),
		FeatArtists: []uuid.UUID{uuid.New(), uuid.New()},
	}
}

func validArtist() users.Artist {
	return users.Artist{
		Id:    uuid.New(),
		Name:  gofakeit.Name(),
		Label: gofakeit.Company(),
	}
}

func validArtists(count int) []users.Artist {
	artists := make([]users.Artist, count)
	for i := 0; i < count; i++ {
		artists[i] = validArtist()
	}

	return artists
}
