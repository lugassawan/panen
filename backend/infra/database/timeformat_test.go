package database

import (
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/shared"
)

// sqlResultZero is a sql.Result stub returning 0 rows affected.
type sqlResultZero struct{}

func (sqlResultZero) LastInsertId() (int64, error) { return 0, nil }
func (sqlResultZero) RowsAffected() (int64, error) { return 0, nil }

// sqlResultOne is a sql.Result stub returning 1 row affected.
type sqlResultOne struct{}

func (sqlResultOne) LastInsertId() (int64, error) { return 1, nil }
func (sqlResultOne) RowsAffected() (int64, error) { return 1, nil }

func TestFormatParseTimeRoundtrip(t *testing.T) {
	original := time.Date(2025, 6, 15, 14, 30, 45, 0, time.UTC)
	formatted := formatTime(original)
	parsed, err := parseTime(formatted)
	if err != nil {
		t.Fatalf("parseTime(%q) error: %v", formatted, err)
	}
	if !parsed.Equal(original) {
		t.Errorf("roundtrip failed: got %v, want %v", parsed, original)
	}
}

func TestFormatTimeConvertsToUTC(t *testing.T) {
	loc := time.FixedZone("WIB", 7*3600)
	wib := time.Date(2025, 6, 15, 21, 30, 45, 0, loc)
	formatted := formatTime(wib)
	want := "2025-06-15 14:30:45"
	if formatted != want {
		t.Errorf("formatTime() = %q, want %q", formatted, want)
	}
}

func TestParseTimeInvalidInput(t *testing.T) {
	_, err := parseTime("not-a-date")
	if err == nil {
		t.Error("expected error for invalid time string")
	}
}

func TestBoolToInt(t *testing.T) {
	tests := []struct {
		input bool
		want  int
	}{
		{true, 1},
		{false, 0},
	}
	for _, tt := range tests {
		got := boolToInt(tt.input)
		if got != tt.want {
			t.Errorf("boolToInt(%v) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestCheckRowsAffectedZeroRows(t *testing.T) {
	res := sqlResultZero{}
	err := checkRowsAffected(res)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("checkRowsAffected() = %v, want ErrNotFound", err)
	}
}

func TestCheckRowsAffectedOneRow(t *testing.T) {
	res := sqlResultOne{}
	err := checkRowsAffected(res)
	if err != nil {
		t.Errorf("checkRowsAffected() unexpected error: %v", err)
	}
}
