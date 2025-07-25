package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
	envcfg "github.com/caarlos0/env/v11"
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"

	embed "github.com/davidonium/namemyserver"
	"github.com/davidonium/namemyserver/internal/bg"
	"github.com/davidonium/namemyserver/internal/env"
	"github.com/davidonium/namemyserver/internal/namemyserver"
	"github.com/davidonium/namemyserver/internal/server"
	"github.com/davidonium/namemyserver/internal/store/sqlitestore"
	"github.com/davidonium/namemyserver/internal/vite"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "could not start namemyserver app.\nerror: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	var cfg env.Config
	if err := envcfg.Parse(&cfg); err != nil {
		return fmt.Errorf("failed to parse environment variables into a config struct: %w", err)
	}
	level := cfg.LogLevel

	var l slog.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		return fmt.Errorf("failed to parse log level '%s': %w", level, err)
	}

	var logger *slog.Logger

	if cfg.LogFormat == "json" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: l}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: l}))
	}

	if len(args) < 2 {
		return errors.New("a command needs to be specified to run the app")
	}

	switch args[1] {
	case "server":
		return runServer(logger, cfg)
	case "seed":
		return runSeed(logger, cfg)
	}

	return fmt.Errorf("unknown command %q", args[1])
}

func runServer(logger *slog.Logger, cfg env.Config) error {
	ctx := context.Background()

	db, err := sqlitestore.Connect(ctx, cfg.DatabaseURL.String())
	if err != nil {
		return err
	}

	defer db.Close()

	dbm := dbmate.New(cfg.DatabaseURL)
	dbm.AutoDumpSchema = false
	dbm.FS = embed.MigrationsFS

	logger.Info("applying migrations...")
	if err := dbm.Migrate(); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	assets := vite.NewAssets(vite.AssetsConfig{
		RootURL:     cfg.AssetsRootURL.String(),
		UseManifest: cfg.AssetsUseManifest,
	})

	if cfg.AssetsUseManifest {
		logger.Info("assets manifest is enabled, loading manifest from embed fs")
		if err := assets.LoadManifestFromFS(embed.FrontendFS, cfg.AssetsManifestLocation); err != nil {
			return fmt.Errorf(
				"failed to load assets from fs at %s: %w",
				cfg.AssetsManifestLocation,
				err,
			)
		}
	}

	pairStore := sqlitestore.NewPairStore(db)
	bucketStore := sqlitestore.NewBucketStore(logger, db)

	generator := namemyserver.NewGenerator(pairStore)

	runner := bg.NewRunner(logger, bucketStore)
	runner.Start()

	s := server.New(&server.Services{
		Logger:      logger.With("service", "server"),
		Config:      cfg,
		Assets:      assets,
		Generator:   generator,
		PairStore:   pairStore,
		BucketStore: bucketStore,
	})

	logger.Info("starting http server", "addr", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start the http server: %w", err)
	}

	return nil
}

func runSeed(logger *slog.Logger, cfg env.Config) error {
	ctx := context.Background()

	db, err := sqlitestore.Connect(ctx, cfg.DatabaseURL.String())
	if err != nil {
		return err
	}

	defer db.Close()

	logger.Info("running seed")
	tables := []string{
		"nouns",
		"adjectives",
	}

	for _, t := range tables {
		logger.Info("seeding table", "table", t)
		if err := seedByTable(ctx, logger, db.DB, t); err != nil {
			logger.Error("failure running seed", "error", err, "table", t)
		}
	}

	logger.Info("seed finished")

	return nil
}

func seedByTable(ctx context.Context, logger *slog.Logger, db *sqlx.DB, table string) error {
	f, err := os.Open(fmt.Sprintf("./db/seed/%s.txt", table))
	if err != nil {
		return err
	}

	defer f.Close()

	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	tempTable := "temporary_" + table

	const tempTableSQLTempl = `CREATE TEMP TABLE %s (value TEXT NOT NULL UNIQUE)`
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(tempTableSQLTempl, tempTable)); err != nil {
		return fmt.Errorf("failed to create temporary table for %s: %w", table, err)
	}

	err = func() error {
		s := bufio.NewScanner(f)

		inserter := sqlitestore.NewChunkInserter(logger, tx, 1000, tempTable)
		for s.Scan() {
			inserter.AddAndFlushIfNeeded(ctx, goqu.Record{"value": s.Text()})
		}

		if inserter.Err != nil {
			return inserter.Err
		}

		// flush remaining chunk
		if err := inserter.Flush(ctx); err != nil {
			return err
		}

		const insTemplSQL = `INSERT INTO %s (value, from_seed)
							 SELECT value, 1 AS from_seed FROM %s WHERE 1
							 ON CONFLICT(value) DO NOTHING`
		insSQL := fmt.Sprintf(insTemplSQL, table, tempTable)
		insResult, err := tx.ExecContext(ctx, insSQL)
		if err != nil {
			return fmt.Errorf(
				"failed to move the values from temporary table to the real table %q: %w",
				table,
				err,
			)
		}

		nins, err := insResult.RowsAffected()
		if err != nil {
			return err
		}

		if nins > 0 {
			logger.Info("seed insertions", slog.String("table", table), slog.Int64("amount", nins))
		}

		const delTemplSQL = `DELETE FROM %s
				             WHERE from_seed = 1 AND value NOT IN (
						 		SELECT value
								FROM %s
						 	 )`
		delSQL := fmt.Sprintf(delTemplSQL, table, tempTable)
		delResult, err := tx.ExecContext(ctx, delSQL)
		if err != nil {
			return fmt.Errorf(
				"failed to remove values that ceased to exist in the seed from table %q: %w",
				table,
				err,
			)
		}
		ndel, err := delResult.RowsAffected()
		if err != nil {
			return err
		}

		if ndel > 0 {
			logger.Info("seed removals", slog.String("table", table), slog.Int64("amount", ndel))
		}

		return nil
	}()
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return fmt.Errorf(
				"failed to rollback on seed error: original - %w, transaction error - %v",
				err,
				txErr,
			)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit seed transaction: %w", err)
	}

	return nil
}
