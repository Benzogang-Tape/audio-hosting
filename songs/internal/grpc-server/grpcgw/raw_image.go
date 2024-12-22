package grpcgw

import (
	"io"
	"net/http"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/services/raw"
	"github.com/google/uuid"
)

func (s RawHandlers) UploadRawSongImageHandler() HandlerErrFunc {
	type response struct {
		ImageUrl string `json:"imageUrl"`
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

		out, err := s.Service.UploadRawSongImage(r.Context(), raw.UploadRawSongImageInput{
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
			ImageUrl: out.ImageUrl,
		})
	}
}

func (s RawHandlers) GetRawSongImageHandler() HandlerErrFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) error {
		songId := pathParams["id"]

		output, err := s.Service.GetRawSongImage(r.Context(), songId)
		if err != nil {
			return err
		}

		_, err = io.Copy(w, output.Content)
		if err != nil {
			return ErrFileNotFound.Wrap(err)
		}

		w.Header().Set("Content-Type", "audio/"+output.Extension)

		return nil
	}
}
