.PHONY: generate build

# Build the server binary.
# SQL scripts in sql/ are regenerated first so they stay in sync with the code.
build: generate
	cd backend && go build -o ../bin/server .

# Regenerate the SQL scripts in sql/ from the Go migration definitions.
# Re-run this after adding or modifying any migration file.
generate:
	cd backend && go generate ./...
