package fields

import (
	"fmt"

	"github.com/shopspring/decimal"
)

const (
	DataTypeStr     = "string"
	DataTypeNumeric = "numeric"

	OperatorEq            = "eq"
	OperatorNotEq         = "neq"
	OperatorLowerThan     = "lt"
	OperatorLowerThanEq   = "lte"
	OperatorGreaterThan   = "gt"
	OperatorGreaterThanEq = "gte"
	OperatorLike          = "like"
)

type F struct {
	Name  string
	Op    string
	Value string
	Type  string
}

func (f F) FormattedValue() string {
	if f.Op == OperatorLike {
		return fmt.Sprint("%", f.Value, "%")
	}

	return f.Value
}

func (f F) ToQueryWithParameter() string {
	if f.Op == OperatorLike && f.Type == DataTypeNumeric {
		return fmt.Sprintf("%s::text %s ?", f.Name, f.FormattedOperator())
	}

	return fmt.Sprintf("%s %s ?", f.Name, f.FormattedOperator())
}

func (f F) FormattedOperator() string {
	switch f.Op {
	case OperatorEq:
		return "="
	case OperatorNotEq:
		return "!="
	case OperatorLowerThan:
		return "<"
	case OperatorLowerThanEq:
		return "<="
	case OperatorGreaterThan:
		return ">"
	case OperatorGreaterThanEq:
		return ">="
	case OperatorLike:
		return "like"
	default:
		return ""
	}
}

func ParseOperator(op string) (string, error) {
	switch op {
	case OperatorEq:
		return "=", nil
	case OperatorNotEq:
		return "!=", nil
	case OperatorLowerThan:
		return "<", nil
	case OperatorLowerThanEq:
		return "<=", nil
	case OperatorGreaterThan:
		return ">", nil
	case OperatorGreaterThanEq:
		return ">=", nil
	case OperatorLike:
		return "like", nil
	}
	return "", fmt.Errorf("bad operator")
}

func ValidateOperator(op string) error {
	switch op {
	case OperatorEq,
		OperatorNotEq,
		OperatorLowerThan,
		OperatorLowerThanEq,
		OperatorGreaterThan,
		OperatorGreaterThanEq,
		OperatorLike:
		return nil
	}
	return fmt.Errorf("bad operator")
}

func ValidateValueWithType(value string, type_ string) error {
	switch type_ {
	case DataTypeStr:
		return nil
	case DataTypeNumeric:
		_, err := decimal.NewFromString(value)
		if err != nil {
			return fmt.Errorf("invalid numeric value")
		}
		return nil
	}

	return fmt.Errorf("bad type")
}
