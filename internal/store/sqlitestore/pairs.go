package sqlitestore

import (
	"context"

	"github.com/jmoiron/sqlx"
)
const singlePairSQL = `
SELECT
    LOWER(adjectives.value || '-' || nouns.value) AS pair
FROM
    adjectives
CROSS JOIN
    nouns
ORDER BY
    RANDOM()
LIMIT 1`

type PairStore struct {
	db *sqlx.DB
}

func NewPairStore(db *sqlx.DB) *PairStore {
	return &PairStore{db: db}
}

func (s *PairStore) FindSinglePair(ctx context.Context) (string, error) {
	var row struct {
		Pair string `db:"pair"`
	}
	if err := s.db.GetContext(ctx, &row, singlePairSQL); err != nil {
		return "", err
	}

	return row.Pair, nil
}
