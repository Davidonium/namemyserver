package sqlitestore

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/jmoiron/sqlx"
)

const perfTuneSQL = `
PRAGMA journal_mode = WAL;
PRAGMA synchronous = normal;
PRAGMA temp_store = memory;
PRAGMA mmap_size = 30000000000;`

// Connect opens a connection to the sqlite database pointed by url. The url argument must be prefixed by `sqlite:` to
// make the url compatible with both dbmate and the sql driver.
func Connect(ctx context.Context, url string) (*sqlx.DB, error) {
	parts := strings.SplitN(url, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("database url '%s' has an unexpected format, expected 'sqlite:<path_to_file>'", url)
	}

	return connectRaw(ctx, parts[1])
}

func ConnectWithImmediate(ctx context.Context, u string) (*sqlx.DB, error) {
	parts := strings.SplitN(u, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("database url '%s' has an unexpected format, expected 'sqlite:<path_to_file>'", u)
	}

	parsed, err := url.Parse(parts[1])
	if err != nil {
		return nil, fmt.Errorf("could not parse url from '%s': %w", u, err)
	}

	q := parsed.Query()
	q.Add("_txlock", "immediate")
	parsed.RawQuery = q.Encode()

	return connectRaw(ctx, parsed.String())
}

func connectRaw(ctx context.Context, url string) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "sqlite3", url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if _, err = db.Exec(perfTuneSQL); err != nil {
		return nil, fmt.Errorf("failed to apply performance configuration to the database: %w", err)
	}

	return db, nil
}
