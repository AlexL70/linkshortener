.PHONY: generate build generate-spec

# Build the server binary.
# SQL scripts in sql/ are regenerated first so they stay in sync with the code.
build: generate
	cd backend && go build -o ../bin/server .

# Regenerate the SQL scripts in sql/ from the Go migration definitions.
# Re-run this after adding or modifying any migration file.
generate:
	cd backend && go generate ./...

# Regenerate the OpenAPI spec from the live route graph.
# Re-run this after adding, removing, or modifying any Huma-registered route.
generate-spec:
	cd backend && go run ./cmd/genopenapi

# Run the PostgreSQL database in a Docker container for development.
run-pg-dev:
	docker compose up -d db

# Run the backend server in development mode with live reloading.
run-backend-dev:
	cd backend && go tool air -c .air.toml &

# Run the frontend development server with live reloading.
run-frontend-dev:
	cd frontend && npm run dev &

# Run all development environments
run-dev: run-pg-dev run-backend-dev run-frontend-dev

# Stop the backend server run by run-backend-dev command.
stop-backend-dev:
	pkill -f "go tool air -c .air.toml"

# Stop the frontend development server run by run-frontend-dev command.
stop-frontend-dev:
	pkill -f "npm run dev"

# Stop the PostgreSQL database container started by run-pg-dev command.
stop-pg-dev:
	docker compose down db

# Stop all development services started by run-dev command.
stop-dev: stop-backend-dev stop-frontend-dev stop-pg-dev
