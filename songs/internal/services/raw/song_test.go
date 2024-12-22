package raw_test

import (
	"context"
	"strings"
	"testing"
	"time"

	rawmocks "github.com/Benzogang-Tape/audio-hosting/songs/internal/mocks/raw"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/services/raw"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/pgconv"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/repoerrs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UploadRawSongSuite struct {
	suite.Suite

	om *rawmocks.ObjectStorage
	sm *rawmocks.SongRepo
	dm *rawmocks.SoundDecoder

	s     *raw.ServiceRaw
	ctx   context.Context
	input raw.UploadRawSongInput
}

func (s *UploadRawSongSuite) SetupTest() {
	s.om = rawmocks.NewObjectStorage(s.T())
	s.sm = rawmocks.NewSongRepo(s.T())
	s.dm = rawmocks.NewSoundDecoder(s.T())

	s.s = raw.NewWithConfig(raw.Config{
		Dependencies: raw.Dependencies{
			ObjectStorage: s.om,
			SongRepo:      s.sm,
			SoundDecoder:  s.dm,
		},
		HostUsesTls: true,
		Host:        gofakeit.DomainName(),
	})

	s.ctx = context.Background()
	s.input = validUploadRawSongInput()
}

func (s *UploadRawSongSuite) TestHappyPath() {
	s.sm.EXPECT().MySong(mock.Anything, mock.Anything).Return(validMySongRow(s.input.SongId), nil).Once()
	s.dm.EXPECT().GetMp3Duration(mock.Anything, mock.Anything).Return(time.Minute, nil).Once()
	s.sm.EXPECT().Begin(mock.Anything).Return(s.sm, nil).Once()
	s.sm.EXPECT().PatchSong(mock.Anything, mock.Anything).Return(postgres.Song(validMySongRow(s.input.SongId).Song), nil).Once()
	s.om.EXPECT().PutSongObject(mock.Anything, mock.Anything).Return(nil).Once()
	s.sm.EXPECT().Commit(mock.Anything).Return(nil).Once()
	s.sm.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	_, err := s.s.UploadRawSong(s.ctx, s.input)
	s.NoError(err)
}

func (s *UploadRawSongSuite) TestExtensionNotMp3() {
	s.input.Extension = "aac"

	_, err := s.s.UploadRawSong(s.ctx, s.input)
	s.ErrorIs(err, raw.ErrInvalidExtension)
}

func (s *UploadRawSongSuite) TestMySong_EmptyResult() {
	s.sm.EXPECT().MySong(mock.Anything, mock.Anything).Return(postgres.MySongRow{}, repoerrs.ErrEmptyResult).Once()

	_, err := s.s.UploadRawSong(s.ctx, s.input)
	s.ErrorIs(err, raw.ErrSongNotExists)
}

func (s *UploadRawSongSuite) TestMySongError() {
	s.sm.EXPECT().MySong(mock.Anything, mock.Anything).Return(postgres.MySongRow{}, gofakeit.ErrorDatabase()).Once()

	_, err := s.s.UploadRawSong(s.ctx, s.input)
	s.Error(err)
}

func (s *UploadRawSongSuite) TestReleasedSong() {
	// It is able to upload raw song even if it was before.
	// But as soon as the song get released, you are not able anymore.
	row := validMySongRow(s.input.SongId)
	row.Song.ReleasedAt = pgconv.Timestamptz(gofakeit.Date())

	s.sm.EXPECT().MySong(mock.Anything, mock.Anything).Return(row, nil).Once()

	_, err := s.s.UploadRawSong(s.ctx, s.input)
	s.ErrorIs(err, raw.ErrSongAlreadyReleased)
}

func (s *UploadRawSongSuite) TestMp3DurationError() {
	s.sm.EXPECT().MySong(mock.Anything, mock.Anything).Return(validMySongRow(s.input.SongId), nil).Once()
	s.dm.EXPECT().GetMp3Duration(mock.Anything, mock.Anything).Return(0, gofakeit.Error()).Once()

	_, err := s.s.UploadRawSong(s.ctx, s.input)
	s.Error(err)
}

func (s *UploadRawSongSuite) TestBeginError() {
	s.sm.EXPECT().MySong(mock.Anything, mock.Anything).Return(validMySongRow(s.input.SongId), nil).Once()
	s.dm.EXPECT().GetMp3Duration(mock.Anything, mock.Anything).Return(time.Minute, nil).Once()
	s.sm.EXPECT().Begin(mock.Anything).Return(nil, gofakeit.Error()).Once()

	_, err := s.s.UploadRawSong(s.ctx, s.input)
	s.Error(err)
}

func (s *UploadRawSongSuite) TestPatchSongError() {
	s.sm.EXPECT().MySong(mock.Anything, mock.Anything).Return(validMySongRow(s.input.SongId), nil).Once()
	s.dm.EXPECT().GetMp3Duration(mock.Anything, mock.Anything).Return(time.Minute, nil).Once()
	s.sm.EXPECT().Begin(mock.Anything).Return(s.sm, nil).Once()
	s.sm.EXPECT().PatchSong(mock.Anything, mock.Anything).Return(postgres.Song{}, gofakeit.ErrorDatabase()).Once()
	s.sm.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	_, err := s.s.UploadRawSong(s.ctx, s.input)
	s.Error(err)
}

func (s *UploadRawSongSuite) TestPutSongObjectError() {
	row := validMySongRow(s.input.SongId)
	row.Song.ReleasedAt = pgconv.NullTimestamptz()
	s.sm.EXPECT().MySong(mock.Anything, mock.Anything).Return(row, nil).Once()
	s.dm.EXPECT().GetMp3Duration(mock.Anything, mock.Anything).Return(time.Minute, nil).Once()
	s.sm.EXPECT().Begin(mock.Anything).Return(s.sm, nil).Once()
	s.sm.EXPECT().PatchSong(mock.Anything, mock.Anything).Return(postgres.Song(validMySongRow(s.input.SongId).Song), nil).Once()
	s.om.EXPECT().PutSongObject(mock.Anything, mock.Anything).Return(gofakeit.ErrorDatabase()).Once()
	s.sm.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	_, err := s.s.UploadRawSong(s.ctx, s.input)
	s.Error(err)
}

func (s *UploadRawSongSuite) TestCommitError() {
	s.sm.EXPECT().MySong(mock.Anything, mock.Anything).Return(validMySongRow(s.input.SongId), nil).Once()
	s.dm.EXPECT().GetMp3Duration(mock.Anything, mock.Anything).Return(time.Minute, nil).Once()
	s.sm.EXPECT().Begin(mock.Anything).Return(s.sm, nil).Once()
	s.sm.EXPECT().PatchSong(mock.Anything, mock.Anything).Return(postgres.Song(validMySongRow(s.input.SongId).Song), nil).Once()
	s.om.EXPECT().PutSongObject(mock.Anything, mock.Anything).Return(nil).Once()
	s.sm.EXPECT().Commit(mock.Anything).Return(gofakeit.Error()).Once()
	s.sm.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	_, err := s.s.UploadRawSong(s.ctx, s.input)
	s.Error(err)
}

func TestUploadRawSong(t *testing.T) {
	suite.Run(t, new(UploadRawSongSuite))
}

func validUploadRawSongInput() raw.UploadRawSongInput {
	return raw.UploadRawSongInput{
		ArtistId:    uuid.New(),
		SongId:      uuid.New(),
		Extension:   "mp3",
		WeightBytes: int32(gofakeit.IntRange(1024, 1024*1024*1024)),
		Content:     strings.NewReader(gofakeit.LoremIpsumSentence(10)),
	}
}

func validMySongRow(id uuid.UUID) postgres.MySongRow {
	return postgres.MySongRow{
		Song: postgres.Song{
			SongID:       id,
			SingerFk:     uuid.New(),
			Name:         gofakeit.Sentence(3),
			S3ObjectName: pgconv.Text(gofakeit.HexColor()),
			ImageUrl:     pgconv.Text(gofakeit.URL()),
			Duration:     pgconv.Interval(time.Minute),
			WeightBytes:  pgconv.Int4(1024),
			UploadedAt:   gofakeit.Date(),
			ReleasedAt:   pgconv.NullTimestamptz(),
		},
	}
}

type GetRawSongSuite struct {
	suite.Suite

	om  *rawmocks.ObjectStorage
	s   *raw.ServiceRaw
	ctx context.Context
}

func (s *GetRawSongSuite) SetupTest() {
	s.om = rawmocks.NewObjectStorage(s.T())
	s.s = raw.NewWithConfig(raw.Config{
		Dependencies: raw.Dependencies{
			ObjectStorage: s.om,
		},
		HostUsesTls: true,
		Host:        gofakeit.DomainName(),
	})
	s.ctx = context.Background()
}

func (s *GetRawSongSuite) TestHappyPath() {
	s.om.EXPECT().GetSongObject(mock.Anything, mock.Anything).Return(strings.NewReader(gofakeit.LoremIpsumSentence(10)), nil).Once()

	reader, err := s.s.GetRawSong(s.ctx, gofakeit.UUID())
	s.NoError(err)
	s.NotNil(reader)
}

func (s *GetRawSongSuite) TestError() {
	s.om.EXPECT().GetSongObject(mock.Anything, mock.Anything).Return(nil, gofakeit.Error()).Once()

	_, err := s.s.GetRawSong(s.ctx, uuid.NewString())
	s.Error(err)
}

func (s *GetRawSongSuite) TestNilReader() {
	s.om.EXPECT().GetSongObject(mock.Anything, mock.Anything).Return(nil, nil).Once()

	reader, err := s.s.GetRawSong(s.ctx, uuid.NewString())
	s.NoError(err)
	s.Nil(reader)
}

func TestGetRawSong(t *testing.T) {
	suite.Run(t, new(GetRawSongSuite))
}
