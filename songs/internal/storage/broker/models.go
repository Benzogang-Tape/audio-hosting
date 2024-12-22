package broker

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SongReleasedMessage struct {
	SongId     uuid.UUID `json:"song_id"`
	ArtistId   uuid.UUID `json:"artist_id"`
	Name       string    `json:"name"`
	ReleasedAt time.Time `json:"released_at"`
}

func (m SongReleasedMessage) Bytes() []byte {
	bytes, err := json.Marshal(m)
	if err != nil {
		return []byte{}
	}

	return bytes
}
