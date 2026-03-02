package database

import (
	"database/sql"
	"time"

	"github.com/lugassawan/panen/backend/internal/domain/shared"
)

const timeLayout = "2006-01-02 15:04:05"

func formatTime(t time.Time) string {
	return t.UTC().Format(timeLayout)
}

func parseTime(s string) (time.Time, error) {
	return time.Parse(timeLayout, s)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func checkRowsAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return shared.ErrNotFound
	}
	return nil
}
