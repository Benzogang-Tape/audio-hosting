package models

import (
	client "github.com/Benzogang-Tape/audio-hosting/playlists/internal/client/songs"
	"time"
)

type PlaylistMetadata struct {
	ID             string    `json:"id" db:"id"`
	Title          string    `json:"title" db:"title"`
	AuthorID       string    `json:"author_id" db:"author_id"`
	CoverURL       string    `json:"cover_url" db:"cover_url"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	ReleasedAt     time.Time `json:"released_at" db:"released_at"`
	IsAlbum        bool      `json:"is_album" db:"is_album"`
	IsMyCollection bool      `json:"is_my_collection" db:"-"`
	IsPublic       bool      `json:"is_public" db:"is_public"`
}

type Playlist struct {
	Metadata PlaylistMetadata
	Songs    []client.Song
}
