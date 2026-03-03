# Agent Instructions — Link Shortener Project

This file contains coding standards, conventions, and constraints that all AI agents must follow when contributing to this project. Read it in full before making any changes. More granular, topic-specific guidelines live in the `/ai-instructions/` directory as separate markdown files. When a rule in `/ai-instructions/` conflicts with a rule here, the more specific file wins but you must report the contradiction as soon as you found it.

---

## 1. Project Overview

A URL shortener web application. Registered users authenticate via third-party OAuth2/OIDC providers (no passwords stored), manage their shortened URLs through a dashboard, and share short links publicly. All architectural decisions are documented in [app_plan.md](app_plan.md). **Do not modify `app_plan.md` unless you are explicitly instructed to do so and the user's instruction contains a direct link to it. If you find any contradiction between other instructions and `app_plan.md`, report it immediately.**

---

## 2. Repository Structure

```
/
├── AGENTS.md               ← This file
├── app_plan.md             ← Architecture & design reference (must be updated if and only if user explicitly instructs you to do so and their instruction contains a direct link to this file)
├── ai-instructions/        ← Topic-specific agent guidelines
├── openapi/
│   └── LinkShortener.json  ← Auto-generated OpenAPI 3.1 spec (do not edit manually)
├── backend/                ← Go application (Gin + Huma + Bun + PostgreSQL)
│   ├── go.mod
│   ├── business-logic/
│   │   ├── handlers/       ← HTTP request handlers (Huma operations)
│   │   ├── interfaces/     ← Repository & service interfaces
│   │   └── models/         ← business layer models that might differ from DB schema and view-models used on the web layer. They must be used in handlers and appropriate interfaces, but each level might have mapping to convert between them and DB/view or other types of models.
│   ├── infrastructure/
│   │   └── pg/             ← Bun ORM setup and repository implementations
│   │       ├── migrations/ ← Numbered Go migration functions
│   │       ├── models/     ← Bun model structs (DB schema)
│   │       └── mappings/   ← ToBusinessModel / ToDbModel conversion functions
│   └── web/
│       ├── viewmodels/     ← Request/response types used by Huma
│       └── mappers/        ← Viewmodel ↔ business logic model converters
└── frontend/               ← Vue 3 + TypeScript + Vite + shadcn/vue
    ├── src/
    │   ├── components/     ← shadcn/vue components only (custom components, if any, will go into separate folder custom-components/)
    │   ├── views/          ← Page-level components (one per route)
    │   ├── stores/         ← Pinia stores
    │   ├── router/         ← Vue Router setup
    │   └── lib/            ← API client and other utilities
    └── ...
```

---

## 3. Technology Stack (Mandatory — Do Not Substitute, unless explicitly instructed by the user to make a specific substitution; still, you may report your concerns if you find particular choice problematic for some reason, but do not substitute it without user approval)

### Backend

| Concern               | Technology                                                           |
| --------------------- | -------------------------------------------------------------------- |
| Language              | Go (use latest stable version declared in `go.mod`)                  |
| HTTP framework        | `github.com/gin-gonic/gin`                                           |
| OpenAPI / validation  | `github.com/danielgtaylor/huma/v2` with `humagin` adapter            |
| ORM / DB access       | `github.com/uptrace/bun` + `pgdialect` + `pgdriver`                  |
| Migrations            | `bun/migrate` (code-first, Go migration functions)                   |
| Auth (OAuth2)         | `github.com/markbates/goth` + `gothic`                               |
| Session (OAuth state) | `github.com/gorilla/sessions` (transient OAuth state only)           |
| JWT                   | `github.com/golang-jwt/jwt/v5`                                       |
| Logging               | Go standard library `log/slog` (no third-party logger)               |
| Testing               | `testing` + `github.com/stretchr/testify` + `github.com/golang/mock` |

### Frontend

| Concern          | Technology                                                        |
| ---------------- | ----------------------------------------------------------------- |
| Language         | TypeScript (strict mode)                                          |
| Framework        | Vue 3 (Composition API with `<script setup>`)                     |
| Build tool       | Vite                                                              |
| UI components    | shadcn/vue + Tailwind CSS                                         |
| State management | Pinia                                                             |
| Router           | Vue Router 4                                                      |
| API client       | Auto-generated from OpenAPI spec via `openapi-typescript-codegen` |
| Testing          | Vitest + `@vue/test-utils`                                        |

Never introduce a dependency that replaces any item listed above without explicit instruction from the user.

---

## 4. General Coding Principles

### 4.1 Clarity and Explicitness

- Write code that is easy to read and understand without comments. Only add comments when the _why_ is non-obvious.
- Prefer explicit over implicit. Avoid magic, reflective tricks, or clever shortcuts that obscure intent.
- Keep functions and methods small and focused on a single responsibility; keep names considerably short, but descriptive. Descriptiveness goes over brevity.

### 4.2 Strict Typing

- **Go:** Use concrete types. Avoid `interface{}` / `any` except where the API (e.g., Huma, Bun) requires it explicitly.
- **TypeScript:** Enable and maintain `strict: true` in all `tsconfig*.json` files. Never use `any`; use `unknown` and narrow appropriately. Prefer specific types and interfaces over unions of broad types. Always use as narrow types as possible under the circumstances.

### 4.3 Error Handling

- **Go:** Always handle errors explicitly. Never ignore an error with `_` unless there is a clear documented reason. Always leave at least a short comment explaining why you ignored the particular error. Wrap errors with context using `fmt.Errorf("...: %w", err)`.
- **TypeScript/Vue:** Never swallow errors silently. Surface errors to the user or propagate them to an appropriate error boundary. For business logic errors (e.g., validation failures, quota exceeded), return structured error responses from the backend and display user-friendly messages on the frontend. Log technical details to the console for debugging, but do not expose them in the UI.

### 4.4 No Raw SQL

- All database access must go through Bun's query builder or model-level APIs. Raw SQL strings are prohibited except for migration scripts where the query builder is insufficient. If you find the situation where avoidance of raw SQL leads to essentially bad performance, report this case to the user and initialize a discussion about whether an exception is warranted. All exceptions must be approved by the user, commented and documented in the appropriate Readme section.

### 4.5 Code-First Schema

- Database schema is derived from Go model structs in `backend/infrastructure/pg/models`. Schema changes start there, followed by a new Bun migration function. Never alter the database schema directly or via external tools. DB models should coincide with business logic models in `backend/business-logic/models`, but not necessarily repeat them. They are storage models, that might differ from the business logic models if it seems profitable to do so. Mapping functions must be placed in the `infrastructure/pg/mappings` directory. Functions must be named as ToBusinessModel to map DB to business logic models and ToDbModel to map business logic models to DB models.

### 4.6 OpenAPI as the Contract

- All HTTP endpoints must be registered through Huma so they appear in the generated OpenAPI spec at `openapi/LinkShortener.json`. Do not register bare Gin routes that bypass Huma for anything other than the OAuth callback and redirect endpoints (which have special non-JSON response shapes). If in the future we need any other end-points for pure service purposes that are not supposed to be part of the public API, you'll be explicitly instructed what to do. If you find any contradiction about it, report it immediately and start a discussion how to handle the situation.

### 4.7 DRY — Don't Repeat Yourself

- Extract shared logic into helper functions, middleware, or shared packages. Do not copy-paste logic across handlers or components. You still may have some code similarity and even duplication in some places in cases where they cover considerably different functionalities and might be changed independently in the future. If you have doubts about code duplication, report it and start a discussion about whether it should be refactored or left as is.

### 4.8 No Hardcoded Configuration

- All environment-sensitive values (DB connection strings, JWT secret, OAuth client IDs/secrets, quota limits, rate limits, etc.) must be read from environment variables at startup. Provide sensible defaults using the values documented in `app_plan.md`. Never commit secrets. All secrets must be held either in .env.<env_name> or just .env files or in the OS environment variables. env.<env_name> files override .env files if <env_name> environment variable is defined on the OS level. OS environment variables override ones defined in .env* files. It is essential that you DO NOT READ OR OTHERWISE USE any .env* files except initially creating them when you are explicitly instructed to do so. There is NO EXCEPTION to this rule. DO NOT TOUCH any existing files whose names start with .env whatsoever.

---

## 5. Backend Standards

### 5.1 Package Layout

- `business-logic/models/` — business logic models.
- `business-logic/interfaces/` — Go interfaces business layer depends on (repositories, services). These are pure interfaces with no implementation details. Business layer must depend on these interfaces only and NOTHING ELSE.
- `business-logic/handlers/` — business logic handlers. Handlers depend on interfaces, never on concrete implementations directly.
- `infrastructure/pg/` — Concrete Bun-based implementations of repository interfaces. All database interaction lives here.
- `infrastructure/pg/migrations/` — Database migration functions. Each migration is a separate Go function with a unique sequential number prefix (e.g., `001_initial_schema.go`).
- `infrastructure/pg/models/` — Bun model structs that define the database schema. These may differ from business logic models and viewmodels.
- `infrastructure/pg/mappings/` — Conversion functions between DB models and business logic models. `ToBusinessModel` maps a DB model to a business logic model; `ToDbModel` maps a business logic model to a DB model.
- `web/viewmodels/` — Huma input/output structs (request bodies, response schemas). These are not model structs; do not embed Bun model structs inside them.
- `web/mappers/` — Functions to convert between viewmodels and business logic models. Web layer should use these mappers to pass data to business logic handlers and to convert handler responses to viewmodels. Business logic layer should not know anything about viewmodels, so it should not use these mappers directly.

### 5.2 Dependency Direction

```
handlers → interfaces ← infrastructure/pg
         ↘ viewmodels
```

Handlers must not import `infrastructure/pg` directly. Wire dependencies via interfaces.

### 5.3 Huma web layer conventions

- Register every web handler as a Huma operation with a descriptive `OperationID`, `Summary`, and correct HTTP status code(s).
- Validate all input through Huma's declared types (use struct tags like `validate:"required,url"` and `json` tags). Do not repeat validation logic inside the handler body if Huma already enforces it.
- Return structured error responses consistently. Use Huma's error helpers (`huma.Error400BadRequest`, `huma.Error404NotFound`, etc.).

### 5.4 Authentication & JWT

- JWT tokens are stateless. Do not store session state server-side beyond the transient OAuth state cookie.
- JWT claims must include: `user_id`, `user_name`, `sub`, `iat`, `exp`.
- All protected endpoints must be guarded by JWT middleware that validates the `Authorization: Bearer <token>` header.
- `gorilla/sessions` store is used exclusively for the duration of the OAuth2 redirect/callback flow and must be cleared immediately after the callback completes. The `gorilla/sessions` store MUST NEVER BE USED OTHERWISE in the application.

### 5.5 Logging

- Use `log/slog` with a JSON handler targeting stdout.
- Log levels in production: `ERROR`, `WARN`, `INFO`. No `DEBUG` in production builds.
- Every log entry for a request must include at minimum: `user_id` (if authenticated), relevant entity IDs, and the error value (if one occurred).
- DO NOT log sensitive data: passwords, JWT tokens, OAuth secrets, full IP addresses in error-level logs EVER.

### 5.6 Performance

- Never use `SELECT *`. Always specify columns in Bun queries.
- Write URL click records to an in-memory batch buffer and flush to the DB on a schedule (see `app_plan.md` §4 for defaults).
- All list endpoints must be paginated. Default page size is controlled by the `DEFAULT_PAGE_SIZE` env var (default: 20).
- Read connection pool limits from env vars at startup; do not hardcode them.

---

## 6. Frontend Standards

### 6.1 Composition API Only

- All components use `<script setup lang="ts">`. Options API is not permitted.
- Reactive state lives in `ref` or `reactive`. Avoid direct mutation of Pinia state outside of store actions.

### 6.2 API Client

- Never write raw `fetch` or `axios` calls against the backend API. Use the typed client generated from the OpenAPI spec.
- Regenerate the client whenever the backend OpenAPI spec changes.
- The API client should be the single source of truth for all backend interactions. Do not bypass it with direct HTTP calls.
- Whenever compilation errors occur in the frontend after backend changes, check if the API client needs to be regenerated. If you find any contradiction about it, report it immediately and start a discussion about how to handle the situation.

### 6.3 Component Structure

- shadcn/vue components belong in `src/components/`.
- Page-level components (one per route) belong in `src/views/`.
- shadcn/vue components are copied into the project and used "as is" without any customization. If you find the case where sticking to this rule leads to considerable code duplication or complexity, report it immediately and start the discussion about how to handle the situation.

### 6.4 State Management

- Global shared state (auth token, user info, URL list) lives in Pinia stores under `src/stores/`.
- Do not use `localStorage` directly in components. Wrap all persistence in Pinia store actions.
- JWT tokens may be stored in `localStorage` only through the auth store.

### 6.5 Routing & Guards

- All routes that require authentication must have a `beforeEnter` guard (or a global `beforeEach` guard) that checks for a valid JWT via the auth store.
- Unauthenticated users must be redirected to the home page.
- After successful login, users must be redirected to the dashboard page.

### 6.6 Styling

- Use Tailwind CSS utility classes exclusively. Do not write custom CSS except in `src/style.css` for global base styles.
- Do not use inline `style` attributes for anything other than truly dynamic values that cannot be expressed with Tailwind.

---

## 7. Testing Standards

### 7.1 Backend

- Minimum 95% code coverage for all packages under `business-logic/` and `infrastructure/`.
- `main()` and auto-generated code are excluded from coverage requirements.
- Tests follow the **Arrange / Act / Assert** pattern.
- Use mock implementations (generated by `mockgen`) for all repository and service interfaces in handler tests.
- Test files live alongside the code they test (e.g., `handlers/urls_test.go`).

### 7.2 Frontend

- Minimum 80% code coverage for component logic (branches, functions, statements, lines).
- Use `@vue/test-utils` for component mounting. Mock all API client calls with `vi.mock()`.
- Test files use the `.spec.ts` suffix and live alongside the component or store they test.
- CSS, animations, and third-party SDK integrations are excluded from coverage requirements.
- Imported shadcn/vue that are installed into `frontend/src/components/` are excluded from testing as they are used "as is" without any customization. Still they might be covered by tests if they are used in a way that requires some logic to be tested (e.g., conditional rendering based on props). If you find any contradiction about it, report it immediately and start a discussion about how to handle the situation.

### 7.3 When to Write Tests

- Every new handler, repository method, utility function, Pinia store action, and Vue component must be accompanied by tests in the same PR/commit.
- Bug fixes must include a regression test that reproduces the bug before the fix, unless they are already covered by existing tests.

---

## 8. Security Rules (must be implemented before any public release)

- **No passwords stored.** Authentication is exclusively through OAuth2/OIDC third-party providers.
- **Input validation:** All URLs submitted by users must be validated: `http`/`https` scheme only, no `localhost`, no RFC-1918 private IP ranges (SSRF prevention), max length enforced.
- **Shortcode validation:** User-supplied shortcodes must be alphanumeric + hyphens, exactly 6 characters, and checked against a reserved-words blocklist.
- **Rate limiting:** Per-minute click rate limiting is enforced in middleware. Violations are recorded and may trigger temporary account suspension.
- **Monthly quotas:** Redirect processing enforces the monthly click quota per user before recording the click.
- **HTTP-only, SameSite cookies** must be used for the OAuth state session.
- Never expose stack traces or internal error details in API responses. Log them server-side and return a generic error message to the client.

---

## 9. Database Conventions

- Table names are `PascalCase` plural, matching the Bun model struct name (e.g., `Users`, `ShortenedUrls`).
- All tables must have `created_at TIMESTAMPTZ` and (where mutable) `updated_at TIMESTAMPTZ` columns, populated automatically by Bun hooks or default values.
- Foreign keys always declare `ON DELETE CASCADE` unless there is a documented reason to prefer `SET NULL` or `RESTRICT`.
- Every migration is a numbered Go function in `infrastructure/pg/migrations/`. Migrations are append-only; NEVER edit a migration that has already been applied unless you are EXPLICITLY told to do so. If you find any contradiction about it, report it immediately and start a discussion about how to handle the situation.

---

## 10. Git & Commit Conventions

- Commit messages follow the **Conventional Commits** format: `<type>(<scope>): <description>`. Common types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`.
- Each commit should represent one logical, self-contained change.
- ALWAYS MAKE SURE THAT you do not commit secrets, `.env` files, build artifacts, or vendor directories. ESPECIALLY it concerns `.env` files and other secrets. If you have any doubts about whether a particular file you intend to commit contains any secrets, ask explicitly about it before committing. There is NO EXCEPTION to this rule.
- All code must pass `go vet ./...` and `go test ./...` (backend) and `npm run test` (frontend) before being committed.

---

## 11. What Agents Must Not Do

- Do not modify `app_plan.md` unless you are EXPLICITLY INSTRUCTED to update this file and user's instruction contains DIRECT LINK to it. If you find any contradiction between `app_plan.md` and other instructions, report it immediately and start a discussion about how to handle the situation.
- Do not introduce new dependencies or install any packages/libraries without explicit user approval.
- Do not remove or weaken existing input validation or security checks unless you are explicitly told to do so and made sure that the user is aware of potential risks.
- DO NOT hardcode secrets, credentials, or environment-specific URLs.
- Do not write raw SQL (except in migration files where unavoidable).
- Do not bypass Huma registration for API endpoints.
- Do not use the Options API in Vue components.
- Do not alter existing database migrations; create new ones instead.
- Do not silently swallow errors in either Go or TypeScript code.
