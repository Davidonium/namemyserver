package sqlitestore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"runtime"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Connect opens a connection to the sqlite database pointed by url. The url argument must be prefixed by `sqlite:` to
// make the url compatible with both dbmate and the sql driver.
// Returns a DBPool with separate read and write connection pools.
func Connect(ctx context.Context, connStr string) (*DBPool, error) {
	parts := strings.SplitN(connStr, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf(
			"database url '%s' has an unexpected format, expected 'sqlite:<path_to_file>'",
			connStr,
		)
	}

	parsed, err := url.Parse(parts[1])
	if err != nil {
		return nil, fmt.Errorf("could not parse url from '%s': %w", connStr, err)
	}

	writeDB, err := connectWritePool(ctx, parsed)
	if err != nil {
		return nil, err
	}

	readDB, err := connectReadPool(ctx, parsed)
	if err != nil {
		return nil, err
	}

	return NewDBPool(writeDB, readDB), nil
}

// sqlitePragmas returns the common SQLite PRAGMA configuration as URL query values.
func sqlitePragmas() url.Values {
	q := url.Values{}
	q.Set("_journal_mode", "WAL")
	q.Set("_synchronous", "NORMAL")
	q.Set("_busy_timeout", "5000")
	q.Set("_cache_size", "-268435456") // 256MB
	q.Set("_foreign_keys", "ON")
	q.Set("_temp_store", "MEMORY")
	q.Set("_mmap_size", "268435456") // 256MB
	return q
}

func connectWritePool(ctx context.Context, connURL *url.URL) (*sqlx.DB, error) {
	q := sqlitePragmas()
	q.Set("_txlock", "immediate")
	connURL.RawQuery = q.Encode()

	db, err := sqlx.ConnectContext(ctx, "sqlite3", connURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to write database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	return db, nil
}

func connectReadPool(ctx context.Context, connURL *url.URL) (*sqlx.DB, error) {
	connURL.RawQuery = sqlitePragmas().Encode()

	db, err := sqlx.ConnectContext(ctx, "sqlite3", connURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to read database: %w", err)
	}

	db.SetMaxOpenConns(max(4, runtime.NumCPU()))

	return db, nil
}

// DBPool holds separate read and write connection pools for SQLite.
// Callers must explicitly choose Read() or Write() to prevent accidental routing.
type DBPool struct {
	writeDB *DB
	readDB  *DB
}

// NewDBPool creates a new connection pool with separate read and write pools.
func NewDBPool(writeDB *sqlx.DB, readDB *sqlx.DB) *DBPool {
	return &DBPool{
		writeDB: &DB{DB: writeDB},
		readDB:  &DB{DB: readDB},
	}
}

// Write returns a DB wrapper for the write pool.
// Use this for INSERT, UPDATE, DELETE, and other write operations.
func (p *DBPool) Write() *DB {
	return p.writeDB
}

// Read returns a DB wrapper for the read pool.
// Use this for SELECT and other read-only operations.
func (p *DBPool) Read() *DB {
	return p.readDB
}

// Close closes both read and write connection pools.
func (p *DBPool) Close() error {
	return errors.Join(
		p.writeDB.Close(),
		p.readDB.Close(),
	)
}

// DB is a thin wrapper around *sqlx.DB that delegates all methods directly.
// Instances are created by DBPool.Read() or DBPool.Write() to ensure explicit routing.
type DB struct {
	*sqlx.DB
}

// WithTx executes a function within a transaction using the write pool.
// Transactions default to immediate mode for write operations.
func (d *DB) WithTx(
	ctx context.Context,
	opts *sql.TxOptions,
	f func(context.Context, *sqlx.Tx) error,
) (err error) {
	tx, err := d.BeginTxx(ctx, opts)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				err = fmt.Errorf(
					"failed to rollback on tx error: %v - source: %w",
					txErr,
					err,
				)
			}
			return
		}

		if txErr := tx.Commit(); txErr != nil {
			err = fmt.Errorf("failed to commit tx: %w", txErr)
		}
	}()

	err = f(ctx, tx)

	return
}
