package service

import "github.com/Benzogang-Tape/audio-hosting/playlists/pkg/erix"

var (
	ErrSavingPlaylist   = erix.NewStatus("can't save playlist", erix.CodeInternalServerError)
	ErrPlaylistNotFound = erix.NewStatus("playlist not found", erix.CodeNotFound)
	ErrAlbumNotFound    = erix.NewStatus("album not found", erix.CodeNotFound)
	ErrGetSongs         = erix.NewStatus("failed to get song", erix.CodeInternalServerError)
	ErrSongDoesNotExist = erix.NewStatus("song does not exist", erix.CodeNotFound)
	ErrUpdatePlaylist   = erix.NewStatus(
		"can't update playlist: playlist not found or user is not playlist's owner",
		erix.CodeBadRequest,
	)
	ErrNoPlaylistToLike      = erix.NewStatus("no playlist to like", erix.CodeBadRequest)
	ErrInvalidCoverExtension = erix.NewStatus("invalid extension, only jpg, png, jpeg supported", erix.CodeBadRequest)

	ErrNoFilters       = erix.NewStatus("no filters", erix.CodeBadRequest)
	ErrMultipleFilters = erix.NewStatus("multiple filters", erix.CodeBadRequest)
)
