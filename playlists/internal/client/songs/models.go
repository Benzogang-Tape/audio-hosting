package songs

import (
	songs "github.com/Benzogang-Tape/audio-hosting/playlists/api/protogen/clients/songs"
	users "github.com/Benzogang-Tape/audio-hosting/playlists/api/protogen/clients/users"
	"time"
)

type Song struct {
	ID          string
	Singer      Artist
	Artists     []Artist
	Name        string
	SongURL     string
	ImageURL    string
	Duration    time.Duration
	WeightBytes int32
	ReleasedAt  time.Time
}

func convSong(songs *songs.Song) Song {
	artists := make([]Artist, 0, len(songs.Artists))
	for _, artist := range songs.Artists {
		artists = append(artists, convArtist(artist))
	}

	return Song{
		ID:          songs.GetId(),
		Singer:      convArtist(songs.GetSinger()),
		Artists:     artists,
		Name:        songs.GetName(),
		SongURL:     songs.GetSongUrl(),
		ImageURL:    songs.GetImageUrl(),
		Duration:    songs.GetDuration().AsDuration(),
		WeightBytes: songs.GetWeightBytes(),
		ReleasedAt:  songs.GetReleasedAt().AsTime(),
	}
}

//
//type MySong struct {
//	ID          string
//	Singer      Artist
//	Artists     []Artist
//	Name        string
//	SongURL     string
//	ImageURL    string
//	Duration    time.Duration
//	WeightBytes int32
//	ReleasedAt  time.Time
//	UploadedAt  time.Time
//}

type Artist struct {
	ID        string
	Name      string
	Label     string
	AvatarURL string
}

func convArtist(artist *users.Artist) Artist {
	return Artist{
		ID:        artist.GetId(),
		Name:      artist.GetName(),
		Label:     artist.GetLabel(),
		AvatarURL: artist.GetAvatarUrl(),
	}
}
