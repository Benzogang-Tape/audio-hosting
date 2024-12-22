package handlers

import (
	"github.com/Benzogang-Tape/audio-hosting/playlists/api/protogen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (*PlaylistsService) UploadRawSong(
	_ grpc.ClientStreamingServer[protogen.UploadRawPlaylistCoverRequest, protogen.UploadRawPlaylistCoverResponse],
) error {
	return status.Errorf(codes.Unimplemented,
		"not available through gRPC, use HTTP %s instead",
		"/playlists/api/v1/playlist/{playlist_id}/cover")
}

func (*PlaylistsService) GetRawSong(
	_ *protogen.GetRawPlaylistCoverRequest, _ grpc.ServerStreamingServer[protogen.GetRawPlaylistCoverResponse],
) error {
	return status.Errorf(codes.Unimplemented,
		"not available through gRPC, use HTTP %s instead",
		"/playlists/api/v1/playlist/{playlist_id}/cover")
}
