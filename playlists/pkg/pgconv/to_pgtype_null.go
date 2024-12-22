package pgconv

import (
	"github.com/jackc/pgx/v5/pgtype"
)

func NullText() pgtype.Text {
	return pgtype.Text{Valid: false}
}

func NullInt4() pgtype.Int4 {
	return pgtype.Int4{
		Int32: 0,
		Valid: false,
	}
}

func NullInterval() pgtype.Interval {
	return pgtype.Interval{
		Microseconds: 0,
		Days:         0,
		Months:       0,
		Valid:        false,
	}
}

func NullTimestamptz() pgtype.Timestamptz {
	return pgtype.Timestamptz{
		InfinityModifier: pgtype.Finite,
		Valid:            false,
	}
}

func NullBool() pgtype.Bool {
	return pgtype.Bool{
		Bool:  false,
		Valid: false,
	}
}

func NullFloat8() pgtype.Float8 {
	return pgtype.Float8{
		Float64: 0.0,
		Valid:   false,
	}
}

func NullUUID() pgtype.UUID {
	return pgtype.UUID{Valid: false}
}
