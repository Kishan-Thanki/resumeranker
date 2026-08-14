package pgutil

import (
	"math"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestToPgText(t *testing.T) {
	t.Run("nil returns invalid text", func(t *testing.T) {
		got := ToPgText(nil)
		if got.Valid {
			t.Fatal("expected invalid text")
		}
	})

	t.Run("value returns valid text", func(t *testing.T) {
		value := "hello"

		got := ToPgText(&value)

		if !got.Valid {
			t.Fatal("expected valid text")
		}
		if got.String != value {
			t.Fatalf("expected %q, got %q", value, got.String)
		}
	})
}

func TestFromPgText(t *testing.T) {
	t.Run("invalid text returns nil", func(t *testing.T) {
		got := FromPgText(pgtype.Text{Valid: false})
		if got != nil {
			t.Fatal("expected nil")
		}
	})

	t.Run("valid text returns value", func(t *testing.T) {
		got := FromPgText(pgtype.Text{
			String: "hello",
			Valid:  true,
		})

		if got == nil {
			t.Fatal("expected non-nil value")
		}
		if *got != "hello" {
			t.Fatalf("expected %q, got %q", "hello", *got)
		}
	})
}

func TestToPgInt8(t *testing.T) {
	t.Run("nil returns invalid int8", func(t *testing.T) {
		got := ToPgInt8(nil)
		if got.Valid {
			t.Fatal("expected invalid int8")
		}
	})

	t.Run("valid value returns valid int8", func(t *testing.T) {
		value := uint64(12345)

		got := ToPgInt8(&value)

		if !got.Valid {
			t.Fatal("expected valid int8")
		}
		if got.Int64 != int64(value) {
			t.Fatalf("expected %d, got %d", value, got.Int64)
		}
	})

	t.Run("max int64 returns valid int8", func(t *testing.T) {
		value := uint64(math.MaxInt64)

		got := ToPgInt8(&value)

		if !got.Valid {
			t.Fatal("expected valid int8")
		}
		if got.Int64 != math.MaxInt64 {
			t.Fatalf("expected %d, got %d", math.MaxInt64, got.Int64)
		}
	})

	t.Run("value above max int64 returns invalid int8", func(t *testing.T) {
		value := uint64(math.MaxInt64) + 1

		got := ToPgInt8(&value)

		if got.Valid {
			t.Fatal("expected invalid int8")
		}
	})
}

func TestFromPgInt8(t *testing.T) {
	t.Run("invalid int8 returns nil", func(t *testing.T) {
		got := FromPgInt8(pgtype.Int8{Valid: false})
		if got != nil {
			t.Fatal("expected nil")
		}
	})

	t.Run("negative int8 returns nil", func(t *testing.T) {
		got := FromPgInt8(pgtype.Int8{
			Int64: -1,
			Valid: true,
		})

		if got != nil {
			t.Fatal("expected nil")
		}
	})

	t.Run("zero returns zero", func(t *testing.T) {
		got := FromPgInt8(pgtype.Int8{
			Int64: 0,
			Valid: true,
		})

		if got == nil {
			t.Fatal("expected non-nil value")
		}
		if *got != 0 {
			t.Fatalf("expected 0, got %d", *got)
		}
	})

	t.Run("valid positive value returns uint64", func(t *testing.T) {
		got := FromPgInt8(pgtype.Int8{
			Int64: 12345,
			Valid: true,
		})

		if got == nil {
			t.Fatal("expected non-nil value")
		}
		if *got != 12345 {
			t.Fatalf("expected 12345, got %d", *got)
		}
	})

	t.Run("max int64 returns corresponding uint64", func(t *testing.T) {
		got := FromPgInt8(pgtype.Int8{
			Int64: math.MaxInt64,
			Valid: true,
		})

		if got == nil {
			t.Fatal("expected non-nil value")
		}
		if *got != uint64(math.MaxInt64) {
			t.Fatalf("expected %d, got %d", uint64(math.MaxInt64), *got)
		}
	})
}

func TestToPgTimestamptz(t *testing.T) {
	t.Run("nil returns invalid timestamp", func(t *testing.T) {
		got := ToPgTimestamptz(nil)
		if got.Valid {
			t.Fatal("expected invalid timestamp")
		}
	})

	t.Run("value returns valid timestamp", func(t *testing.T) {
		value := time.Date(2026, 8, 14, 16, 30, 0, 0, time.UTC)

		got := ToPgTimestamptz(&value)

		if !got.Valid {
			t.Fatal("expected valid timestamp")
		}
		if !got.Time.Equal(value) {
			t.Fatalf("expected %v, got %v", value, got.Time)
		}
	})
}

func TestFromPgTimestamptz(t *testing.T) {
	t.Run("invalid timestamp returns nil", func(t *testing.T) {
		got := FromPgTimestamptz(pgtype.Timestamptz{Valid: false})
		if got != nil {
			t.Fatal("expected nil")
		}
	})

	t.Run("valid timestamp returns value", func(t *testing.T) {
		value := time.Date(2026, 8, 14, 16, 30, 0, 0, time.UTC)

		got := FromPgTimestamptz(pgtype.Timestamptz{
			Time:  value,
			Valid: true,
		})

		if got == nil {
			t.Fatal("expected non-nil value")
		}
		if !got.Equal(value) {
			t.Fatalf("expected %v, got %v", value, *got)
		}
	})
}
