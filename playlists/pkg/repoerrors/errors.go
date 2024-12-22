package repoerrors

import "dev.gaijin.team/go/golib/e"

var (
	ErrEmptyResult = e.New("empty result")
	ErrUnique      = e.New("unique constraint violation")
	ErrChecked     = e.New("check constraint violation")
	ErrFK          = e.New("foreign key constraint violation")
	ErrNull        = e.New("not null constraint violation")
)
