package pgutil

import (
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// ToPgText converts a *string to pgtype.Text
func ToPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// FromPgText converts a pgtype.Text to *string
func FromPgText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}

// ToPgInt8 converts a *uint64 to pgtype.Int8
func ToPgInt8(id *uint64) pgtype.Int8 {
	if id == nil || *id > math.MaxInt64 {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{
		Int64: int64(*id),
		Valid: true,
	}
}

// FromPgInt8 converts a pgtype.Int8 to *uint64
func FromPgInt8(i pgtype.Int8) *uint64 {
	if !i.Valid || i.Int64 < 0 {
		return nil
	}
	val := uint64(i.Int64)
	return &val
}

// ToPgTimestamptz converts a *time.Time to pgtype.Timestamptz
func ToPgTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// FromPgTimestamptz converts a pgtype.Timestamptz to *time.Time
func FromPgTimestamptz(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
