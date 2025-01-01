package sqlitestore

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/davidonium/namemyserver/internal/namemyserver"
)

const singlePairSQL = `
SELECT
    adjectives.value as adjective,
	nouns.value as noun
FROM
    adjectives
JOIN
    nouns
WHERE
    adjectives.id >= (SELECT ABS(RANDOM()) % (SELECT COUNT(*) FROM adjectives) + 1)
AND
    nouns.id >= (SELECT ABS(RANDOM()) % (SELECT COUNT(*) FROM nouns) + 1)
LIMIT 1`

type PairStore struct {
	db *sqlx.DB
}

func NewPairStore(db *sqlx.DB) *PairStore {
	return &PairStore{db: db}
}

func (s *PairStore) GetSinglePair(ctx context.Context) (namemyserver.Pair, error) {
	var row struct {
		Adjective string `db:"adjective"`
		Noun      string `db:"noun"`
	}
	if err := s.db.GetContext(ctx, &row, singlePairSQL); err != nil {
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
		NounCount int `db:"noun_count"`
	}
	if err := s.db.GetContext(ctx, &row, statsSQL); err != nil {
		return namemyserver.Stats{}, err
	}
	return namemyserver.Stats{
		PairCount: row.PairCount,
		AdjectiveCount: row.AdjectiveCount,
		NounCount: row.NounCount,
	}, nil
}
