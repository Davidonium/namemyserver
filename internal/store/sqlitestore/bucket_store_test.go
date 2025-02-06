package sqlitestore_test

import (
	"context"
	"testing"

	"github.com/davidonium/namemyserver/internal/dbtesting"
	"github.com/davidonium/namemyserver/internal/namemyserver"
	"github.com/davidonium/namemyserver/internal/store/sqlitestore"
)

func TestBucketStore(t *testing.T) {
	dbtesting.Run(t, func(t *testing.T, db *sqlitestore.DB) {
		ctx := context.Background()
		store := sqlitestore.NewBucketStore(db)

		b := &namemyserver.Bucket{
			Name: "test-bucket",
		}

		if err := store.Create(ctx, b); err != nil {
			t.Errorf("Create() = expected to succeed but got err: %v", err)
		}

		if err := store.Create(ctx, b); err == nil {
			t.Error("Create() = expected to fail when creating a bucket with an already existing name but succeeded")
		}

		bk, err := store.OneByName(ctx, "test-bucket")
		if err != nil {
			t.Errorf("OneByName() = expected to retrieve an existing bucket but failed: %v", err)
		}

		if bk.Name != "test-bucket" {
			t.Errorf("OneByName() = unexpected bucket name retrieved. got '%s' want '%s'", bk.Name, "test-bucket")
		}
	})
}
