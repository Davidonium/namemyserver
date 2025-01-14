package sqlitestore

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/doug-martin/goqu/v9"
)

type DBExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type ChunkInserter struct {
	logger *slog.Logger
	db     DBExecer
	size   int
	table  string
	idx    int
	chunk  []any
	Err    error
}

func NewChunkInserter(
	logger *slog.Logger,
	db DBExecer,
	size int,
	table string,
) *ChunkInserter {
	return &ChunkInserter{
		logger: logger,
		db:     db,
		size:   size,
		table:  table,
	}
}

func (ci *ChunkInserter) AddAndFlushIfNeeded(ctx context.Context, record any) {
	if ci.Err != nil {
		return
	}

	ci.chunk = append(ci.chunk, record)
	if len(ci.chunk) >= ci.size {
		if err := ci.Flush(ctx); err != nil {
			ci.Err = err
			return
		}
	}
}

func (ci *ChunkInserter) Flush(ctx context.Context) error {
	if len(ci.chunk) == 0 {
		ci.logger.Info("empty chunk, nothing to flush")
		return nil
	}

	ds := goqu.Insert(ci.table).Rows(ci.chunk...).OnConflict(goqu.DoNothing())
	sql, args, _ := ds.ToSQL()

	ci.logger.Debug("executing sql", "sql", sql, "args", args)

	if _, err := ci.db.ExecContext(ctx, sql, args...); err != nil {
		return fmt.Errorf("failed to seed database at chunk idx %d: %w", ci.idx, err)
	}

	ci.idx += ci.size
	ci.chunk = ci.chunk[:0]
	return nil
}
