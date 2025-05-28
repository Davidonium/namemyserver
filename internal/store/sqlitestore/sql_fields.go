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
