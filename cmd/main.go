package main

import (
	"fmt"
	"log/slog"
	"os"

	envcfg "github.com/caarlos0/env/v11"
	"github.com/davidonium/namemyserver/internal/env"
	"github.com/davidonium/namemyserver/internal/server"
	"github.com/davidonium/namemyserver/internal/vite"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "could not start namemyserver app.\nerror: %v\n", err)
		os.Exit(1)
	}
}
func run(args []string) error {
	level := slog.LevelInfo

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	var cfg env.Config
	if err := envcfg.Parse(&cfg); err != nil {
		return fmt.Errorf("failed to parse environment variables into a config struct: %w", err)
	}

	assets := vite.NewAssets(vite.AssetsConfig{
		RootURL: cfg.AssetsRootURL.String(),
		UseManifest: cfg.AssetsUseManifest,
	})

	s := server.New(&server.Services{
		Logger: logger,
		Config: cfg,
		Assets: assets,
	})

	if err := s.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start the http server: %w", err)
	}

	return nil
}
