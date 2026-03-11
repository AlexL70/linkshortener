package main

//go:generate go run ./cmd/gensql

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/AlexL70/linkshortener/backend/config"
	"github.com/AlexL70/linkshortener/backend/infrastructure/pg"
	"github.com/AlexL70/linkshortener/backend/infrastructure/pg/seed"
	"github.com/AlexL70/linkshortener/backend/web/routes"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
)

func main() {
	// Set up the JSON logger first so all startup errors are captured via slog.
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := config.LoadEnv(); err != nil {
		slog.Error("failed to load environment", "error", err)
		os.Exit(1)
	}

	if err := config.Validate(); err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	config.LogConfig()

	db, err := pg.Open()
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close() // nolint: errcheck — best-effort close on shutdown

	// Auto-migration runs only in dev mode.
	// In prod, apply the generated sql/schema.sql (or per-migration files) via psql.
	if config.GetAppEnv() == config.EnvDev {
		if err := pg.EnsureDatabase(context.Background()); err != nil {
			slog.Error("failed to ensure database exists", "error", err)
			os.Exit(1)
		}
		if err := pg.RunMigrations(context.Background(), db); err != nil {
			slog.Error("failed to run migrations", "error", err)
			os.Exit(1)
		}
		seed.RunIfEmpty(context.Background(), db)
	} else {
		slog.Info("auto-migration skipped in prod mode")
	}

	router := gin.Default()

	humaConfig := huma.DefaultConfig("Link Shortener API", "0.1.0")
	api := humagin.New(router, humaConfig)

	routes.RegisterHello(api)

	port := os.Getenv("PORT")
	slog.Info("starting server", "port", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
