package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
	envcfg "github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"

	"github.com/davidonium/namemyserver"
	"github.com/davidonium/namemyserver/internal/env"
	"github.com/davidonium/namemyserver/internal/server"
	"github.com/davidonium/namemyserver/internal/store/sqlitestore"
	"github.com/davidonium/namemyserver/internal/vite"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "could not start namemyserver app.\nerror: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	ctx := context.Background()

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

	db, err := sqlitestore.Connect(ctx, cfg.DatabaseURL.String())
	if err != nil {
		return err
	}

	defer db.Close()

	dbm := dbmate.New(cfg.DatabaseURL)
	dbm.AutoDumpSchema = false
	dbm.FS = namemyserver.MigrationsFS

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
		if err := assets.LoadManifestFromFS(namemyserver.FrontendFS, cfg.AssetsManifestLocation); err != nil {
			return fmt.Errorf("failed to load assets from fs at %s: %w", cfg.AssetsManifestLocation, err)
		}
	}

	s := server.New(&server.Services{
		Logger: logger,
		Config: cfg,
		Assets: assets,
	})

	logger.Info("starting http server", "addr", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start the http server: %w", err)
	}

	return nil
}
