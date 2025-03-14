package namemyserver

import (
	"context"
)

type BucketStore interface {
	All(ctx context.Context) ([]Bucket, error)
	Create(ctx context.Context, b *Bucket) error
	SetCursor(ctx context.Context, bucketID int32, cursor int32) error
	OneByName(ctx context.Context, name string) (Bucket, error)
	OneByID(ctx context.Context, id int32) (Bucket, error)
	FillBucketValues(ctx context.Context, b Bucket, f RandomPairFilters) error
	PopName(ctx context.Context, b Bucket) (string, error)
	Archive(ctx context.Context, b *Bucket) error
}
