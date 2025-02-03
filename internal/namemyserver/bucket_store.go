package namemyserver

import (
	"context"
)

type BucketStore interface {
	Create(ctx context.Context, b *Bucket) error
	SetCursor(ctx context.Context, bucketID int32, cursor int32) error
	OneByName(ctx context.Context, name string) (Bucket, error)
	FillBucketValues(ctx context.Context, b Bucket, f RandomPairFilters) error
	PopName(ctx context.Context, b Bucket) (string, error)
}
