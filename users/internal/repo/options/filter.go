package options

import (
	"fmt"

	"github.com/Benzogang-Tape/audio-hosting/users/pkg/fields"
)

type Filter struct {
	Enable bool

	Fields []fields.F
}

func (f *Filter) AddField(name, op, value, type_ string) error {
	if err := fields.ValidateOperator(op); err != nil {
		return fmt.Errorf("can't add field: %w", err)
	}

	if err := fields.ValidateValueWithType(value, type_); err != nil {
		return err
	}

	f.Fields = append(f.Fields, fields.F{
		Name:  name,
		Op:    op,
		Value: value,
		Type:  type_,
	})

	return nil
}
