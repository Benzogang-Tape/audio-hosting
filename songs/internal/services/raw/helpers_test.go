package raw

import (
	"os"
	"testing"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/pgconv"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func Test_ThisTestIsOnlyForCoverageItDoesNotMakeSense(t *testing.T) {
	a := randomSong()
	b := randomSong()

	obj := songsDiff(a, b)
	log := zerolog.New(os.Stdout)
	log.Info().Object("key", obj)
}

func randomSong() postgres.Song {
	return postgres.Song{
		SongID:       uuid.New(),
		SingerFk:     uuid.New(),
		Name:         gofakeit.Name(),
		S3ObjectName: pgconv.Text(uuid.NewString()),
		ImageUrl:     pgconv.Text(gofakeit.URL() + ".jpg"),
		Duration:     pgconv.Interval(gofakeit.Date().Sub(gofakeit.Date()).Abs()),
		WeightBytes:  pgconv.Int4(gofakeit.IntRange(1024, 1024*1024*1024)),
		UploadedAt:   gofakeit.Date(),
		ReleasedAt:   pgconv.Timestamptz(gofakeit.Date()),
	}
}
