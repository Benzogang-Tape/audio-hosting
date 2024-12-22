package options

type Options struct {
	Filter     Filter
	Sort       Sort
	Pagination Pagination
}

func (o Options) WithPagination(limit, offset int) Options {
	return Options{
		Filter: o.Filter,
		Sort:   o.Sort,
		Pagination: Pagination{ //nolint:exhaustruct
			Enable: true,
			Limit:  limit,
			Offset: offset,
		},
	}
}

func (o Options) WithSort(field, order string) Options {
	return Options{
		Filter: o.Filter,
		Sort: Sort{
			Enable: true,
			Field:  field,
			Order:  order,
		},
		Pagination: o.Pagination,
	}
}

func NewEmpty() Options {
	return Options{
		Filter: Filter{ //nolint:exhaustruct
			Enable: false,
		},
		Sort: Sort{ //nolint:exhaustruct
			Enable: false,
		},
		Pagination: Pagination{ //nolint:exhaustruct
			Enable: false,
		},
	}
}
