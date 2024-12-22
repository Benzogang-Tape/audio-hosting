package handlers

import (
	"github.com/Benzogang-Tape/audio-hosting/playlists/api/protogen"
	protosongs "github.com/Benzogang-Tape/audio-hosting/playlists/api/protogen/clients/songs"
	protousers "github.com/Benzogang-Tape/audio-hosting/playlists/api/protogen/clients/users"
	client "github.com/Benzogang-Tape/audio-hosting/playlists/internal/client/songs"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/models"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convSongs(songs []client.Song) []*protosongs.Song {
	msg := make([]*protosongs.Song, 0, len(songs))

	for _, song := range songs {
		msg = append(msg, &protosongs.Song{
			Id:   song.ID,
			Name: song.Name,
			Singer: &protousers.Artist{
				Id:        song.Singer.ID,
				Name:      song.Singer.Name,
				Label:     song.Singer.Label,
				AvatarUrl: song.Singer.AvatarURL,
			},
			Artists:     convArtists(song.Artists),
			SongUrl:     song.SongURL,
			ImageUrl:    &song.ImageURL,
			Duration:    durationpb.New(song.Duration),
			WeightBytes: song.WeightBytes,
			ReleasedAt:  timestamppb.New(song.ReleasedAt),
		})
	}

	return msg
}

func convArtists(artists []client.Artist) []*protousers.Artist {
	msg := make([]*protousers.Artist, 0, len(artists))

	for _, artist := range artists {
		msg = append(msg, &protousers.Artist{
			Id:        artist.ID,
			Name:      artist.Name,
			Label:     artist.Label,
			AvatarUrl: artist.AvatarURL,
		})
	}

	return msg
}

func convPlaylistMetadata(md models.PlaylistMetadata) *protogen.PlaylistMetadata {
	return &protogen.PlaylistMetadata{
		Id:             md.ID,
		Title:          md.Title,
		AuthorId:       md.AuthorID,
		CoverUrl:       md.CoverURL,
		CreatedAt:      timestamppb.New(md.CreatedAt),
		IsAlbum:        md.IsAlbum,
		IsMyCollection: md.IsMyCollection,
		IsPublic:       md.IsPublic,
	}
}
