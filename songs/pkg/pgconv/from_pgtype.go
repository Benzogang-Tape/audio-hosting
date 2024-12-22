package pgconv

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func ptr[T any](v T) *T {
	return &v
}

func FromText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}

	return &t.String
}

func FromInt4(i pgtype.Int4) *int32 {
	if !i.Valid {
		return nil
	}

	return &i.Int32
}

func FromInterval(i pgtype.Interval) *time.Duration {
	if !i.Valid {
		return nil
	}

	return ptr(time.Duration(
		i.Microseconds*time.Microsecond.Nanoseconds() +
			int64(i.Days)*time.Hour.Nanoseconds()*24 +
			int64(i.Months)*time.Hour.Nanoseconds()*24*30))
}

func FromTimestamptz(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}

	return &t.Time
}

func FromBool(b pgtype.Bool) *bool {
	if !b.Valid {
		return nil
	}

	return &b.Bool
}

func FromFloat8(f pgtype.Float8) *float64 {
	if !f.Valid {
		return nil
	}

	return &f.Float64
}

func FromUUID(s pgtype.UUID) *uuid.UUID {
	if !s.Valid {
		return nil
	}

	return ptr(uuid.UUID(s.Bytes))
}
