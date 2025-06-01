package namemyserver

import (
	"context"
	"time"
)

type BucketStore interface {
	List(ctx context.Context, opts ListOptions) ([]Bucket, error)
	Create(ctx context.Context, b *Bucket) error
	SetCursor(ctx context.Context, bucketID int32, cursor int32) error
	OneByName(ctx context.Context, name string) (Bucket, error)
	OneByID(ctx context.Context, id int32) (Bucket, error)
	FillBucketValues(ctx context.Context, b Bucket, f RandomPairFilters) error
	RemainingValuesTotal(ctx context.Context, b Bucket) (int64, error)
	PopName(ctx context.Context, b Bucket) (string, error)
	Save(ctx context.Context, b *Bucket) error
	RemoveBucketsArchivedForMoreThan(ctx context.Context, t time.Duration) (int64, error)
}

type ListOptions struct {
	ArchivedOnly bool
}
