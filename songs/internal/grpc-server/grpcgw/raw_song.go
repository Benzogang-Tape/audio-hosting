package grpcgw

import (
	"context"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/services/raw"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/erix"
)

type RawService interface {
	UploadRawSong(ctx context.Context, in raw.UploadRawSongInput) (raw.UploadRawSongOutput, error)
	GetRawSong(ctx context.Context, songId string) (io.Reader, error)
	UploadRawSongImage(ctx context.Context, input raw.UploadRawSongImageInput) (raw.UploadRawSongImageOutput, error)
	GetRawSongImage(ctx context.Context, songId string) (raw.GetRawSongImageOutput, error)
}

type RawHandlers struct {
	Service RawService
}

var (
	ErrFileNotFound = erix.NewStatus("file not found", erix.CodeNotFound)
)

func (s RawHandlers) UploadRawSongHandler() HandlerErrFunc {
	type response struct {
		SongUrl string `json:"songUrl"`
	}

	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) error {
		token := TokenFromCtx(r.Context())

		form, err := getFileFromSongForm(r)
		if err != nil {
			return err
		}

		songId, err := uuid.Parse(pathParams["song_id"])
		if err != nil {
			return ErrNotUuid.Wrap(err)
		}

		out, err := s.Service.UploadRawSong(r.Context(), raw.UploadRawSongInput{
			ArtistId:    token.Subject,
			SongId:      songId,
			Extension:   form.Ext,
			WeightBytes: form.Size,
			Content:     form.File,
		})
		if err != nil {
			return err
		}

		return jsonResp(w, response{
			SongUrl: out.SongUrl,
		})
	}
}

func (s RawHandlers) GetRawSongHandler() HandlerErrFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) error {
		songId := pathParams["id"]

		reader, err := s.Service.GetRawSong(r.Context(), songId)
		if err != nil {
			return err
		}

		_, err = io.Copy(w, reader)
		if err != nil {
			return ErrFileNotFound.Wrap(err)
		}

		w.Header().Set("Content-Type", "audio/mpeg")

		return nil
	}
}
