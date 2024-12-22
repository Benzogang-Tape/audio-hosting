package logger

import (
	"fmt"

	"github.com/rs/zerolog"
)

// ObjectFunc provides a way to marshal custom objects into zerolog.Event
// using just a lambda.
type ObjectFunc func(*zerolog.Event)

func (obj ObjectFunc) MarshalZerologObject(e *zerolog.Event) {
	obj(e)
}

// ArrayFunc provides a way to marshal custom arrays into zerolog.Event
// using just a lambda.
type ArrayFunc func(*zerolog.Array)

func (arr ArrayFunc) MarshalZerologArray(a *zerolog.Array) {
	arr(a)
}

// Stringers provides a zerolog.ArrayMarshaler that allows to
// marshal a slice of fmt.Stringer compatible items.
//
// For other types can use [ArrayFunc].
type Stringers[T fmt.Stringer] []T

func (s Stringers[T]) MarshalZerologArray(a *zerolog.Array) {
	for _, v := range s {
		a.Str(v.String())
	}
}
