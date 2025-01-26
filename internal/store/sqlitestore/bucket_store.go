package sqlitestore

import (
	"context"
	"fmt"

	"github.com/davidonium/namemyserver/internal/namemyserver"
	"github.com/jmoiron/sqlx"
)

type BucketStore struct {
	db *sqlx.DB
}

func NewBucketStore(db *sqlx.DB) *BucketStore {
	return &BucketStore{db: db}
}

const createBucketSQL = `
INSERT INTO buckets
	(name)
VALUES
	(:name)`

func (s *BucketStore) Create(ctx context.Context, b *namemyserver.Bucket) error {
	args := map[string]any{"name": b.Name}
	r, err := s.db.NamedExecContext(ctx, createBucketSQL, args)
	if err != nil {
		return err
	}

	id, err := r.LastInsertId()
	if err != nil {
		return err
	}

	b.ID = int(id)


	return nil
}

const fillBucketValuesSQL = `
INSERT INTO bucket_values
	(bucket_id, value, order_id)
SELECT
	:bucket_id AS bucket_id,
	a.value || '-' || n.value AS value,
	ROW_NUMBER() OVER (ORDER BY RANDOM()) AS order_id
FROM
    adjectives a
JOIN
    nouns n
WHERE
	%s`
func (s *BucketStore) FillBucketValues(ctx context.Context, b namemyserver.Bucket, f namemyserver.RandomPairFilters) error {
	whereSQL, args := buildPairFilterWhereSQL(f)
	// TODO maybe the two operations should be done in a transaction
	args["bucket_id"] = b.ID
	sql := fmt.Sprintf(fillBucketValuesSQL, whereSQL)
	if _, err := s.db.NamedExecContext(ctx, sql, args); err != nil {
		return err
	}

	if err := s.SetCursor(ctx, b.ID, 1); err != nil {
		return err
	}

	return nil
}

const setCursorSQL = `
UPDATE
	buckets
SET
	cursor = :cursor,
	updated_at = CURRENT_TIMESTAMP
WHERE
	id = :bucket_id`
func (s *BucketStore) SetCursor(ctx context.Context, bucketID int, cursor int) error {
	args := map[string]any{
		"bucket_id": bucketID,
		"cursor": cursor,
	}
	if _, err := s.db.NamedExecContext(ctx, setCursorSQL, args); err != nil {
		return err
	}

	return nil

}
