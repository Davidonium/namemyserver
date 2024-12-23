package sqlitestore

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

const perfTuneSQL = `
pragma journal_mode = WAL;
pragma synchronous = normal;
pragma temp_store = memory;
pragma mmap_size = 30000000000;`

func Connect(ctx context.Context, url string) (*sqlx.DB, error) {
	parts := strings.SplitN(url, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("database url '%s' has an unexpected format, expected 'sqlite:<path_to_file>'", url)
	}

	db, err := sqlx.ConnectContext(ctx, "sqlite3", parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if _, err = db.Exec(perfTuneSQL); err != nil {
		return nil, fmt.Errorf("failed to apply performance configuration to the database: %w", err)
	}

	return db, nil
}
