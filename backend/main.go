package main

//go:generate go run ./cmd/gensql

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"
	"github.com/AlexL70/linkshortener/backend/business-logic/handlers"
	"github.com/AlexL70/linkshortener/backend/config"
	"github.com/AlexL70/linkshortener/backend/infrastructure/pg"
	pgrepositories "github.com/AlexL70/linkshortener/backend/infrastructure/pg/repositories"
	"github.com/AlexL70/linkshortener/backend/infrastructure/pg/seed"
	"github.com/AlexL70/linkshortener/backend/web/routes"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
)

// appVersion is set at build time via -ldflags "-X main.appVersion=<version>".
// The default "dev" is used for local / test builds without the flag.
var appVersion = "dev"

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
	humaConfig.Info.Description = "App version: " + appVersion
	api := humagin.New(router, humaConfig)

	userRepo := pgrepositories.NewUserRepository(db)
	isDevMode := config.GetAppEnv() == config.EnvDev
	adminEmail := os.Getenv("SUPER_ADMIN_EMAIL")
	authHandler := handlers.NewAuthHandler(userRepo, isDevMode, adminEmail)

	urlRepo := pgrepositories.NewUrlRepository(db)
	maxUrlLen, _ := strconv.Atoi(os.Getenv("MAX_URL_LENGTH"))
	minShortcodeLen, _ := strconv.Atoi(os.Getenv("MIN_SHORTCODE_LENGTH"))
	maxShortcodeLen, _ := strconv.Atoi(os.Getenv("MAX_SHORTCODE_LENGTH"))
	maxShortcodeRetries, _ := strconv.Atoi(os.Getenv("MAX_SHORTCODE_RETRIES"))
	shortcodeGen := businesslogic.NewShortcodeGenerator(maxShortcodeLen)
	urlHandler := handlers.NewUrlHandler(urlRepo, shortcodeGen, maxUrlLen, minShortcodeLen, maxShortcodeLen, maxShortcodeRetries)

	blacklist := routes.NewTokenBlacklist()
	router.Use(routes.CORSMiddleware())
	router.Use(routes.RequireJWTGlobal(blacklist, routes.DefaultPublicPaths))

	routes.RegisterAuthRoutes(router, api, authHandler, blacklist)
	routes.RegisterUrlRoutes(api, urlHandler)
	routes.RegisterRedirectRoute(router, api, urlHandler)

	port := os.Getenv("PORT")
	slog.Info("starting server", "port", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
