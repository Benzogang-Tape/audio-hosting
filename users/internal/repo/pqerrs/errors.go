package pqerrs

import "errors"

const (
	ErrUniqueViolationCode = "23505"
	ErrFKViolationCode     = "23503"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrUniqueViolation = errors.New("unique violation")
	ErrFKViolation     = errors.New("foreign key violation")
)
