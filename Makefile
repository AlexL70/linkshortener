.PHONY: generate build generate-spec \
        build-backend-prod build-frontend-prod build-prod deploy bump

APP_VERSION := $(shell cat version)
_SSH_OPTS   := $(if $(SSH_KEY),-i $(SSH_KEY))
_RSYNC_E    := $(if $(SSH_KEY),-e "ssh -i $(SSH_KEY)")

# Build the server binary.
# SQL scripts in sql/ are regenerated first so they stay in sync with the code.
build: generate
	cd backend && go build -ldflags="-X main.appVersion=$(APP_VERSION)" -o ../bin/server .

# Regenerate the SQL scripts in sql/ from the Go migration definitions.
# Re-run this after adding or modifying any migration file.
generate:
	cd backend && go generate ./...

# Regenerate the OpenAPI spec from the live route graph.
# Re-run this after adding, removing, or modifying any Huma-registered route.
generate-spec:
	cd backend && go run -ldflags="-X main.appVersion=$(APP_VERSION)" ./cmd/genopenapi

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
	pkill -f "[g]o tool air -c .air.toml" || true

# Stop the frontend development server run by run-frontend-dev command.
stop-frontend-dev:
	pkill -f "[n]ode_modules/.bin/vite" || true
	pkill esbuild || true

# Stop the PostgreSQL database container started by run-pg-dev command.
stop-pg-dev:
	docker compose down db

# Stop all development services started by run-dev command.
stop-dev: stop-frontend-dev stop-backend-dev stop-pg-dev

# Build the backend binary for production (linux/amd64, symbols stripped).
build-backend-prod:
	mkdir -p bin
	cd backend && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.appVersion=$(APP_VERSION)" -o ../bin/server .

# Build the frontend for production.
# Requires APP_BASE_URL, e.g.:  make build-frontend-prod APP_BASE_URL=https://yourdomain.com
build-frontend-prod:
	cd frontend && LINKSHORTENER_ENV=prod APP_BASE_URL=$(APP_BASE_URL) APP_VERSION=$(APP_VERSION) npm run build

# Build both backend and frontend for production.
# Requires APP_BASE_URL.  Example:  make build-prod APP_BASE_URL=https://yourdomain.com
build-prod: build-backend-prod build-frontend-prod

# Build and deploy to the remote server via rsync + SSH, then restart the service.
# Requires SSH_HOST and APP_BASE_URL.  SSH_KEY is optional (path to identity file).  Examples:
#   make deploy SSH_HOST=user@your-server APP_BASE_URL=https://yourdomain.com
#   make deploy SSH_HOST=user@your-server SSH_KEY=~/.ssh/id_rsa APP_BASE_URL=https://yourdomain.com
deploy: build-prod
	rsync -avz $(_RSYNC_E) --rsync-path="sudo rsync" bin/server $(SSH_HOST):/opt/linkshortener/server
	rsync -avz --delete $(_RSYNC_E) --rsync-path="sudo rsync" frontend/dist/ $(SSH_HOST):/opt/linkshortener/frontend/
	ssh $(_SSH_OPTS) $(SSH_HOST) "sudo systemctl restart linkshortener"

# Bump the application version, commit all staged changes, and create a git tag.
# Usage:  make bump BUMP=patch   (or major, minor)
bump:
	@if [ "$(BUMP)" != "major" ] && [ "$(BUMP)" != "minor" ] && [ "$(BUMP)" != "patch" ]; then \
		echo "Error: BUMP must be one of: major, minor, patch"; \
		echo "Usage: make bump BUMP=patch"; \
		exit 1; \
	fi
	@CURRENT=$$(cat version); \
	MAJOR=$$(echo $$CURRENT | cut -d. -f1); \
	MINOR=$$(echo $$CURRENT | cut -d. -f2); \
	PATCH=$$(echo $$CURRENT | cut -d. -f3); \
	if [ "$(BUMP)" = "major" ]; then \
		MAJOR=$$((MAJOR + 1)); MINOR=0; PATCH=0; \
	elif [ "$(BUMP)" = "minor" ]; then \
		MINOR=$$((MINOR + 1)); PATCH=0; \
	else \
		PATCH=$$((PATCH + 1)); \
	fi; \
	NEW_VERSION=$$MAJOR.$$MINOR.$$PATCH; \
	echo $$NEW_VERSION > version; \
	git add -A; \
	git commit -m "Version bumped to $$NEW_VERSION"; \
	git tag "v$$NEW_VERSION"; \
	echo "Bumped to $$NEW_VERSION — run: git push && git push --tags"
