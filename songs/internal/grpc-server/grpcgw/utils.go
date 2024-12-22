package grpcgw

import (
	"encoding/json"
	"io"
	"math"
	"net/http"
	"path/filepath"

	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/erix"
)

var (
	ErrParsingForm  = erix.NewStatus("failed to parse form, must be multipart/form-data", erix.CodeBadRequest)
	ErrInvalidForm  = erix.NewStatus("form must contain file 'attachment'", erix.CodeBadRequest)
	ErrFileTooLarge = erix.NewStatus("file too large, max size is 2GB", erix.CodeBadRequest)
	ErrNotUuid      = erix.NewStatus("song_id path param must be uuid", erix.CodeBadRequest)
)

func jsonResp[T any](w http.ResponseWriter, resp T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	bytes, err := json.Marshal(resp)
	if err != nil {
		return err //nolint:wrapcheck
	}

	_, err = w.Write(bytes)

	return err //nolint:wrapcheck
}

func jsonErr(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(erix.HttpCode(err))

	_, _ = w.Write([]byte("{\"message\": \"" + erix.LastReason(err) + "\"}"))
}

type FormFile struct {
	Ext  string
	Size int32
	File io.Reader
}

func getFileFromSongForm(r *http.Request) (FormFile, error) {
	err := r.ParseForm()
	if err != nil {
		return FormFile{}, ErrParsingForm.Wrap(err)
	}

	file, header, err := r.FormFile("attachment")
	if err != nil {
		return FormFile{}, ErrInvalidForm.Wrap(err)
	}
	defer file.Close()

	if header.Size > math.MaxInt32 {
		return FormFile{}, ErrFileTooLarge
	}

	return FormFile{
		Ext:  filepath.Ext(header.Filename)[1:],
		Size: int32(header.Size), //nolint:gosec
		File: file,
	}, nil
}
