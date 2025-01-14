package sqlitestore

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/davidonium/namemyserver/internal/namemyserver"
)

const singlePairSQL = `
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
%s
LIMIT 1`

type PairStore struct {
	db *sqlx.DB
}

func NewPairStore(db *sqlx.DB) *PairStore {
	return &PairStore{db: db}
}

func (s *PairStore) OneRandom(ctx context.Context, f namemyserver.RandomPairFilters) (namemyserver.Pair, error) {
	wheres := []string{"1=1"}
	params := struct {
		Length int `db:"length"`
	}{}

	if f.Length > 0 {
		params.Length = f.Length
		switch f.LengthMode {
		case namemyserver.LengthModeExactly:
			wheres = append(wheres, "(LENGTH(a.value) + LENGTH(n.value) + 1) = :length")
		case namemyserver.LengthModeUpto:
			wheres = append(wheres, "(LENGTH(a.value) + LENGTH(n.value) + 1) <= :length")
		}
	}

	sql := fmt.Sprintf(singlePairSQL, "AND "+strings.Join(wheres, " AND "))

	stmt, err := s.db.PrepareNamedContext(ctx, sql)
	if err != nil {
		return namemyserver.Pair{}, err
	}

	var row struct {
		Adjective string `db:"adjective"`
		Noun      string `db:"noun"`
	}
	if err := stmt.GetContext(ctx, &row, params); err != nil {
		return namemyserver.Pair{}, err
	}

	return namemyserver.Pair{
		Adjective: row.Adjective,
		Noun:      row.Noun,
	}, nil
}

const statsSQL = `
SELECT
    (SELECT count(*) FROM nouns) AS noun_count,
    (SELECT count(*) FROM adjectives) AS adjective_count,
    (SELECT count(*) FROM nouns CROSS JOIN adjectives) AS pair_count`

func (s *PairStore) Stats(ctx context.Context) (namemyserver.Stats, error) {
	var row struct {
		PairCount      int `db:"pair_count"`
		AdjectiveCount int `db:"adjective_count"`
		NounCount      int `db:"noun_count"`
	}
	if err := s.db.GetContext(ctx, &row, statsSQL); err != nil {
		return namemyserver.Stats{}, err
	}
	return namemyserver.Stats{
		PairCount:      row.PairCount,
		AdjectiveCount: row.AdjectiveCount,
		NounCount:      row.NounCount,
	}, nil
}
