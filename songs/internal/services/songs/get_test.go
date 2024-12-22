package songs_test

import (
	"context"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/clients/users"
	songsmocks "github.com/Benzogang-Tape/audio-hosting/songs/internal/mocks/songs"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/services/songs"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/pgconv"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/repoerrs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type GetSongSuite struct {
	suite.Suite

	sm *songsmocks.SongRepo
	um *songsmocks.UserRepo

	s     *songs.Service
	ctx   context.Context
	input songs.GetSongInput
}

func (s *GetSongSuite) SetupTest() {
	s.sm = songsmocks.NewSongRepo(s.T())
	s.um = songsmocks.NewUserRepo(s.T())

	s.s = songs.NewWithConfig(songs.Config{
		Dependencies: songs.Dependencies{
			SongRepo:   s.sm,
			UserRepo:   s.um,
			RawService: newFakeRawService(),
		},
	})

	s.ctx = context.Background()
	s.input = validGetSongInput()
}

func (s *GetSongSuite) TestHappyPath() {
	s.sm.EXPECT().Song(mock.Anything, s.input.Id).Return(validSongRow(), nil).Once()
	s.um.EXPECT().ArtistsByIds(mock.Anything, mock.Anything).Return(validArtists(2), nil).Once()

	_, err := s.s.GetSong(s.ctx, s.input)
	s.NoError(err)
}

func (s *GetSongSuite) TestSongRepoError() {
	s.sm.EXPECT().Song(mock.Anything, s.input.Id).Return(postgres.SongRow{}, gofakeit.ErrorDatabase()).Once()

	_, err := s.s.GetSong(s.ctx, s.input)
	s.Error(err)
}

func (s *GetSongSuite) TestSongRepo_NotFoundError() {
	s.sm.EXPECT().Song(mock.Anything, s.input.Id).Return(postgres.SongRow{}, repoerrs.ErrEmptyResult).Once()

	_, err := s.s.GetSong(s.ctx, s.input)
	s.ErrorIs(err, songs.ErrSongNotFound)
}

func (s *GetSongSuite) TestUserRepoError() {
	s.sm.EXPECT().Song(mock.Anything, s.input.Id).Return(validSongRow(), nil).Once()
	s.um.EXPECT().ArtistsByIds(mock.Anything, mock.Anything).Return(nil, gofakeit.ErrorDatabase()).Once()

	_, err := s.s.GetSong(s.ctx, s.input)
	s.Error(err)
}

func validGetSongInput() songs.GetSongInput {
	return songs.GetSongInput{
		Id: uuid.New(),
	}
}

func validSongRow() postgres.SongRow {
	return postgres.SongRow{
		Song: postgres.Song{
			SongID:       uuid.New(),
			SingerFk:     uuid.New(),
			Name:         gofakeit.Sentence(3),
			S3ObjectName: pgconv.Text(gofakeit.HexColor()),
			ImageUrl:     pgconv.Text(gofakeit.URL()),
			Duration:     pgconv.Interval(time.Minute),
			WeightBytes:  pgconv.Int4(1024),
			UploadedAt:   gofakeit.Date(),
			ReleasedAt:   pgconv.Timestamptz(gofakeit.Date()),
		},
		ArtistsIds: []uuid.UUID{
			uuid.New(),
			uuid.New(),
		},
	}
}

func TestGetSong(t *testing.T) {
	suite.Run(t, new(GetSongSuite))
}

type GetSongsSuite struct {
	suite.Suite

	sm *songsmocks.SongRepo
	um *songsmocks.UserRepo

	s     *songs.Service
	ctx   context.Context
	input songs.GetSongsInput
}

func (s *GetSongsSuite) SetupTest() {
	s.sm = songsmocks.NewSongRepo(s.T())
	s.um = songsmocks.NewUserRepo(s.T())

	s.s = songs.NewWithConfig(songs.Config{
		Dependencies: songs.Dependencies{
			SongRepo:   s.sm,
			UserRepo:   s.um,
			RawService: newFakeRawService(),
		},
	})

	s.ctx = context.Background()
	s.input = validGetSongsInput()
}

func (s *GetSongsSuite) TestHappyPath() {
	s.sm.EXPECT().ReleasedSongs(mock.Anything, mock.Anything).Return(validReleasedSongsRows(3), nil).Once()
	s.um.EXPECT().ArtistsByIds(mock.Anything, mock.Anything).Return(validArtists(2), nil).Times(3)

	output, err := s.s.GetSongs(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *GetSongsSuite) TestMultipleFilters() {
	s.input.MatchName = ptr("name")

	_, err := s.s.GetSongs(s.ctx, s.input)
	s.ErrorIs(err, songs.ErrMultipleFilters)
}

func (s *GetSongsSuite) TestNoFilters() {
	s.input.Ids = nil

	_, err := s.s.GetSongs(s.ctx, s.input)
	s.ErrorIs(err, songs.ErrNoFilters)
}

func (s *GetSongsSuite) TestFilter_MatchName() {
	s.input.Ids = nil
	s.input.MatchName = ptr("name")

	s.sm.EXPECT().ReleasedSongs(mock.Anything, mock.MatchedBy(func(p postgres.ReleasedSongsParams) bool {
		return p.ByName && s.Equal(*s.input.MatchName, p.MatchName) && !p.ByIds && !p.BySinger && !p.WithArtist
	})).Return(validReleasedSongsRows(3), nil).Once()
	s.sm.EXPECT().CountSongsMatchName(mock.Anything, mock.Anything).Return(int32(100), nil).Once()
	s.um.EXPECT().ArtistsByIds(mock.Anything, mock.Anything).Return(validArtists(2), nil).Times(3)

	output, err := s.s.GetSongs(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *GetSongsSuite) TestFilter_ByIds() {
	// Set by default in validGetSongsInput

	s.sm.EXPECT().ReleasedSongs(mock.Anything, mock.MatchedBy(func(p postgres.ReleasedSongsParams) bool {
		return p.ByIds && s.Equal(s.input.Ids, p.Ids) && !p.ByName && !p.BySinger && !p.WithArtist
	})).Return(validReleasedSongsRows(3), nil).Once()
	s.um.EXPECT().ArtistsByIds(mock.Anything, mock.Anything).Return(validArtists(2), nil).Times(3)

	output, err := s.s.GetSongs(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *GetSongsSuite) TestFilter_BySinger() {
	s.input.Ids = nil
	s.input.ArtistId = &uuid.Max

	s.sm.EXPECT().ReleasedSongs(mock.Anything, mock.MatchedBy(func(p postgres.ReleasedSongsParams) bool {
		return p.WithArtist && s.Contains(p.SingersIds, *s.input.ArtistId) && !p.ByIds && !p.ByName && !p.BySinger
	})).Return(validReleasedSongsRows(3), nil).Once()
	s.sm.EXPECT().CountSongsWithArtistsIds(mock.Anything, mock.Anything).Return(int32(100), nil).Once()
	s.um.EXPECT().ArtistsByIds(mock.Anything, mock.Anything).Return(validArtists(2), nil).Times(3)

	output, err := s.s.GetSongs(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *GetSongsSuite) TestFilter_MatchArtist() {
	s.input.Ids = nil
	s.input.MatchArtist = ptr("name")
	matchedArtists := validArtists(3)

	matchedArtistsIds := make([]uuid.UUID, 0, len(matchedArtists))
	for _, a := range matchedArtists {
		matchedArtistsIds = append(matchedArtistsIds, a.Id)
	}

	s.um.EXPECT().ArtistsMatchingName(mock.Anything, *s.input.MatchArtist).Return(matchedArtists, nil).Once()
	s.sm.EXPECT().ReleasedSongs(mock.Anything, mock.MatchedBy(func(p postgres.ReleasedSongsParams) bool {
		return p.WithArtist && s.Equal(p.SingersIds, matchedArtistsIds) && !p.ByIds && !p.ByName && !p.BySinger
	})).Return(validReleasedSongsRows(3), nil).Once()
	s.sm.EXPECT().CountSongsWithArtistsIds(mock.Anything, mock.Anything).Return(int32(100), nil).Once()
	s.um.EXPECT().ArtistsByIds(mock.Anything, mock.Anything).Return(validArtists(2), nil).Times(3)

	output, err := s.s.GetSongs(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *GetSongsSuite) TestFilter_MatchArtistError() {
	s.input.Ids = nil
	s.input.MatchArtist = ptr("name")

	s.um.EXPECT().ArtistsMatchingName(mock.Anything, *s.input.MatchArtist).Return(nil, gofakeit.ErrorDatabase()).Once()

	_, err := s.s.GetSongs(s.ctx, s.input)
	s.Error(err)
}

func (s *GetSongsSuite) TestSongRepoError() {
	s.sm.EXPECT().ReleasedSongs(mock.Anything, mock.Anything).Return(nil, gofakeit.ErrorDatabase()).Once()

	_, err := s.s.GetSongs(s.ctx, s.input)
	s.Error(err)
}

func (s *GetSongsSuite) TestSongRepo_EmptyResultError() {
	s.sm.EXPECT().ReleasedSongs(mock.Anything, mock.Anything).Return(nil, repoerrs.ErrEmptyResult).Once()

	output, err := s.s.GetSongs(s.ctx, s.input)
	s.NoError(err)
	s.Empty(output.Songs)
}

func (s *GetSongsSuite) TestUserRepoError() {
	var calls, failed = 4000, 0

	rows := validReleasedSongsRows(calls)

	s.sm.EXPECT().ReleasedSongs(mock.Anything, mock.Anything).Return(rows, nil).Once()
	s.um.EXPECT().ArtistsByIds(mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, ids []uuid.UUID) ([]users.Artist, error) {
			if rand.Uint()%16 == 0 {
				failed++
				return nil, gofakeit.Error()
			}

			return validArtists(len(ids)), nil
		})

	output, err := s.s.GetSongs(s.ctx, s.input)
	s.Len(output.Songs, calls-failed)
	s.NoError(err)
}

func validGetSongsInput() songs.GetSongsInput {
	return songs.GetSongsInput{
		Ids:      []uuid.UUID{uuid.New(), uuid.New(), uuid.New()},
		Page:     1,
		PageSize: 10,
	}
}

func validReleasedSongsRows(count int) []postgres.ReleasedSongsRow {
	songs := make([]postgres.ReleasedSongsRow, count)
	for i := 0; i < count; i++ {
		songs[i] = postgres.ReleasedSongsRow(validSongRow())
	}

	return songs
}

func TestGetSongs(t *testing.T) {
	suite.Run(t, new(GetSongsSuite))
}

type GetMySongsSuite struct {
	suite.Suite

	sm *songsmocks.SongRepo
	um *songsmocks.UserRepo

	s     *songs.Service
	ctx   context.Context
	input songs.GetMySongsInput
}

func (s *GetMySongsSuite) SetupTest() {
	s.sm = songsmocks.NewSongRepo(s.T())
	s.um = songsmocks.NewUserRepo(s.T())

	s.s = songs.NewWithConfig(songs.Config{
		Dependencies: songs.Dependencies{
			SongRepo:   s.sm,
			UserRepo:   s.um,
			RawService: newFakeRawService(),
		},
	})

	s.ctx = context.Background()
	s.input = validGetMySongsInput()
}

func (s *GetMySongsSuite) TestHappyPath() {
	s.sm.EXPECT().MySongs(mock.Anything, mock.Anything).Return(validSongRows(3), nil).Once()
	s.sm.EXPECT().CountMySongs(mock.Anything, s.input.UserId).Return(int32(100), nil).Once()
	s.um.EXPECT().ArtistsByIds(mock.Anything, mock.Anything).Return(validArtists(2), nil).Times(3)

	output, err := s.s.GetMySongs(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *GetMySongsSuite) TestHappyPath_ByIds() {
	s.input.ByIds = true
	s.input.Ids = []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	s.sm.EXPECT().MySongs(mock.Anything, mock.Anything).Return(validSongRows(3), nil).Once()
	s.um.EXPECT().ArtistsByIds(mock.Anything, mock.Anything).Return(validArtists(2), nil).Times(3)

	output, err := s.s.GetMySongs(s.ctx, s.input)
	s.NoError(err)
	s.NotEmpty(output)
}

func (s *GetMySongsSuite) TestSongRepo_EmptyResultError() {
	s.sm.EXPECT().MySongs(mock.Anything, mock.Anything).Return(nil, repoerrs.ErrEmptyResult).Once()

	output, err := s.s.GetMySongs(s.ctx, s.input)
	s.NoError(err)
	s.Empty(output.Songs)
}

func (s *GetMySongsSuite) TestSongRepoError() {
	s.sm.EXPECT().MySongs(mock.Anything, mock.Anything).Return(nil, gofakeit.ErrorDatabase()).Once()

	_, err := s.s.GetMySongs(s.ctx, s.input)
	s.Error(err)
}

func (s *GetMySongsSuite) TestCountMySongsError() {
	s.sm.EXPECT().MySongs(mock.Anything, mock.Anything).Return(validSongRows(3), nil).Once()
	s.sm.EXPECT().CountMySongs(mock.Anything, s.input.UserId).Return(0, gofakeit.ErrorDatabase()).Once()

	_, err := s.s.GetMySongs(s.ctx, s.input)
	s.Error(err)
}

func (s *GetMySongsSuite) TestUserRepoError() {
	var calls, failed = 4000, 0

	rows := validSongRows(calls)

	s.sm.EXPECT().MySongs(mock.Anything, mock.Anything).Return(rows, nil).Once()
	s.sm.EXPECT().CountMySongs(mock.Anything, s.input.UserId).Return(int32(10000), nil).Once()
	s.um.EXPECT().ArtistsByIds(mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, ids []uuid.UUID) ([]users.Artist, error) {
			if rand.Uint()%16 == 0 {
				failed++
				return nil, gofakeit.Error()
			}

			return validArtists(len(ids)), nil
		})

	output, err := s.s.GetMySongs(s.ctx, s.input)
	s.Len(output.Songs, calls-failed)
	s.NoError(err)
}

func validGetMySongsInput() songs.GetMySongsInput {
	return songs.GetMySongsInput{
		UserId:   uuid.New(),
		Page:     1,
		PageSize: 10,
	}
}

func validSongRows(count int) []postgres.MySongsRow {
	songs := make([]postgres.MySongsRow, count)
	for i := 0; i < count; i++ {
		songs[i] = validSongsRow()
	}

	return songs
}

func validSongsRow() postgres.MySongsRow {
	return postgres.MySongsRow(validSongRow())
}

func TestGetMySongs(t *testing.T) {
	suite.Run(t, new(GetMySongsSuite))
}

func ptr[T any](v T) *T {
	return &v
}

type fakeRawService struct{}

func newFakeRawService() fakeRawService {
	return fakeRawService{}
}

func (fakeRawService) SongUrl(rawSongId string) string {
	return gofakeit.URL()
}
