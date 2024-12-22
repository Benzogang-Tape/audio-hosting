package gateway

import (
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service/covers"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/erix"
	"io"
	"net/http"
)

var (
	ErrTokenNotFound = erix.NewStatus("token not found", erix.CodeInternalServerError)
	ErrFileNotFound  = erix.NewStatus("file not found", erix.CodeNotFound)
)

func (s CoversHandlers) UploadRawCoverHandler() HandlerErrFunc {
	type response struct {
		CoverURL string `json:"coverUrl"`
	}

	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) error {
		token, ok := TokenFromCtx(r.Context())
		if !ok {
			return ErrTokenNotFound
		}

		form, err := getFileFromPlaylistForm(r)
		if err != nil {
			return err
		}

		out, err := s.service.UploadRawCover(r.Context(), covers.UploadRawCoverInput{
			UserId:      token.Subject,
			PlaylistId:  form.PlaylistId,
			Extension:   form.Ext,
			WeightBytes: form.Size,
			Content:     form.File,
		})
		if err != nil {
			return err //nolint:wrapcheck
		}

		return jsonResp(w, response{
			CoverURL: out.CoverUrl,
		})
	}
}

func (s CoversHandlers) GetRawCoverHandler() HandlerErrFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) error {
		playlistID := pathParams["cover_id"]

		output, err := s.service.GetRawCover(r.Context(), playlistID)
		if err != nil {
			return err
		}

		_, err = io.Copy(w, output.Content)
		if err != nil {
			return ErrFileNotFound.Wrap(err)
		}

		w.Header().Set("Content-Type", "image/"+output.Extension)

		return nil
	}
}
