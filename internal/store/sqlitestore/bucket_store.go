package sqlitestore

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/davidonium/namemyserver/internal/namemyserver"
)

type NamedExecContexter interface {
	NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error)
}

type bucketRow struct {
	ID                  int32          `db:"id"`
	Name                string         `db:"name"`
	Description         sql.NullString `db:"description"`
	Cursor              sql.NullInt32  `db:"cursor"`
	ArchivedAt          sql.NullTime   `db:"archived_at"`
	CreatedAt           time.Time      `db:"created_at"`
	UpdatedAt           sql.NullTime   `db:"updated_at"`
	FilterLengthEnabled int            `db:"filter_length_enabled"`
	FilterLengthMode    sql.NullString `db:"filter_length_mode"`
	FilterLengthValue   sql.NullInt32  `db:"filter_length_value"`
}

type BucketStore struct {
	db     *DBPool
	logger *slog.Logger
}

func NewBucketStore(logger *slog.Logger, db *DBPool) *BucketStore {
	return &BucketStore{logger: logger, db: db}
}

const createBucketSQL = `
INSERT INTO buckets
	(name, description, filter_length_enabled, filter_length_mode, filter_length_value)
VALUES
	(:name, :description, :filter_length_enabled, :filter_length_mode, :filter_length_value)`

func (s *BucketStore) Create(ctx context.Context, b *namemyserver.Bucket) error {
	args := map[string]any{
		"name":                  b.Name,
		"description":           b.Description,
		"filter_length_enabled": boolToInt(b.FilterLengthEnabled),
		"filter_length_mode":    nullableString(string(b.FilterLengthMode)),
		"filter_length_value":   nullableInt(b.FilterLengthValue, b.FilterLengthEnabled),
	}
	r, err := s.db.Write().NamedExecContext(ctx, createBucketSQL, args)
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

func (s *BucketStore) FillBucketValues(
	ctx context.Context,
	b namemyserver.Bucket,
	f namemyserver.RandomPairFilters,
) error {
	return s.db.Write().WithTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx *sqlx.Tx) error {
		whereSQL, args := buildPairFilterWhereSQL(f)
		args["bucket_id"] = b.ID
		sql := fmt.Sprintf(fillBucketValuesSQL, whereSQL)
		if _, err := tx.NamedExecContext(ctx, sql, args); err != nil {
			return err
		}

		if err := s.setCursor(ctx, tx, b.ID, 1); err != nil {
			return err
		}

		return nil
	})
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
	return s.setCursor(ctx, s.db.Write(), bucketID, cursor)
}

func (s *BucketStore) setCursor(
	ctx context.Context,
	db NamedExecContexter,
	bucketID int32,
	cursor int32,
) error {
	args := map[string]any{
		"bucket_id": bucketID,
		"cursor":    cursor,
	}
	if _, err := db.NamedExecContext(ctx, setCursorSQL, args); err != nil {
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

	err := s.db.Write().WithTx(
		ctx,
		&sql.TxOptions{},
		func(ctx context.Context, tx *sqlx.Tx) error {
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
		},
	)
	if err != nil {
		return "", err
	}

	return row.Name, nil
}

const oneByNameSQL = `
SELECT
	id,
	name,
	description,
	cursor,
	archived_at,
	created_at,
	updated_at,
	filter_length_enabled,
	filter_length_mode,
	filter_length_value
FROM
	buckets
WHERE
	name = :name`

func (s *BucketStore) OneByName(ctx context.Context, name string) (namemyserver.Bucket, error) {
	stmt, err := s.db.Read().PrepareNamedContext(ctx, oneByNameSQL)
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
SELECT
	id,
	name,
	description,
	cursor,
	archived_at,
	created_at,
	updated_at,
	filter_length_enabled,
	filter_length_mode,
	filter_length_value
FROM
	buckets
WHERE
	id = :id`

func (s *BucketStore) OneByID(ctx context.Context, id int32) (namemyserver.Bucket, error) {
	stmt, err := s.db.Read().PrepareNamedContext(ctx, oneByIDSQL)
	if err != nil {
		return namemyserver.Bucket{}, err
	}

	var row bucketRow
	if err := stmt.GetContext(ctx, &row, map[string]any{"id": id}); err != nil {
		return namemyserver.Bucket{}, err
	}

	return rowToBucket(row), nil
}

const listBucketsSQLTpl = `
SELECT
	id,
	name,
	description,
	cursor,
	archived_at,
	created_at,
	updated_at,
	filter_length_enabled,
	filter_length_mode,
	filter_length_value
FROM
	buckets
WHERE
	%s`

func (s *BucketStore) List(
	ctx context.Context,
	opts namemyserver.ListOptions,
) ([]namemyserver.Bucket, error) {
	wheres := []string{"1=1"}

	if opts.ArchivedOnly {
		wheres = append(wheres, "archived_at IS NOT NULL")
	} else {
		wheres = append(wheres, "archived_at IS NULL")
	}

	var rows []bucketRow
	if err := s.db.Read().SelectContext(ctx, &rows, fmt.Sprintf(listBucketsSQLTpl, strings.Join(wheres, " AND "))); err != nil {
		return nil, err
	}

	buckets := make([]namemyserver.Bucket, 0, len(rows))
	for _, r := range rows {
		buckets = append(buckets, rowToBucket(r))
	}

	return buckets, nil
}

const saveBucketSQL = `
UPDATE
	buckets
SET
	description = :description,
	archived_at = :archived_at,
	updated_at = CURRENT_TIMESTAMP
WHERE
	id = :id`

func (s *BucketStore) Save(ctx context.Context, b *namemyserver.Bucket) error {
	params := map[string]any{
		"id":          b.ID,
		"archived_at": b.ArchivedAt,
		"description": b.Description,
	}
	if _, err := s.db.Write().NamedExecContext(ctx, saveBucketSQL, params); err != nil {
		return err
	}

	return nil
}

const removeBucketValuesFromArchivedSQL = `
DELETE FROM
	bucket_values
WHERE
	bucket_id IN (
		SELECT
			id
		FROM
			buckets
		WHERE
			archived_at < :cutoff
	)`
const removeArchivedBucketsSQL = `DELETE FROM buckets WHERE archived_at < :cutoff`

func (s *BucketStore) RemoveBucketsArchivedForMoreThan(
	ctx context.Context,
	t time.Duration,
) (amount int64, err error) {
	tx, err := s.db.Write().BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				s.logger.Error(
					"failure rolling back after failure",
					slog.Any("err", txErr),
					slog.Any("err.original", err),
				)
			}
			return
		}

		if txErr := tx.Commit(); txErr != nil {
			err = txErr
		}
	}()

	cutoff := time.Now().Add(-t)

	params := map[string]any{
		"cutoff": cutoff,
	}
	if _, err = tx.NamedExecContext(ctx, removeBucketValuesFromArchivedSQL, params); err != nil {
		return
	}

	result, err := tx.NamedExecContext(ctx, removeArchivedBucketsSQL, params)
	if err != nil {
		return
	}

	amount, err = result.RowsAffected()
	if err != nil {
		return
	}

	return
}

const remainingValuesSQL = `
SELECT
	count(*) as count
FROM
	bucket_values
WHERE
	bucket_id = :id
AND
	order_id >= :cursor`

func (s *BucketStore) RemainingValuesTotal(
	ctx context.Context,
	b namemyserver.Bucket,
) (int64, error) {
	stmt, err := s.db.Read().PrepareNamedContext(ctx, remainingValuesSQL)
	if err != nil {
		return 0, err
	}

	var count int64
	if err := stmt.GetContext(ctx, &count, map[string]any{"id": b.ID, "cursor": b.Cursor}); err != nil {
		return 0, err
	}

	return count, nil
}

func rowToBucket(row bucketRow) namemyserver.Bucket {
	return namemyserver.Bucket{
		ID:                  row.ID,
		Name:                row.Name,
		Description:         row.Description.String,
		Cursor:              row.Cursor.Int32,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           sqlTimeToPtr(row.UpdatedAt),
		ArchivedAt:          sqlTimeToPtr(row.ArchivedAt),
		FilterLengthEnabled: row.FilterLengthEnabled == 1,
		FilterLengthMode:    namemyserver.LengthMode(row.FilterLengthMode.String),
		FilterLengthValue:   int(row.FilterLengthValue.Int32),
	}
}
