package options

type Sort struct {
	Enable bool

	Field string
	Order string
}

func (s Sort) QuottedField() string {
	return "\"" + s.Field + "\""
}

func (s Sort) QuottedOrder() string {
	return "" + s.Order + ""
}
