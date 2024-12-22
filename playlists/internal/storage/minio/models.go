package minio

import "io"

type CoverObject struct {
	ID          string
	Extension   string
	WeightBytes int32
	Content     io.Reader
}
