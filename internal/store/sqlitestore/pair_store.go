package sqlitestore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/davidonium/serverplate/internal/serverplate"
)

type PairStore struct {
	db *DBPool
}

func NewPairStore(db *DBPool) *PairStore {
	return &PairStore{db: db}
}

const singlePairSQLTpl = `
SELECT
    a.value as adjective,
	n.value as noun
FROM
    adjectives a
JOIN
    nouns n
WHERE
    a.id >= (SELECT (ABS(RANDOM()) %% (MAX(id) - MIN(id) + 1)) + MIN(id) FROM adjectives)
AND
    n.id >= (SELECT (ABS(RANDOM()) %% (MAX(id) - MIN(id) + 1)) + MIN(id) FROM nouns)
AND
	%s
LIMIT 1`

func (s *PairStore) OneRandom(
	ctx context.Context,
	f serverplate.RandomPairFilters,
) (serverplate.Pair, error) {
	whereSQL, args := buildPairFilterWhereSQL(f)
	query := fmt.Sprintf(singlePairSQLTpl, whereSQL)

	stmt, err := s.db.Read().PrepareNamedContext(ctx, query)
	if err != nil {
		return serverplate.Pair{}, err
	}

	var row struct {
		Adjective string `db:"adjective"`
		Noun      string `db:"noun"`
	}
	if err := stmt.GetContext(ctx, &row, args); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return serverplate.Pair{}, serverplate.ErrNoMatchingPairs
		}
		return serverplate.Pair{}, err
	}

	return serverplate.Pair{
		Adjective: row.Adjective,
		Noun:      row.Noun,
	}, nil
}

const statsSQLTpl = `
SELECT
    (SELECT count(*) FROM nouns) AS noun_count,
    (SELECT count(*) FROM adjectives) AS adjective_count,
    (SELECT count(*) FROM adjectives a CROSS JOIN nouns n WHERE %s) AS pair_count`

const dbSizeSQL = `
SELECT page_count * page_size as size
FROM pragma_page_count(), pragma_page_size()`

func (s *PairStore) Stats(
	ctx context.Context,
	f serverplate.RandomPairFilters,
) (serverplate.Stats, error) {
	whereSQL, args := buildPairFilterWhereSQL(f)
	sql := fmt.Sprintf(statsSQLTpl, whereSQL)
	var row struct {
		PairCount      int `db:"pair_count"`
		AdjectiveCount int `db:"adjective_count"`
		NounCount      int `db:"noun_count"`
	}
	stmt, err := s.db.Read().PrepareNamedContext(ctx, sql)
	if err != nil {
		return serverplate.Stats{}, err
	}

	if err := stmt.GetContext(ctx, &row, args); err != nil {
		return serverplate.Stats{}, err
	}

	// Query database size
	var dbSizeRow struct {
		Size int64 `db:"size"`
	}
	if err := s.db.Read().QueryRowContext(ctx, dbSizeSQL).Scan(&dbSizeRow.Size); err != nil {
		return serverplate.Stats{}, err
	}

	return serverplate.Stats{
		DatabaseSizeBytes: dbSizeRow.Size,
		PairCount:         row.PairCount,
		AdjectiveCount:    row.AdjectiveCount,
		NounCount:         row.NounCount,
	}, nil
}

// buildPairFilterWhereSQL returns the sql based on the serverplate.RandomPairFilters. Assumes the query using the
// resulting sql sets up aliases 'a' for adjectives table and 'n' for nouns table, these should potentially
// be passed as function arguments instead of making this assumption but it works for now.
func buildPairFilterWhereSQL(f serverplate.RandomPairFilters) (string, map[string]any) {
	wheres := []string{"1=1"}
	args := map[string]any{}

	if f.Length > 0 {
		args["length"] = f.Length
		switch f.LengthMode {
		case serverplate.LengthModeExactly:
			wheres = append(wheres, "(LENGTH(a.value) + LENGTH(n.value) + 1) = :length")
		case serverplate.LengthModeUpto:
			wheres = append(wheres, "(LENGTH(a.value) + LENGTH(n.value) + 1) <= :length")
		}
	}

	return strings.Join(wheres, " AND "), args
}
