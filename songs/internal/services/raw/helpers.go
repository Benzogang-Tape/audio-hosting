package raw

import (
	"crypto/sha1" //nolint:gosec
	"encoding/hex"
	"reflect"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/pgconv"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func getObjectId(artistId uuid.UUID, songName, fileExt string) string {
	objectIdHash := sha1.Sum([]byte(artistId.String() + "\u0002" + songName)) //nolint:gosec
	return hex.EncodeToString(objectIdHash[:]) + "." + fileExt
}

func (s *ServiceRaw) SongUrl(rawSongId string) string {
	return s.songUrlTpl + rawSongId
}

func (s *ServiceRaw) ImageUrl(rawImageId string) string {
	return s.imageUrlTpl + rawImageId
}

func songsDiff(a, b postgres.Song) zerolog.LogObjectMarshaler {
	return logger.ObjectFunc(func(e *zerolog.Event) {
		if a.SongID != b.SongID {
			e.Stringer("old_song_id", a.SongID)
			e.Stringer("song_id", b.SongID)
		}

		if a.SingerFk != b.SingerFk {
			e.Stringer("old_singer_fk", b.SingerFk)
			e.Stringer("singer_fk", b.SingerFk)
		}

		if a.Name != b.Name {
			e.Str("old_name", a.Name)
			e.Str("name", b.Name)
		}

		if a.S3ObjectName.String != b.S3ObjectName.String {
			e.Str("old_s3_object_name", a.S3ObjectName.String)
			e.Str("s3_object_name", b.S3ObjectName.String)
		}

		if a.ImageUrl.String != b.ImageUrl.String {
			e.Str("old_image_url", a.ImageUrl.String)
			e.Str("image_url", b.ImageUrl.String)
		}

		if !reflect.DeepEqual(a.Duration, b.Duration) {
			e.Dur("old_duration", pointer.Get(pgconv.FromInterval(a.Duration)))
			e.Dur("duration", pointer.Get(pgconv.FromInterval(b.Duration)))
		}

		if !reflect.DeepEqual(a.WeightBytes, b.WeightBytes) {
			e.Int32("old_weight_bytes", a.WeightBytes.Int32)
			e.Int32("weight_bytes", b.WeightBytes.Int32)
		}

		if !reflect.DeepEqual(a.UploadedAt, b.UploadedAt) {
			e.Time("old_uploaded_at", a.UploadedAt)
			e.Time("uploaded_at", b.UploadedAt)
		}

		if a.ReleasedAt.Time != b.ReleasedAt.Time {
			e.Time("old_released_at", a.ReleasedAt.Time)
			e.Time("released_at", b.ReleasedAt.Time)
		}
	})
}
