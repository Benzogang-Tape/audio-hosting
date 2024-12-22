package raw_test

import (
	"context"
	"strings"
	"testing"

	rawmocks "github.com/Benzogang-Tape/audio-hosting/songs/internal/mocks/raw"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/services/raw"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/repoerrs"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UploadRawSongImageSuite struct {
	suite.Suite

	om *rawmocks.ObjectStorage
	sm *rawmocks.SongRepo

	s     *raw.ServiceRaw
	ctx   context.Context
	input raw.UploadRawSongImageInput
}

func (s *UploadRawSongImageSuite) SetupTest() {
	s.om = rawmocks.NewObjectStorage(s.T())
	s.sm = rawmocks.NewSongRepo(s.T())

	s.s = raw.NewWithConfig(raw.Config{
		Dependencies: raw.Dependencies{
			ObjectStorage: s.om,
			SongRepo:      s.sm,
		},
		HostUsesTls: true,
		Host:        gofakeit.DomainName(),
	})

	s.ctx = context.Background()
	s.input = validUploadRawSongImageInput()
}

func (s *UploadRawSongImageSuite) TestHappyPath() {
	s.sm.EXPECT().MySong(mock.Anything, mock.Anything).Return(validMySongRow(s.input.SongId), nil).Once()
	s.sm.EXPECT().Begin(mock.Anything).Return(s.sm, nil).Once()
	s.sm.EXPECT().PatchSong(mock.Anything, mock.Anything).Return(postgres.Song(validMySongRow(s.input.SongId).Song), nil).Once()
	s.om.EXPECT().PutImageObject(mock.Anything, mock.Anything).Return(nil).Once()
	s.sm.EXPECT().Commit(mock.Anything).Return(nil).Once()
	s.sm.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	_, err := s.s.UploadRawSongImage(s.ctx, s.input)
	s.NoError(err)
}

func (s *UploadRawSongImageSuite) TestExtensionNotJpgPngJpeg() {
	s.input.Extension = "mp4"

	_, err := s.s.UploadRawSongImage(s.ctx, s.input)
	s.ErrorIs(err, raw.ErrInvalidImageExtension)
}

func (s *UploadRawSongImageSuite) TestMySong_EmptyResult() {
	s.sm.EXPECT().MySong(mock.Anything, mock.Anything).Return(postgres.MySongRow{}, repoerrs.ErrEmptyResult).Once()

	_, err := s.s.UploadRawSongImage(s.ctx, s.input)
	s.ErrorIs(err, raw.ErrSongNotExists)
}

func (s *UploadRawSongImageSuite) TestMySongError() {
	s.sm.EXPECT().MySong(mock.Anything, mock.Anything).Return(postgres.MySongRow{}, gofakeit.ErrorDatabase()).Once()

	_, err := s.s.UploadRawSongImage(s.ctx, s.input)
	s.Error(err)
}

func (s *UploadRawSongImageSuite) TestBeginError() {
	s.sm.EXPECT().MySong(mock.Anything, mock.Anything).Return(validMySongRow(s.input.SongId), nil).Once()
	s.sm.EXPECT().Begin(mock.Anything).Return(nil, gofakeit.Error()).Once()

	_, err := s.s.UploadRawSongImage(s.ctx, s.input)
	s.Error(err)
}

func (s *UploadRawSongImageSuite) TestPatchSongError() {
	s.sm.EXPECT().MySong(mock.Anything, mock.Anything).Return(validMySongRow(s.input.SongId), nil).Once()
	s.sm.EXPECT().Begin(mock.Anything).Return(s.sm, nil).Once()
	s.sm.EXPECT().PatchSong(mock.Anything, mock.Anything).Return(postgres.Song{}, gofakeit.ErrorDatabase()).Once()
	s.sm.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	_, err := s.s.UploadRawSongImage(s.ctx, s.input)
	s.Error(err)
}

func (s *UploadRawSongImageSuite) TestPutSongObjectError() {
	s.sm.EXPECT().MySong(mock.Anything, mock.Anything).Return(validMySongRow(s.input.SongId), nil).Once()
	s.sm.EXPECT().Begin(mock.Anything).Return(s.sm, nil).Once()
	s.sm.EXPECT().PatchSong(mock.Anything, mock.Anything).Return(postgres.Song(validMySongRow(s.input.SongId).Song), nil).Once()
	s.om.EXPECT().PutImageObject(mock.Anything, mock.Anything).Return(gofakeit.ErrorDatabase()).Once()
	s.sm.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	_, err := s.s.UploadRawSongImage(s.ctx, s.input)
	s.Error(err)
}

func (s *UploadRawSongImageSuite) TestCommitError() {
	s.sm.EXPECT().MySong(mock.Anything, mock.Anything).Return(validMySongRow(s.input.SongId), nil).Once()
	s.sm.EXPECT().Begin(mock.Anything).Return(s.sm, nil).Once()
	s.sm.EXPECT().PatchSong(mock.Anything, mock.Anything).Return(postgres.Song(validMySongRow(s.input.SongId).Song), nil).Once()
	s.om.EXPECT().PutImageObject(mock.Anything, mock.Anything).Return(nil).Once()
	s.sm.EXPECT().Commit(mock.Anything).Return(gofakeit.Error()).Once().Once()
	s.sm.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	_, err := s.s.UploadRawSongImage(s.ctx, s.input)
	s.Error(err)
}

func TestUploadRawSongImage(t *testing.T) {
	suite.Run(t, new(UploadRawSongImageSuite))
}

func validUploadRawSongImageInput() raw.UploadRawSongImageInput {
	return raw.UploadRawSongImageInput{
		ArtistId:    uuid.New(),
		SongId:      uuid.New(),
		Extension:   "jpg",
		WeightBytes: int32(gofakeit.IntRange(1024, 1024*1024*1024)),
		Content:     strings.NewReader(gofakeit.LoremIpsumSentence(10)),
	}
}

type GetRawSongImageSuite struct {
	suite.Suite

	om  *rawmocks.ObjectStorage
	s   *raw.ServiceRaw
	ctx context.Context
}

func (s *GetRawSongImageSuite) SetupTest() {
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

func (s *GetRawSongImageSuite) TestHappyPath() {
	s.om.EXPECT().GetImageObject(mock.Anything, mock.Anything).Return(strings.NewReader(gofakeit.LoremIpsumSentence(10)), nil).Once()

	output, err := s.s.GetRawSongImage(s.ctx, gofakeit.Fruit()+".jpg")
	s.NoError(err)
	s.NotNil(output.Content)
	s.Equal("jpg", output.Extension)
}

func (s *GetRawSongImageSuite) TestError() {
	s.om.EXPECT().GetImageObject(mock.Anything, mock.Anything).Return(nil, gofakeit.Error()).Once()

	_, err := s.s.GetRawSongImage(s.ctx, gofakeit.Fruit()+".jpg")
	s.Error(err)
}

func TestGetRawSongImage(t *testing.T) {
	suite.Run(t, new(GetRawSongImageSuite))
}
