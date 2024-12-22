package handlers

import (
	"fmt"

	"github.com/Benzogang-Tape/audio-hosting/users/api/protogen"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/repo/options"
	"github.com/Benzogang-Tape/audio-hosting/users/pkg/fields"
)

const (
	defaultLimit  = 20
	defaultOffset = 0
)

func ParsePagination(pagination *protogen.PaginationRequest) (options.Pagination, error) {
	if pagination == nil {
		return options.Pagination{
			Enable: true,
			Limit:  defaultLimit,
			Offset: defaultOffset,
		}, nil
	}

	if pagination.Limit == 0 {
		pagination.Limit = defaultLimit
	}

	if pagination.Offset == 0 {
		pagination.Offset = defaultOffset
	}

	return options.Pagination{
		Enable: true,
		Limit:  int(pagination.Limit),
		Offset: int(pagination.Offset),
	}, nil
}

const (
	ArtistEntityName = "artist"
)

func ParseSort(sort *protogen.Sort, entityName string) (options.Sort, error) {
	if sort == nil {
		return options.Sort{
			Enable: false,
		}, nil
	}

	var s options.Sort
	switch entityName {
	case ArtistEntityName:
		field, err := mapArtistsFieldName(sort.Field)
		if err != nil {
			return options.Sort{}, err
		}

		if sort.Order != "asc" && sort.Order != "desc" {
			return options.Sort{}, fmt.Errorf("unknown order: %s", sort.Order)
		}

		s = options.Sort{
			Enable: true,
			Field:  field,
			Order:  sort.Order,
		}

	default:
		return options.Sort{}, fmt.Errorf("unknown entity: %s", entityName)
	}

	s.Enable = true

	return s, nil
}

func ParseFilters(filters []*protogen.Filter, entityName string) (options.Filter, error) {
	if filters == nil {
		return options.Filter{
			Enable: false,
		}, nil
	}

	var f options.Filter
	switch entityName {
	case ArtistEntityName:
		for _, filter := range filters {
			field, err := mapArtistsFieldName(filter.Field)
			if err != nil {
				return options.Filter{}, err
			}

			type_, err := getArtistsFieldType(filter.Field)
			if err != nil {
				return options.Filter{}, err
			}

			if err := f.AddField(field, filter.Operator, filter.Value, type_); err != nil {
				return options.Filter{}, err
			}
		}
	default:
		return options.Filter{}, fmt.Errorf("unknown entity: %s", entityName)
	}

	f.Enable = true

	return f, nil
}

func getArtistsFieldType(field string) (string, error) {
	switch field {
	case "name", "label", "id":
		return fields.DataTypeStr, nil
	default:
		return "", fmt.Errorf("unknown field: %s", field)
	}
}

func mapArtistsFieldName(fieldName string) (string, error) {
	fields := map[string]string{
		"name":  "name",
		"id":    "users.id",
		"label": "label",
	}

	field, ok := fields[fieldName]
	if !ok {
		return "", fmt.Errorf("unknown field name: %s", fieldName)
	}

	return field, nil
}
