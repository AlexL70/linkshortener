// genopenapi dumps the application's OpenAPI 3.1 spec to ../openapi/LinkShortener.json.
//
// It constructs the full Gin + Huma route graph in-process (no database or
// real OAuth credentials required) and writes the JSON spec that Huma produces.
//
// Invoke via:
//
// go run ./cmd/genopenapi
//
// or add it to the Makefile / CI pipeline so the spec stays in sync whenever
// routes change.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"

	"github.com/AlexL70/linkshortener/backend/business-logic/handlers"
	bizinterfaces "github.com/AlexL70/linkshortener/backend/business-logic/interfaces"
	bizmodels "github.com/AlexL70/linkshortener/backend/business-logic/models"
	"github.com/AlexL70/linkshortener/backend/web/routes"
)

func main() {
	// Stub env vars so route registration (which reads them at startup) doesn't panic.
	os.Setenv("SESSION_SECRET", "stub-session-secret-for-spec-generation")
	os.Setenv("JWT_SECRET", "stub-jwt-secret-for-spec-generation")
	os.Setenv("APP_BASE_URL", "http://localhost:8080")
	os.Setenv("GOOGLE_CLIENT_ID", "stub-client-id")
	os.Setenv("GOOGLE_CLIENT_SECRET", "stub-client-secret")
	os.Setenv("LINKSHORTENER_ENV", "dev")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	humaConfig := huma.DefaultConfig("Link Shortener API", "0.1.0")
	api := humagin.New(router, humaConfig)

	authHandler := handlers.NewAuthHandler(stubUserRepo{}, false, "")
	urlHandler := handlers.NewUrlHandler(stubUrlRepo{})
	routes.RegisterHello(api)
	routes.RegisterAuthRoutes(router, api, authHandler, routes.NewTokenBlacklist())
	routes.RegisterUrlRoutes(api, urlHandler)

	specBytes, err := json.MarshalIndent(api.OpenAPI(), "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "genopenapi: marshal spec: %v\n", err)
		os.Exit(1)
	}

	outPath := filepath.Join("..", "openapi", "LinkShortener.json")
	if err := os.WriteFile(outPath, specBytes, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "genopenapi: write file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("genopenapi: wrote spec to %s\n", outPath)
}

// stubUserRepo is a no-op implementation of UserRepository used only during
// spec generation — no actual database calls are ever made.
type stubUserRepo struct{}

func (stubUserRepo) FindByProviderID(_ context.Context, _ bizmodels.Provider, _ string) (*bizmodels.User, *bizmodels.UserProvider, error) {
	return nil, nil, nil
}

func (stubUserRepo) FindByProviderEmailWithSeedID(_ context.Context, _ bizmodels.Provider, _ string) (*bizmodels.User, *bizmodels.UserProvider, error) {
	return nil, nil, nil
}

func (stubUserRepo) UpdateProviderUserID(_ context.Context, _ int64, _ string) error {
	return nil
}

func (stubUserRepo) CreateUserWithProvider(_ context.Context, _ string, _ *bizmodels.UserProvider) (*bizmodels.User, error) {
	return nil, nil
}

var _ bizinterfaces.UserRepository = stubUserRepo{}

// stubUrlRepo is a no-op implementation of UrlRepository used only during
// spec generation — no actual database calls are ever made.
type stubUrlRepo struct{}

func (stubUrlRepo) FindByUserID(_ context.Context, _ int64, _, _ int) ([]*bizmodels.ShortenedUrl, int, error) {
	return nil, 0, nil
}

var _ bizinterfaces.UrlRepository = stubUrlRepo{}
