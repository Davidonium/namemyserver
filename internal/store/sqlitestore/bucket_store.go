package sqlitestore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/davidonium/namemyserver/internal/namemyserver"
	"github.com/jmoiron/sqlx"
)

type bucketRow struct {
	ID          int32          `db:"id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	Cursor      sql.NullInt32  `db:"cursor"`
	ArchivedAt  sql.NullTime   `db:"archived_at"`
}

type BucketStore struct {
	db *DB
}

func NewBucketStore(db *DB) *BucketStore {
	return &BucketStore{db: db}
}

const createBucketSQL = `
INSERT INTO buckets
	(name, description)
VALUES
	(:name, :description)`

func (s *BucketStore) Create(ctx context.Context, b *namemyserver.Bucket) error {
	args := map[string]any{
		"name":        b.Name,
		"description": b.Description,
	}
	r, err := s.db.NamedExecContext(ctx, createBucketSQL, args)
	if err != nil {
		return err
	}

	id, err := r.LastInsertId()
	if err != nil {
		return err
	}

	b.ID = int32(id)
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

func (s *BucketStore) SetCursor(ctx context.Context, bucketID int32, cursor int32) error {
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
	var row struct {
		Name string `db:"value"`
	}

	err := s.db.WithImmediateTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx *sqlx.Tx) error {
		stmt, err := tx.PrepareNamedContext(ctx, currentBucketNameValueSQL)
		if err != nil {
			return fmt.Errorf("failed to prepare query to retrieve cursor name: %w", err)
		}

		args := map[string]any{
			"bucket_id": b.ID,
			"cursor":    b.Cursor,
		}
		if err := stmt.GetContext(ctx, &row, args); err != nil {
			return fmt.Errorf("failed to retrieve name from the cursor: %w", err)
		}

		args = map[string]any{
			"bucket_id": b.ID,
		}
		if _, err := tx.NamedExecContext(ctx, advanceCursorSQL, args); err != nil {
			return fmt.Errorf("failed to advance the cursor to the next position: %w", err)
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return row.Name, nil
}

const oneByNameSQL = `
SELECT id, name, description, cursor, archived_at
FROM buckets
WHERE name = :name`

func (s *BucketStore) OneByName(ctx context.Context, name string) (namemyserver.Bucket, error) {
	stmt, err := s.db.PrepareNamedContext(ctx, oneByNameSQL)
	if err != nil {
		return namemyserver.Bucket{}, err
	}

	var row bucketRow
	if err := stmt.GetContext(ctx, &row, map[string]any{"name": name}); err != nil {
		return namemyserver.Bucket{}, err
	}

	return rowToBucket(row), nil
}

const oneByIDSQL = `
SELECT id, name, description, cursor, archived_at
FROM buckets
WHERE id = :id`

func (s *BucketStore) OneByID(ctx context.Context, id int32) (namemyserver.Bucket, error) {
	stmt, err := s.db.PrepareNamedContext(ctx, oneByIDSQL)
	if err != nil {
		return namemyserver.Bucket{}, err
	}

	var row bucketRow
	if err := stmt.GetContext(ctx, &row, map[string]any{"id": id}); err != nil {
		return namemyserver.Bucket{}, err
	}

	return rowToBucket(row), nil
}

const allBucketsSQL = `
SELECT id, name, description, cursor, archived_at
FROM buckets`

func (s *BucketStore) All(ctx context.Context) ([]namemyserver.Bucket, error) {
	var rows []bucketRow
	if err := s.db.SelectContext(ctx, &rows, allBucketsSQL); err != nil {
		return nil, err
	}

	var buckets []namemyserver.Bucket
	for _, r := range rows {
		buckets = append(buckets, rowToBucket(r))
	}

	return buckets, nil
}

const archiveBucketSQL = `
UPDATE
	buckets
SET
	archived_at = :archived_at,
	updated_at = CURRENT_TIMESTAMP
WHERE
	id = :id`

func (s *BucketStore) Archive(ctx context.Context, b *namemyserver.Bucket) error {
	b.ArchivedAt = time.Now()
	params := map[string]any{
		"archived_at": b.ArchivedAt,
		"id":          b.ID,
	}
	if _, err := s.db.NamedExecContext(ctx, archiveBucketSQL, params); err != nil {
		return err
	}

	return nil
}

func rowToBucket(row bucketRow) namemyserver.Bucket {
	return namemyserver.Bucket{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description.String,
		Cursor:      row.Cursor.Int32,
		ArchivedAt:  row.ArchivedAt.Time,
	}
}
