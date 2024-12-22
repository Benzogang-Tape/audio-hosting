package pgconv

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func Text(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func TextPtr(s *string) pgtype.Text {
	if s == nil {
		return NullText()
	}

	return pgtype.Text{String: *s, Valid: true}
}

func Int4[T interface{ int | int32 }](i T) pgtype.Int4 {
	return pgtype.Int4{
		Int32: int32(i),
		Valid: true,
	}
}

func Interval(dur time.Duration) pgtype.Interval {
	return pgtype.Interval{
		Microseconds: dur.Microseconds(),
		Days:         0,
		Months:       0,
		Valid:        true,
	}
}

func Timestamptz(i time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:             i,
		InfinityModifier: pgtype.Finite,
		Valid:            true,
	}
}

func Bool(b bool) pgtype.Bool {
	return pgtype.Bool{
		Bool:  b,
		Valid: true,
	}
}

func Float8(f float64) pgtype.Float8 {
	return pgtype.Float8{
		Float64: f,
		Valid:   true,
	}
}

func UUID(s uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: s, Valid: true}
}
