package sqlitestore

import (
	"context"
	"fmt"
	"strings"

	"github.com/davidonium/namemyserver/internal/namemyserver"
)

type PairStore struct {
	db *DB
}

func NewPairStore(db *DB) *PairStore {
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

func (s *PairStore) OneRandom(ctx context.Context, f namemyserver.RandomPairFilters) (namemyserver.Pair, error) {
	whereSQL, args := buildPairFilterWhereSQL(f)
	sql := fmt.Sprintf(singlePairSQLTpl, whereSQL)

	stmt, err := s.db.PrepareNamedContext(ctx, sql)
	if err != nil {
		return namemyserver.Pair{}, err
	}

	var row struct {
		Adjective string `db:"adjective"`
		Noun      string `db:"noun"`
	}
	if err := stmt.GetContext(ctx, &row, args); err != nil {
		return namemyserver.Pair{}, err
	}

	return namemyserver.Pair{
		Adjective: row.Adjective,
		Noun:      row.Noun,
	}, nil
}

const statsSQLTpl = `
SELECT
    (SELECT count(*) FROM nouns) AS noun_count,
    (SELECT count(*) FROM adjectives) AS adjective_count,
    (SELECT count(*) FROM adjectives a CROSS JOIN nouns n WHERE %s) AS pair_count`

func (s *PairStore) Stats(ctx context.Context, f namemyserver.RandomPairFilters) (namemyserver.Stats, error) {
	whereSQL, args := buildPairFilterWhereSQL(f)
	sql := fmt.Sprintf(statsSQLTpl, whereSQL)
	var row struct {
		PairCount      int `db:"pair_count"`
		AdjectiveCount int `db:"adjective_count"`
		NounCount      int `db:"noun_count"`
	}
	stmt, err := s.db.PrepareNamedContext(ctx, sql)
	if err != nil {
		return namemyserver.Stats{}, err
	}

	if err := stmt.GetContext(ctx, &row, args); err != nil {
		return namemyserver.Stats{}, err
	}

	return namemyserver.Stats{
		PairCount:      row.PairCount,
		AdjectiveCount: row.AdjectiveCount,
		NounCount:      row.NounCount,
	}, nil
}

func buildPairFilterWhereSQL(f namemyserver.RandomPairFilters) (string, map[string]any) {
	wheres := []string{"1=1"}
	args := map[string]any{}

	if f.Length > 0 {
		args["length"] = f.Length
		// TODO assumes the query using the resulting sql sets up aliases 'a' for adjectives table
		// and 'n' for nouns table, these should potentially be passed as function arguments
		// instead of making this assumption.
		switch f.LengthMode {
		case namemyserver.LengthModeExactly:
			wheres = append(wheres, "(LENGTH(a.value) + LENGTH(n.value) + 1) = :length")
		case namemyserver.LengthModeUpto:
			wheres = append(wheres, "(LENGTH(a.value) + LENGTH(n.value) + 1) <= :length")
		}
	}

	return strings.Join(wheres, " AND "), args
}
