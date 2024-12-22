package s3minio

import (
	"context"
	"io"

	"dev.gaijin.team/go/golib/e"
	"dev.gaijin.team/go/golib/fields"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Storage struct {
	m            *minio.Client
	songsBucket  string
	imagesBucket string
}

type Config struct {
	Endpoint     string
	AccessKey    string
	SecretKey    string
	UseSsl       bool
	SongsBucket  string
	ImagesBucket string
}

func Connect(ctx context.Context, conf Config) (*S3Storage, error) {
	client, err := minio.New(conf.Endpoint, &minio.Options{ //nolint:exhaustruct
		Creds:  credentials.NewStaticV4(conf.AccessKey, conf.SecretKey, ""),
		Secure: conf.UseSsl,
	})
	if err != nil {
		return nil, e.NewFrom("connecting to minio", err)
	}

	m := &S3Storage{
		m:            client,
		songsBucket:  conf.SongsBucket,
		imagesBucket: conf.ImagesBucket,
	}

	err = m.createBuckets(ctx, conf.SongsBucket, conf.ImagesBucket)
	if err != nil {
		return nil, e.NewFrom("creating buckets", err)
	}

	return m, nil
}

func (m *S3Storage) createBuckets(ctx context.Context, buckets ...string) error {
	for _, bucket := range buckets {
		ok, err := m.m.BucketExists(ctx, bucket)
		if err != nil {
			return e.NewFrom("checking if bucket exists", err, fields.F("bucket", bucket))
		}

		if !ok {
			err = m.m.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}) //nolint:exhaustruct
			if err != nil {
				return e.NewFrom("creating bucket", err, fields.F("bucket", bucket))
			}
		}
	}

	return nil
}

func (m *S3Storage) PutSongObject(ctx context.Context, song SongObject) error {
	_, err := m.m.PutObject(ctx, m.songsBucket, song.Id,
		song.Content, int64(song.WeightBytes), minio.PutObjectOptions{}) //nolint:exhaustruct
	if err != nil {
		return e.NewFrom("saving song to minio", err,
			fields.F("song_id", song.Id), fields.F("weight", song.WeightBytes))
	}

	return nil
}

func (m *S3Storage) GetSongObject(ctx context.Context, id string) (io.Reader, error) {
	object, err := m.m.GetObject(ctx, m.songsBucket, id, minio.GetObjectOptions{}) //nolint:exhaustruct
	if err != nil {
		return nil, e.NewFrom("getting song from minio", err, fields.F("song_id", id))
	}

	return object, nil
}

func (m *S3Storage) PutImageObject(ctx context.Context, image ImageObject) error {
	_, err := m.m.PutObject(ctx, m.imagesBucket, image.Id,
		image.Content, int64(image.WeightBytes), minio.PutObjectOptions{}) //nolint:exhaustruct
	if err != nil {
		return e.NewFrom("saving song image to minio", err,
			fields.F("image_id", image.Id), fields.F("weight", image.WeightBytes))
	}

	return nil
}

func (m *S3Storage) GetImageObject(ctx context.Context, id string) (io.Reader, error) {
	object, err := m.m.GetObject(ctx, m.imagesBucket, id, minio.GetObjectOptions{}) //nolint:exhaustruct
	if err != nil {
		return nil, e.NewFrom("getting song image from minio", err, fields.F("image_id", id))
	}

	return object, nil
}
