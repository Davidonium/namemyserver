package sqlitestore

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/davidonium/namemyserver/internal/namemyserver"
)

type BucketStore struct {
	db *sqlx.DB
	imDB *sqlx.DB
}

func NewBucketStore(db *sqlx.DB, imDB *sqlx.DB) *BucketStore {
	return &BucketStore{db: db, imDB: imDB}
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
		"cursor":    cursor,
	}
	if _, err := s.db.NamedExecContext(ctx, setCursorSQL, args); err != nil {
		return err
	}

	return nil
}

const currentBucketNameValueSQL = `
SELECT
	value
FROM
	bucket_values
WHERE
	bucket_id = :bucket_id
AND
	order_id = :cursor`

const advanceCursorSQL = `
UPDATE
	buckets
SET
	cursor = (
		SELECT
			order_id
		FROM
			bucket_values
		WHERE
			bucket_id = :bucket_id
		AND
			order_id > buckets.cursor
		ORDER BY
			order_id ASC
		LIMIT 1
	),
	updated_at = CURRENT_TIMESTAMP
WHERE
	id = :bucket_id`

func (s *BucketStore) PopName(ctx context.Context, b namemyserver.Bucket) (string, error) {
	tx, err := s.imDB.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return "", err
	}

	stmt, err := tx.PrepareNamedContext(ctx, currentBucketNameValueSQL)
	if err != nil {
		return "", fmt.Errorf("failed to prepare query to retrieve cursor name: %w", err)
	}

	args := map[string]any{
		"bucket_id": b.ID,
		"cursor":    b.Cursor,
	}
	var row struct {
		Name string `db:"value"`
	}
	if err := stmt.GetContext(ctx, &row, args); err != nil {
		return "", fmt.Errorf("failed to retrieve name from the cursor: %w", err)
	}

	args = map[string]any{
		"bucket_id": b.ID,
	}
	if _, err := tx.NamedExecContext(ctx, advanceCursorSQL, args); err != nil {
		return "", fmt.Errorf("failed to advance the cursor to the next position: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit advance cursor update: %w", err)
	}

	return row.Name, nil
}

const oneByNameSQL = `
SELECT id, name, cursor
FROM buckets
WHERE name = :name`

func (s *BucketStore) OneByName(ctx context.Context, name string) (namemyserver.Bucket, error) {
	stmt, err := s.db.PrepareNamedContext(ctx, oneByNameSQL)
	if err != nil {
		return namemyserver.Bucket{}, err
	}

	var row struct {
		ID     int    `db:"id"`
		Name   string `db:"name"`
		Cursor int    `db:"cursor"`
	}
	if err := stmt.GetContext(ctx, &row, map[string]any{"name": name}); err != nil {
		return namemyserver.Bucket{}, err
	}

	return namemyserver.Bucket{
		ID:     row.ID,
		Name:   row.Name,
		Cursor: row.Cursor,
	}, nil
}
