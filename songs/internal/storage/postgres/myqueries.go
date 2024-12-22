package postgres

import "github.com/google/uuid"

// Implementing songs.songRow interface.
func (r ReleasedSongsRow) GetSingerFk() uuid.UUID {
	return r.Song.SingerFk
}

func (r ReleasedSongsRow) GetArtistsIds() []uuid.UUID {
	return r.ArtistsIds
}

func (r MySongsRow) GetSingerFk() uuid.UUID {
	return r.Song.SingerFk
}

func (r MySongsRow) GetArtistsIds() []uuid.UUID {
	return r.ArtistsIds
}
