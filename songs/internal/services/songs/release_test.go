package songs_test

import (
	"context"
	"testing"

	songsmocks "github.com/Benzogang-Tape/audio-hosting/songs/internal/mocks/songs"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/services/songs"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/pgconv"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/repoerrs"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ReleaseSongsSuite struct {
	suite.Suite

	sm *songsmocks.SongRepo
	bm *songsmocks.Broker

	s     *songs.Service
	ctx   context.Context
	input songs.ReleaseSongsInput
}

func (s *ReleaseSongsSuite) SetupTest() {
	s.sm = songsmocks.NewSongRepo(s.T())
	s.bm = songsmocks.NewBroker(s.T())

	s.s = songs.NewWithConfig(songs.Config{
		Dependencies: songs.Dependencies{
			SongRepo: s.sm,
			Broker:   s.bm,
		},
	})

	s.ctx = context.Background()
	s.input = validReleaseSongsInput()
}

func (s *ReleaseSongsSuite) TestHappyPath() {
	s.sm.EXPECT().PatchSongs(mock.Anything, mock.Anything).Return(nil).Once()
	s.sm.EXPECT().MySongs(mock.Anything, mock.Anything).Return(validMySongsRows(2), nil).Once()
	s.bm.EXPECT().SendReleasedMessages(mock.Anything, mock.Anything).Return(nil).Once()

	_, err := s.s.ReleaseSongs(s.ctx, s.input)
	s.NoError(err)
}

func (s *ReleaseSongsSuite) TestHappyPathWithoutNotify() {
	s.input.Notify = false

	s.sm.EXPECT().PatchSongs(mock.Anything, mock.Anything).Return(nil).Once()
	s.sm.EXPECT().MySongs(mock.Anything, mock.Anything).Return(validMySongsRows(2), nil).Once()

	_, err := s.s.ReleaseSongs(s.ctx, s.input)
	s.NoError(err)
}

func (s *ReleaseSongsSuite) TestMySongsError() {
	s.sm.EXPECT().MySongs(mock.Anything, mock.Anything).Return(nil, gofakeit.ErrorDatabase()).Once()

	_, err := s.s.ReleaseSongs(s.ctx, s.input)
	s.Error(err)
}

func (s *ReleaseSongsSuite) TestMySongs_EmptyResultError() {
	s.sm.EXPECT().MySongs(mock.Anything, mock.Anything).Return(nil, repoerrs.ErrEmptyResult).Once()

	_, err := s.s.ReleaseSongs(s.ctx, s.input)
	s.ErrorIs(err, songs.ErrSongNotFound)
}

func (s *ReleaseSongsSuite) TestAlreadyReleased() {
	rows := validMySongsRows(2)
	rows[0].Song.ReleasedAt = pgconv.Timestamptz(gofakeit.Date())

	s.sm.EXPECT().MySongs(mock.Anything, mock.Anything).Return(rows, nil).Once()

	_, err := s.s.ReleaseSongs(s.ctx, s.input)
	s.Error(err)
}

func (s *ReleaseSongsSuite) TestNotLoaded() {
	rows := validMySongsRows(2)
	rows[0].Song.S3ObjectName = pgconv.NullText()

	s.sm.EXPECT().MySongs(mock.Anything, mock.Anything).Return(rows, nil).Once()

	_, err := s.s.ReleaseSongs(s.ctx, s.input)
	s.Error(err)
}

func (s *ReleaseSongsSuite) TestPreconditionFailedError() {
	rows := validMySongsRows(3)
	rows[0].Song.ReleasedAt = pgconv.Timestamptz(gofakeit.Date())
	rows[0].Song.S3ObjectName = pgconv.NullText()
	rows[1].Song.ReleasedAt = pgconv.Timestamptz(gofakeit.Date())

	s.sm.EXPECT().MySongs(mock.Anything, mock.Anything).Return(rows, nil).Once()

	_, err := s.s.ReleaseSongs(s.ctx, s.input)
	if s.Error(err) {
		s.Equal(err.Error(),
			rows[0].Song.Name+" ("+rows[0].Song.SongID.String()+") already released | "+
				rows[0].Song.Name+" ("+rows[0].Song.SongID.String()+") has not been loaded | "+
				rows[1].Song.Name+" ("+rows[1].Song.SongID.String()+") already released | ",
		)
	}
}

func (s *ReleaseSongsSuite) TestPatchSongsError() {
	s.sm.EXPECT().MySongs(mock.Anything, mock.Anything).Return(validMySongsRows(3), nil).Once()
	s.sm.EXPECT().PatchSongs(mock.Anything, mock.Anything).Return(gofakeit.ErrorDatabase()).Once()

	_, err := s.s.ReleaseSongs(s.ctx, s.input)
	s.Error(err)
}

func (s *ReleaseSongsSuite) TestBrokerError() {
	s.sm.EXPECT().MySongs(mock.Anything, mock.Anything).Return(validMySongsRows(2), nil).Once()
	s.sm.EXPECT().PatchSongs(mock.Anything, mock.Anything).Return(nil).Once()
	s.bm.EXPECT().SendReleasedMessages(mock.Anything, mock.Anything).Return(gofakeit.Error()).Once()

	_, err := s.s.ReleaseSongs(s.ctx, s.input)
	s.Error(err)
}

func validReleaseSongsInput() songs.ReleaseSongsInput {
	return songs.ReleaseSongsInput{
		UserId:   uuid.New(),
		SongsIds: []uuid.UUID{uuid.New(), uuid.New()},
		Notify:   true,
	}
}

func validMySongsRows(count int) []postgres.MySongsRow {
	rows := make([]postgres.MySongsRow, count)
	for i := 0; i < count; i++ {
		rows[i] = validSongsRow()
		rows[i].Song.ReleasedAt = pgconv.NullTimestamptz()
	}

	return rows
}

func TestReleaseSongs(t *testing.T) {
	suite.Run(t, new(ReleaseSongsSuite))
}
