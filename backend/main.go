package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/AlexL70/linkshortener/backend/config"
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
