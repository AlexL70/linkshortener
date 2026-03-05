# Backend DB Layer — Agent Instructions

> **Scope:** Backend
> These coding standards and conventions **must be strictly followed** whenever you work on any database-layer code, regardless of the nature of the work (new features, bug fixes, refactoring, or any other changes). If you find any contradiction between the rules in this file and those in any other instruction file (including `AGENTS.md`), **report it immediately and start a discussion** before making any changes.

---

## 1. Technology — Non-Negotiable Constraints

| Concern    | Technology                                         |
| ---------- | -------------------------------------------------- |
| ORM        | `github.com/uptrace/bun` with `pgdialect`          |
| DB driver  | `github.com/uptrace/bun/driver/pgdriver`           |
| Database   | PostgreSQL                                         |
| Migrations | `bun/migrate` (Go migration functions, code-first) |

Do not replace or supplement any of these with alternative libraries or tools without explicit user approval.

---

## 2. Directory Layout

All DB layer code lives under `backend/infrastructure/pg/`:

| Path                              | Purpose                                              |
| --------------------------------- | ---------------------------------------------------- |
| `infrastructure/pg/`              | Bun DB setup, connection pool initialization         |
| `infrastructure/pg/models/`       | Bun model structs defining the database schema       |
| `infrastructure/pg/migrations/`   | Numbered, append-only Go migration functions         |
| `infrastructure/pg/mappings/`     | `ToBusinessModel` / `ToDbModel` conversion functions |
| `infrastructure/pg/repositories/` | Concrete repository implementations                  |

Repository interface definitions belong in `business-logic/interfaces/`. Concrete repository implementations belong in `infrastructure/pg/repositories/`.

---

## 3. Bun Model Structs

- Every model struct must use `bun` struct tags to declare the table name, columns, and relationships.
- All tables must have `created_at TIMESTAMPTZ` column, and mutable tables must also have `updated_at TIMESTAMPTZ`. Both must be populated automatically (via Bun hooks or PostgreSQL defaults — choose one approach consistently across all models).
- Table names are **PascalCase plural** and must match the struct name (e.g., struct `User` → table `Users`).
- Use `bun:"pk,autoincrement"` for SERIAL primary keys.
- All foreign keys must declare `ON DELETE CASCADE` in migration scripts unless a documented reason justifies `SET NULL` or `RESTRICT`.
- DB models may differ from business logic models when the storage representation benefits from a different shape. Do not blindly mirror business logic models — model the schema correctly for storage first.

---

## 4. Database Schema

The canonical schema is defined by the structs in `infrastructure/pg/models/`. The four core tables are:

| Table           | Notes                                                                                                                                                        |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `Users`         | `id`, `user_name` (UNIQUE NOT NULL), `created_at`, `updated_at`                                                                                              |
| `UserProviders` | `id`, `user_id` (FK→Users), `provider`, `provider_user_id`, `provider_email`, `created_at`. UNIQUE(`provider`, `provider_user_id`). INDEX on `user_id`.      |
| `ShortenedUrls` | `id`, `user_id` (FK→Users), `shortcode` (UNIQUE NOT NULL), `long_url`, `expires_at` (NULLABLE), `created_at`, `updated_at`. INDEX on `shortcode`, `user_id`. |
| `UrlClicks`     | `id`, `url_id` (FK→ShortenedUrls), `clicked_at`, `ip_address`, `user_agent`, `referer`. INDEX on `url_id`, `clicked_at`, `ip_address`.                       |

Any schema change must start with updating the struct in `infrastructure/pg/models/` and then adding a **new** migration function. Never modify an already-applied migration.

---

## 5. Migrations

- Every migration is a numbered Go function in `infrastructure/pg/migrations/` with a sequential prefix, e.g., `001_initial_schema.go`, `002_add_expires_at.go`.
- Migrations are **append-only**. Never edit a migration file that has already been applied. Add a new migration instead.
- The migration file must register itself with the `bun/migrate` migrator at `init()` time.
- Raw SQL strings are permitted **only** inside migration files, and only when the Bun query builder is insufficient (complex DDL, multi-step data migrations, etc.). Add a comment explaining why raw SQL was necessary.
- Migration state is tracked automatically via the `bun_migrations` table — do not manage it manually.

---

## 6. Repository Interfaces

- Every repository interface lives in `business-logic/interfaces/`. It must be a pure Go interface with no implementation details or infrastructure imports.
- Interface methods must operate exclusively on business logic model types (from `business-logic/models/`) — never on Bun model structs or raw SQL types.
- Name interfaces descriptively: e.g., `UrlRepository`, `UserRepository`, `ClickRepository`.
- Handlers and other business logic code must depend only on these interfaces, never on concrete implementations.

---

## 7. Repository Implementations

- Concrete repository structs live in `infrastructure/pg/repositories/`. They implement the interfaces defined in `business-logic/interfaces/`.
- Constructors must accept a `*bun.DB` and return the interface type (not the concrete type), e.g.:
  ```go
  func NewUrlRepository(db *bun.DB) interfaces.UrlRepository { ... }
  ```
- Use Bun's query builder for all reads and writes. Never write raw SQL strings in repository implementations.
- Always specify columns explicitly in SELECT queries — **never use `SELECT *`**.
- Wrap all database errors with context before returning them: `fmt.Errorf("UrlRepository.FindByShortcode: %w", err)`.

---

## 8. Mapping Functions

- Every pair of DB model ↔ business logic model must have conversion functions in `infrastructure/pg/mappings/`.
- Function naming is mandatory:
  - `ToBusinessModel(dbModel *models.Foo) *bizmodels.Foo` — DB to business logic
  - `ToDbModel(bizModel *bizmodels.Foo) *models.Foo` — business logic to DB
- Mapping functions must be pure functions (no side effects, no DB calls).
- If a slice conversion is needed, add a `SliceToBusinessModel` / `SliceToDbModel` helper in the same file.

---

## 9. Connection Pool

- Open the underlying `*sql.DB` via `pgdriver.NewConnector` and configure the pool from environment variables:

  | Env var                 | Default | Description                  |
  | ----------------------- | ------- | ---------------------------- |
  | `DB_MAX_OPEN_CONNS`     | 25      | Maximum open connections     |
  | `DB_MAX_IDLE_CONNS`     | 5       | Maximum idle connections     |
  | `DB_CONN_MAX_LIFETIME`  | 3600    | Connection max lifetime (s)  |
  | `DB_CONN_MAX_IDLE_TIME` | 600     | Connection max idle time (s) |

- Never hardcode pool values. Read and apply them at startup before any query is executed.
- Wrap the configured `*sql.DB` with `bun.NewDB(sqldb, pgdialect.New())` to obtain the application-wide `*bun.DB` instance.

---

## 10. Click Batching

- URL click records (`UrlClicks`) must **not** be written to the database synchronously on every redirect.
- Buffer click events in an in-memory batch and flush to the DB on a schedule driven by these env vars:

  | Env var                       | Default | Description                          |
  | ----------------------------- | ------- | ------------------------------------ |
  | `CLICK_BATCH_SIZE`            | 1000    | Flush when buffer reaches this size  |
  | `CLICK_BATCH_TIMEOUT_SECONDS` | 5       | Flush after this many seconds anyway |

- The batch flusher must be safe for concurrent access and must not lose records on application shutdown (attempt a final flush before exiting).
- Expose the batch writer as an interface in `business-logic/interfaces/` so the redirect handler does not depend directly on the infrastructure implementation.

---

## 11. Pagination

- All repository methods that return lists must accept pagination parameters (`page`, `pageSize`).
- Default page size is read from the `DEFAULT_PAGE_SIZE` env var (default: 20).
- Repository methods must return both the result slice and the total record count so callers can compute pagination metadata.

---

## 12. Testing

- Minimum **95% code coverage** for all packages under `infrastructure/pg/`.
- Use `github.com/stretchr/testify` (`assert` / `require`) for assertions.
- Test files live alongside the code they test (e.g., `infrastructure/pg/repositories/url_repository_test.go`).
- Tests follow the **Arrange / Act / Assert** pattern.
- For integration tests that hit the database, use a dedicated test PostgreSQL instance (not the production DB). Isolate tests with transactions that are rolled back after each test case.
- Auto-generated code (mapping stubs, mock implementations) is excluded from the 95% coverage requirement, but hand-written mapping functions are not.

---

## 13. What Agents Must Not Do

- Do not write raw SQL in repository implementations. Migration files are the only permitted exception (with a comment justifying it).
- Do not use `SELECT *` in any Bun query.
- Do not edit migration files that have already been applied. Add new migrations instead.
- Do not let handlers import `infrastructure/pg` directly. Always wire through interfaces.
- Do not hardcode connection pool settings, batch sizes, or page sizes.
- Do not write `UrlClicks` records synchronously in the redirect path — always use the batch buffer.
- Do not embed Bun model structs in viewmodels or business logic models.
- Do not bypass `ToBusinessModel` / `ToDbModel` mappers by casting or assigning structs directly across layers.
