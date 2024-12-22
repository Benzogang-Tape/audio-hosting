package s3minio

import (
	"io"
	"time"
)

type SongObject struct {
	Id          string
	Extension   string
	Duration    time.Duration
	WeightBytes int32
	Content     io.Reader
}

type ImageObject struct {
	Id          string
	Extension   string
	WeightBytes int32
	Content     io.Reader
}
