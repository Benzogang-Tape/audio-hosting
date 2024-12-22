package erix

import (
	"errors"
	"net/http"

	"dev.gaijin.team/go/golib/e"
	"dev.gaijin.team/go/golib/fields"
	"google.golang.org/grpc/codes"
)

type StatusErr struct {
	*e.Err
	code int
}

func NewStatus(reason string, code int) *StatusErr {
	return &StatusErr{
		Err:  e.New(reason),
		code: code,
	}
}

// Wrap creates a new StatusErr that wraps the provided error with the source one.
func (s *StatusErr) Wrap(err error, fls ...fields.Field) *StatusErr {
	return &StatusErr{
		// Maybe not the most convenient way, but that works!
		Err:  e.From(s).Wrap(err, fls...),
		code: s.code,
	}
}

func GrpcCode(err error) codes.Code {
	if err == nil {
		return codes.OK
	}

	statErr := new(StatusErr)
	if errors.As(err, &statErr) {
		return ToGrpc(statErr.code)
	}

	return codes.Internal
}

func HttpCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	statErr := new(StatusErr)
	if errors.As(err, &statErr) {
		return ToHttp(statErr.code)
	}

	return http.StatusInternalServerError
}

func LastReason(err error) string {
	statErr := new(StatusErr)
	if errors.As(err, &statErr) {
		return statErr.Err.Reason()
	}

	eErr := new(e.Err)
	if errors.As(err, &eErr) {
		return eErr.Reason()
	}

	// idc if it panics
	return err.Error()
}
