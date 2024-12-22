package artists

import (
	"context"
	"fmt"

	"github.com/Benzogang-Tape/audio-hosting/users/internal/model"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/repo/options"
)

type DTOGetArtistsInput struct {
	options.Options
}

type DTOGetArtistsOutput struct {
	Artists    []model.Artist
	Pagination options.Pagination
}

func (se *Service) GetArtists(
	ctx context.Context,
	input DTOGetArtistsInput,
) (DTOGetArtistsOutput, error) {
	artistsCount, err := se.artistsRepository.GetArtistsCount(ctx, input.Options)
	if err != nil {
		return DTOGetArtistsOutput{}, fmt.Errorf(
			"artists.Service.GetArtists - get artists count: %w",
			err,
		)
	}

	lastPage := 1
	if input.Pagination.Limit > 0 {
		// Fast ceiling of positive integers division with avoided type overflowing.
		lastPage = (artistsCount-1)/input.Pagination.Limit + 1
		input.Pagination.Limit++
	}

	artists, err := se.artistsRepository.GetArtists(ctx, input.Options)
	if err != nil {
		return DTOGetArtistsOutput{}, fmt.Errorf(
			"artists.Service.GetArtists - get artists: %w",
			err,
		)
	}

	hasNext := false
	if len(artists) == input.Pagination.Limit &&
		len(artists) > 0 {
		hasNext = true
		artists = artists[:len(artists)-1]
	}

	return DTOGetArtistsOutput{
		Artists: artists,
		Pagination: options.Pagination{
			HasNext:  hasNext,
			Total:    artistsCount,
			LastPage: lastPage,
		},
	}, nil
}
