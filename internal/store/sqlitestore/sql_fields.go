package sqlitestore

import (
	"database/sql"
	"time"
)

func sqlTimeToPtr(t sql.NullTime) *time.Time {
	if t.Valid {
		return &t.Time
	}

	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s == ""}
}

func nullableInt(val int, enabled bool) sql.NullInt32 {
	if !enabled {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{Int32: int32(val), Valid: true}
}
