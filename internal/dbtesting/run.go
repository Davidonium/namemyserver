package dbtesting

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	// import sqlite specific driver for running migrations in integration testing.
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"

	embed "github.com/davidonium/namemyserver"

	"github.com/davidonium/namemyserver/internal/store/sqlitestore"
)

func Run(t *testing.T, f func(*testing.T, *sqlitestore.DB)) {
	t.Helper()
	if testing.Short() {
		t.Skip("database tests are skipped for short testing")
	}

	fd, err := os.CreateTemp(t.TempDir(), "namemyserver-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file for the sqlite database, cause: %v", err)
	}

	dbFile := fd.Name()
	// the sqlite driver will open the file again
	fd.Close()


	dbURL := "sqlite:" + dbFile

	ctx := context.Background()

	db, err := sqlitestore.Connect(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to startup a database for the test: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	u, err := url.Parse(dbURL)
	if err != nil {
		t.Fatalf("failed to parse database url: %v", err)
	}

	dbm := dbmate.New(u)
	dbm.AutoDumpSchema = false
	dbm.FS = embed.MigrationsFS

	if err := dbm.Migrate(); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	f(t, db)
}
