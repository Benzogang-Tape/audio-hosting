package minio

import (
	"context"
	"dev.gaijin.team/go/golib/e"
	"dev.gaijin.team/go/golib/fields"
	s3 "github.com/Benzogang-Tape/audio-hosting/playlists/pkg/db/minio"
	"github.com/minio/minio-go/v7"
	"io"
)

type S3Storage struct {
	m            *minio.Client
	coversBucket string
}

func New(ctx context.Context, conf s3.Config) (*S3Storage, error) {
	client, err := s3.Connect(conf)
	if err != nil {
		return nil, e.NewFrom("connecting to minio", err)
	}

	m := &S3Storage{
		m:            client,
		coversBucket: conf.CoversBucket,
	}

	err = m.createBuckets(ctx, m.coversBucket)
	if err != nil {
		return nil, e.NewFrom("creating buckets", err)
	}

	return m, nil
}

func (s *S3Storage) createBuckets(ctx context.Context, buckets ...string) error {
	for _, bucket := range buckets {
		ok, err := s.m.BucketExists(ctx, bucket)
		if err != nil {
			return e.NewFrom("checking for bucket existence", err, fields.F("bucket", bucket))
		}

		if !ok {
			err = s.m.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}) //nolint:exhaustruct
			if err != nil {
				return e.NewFrom("creating bucket", err, fields.F("bucket", bucket))
			}
		}
	}

	return nil
}

func (s *S3Storage) PutCoverObject(ctx context.Context, image CoverObject) error {
	_, err := s.m.PutObject(ctx, s.coversBucket, image.ID,
		image.Content, int64(image.WeightBytes), minio.PutObjectOptions{}) //nolint:exhaustruct
	if err != nil {
		return e.NewFrom("saving song image to minio", err,
			fields.F("image_id", image.ID), fields.F("weight", image.WeightBytes))
	}

	return nil
}

func (s *S3Storage) GetCoverObject(ctx context.Context, id string) (io.Reader, error) {
	object, err := s.m.GetObject(ctx, s.coversBucket, id, minio.GetObjectOptions{}) //nolint:exhaustruct
	if err != nil {
		return nil, e.NewFrom("getting song image from minio", err, fields.F("image_id", id))
	}

	return object, nil
}
