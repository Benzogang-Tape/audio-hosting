package erix_test

import (
	"database/sql"
	"net/http"
	"testing"

	"dev.gaijin.team/go/golib/e"
	"dev.gaijin.team/go/golib/fields"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/erix"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

// The test is temporary.

var ErrRepo = e.New("something repo")
var ErrSomething = erix.NewStatus("something service", erix.CodeBadRequest)
var ErrSomething2 = erix.NewStatus("something service 2", erix.CodeNotFound)
var ErrController = e.New("something controller")

func TestEverything(t *testing.T) {

	var repoErr error = ErrRepo.Wrap(sql.ErrNoRows, fields.F("wtf", "null"))

	var serviceErr error = ErrSomething.Wrap(repoErr, fields.F("what", "query failed"))

	var serviceErr2 error = ErrSomething2.Wrap(serviceErr)

	var err error = ErrController.Wrap(serviceErr2, fields.F("controller", 52))

	assert.Equal(t, "something service 2", erix.LastReason(err))
	assert.Equal(t, http.StatusNotFound, erix.HttpCode(err))
	assert.Equal(t, codes.NotFound, erix.GrpcCode(err))
}

func TestWrap(t *testing.T) {
	statusErr := erix.NewStatus("error B", erix.CodeBadRequest)
	gaijinErr := e.New("error A")

	err := statusErr.Wrap(gaijinErr, fields.F("field", "v"))

	assert.ErrorIs(t, err, statusErr)
	assert.Equal(t, "error B (field=v): error A", err.Error())
}
