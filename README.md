# Link Shortener

A simple URL shortening application that lets users create short links from long URLs.

## About This Project

This is a Link Shortener app that allows registered users to:

- Sign in using their Google, Microsoft, or Facebook accounts
- Create shortened versions of long URLs
- Share their short links easily
- Track basic statistics on their shortened links
- Manage their URLs through a personal dashboard

## Status

⚠️ **This project is currently under construction.** The MVP has not been created yet. This is an ongoing development effort.

## Development Note

This project is developed using **GitHub Copilot**. A key objective of this project is to gain valuable experience with heavily using GitHub Copilot as an AI coding assistant, exploring its capabilities for accelerating development workflows and improving code quality.

## Building

```bash
make build
```

This runs `go generate ./...` first (regenerating the SQL scripts in `sql/`) and then compiles the server binary to `bin/server`.

To regenerate SQL scripts without building:

```bash
make generate
```

## Database Migrations

### Development

In `dev` mode (`LINKSHORTENER_ENV=dev`) the server applies any pending migrations automatically on startup. No manual action is required during day-to-day development.

### Production

In `prod` mode (`LINKSHORTENER_ENV=prod`) the server **never** modifies the database schema automatically. Ready-to-run SQL scripts are generated into the `sql/` directory as part of every build.

Two types of script are generated:

| File             | Purpose                                                          |
| ---------------- | ---------------------------------------------------------------- |
| `sql/schema.sql` | All migrations in order — use to initialise a brand-new database |
| `sql/<name>.sql` | Individual migration — apply only the incremental change         |

**Recommended workflow for a schema change:**

1. Add a new migration file in `backend/infrastructure/pg/migrations/`.
2. Run `make build` (or `make generate`) — the updated SQL scripts are written to `sql/`.
3. Review the generated `sql/<new_migration>.sql` to confirm the DDL.
4. Apply the script against the production database **before** deploying the new server binary:
   ```bash
   psql "$DATABASE_URL" -f sql/<new_migration>.sql
   ```
5. Deploy and start the new server binary.

**To initialise a brand-new production database from scratch:**

```bash
psql "$DATABASE_URL" -f sql/schema.sql
```

> **Security note:** `sql/` contains no secrets. The scripts are safe to commit to version control and review in pull requests.

## Deployment

See [Deploy.md](Deploy.md) for the full step-by-step guide covering one-time server setup and repeatable deployments.

## License

MIT License - See [LICENSE](LICENSE) file for details
