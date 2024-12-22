package gateway

import (
	"dev.gaijin.team/go/golib/fields"
	"dev.gaijin.team/go/golib/stacktrace"
	"encoding/json"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/erix"
	"github.com/google/uuid"
	"go.uber.org/zap/zapcore"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	ErrParsingForm = erix.NewStatus("failed to parse form, must be multipart/form-data", erix.CodeBadRequest)
	ErrInvalidForm = erix.NewStatus(
		"form must contain UUID 'playlist_id' and file 'attachment'", erix.CodeBadRequest)
	ErrFileTooLarge = erix.NewStatus("file too large, max size is 2GB", erix.CodeBadRequest)
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
	PlaylistId uuid.UUID
	Ext        string
	Size       int32
	File       io.Reader
}

func getFileFromPlaylistForm(r *http.Request) (FormFile, error) {
	err := r.ParseForm()
	if err != nil {
		return FormFile{}, ErrParsingForm.Wrap(err)
	}

	file, header, err := r.FormFile("attachment")
	if err != nil {
		return FormFile{}, ErrInvalidForm.Wrap(err, fields.F("form", "attachment"))
	}
	defer file.Close()

	if header.Size > math.MaxInt32 {
		return FormFile{}, ErrFileTooLarge
	}

	playlistID, err := getIDFromPath(r.URL.Path)
	if err != nil {
		return FormFile{}, ErrInvalidForm.Wrap(err, fields.F("form", "playlist_id"), fields.F("value", playlistID))
	}

	return FormFile{
		PlaylistId: playlistID,
		Ext:        filepath.Ext(header.Filename),
		Size:       int32(header.Size), //nolint:gosec
		File:       file,
	}, nil
}

func getIDFromPath(path string) (uuid.UUID, error) {
	p := strings.Split(path, "/")

	id, err := uuid.Parse(p[len(p)-2])
	return id, err
}

// Wrapper for stacktrace frames to use in zap.Array.
type wrappedFrame struct {
	runtime.Frame
}

func (w wrappedFrame) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("file", w.File)
	enc.AddInt("line", w.Line)
	enc.AddString("function", w.Function)

	return nil
}

type wrappedFrames struct {
	frames []wrappedFrame
}

func wrapFramesFromStack(stack *stacktrace.Stack) wrappedFrames {
	var frames []wrappedFrame

	for _, frame := range stack.FramesIter() {
		frames = append(frames, wrappedFrame{Frame: frame})
	}

	return wrappedFrames{frames: frames}
}

func (w wrappedFrames) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, frame := range w.frames {
		err := enc.AppendObject(frame)
		if err != nil {
			return err
		}
	}

	return nil
}
