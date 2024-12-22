package grpcserver

import (
	"github.com/AlekSi/pointer"
	"github.com/Benzogang-Tape/audio-hosting/playlists/api/protogen"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"google.golang.org/protobuf/proto"
	"net/url"
	"strconv"
)

type QueryParser struct{}

func (*QueryParser) Parse(target proto.Message, values url.Values, filter *utilities.DoubleArray) error {
	switch req := target.(type) { //nolint:gocritic
	// Different messages/requests can have different parsers, of course
	case *protogen.GetPlaylistsRequest:
		return populateGetPlaylistsParams(values, req)
	}

	return (&runtime.DefaultQueryParser{}).Parse(target, values, filter) //nolint:wrapcheck
}

// populateGetPlaylistsParams populates the GetPlaylistsRequest with query parameters.
// TODO: filters
func populateGetPlaylistsParams(values url.Values, r *protogen.GetPlaylistsRequest) error {
	if r.GetIds(); len(r.GetIds()) > 0 {
		return nil
	}

	// Parsing pagination
	var pageOptions protogen.PaginationRequest

	if limit := values.Get("limit"); limit != "" {
		if parsedLimit, err := strconv.Atoi(limit); err == nil {
			pageOptions.Limit = int32(parsedLimit)
		}
	}

	if page := values.Get("page"); page != "" {
		if parsedPage, err := strconv.Atoi(page); err == nil {
			pageOptions.Page = int32(parsedPage) //nolint:gosec
		}
	}

	r.Pagination = &pageOptions

	// Parsing filters
	var filterOptions protogen.Filter

	if matchTitle := values.Get("match_title"); matchTitle != "" {
		filterOptions.MatchTitle = pointer.To(matchTitle)
	} else if artistID := values.Get("artist_id"); artistID != "" {
		filterOptions.ArtistId = pointer.To(artistID)
	}

	r.Filter = &filterOptions

	return nil
}
