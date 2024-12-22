package options

type Pagination struct {
	Enable bool

	Limit    int
	Offset   int
	HasNext  bool
	Total    int
	LastPage int
}
