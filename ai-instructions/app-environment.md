# App Environment Variables — Agent Instructions

> **Scope:** Full-stack
> These coding standards and conventions **must be strictly followed** whenever you work on any code that reads, loads, validates, or documents environment variables, regardless of the nature of the work (new features, bug fixes, refactoring, or any other changes). If you find any contradiction between the rules in this file and those in any other instruction file (including `AGENTS.md`), **report it immediately and start a discussion** before making any changes.

---

## 1. Overview

All runtime configuration is supplied through environment variables. The backend reads and validates these variables at startup via `backend/config/` (currently `backend/config/env.go`). **Fail-fast behavior is mandatory:** the application must refuse to start if the configuration is invalid or incomplete, printing a clear, human-readable error message to stderr before exiting.

---

## 2. The `LINKSHORTENER_ENV` Mode Selector

`LINKSHORTENER_ENV` controls which configuration sources are consulted at startup. It is a special variable with the following rules:

- It is **always read from the OS environment only.** `.env` files must never override it, regardless of mode or order.
- It must be set explicitly. There is no default.
- **Accepted values are `dev` and `prod` only.** Any other value — including an empty string — must cause the application to exit immediately with a descriptive error message before any other initialization takes place. Example message:

  ```
  FATAL: LINKSHORTENER_ENV must be "dev" or "prod", got "<actual_value>"
  ```

---

## 3. Configuration Loading Algorithm

### 3.1 `dev` mode

Execute the following steps in order:

1. Parse `../.env` (path relative to the Go binary's working directory). If the file does not exist, silently skip it. If it exists but is malformed, abort with a descriptive error.
2. Parse `../.env.dev`. Entries in this file override matching entries from `../.env`. If the file does not exist, silently skip it. If it exists but is malformed, abort with a descriptive error.
3. For every key/value pair collected from the files, apply it via `os.Setenv` **only if** the OS environment does not already hold a non-empty value for that key. OS-level variables always win.
4. Never apply a file value for `LINKSHORTENER_ENV` — skip it unconditionally regardless of which file it appears in.

### 3.2 `prod` mode

1. Skip all file reading entirely. Do not open, parse, or otherwise access `.env` or any `.env.*` file.
2. Use OS-level environment variables as-is.

### 3.3 Validation (both modes)

After the loading algorithm has run, validate the resolved configuration:

- Check every required variable (see §4.1). If any required variable resolves to an empty string, abort immediately with a message such as:

  ```
  FATAL: required environment variable DATABASE_URL is not set
  ```

- Apply defaults for every optional variable that resolved to an empty string (see §4.2). Numeric defaults must be parsed and validated; abort on non-numeric input with a message identifying the offending variable and its value.
- Validate `SUPER_ADMIN_EMAIL` as a syntactically valid email address (RFC 5322 local-part + `@` + domain). If the value is present but fails validation, abort immediately with a message such as:

  ```
  FATAL: SUPER_ADMIN_EMAIL is not a valid email address: "<actual_value>"
  ```

---

## 4. Variable Catalog

### 4.1 Required variables (no default — abort if absent after loading)

| Variable                  | Description                                                                                                                                                                                                                  |
| ------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `LINKSHORTENER_ENV`       | Mode selector — must be `dev` or `prod` (see §2)                                                                                                                                                                             |
| `DATABASE_URL`            | PostgreSQL connection string (DSN), e.g. `postgres://user:pass@host/db`                                                                                                                                                      |
| `JWT_SECRET`              | Secret key used to sign and verify JWT tokens                                                                                                                                                                                |
| `SESSION_SECRET`          | Secret key used by the Gorilla sessions cookie store                                                                                                                                                                         |
| `GOOGLE_CLIENT_ID`        | Google OAuth2 application client ID                                                                                                                                                                                          |
| `GOOGLE_CLIENT_SECRET`    | Google OAuth2 application client secret                                                                                                                                                                                      |
| `MICROSOFT_CLIENT_ID`     | Microsoft OAuth2 application client ID                                                                                                                                                                                       |
| `MICROSOFT_CLIENT_SECRET` | Microsoft OAuth2 application client secret                                                                                                                                                                                   |
| `FACEBOOK_CLIENT_ID`      | Facebook OAuth2 application client ID                                                                                                                                                                                        |
| `FACEBOOK_CLIENT_SECRET`  | Facebook OAuth2 application client secret                                                                                                                                                                                    |
| `APP_BASE_URL`            | The application's externally-accessible base URL, e.g. `http://localhost:8080`. Used by the backend to construct the OAuth provider callback URL and by the frontend as the API base URL.                                    |
| `SUPER_ADMIN_EMAIL`       | Email address of the super-admin user. Must be a syntactically valid email address. Used to grant admin privileges and, in `dev` mode with an empty database, to seed an initial user record (see `backend-dblayer.md §14`). |

If any of these is absent (resolved to an empty string) after the loading algorithm has run, the application **must not start**.

### 4.2 Optional variables with defaults

| Variable                      | Default     | Description                                                                                                                                                                                                                                                                                                            |
| ----------------------------- | ----------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `PORT`                        | `8080`      | TCP port the HTTP server listens on                                                                                                                                                                                                                                                                                    |
| `GIN_MODE`                    | `release`   | Gin engine mode (`release` or `debug`)                                                                                                                                                                                                                                                                                 |
| `REQUEST_TIMEOUT_SECONDS`     | `30`        | Per-request timeout in seconds                                                                                                                                                                                                                                                                                         |
| `MAX_CONCURRENT_REQUESTS`     | `100`       | Maximum simultaneous in-flight HTTP requests                                                                                                                                                                                                                                                                           |
| `DB_MAX_OPEN_CONNS`           | `25`        | Database connection pool — maximum open connections                                                                                                                                                                                                                                                                    |
| `DB_MAX_IDLE_CONNS`           | `5`         | Database connection pool — maximum idle connections                                                                                                                                                                                                                                                                    |
| `DB_CONN_MAX_LIFETIME`        | `3600`      | Maximum connection lifetime in seconds (1 hour)                                                                                                                                                                                                                                                                        |
| `DB_CONN_MAX_IDLE_TIME`       | `600`       | Maximum connection idle time in seconds (10 minutes)                                                                                                                                                                                                                                                                   |
| `CLICK_BATCH_SIZE`            | `1000`      | Number of click events buffered before a flush to the database                                                                                                                                                                                                                                                         |
| `CLICK_BATCH_TIMEOUT_SECONDS` | `5`         | Maximum seconds between automatic click-batch flushes                                                                                                                                                                                                                                                                  |
| `DEFAULT_PAGE_SIZE`           | `20`        | Default number of items returned by paginated list endpoints                                                                                                                                                                                                                                                           |
| `MAX_URL_LENGTH`              | `2048`      | Maximum allowed length (in characters) for a submitted long URL                                                                                                                                                                                                                                                        |
| `MIN_SHORTCODE_LENGTH`        | `6`         | Minimum length for a user-supplied custom shortcode                                                                                                                                                                                                                                                                    |
| `MAX_SHORTCODE_LENGTH`        | `6`         | Maximum length for both custom and auto-generated shortcodes                                                                                                                                                                                                                                                           |
| `MAX_SHORTCODE_RETRIES`       | `10`        | Maximum number of retry attempts when a generated shortcode collides with an existing one                                                                                                                                                                                                                              |
| `RATE_LIMIT_AUTH_RPM`         | `5`         | Max requests per minute per IP on auth endpoints (`/auth/login`, `/auth/register`)                                                                                                                                                                                                                                     |
| `RATE_LIMIT_API_RPM`          | `60`        | Max requests per minute per IP on authenticated API endpoints (`/user/`)                                                                                                                                                                                                                                               |
| `RATE_LIMIT_REDIRECT_RPM`     | `120`       | Max requests per minute per IP on the public redirect endpoint (`/r/`)                                                                                                                                                                                                                                                 |
| `TRUSTED_PROXIES`             | `127.0.0.1` | Comma-separated list of trusted reverse-proxy IP addresses used by Gin to extract the real client IP from `X-Forwarded-For`                                                                                                                                                                                            |
| `DNS_LOOKUP_FAIL_OPEN`        | `false`     | Controls DNS lookup failure behavior during URL SSRF validation (prod mode only; DNS checks are skipped in dev mode). When `true`, a DNS resolution error is non-fatal and the URL is allowed through (fail-open). When `false`, DNS errors cause the URL to be rejected (fail-closed). Must be `"true"` or `"false"`. |

---

## 5. Implementation Rules

- All environment loading and validation logic lives in `backend/config/`. Do not scatter `os.Getenv` calls across unrelated packages for configuration purposes.
- `LoadEnv()` (file loading + OS merging) must be called **first** in `main()`, before any other initialization.
- A separate `Validate()` function (or equivalent) must be called immediately after `LoadEnv()`. It is responsible for checking required variables and applying defaults for optional ones.
- Once validation passes, all other packages read their configuration values through `os.Getenv` or through a typed config struct populated by the `config` package. Do not store a second copy of env values in global variables outside the `config` package unless there is a documented reason to do so.
- If a new environment variable is introduced anywhere in the codebase, its entry **must** be added to §4.1 or §4.2 in this file in the same commit. Undocumented variables are a defect.

---

## 6. Startup Configuration Logging

Both the backend and the frontend must emit a configuration summary at startup, **before serving any requests**.

- **Required variables** (§4.1) are always included in the log — they have no default, so every value represents an active, non-trivial configuration choice.
- **Optional variables** (§4.2) are logged only if their resolved value differs from the documented default. Variables already at their default are **not** logged.

### 6.1 Backend

Use `slog` at `INFO` level. Emit one structured log entry per variable.

The masking behaviour depends on mode:

| Mode   | Variable type                                         | Logged value                                                                                                                            |
| ------ | ----------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| `dev`  | Any                                                   | Actual value — no masking applied                                                                                                       |
| `prod` | Non-secret variables                                  | Actual value                                                                                                                            |
| `prod` | `JWT_SECRET`, `SESSION_SECRET`, any `*_CLIENT_SECRET` | Fixed string `****` (must **not** reveal actual length)                                                                                 |
| `prod` | `DATABASE_URL`                                        | Parse the DSN; replace only the password component with `****`; log the resulting string (e.g. `postgres://user:****@host:5432/dbname`) |

The DSN masking must handle both URI-format (`postgres://user:pass@host/db`) and key=value-format (`host=... password=... dbname=...`) connection strings.

### 6.2 Frontend

Log to the browser console using `console.info` during app initialization (e.g., in `main.ts`, before `app.mount()`). Frontend environment variables are accessed via `import.meta.env`.

- **Dev mode:** log the actual value of every non-default variable — no masking applied.
- **Prod mode:** apply the same masking rules as the backend. Frontend variables normally do not contain secrets, but any variable classified as a secret in §4.1 must still be masked.

---

## 7. Security Constraints

- **Never log** the values of `JWT_SECRET`, `SESSION_SECRET`, or any `*_CLIENT_SECRET` variable in `prod` mode — not even at `DEBUG` level. In `dev` mode these logging restrictions do not apply; unmasked values may appear in startup logs and debug output (see §6 for the full startup logging rules).
- `DATABASE_URL` must be logged in `prod` mode, but only after masking the password component as described in §6.1. Logging the raw unmasked DSN is forbidden in `prod` mode.
- **Never expose** any environment variable value in an API response, an error message returned to a client, or any output accessible to end users. This rule applies in **both** modes without exception.
- In `prod` mode, the algorithm must **never** attempt to open or read `.env` files. This is a hard constraint, not a soft guideline.
- Secrets (`JWT_SECRET`, `SESSION_SECRET`, `*_CLIENT_SECRET`) must be cryptographically strong random values of at least 32 bytes. Document this requirement in any `.env.example` or setup guide you create.

---

## 8. `.env` File Conventions

These rules govern **creating** `.env` files. Code outside `backend/config/env.go` must never open or read `.env*` files directly — there are no exceptions.

| File       | Purpose                                                                               |
| ---------- | ------------------------------------------------------------------------------------- |
| `.env`     | Base values shared across all non-`prod` environments. Must not contain real secrets. |
| `.env.dev` | Development-only overrides. May contain local secrets (e.g. local DB password).       |

- Both files must be listed in `.gitignore` and must never be committed to the repository.
- A `.env.example` file (containing only placeholder values, never real secrets) may be committed as documentation for required variables.
- When **explicitly instructed** to create a `.env` or `.env.dev` file, populate it with the variables from §4.1 and §4.2, using empty or placeholder values for secrets and documented defaults for optional variables.

---

## 9. Contradictions and Gaps

- If you find any environment variable being read by the code but not listed in §4, add it to the correct table and report the discrepancy to the user immediately.
- If you find any contradiction between this file and `AGENTS.md` or any other instruction file, report it immediately and start a discussion before making any changes.
