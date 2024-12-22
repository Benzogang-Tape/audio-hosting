package covers_test

import (
	"context"
	client "github.com/Benzogang-Tape/audio-hosting/playlists/internal/client/songs"
	coversmocks "github.com/Benzogang-Tape/audio-hosting/playlists/internal/mocks/covers"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/models"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service/covers"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/pgconv"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type GetRawCoverSuite struct {
	suite.Suite

	or *coversmocks.ObjectRepository
	r  *coversmocks.Repository

	s     *covers.ServiceCovers
	ctx   context.Context
	input string
}

func (s *GetRawCoverSuite) SetupTest() {
	s.or = coversmocks.NewObjectRepository(s.T())
	s.r = coversmocks.NewRepository(s.T())

	s.s = covers.New(s.r, s.or, covers.Config{
		HostUsesTLS: true,
		Host:        "localhost:8080",
	})

	log := logger.New("test", "prod")

	s.ctx = context.WithValue(context.Background(), logger.LoggerKey, log)
	s.input = gofakeit.URL() + ".jpg"
}

func (s *GetRawCoverSuite) TestHappyPath() {
	s.or.EXPECT().GetCoverObject(mock.Anything, mock.Anything).Return(FakeReader{}, nil).Once()

	output, err := s.s.GetRawCover(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *GetRawCoverSuite) TestError() {
	s.or.EXPECT().GetCoverObject(mock.Anything, mock.Anything).Return(nil, gofakeit.ErrorDatabase()).Once()

	output, err := s.s.GetRawCover(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func TestGetRawCover(t *testing.T) {
	suite.Run(t, new(GetRawCoverSuite))
}

type FakeReader struct{}

func (f FakeReader) Read(p []byte) (n int, err error) {
	return 0, nil
}

type UploadRawCoverSuite struct {
	suite.Suite

	or *coversmocks.ObjectRepository
	r  *coversmocks.Repository

	s     *covers.ServiceCovers
	ctx   context.Context
	input covers.UploadRawCoverInput
}

func (s *UploadRawCoverSuite) SetupTest() {
	s.or = coversmocks.NewObjectRepository(s.T())
	s.r = coversmocks.NewRepository(s.T())

	s.s = covers.New(s.r, s.or, covers.Config{
		HostUsesTLS: true,
		Host:        "localhost:8080",
	})

	log := logger.New("test", "prod")

	s.ctx = context.WithValue(context.Background(), logger.LoggerKey, log)
	s.input = validUploadInput()
}

func validUploadInput() covers.UploadRawCoverInput {
	return covers.UploadRawCoverInput{
		UserId:      uuid.New(),
		PlaylistId:  uuid.New(),
		Extension:   ".jpg",
		WeightBytes: 1000,
		Content:     FakeReader{},
	}
}

func (s *UploadRawCoverSuite) TestHappyPath() {
	s.r.EXPECT().Playlist(mock.Anything, mock.Anything).Return(validPostgresRow(s.input.PlaylistId.String()), nil).Once()
	s.r.EXPECT().BeginCovers(mock.Anything).Return(s.r, nil).Once()
	s.r.EXPECT().Rollback(mock.Anything).Return(nil).Once()
	s.r.EXPECT().PatchPlaylist(mock.Anything, mock.Anything).Return(validPostgresPlaylist(s.input.PlaylistId.String()), nil).Once()
	s.or.EXPECT().PutCoverObject(mock.Anything, mock.Anything).Return(nil).Once()
	s.r.EXPECT().Commit(mock.Anything).Return(nil).Once()

	output, err := s.s.UploadRawCover(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *UploadRawCoverSuite) TestWrongExtension() {
	s.input = validUploadInput()
	s.input.Extension = ".flac"

	output, err := s.s.UploadRawCover(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func (s *UploadRawCoverSuite) TestGetPlaylistError() {
	s.r.EXPECT().Playlist(mock.Anything, mock.Anything).Return(postgres.PlaylistRow{}, gofakeit.ErrorDatabase()).Once()
	s.r.EXPECT().BeginCovers(mock.Anything).Return(s.r, nil).Maybe()
	s.r.EXPECT().Rollback(mock.Anything).Return(nil).Maybe()

	output, err := s.s.UploadRawCover(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func (s *UploadRawCoverSuite) TestPlaylistNotFound() {
	s.r.EXPECT().Playlist(mock.Anything, mock.Anything).Return(postgres.PlaylistRow{}, pgx.ErrNoRows).Once()
	s.r.EXPECT().BeginCovers(mock.Anything).Return(s.r, nil).Maybe()
	s.r.EXPECT().Rollback(mock.Anything).Return(nil).Maybe()

	output, err := s.s.UploadRawCover(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func (s *UploadRawCoverSuite) TestBeginCoversError() {
	s.r.EXPECT().Playlist(mock.Anything, mock.Anything).Return(validPostgresRow(s.input.PlaylistId.String()), nil).Once()
	s.r.EXPECT().BeginCovers(mock.Anything).Return(nil, gofakeit.ErrorDatabase()).Once()
	s.r.EXPECT().Rollback(mock.Anything).Return(nil).Maybe()

	output, err := s.s.UploadRawCover(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func (s *UploadRawCoverSuite) TestPatchPlaylistError() {
	s.r.EXPECT().Playlist(mock.Anything, mock.Anything).Return(validPostgresRow(s.input.PlaylistId.String()), nil).Once()
	s.r.EXPECT().BeginCovers(mock.Anything).Return(s.r, nil).Once()
	s.r.EXPECT().Rollback(mock.Anything).Return(nil).Once()
	s.r.EXPECT().PatchPlaylist(mock.Anything, mock.Anything).Return(postgres.Playlist{}, gofakeit.ErrorDatabase()).Once()

	output, err := s.s.UploadRawCover(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func (s *UploadRawCoverSuite) TestPutCoverObjectError() {
	s.r.EXPECT().Playlist(mock.Anything, mock.Anything).Return(validPostgresRow(s.input.PlaylistId.String()), nil).Once()
	s.r.EXPECT().BeginCovers(mock.Anything).Return(s.r, nil).Once()
	s.r.EXPECT().Rollback(mock.Anything).Return(nil).Once()
	s.r.EXPECT().PatchPlaylist(mock.Anything, mock.Anything).Return(validPostgresPlaylist(s.input.PlaylistId.String()), nil).Once()
	s.or.EXPECT().PutCoverObject(mock.Anything, mock.Anything).Return(gofakeit.ErrorDatabase()).Once()

	output, err := s.s.UploadRawCover(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func (s *UploadRawCoverSuite) TestCommitError() {
	s.r.EXPECT().Playlist(mock.Anything, mock.Anything).Return(validPostgresRow(s.input.PlaylistId.String()), nil).Once()
	s.r.EXPECT().BeginCovers(mock.Anything).Return(s.r, nil).Once()
	s.r.EXPECT().Rollback(mock.Anything).Return(nil).Once()
	s.r.EXPECT().PatchPlaylist(mock.Anything, mock.Anything).Return(validPostgresPlaylist(s.input.PlaylistId.String()), nil).Once()
	s.or.EXPECT().PutCoverObject(mock.Anything, mock.Anything).Return(nil).Once()
	s.r.EXPECT().Commit(mock.Anything).Return(gofakeit.ErrorDatabase()).Once()

	output, err := s.s.UploadRawCover(s.ctx, s.input)
	s.Error(err)
	s.Empty(output)
}

func TestUploadRawCover(t *testing.T) {
	suite.Run(t, new(UploadRawCoverSuite))
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
