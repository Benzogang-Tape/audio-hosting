package grpcserver

import (
	"github.com/Benzogang-Tape/audio-hosting/songs/api/protogen/api"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (*songsServer) UploadRawSong(
	_ grpc.ClientStreamingServer[api.UploadRawSongRequest, api.UploadRawSongResponse],
) error {
	return status.Errorf(codes.Unimplemented,
		"not available through gRPC, use HTTP %s instead",
		"/songs/api/v1/song/raw")
}

func (*songsServer) GetRawSong(
	_ *api.GetRawSongRequest, _ grpc.ServerStreamingServer[api.GetRawSongResponse],
) error {
	return status.Errorf(codes.Unimplemented,
		"not available through gRPC, use HTTP %s instead",
		"/songs/api/v1/song/raw/{id}")
}

func (*songsServer) UploadRawSongImage(
	_ grpc.ClientStreamingServer[api.UploadRawSongImageRequest, api.UploadRawSongImageResponse],
) error {
	return status.Errorf(codes.Unimplemented,
		"not available through gRPC, use HTTP %s instead",
		"/songs/api/v1/song/image/raw")
}

func (*songsServer) GetRawSongImage(
	_ *api.GetRawSongImageRequest, _ grpc.ServerStreamingServer[api.GetRawSongImageResponse],
) error {
	return status.Errorf(codes.Unimplemented,
		"not available through gRPC, use HTTP %s instead",
		"/songs/api/v1/song/image/raw/{id}")
}
