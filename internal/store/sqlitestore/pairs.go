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

func (s *PairStore) FindSinglePair(ctx context.Context) (namemyserver.Pair, error) {
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
