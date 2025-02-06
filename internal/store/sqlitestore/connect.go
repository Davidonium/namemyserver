package sqlitestore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Connect opens a connection to the sqlite database pointed by url. The url argument must be prefixed by `sqlite:` to
// make the url compatible with both dbmate and the sql driver.
func Connect(ctx context.Context, connStr string) (*DB, error) {
	parts := strings.SplitN(connStr, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("database url '%s' has an unexpected format, expected 'sqlite:<path_to_file>'", connStr)
	}

	parsed, err := url.Parse(parts[1])
	if err != nil {
		return nil, fmt.Errorf("could not parse url from '%s': %w", connStr, err)
	}

	db, err := connectURL(ctx, parsed)
	if err != nil {
		return nil, err
	}

	imDB, err := connectWithImmediate(ctx, parsed)
	if err != nil {
		return nil, err
	}

	return NewDB(db, imDB), nil
}

func connectWithImmediate(ctx context.Context, connURL *url.URL) (*sqlx.DB, error) {
	q := connURL.Query()
	q.Add("_txlock", "immediate")
	connURL.RawQuery = q.Encode()

	return connectURL(ctx, connURL)
}

func connectURL(ctx context.Context, connURL *url.URL) (*sqlx.DB, error) {
	q := connURL.Query()
	q.Set("_journal_mode", "WAL")
	q.Set("_syncronous", "normal")
	q.Set("_temp_store", "memory")
	q.Set("_mmap_size", "30000000000")
	connURL.RawQuery = q.Encode()

	db, err := sqlx.ConnectContext(ctx, "sqlite3", connURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

type DB struct {
	*sqlx.DB
	imDB *sqlx.DB
}

func NewDB(db *sqlx.DB, imDB *sqlx.DB) *DB {
	return &DB{
		DB:   db,
		imDB: imDB,
	}
}

func (db *DB) WithImmediateTx(ctx context.Context, opts *sql.TxOptions, f func(context.Context, *sqlx.Tx) error) (err error) {
	tx, err := db.imDB.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			err = tx.Rollback()
			return
		}

		err = tx.Commit()
	}()

	err = f(ctx, tx)

	return
}

func (db *DB) Close() error {
	return errors.Join(
		db.DB.Close(),
		db.imDB.Close(),
	)
}
