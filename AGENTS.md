# Agent Instructions — Link Shortener Project

This file contains coding standards, conventions, and constraints that all AI agents must follow when contributing to this project. Read it in full before making any changes. More granular, topic-specific guidelines live in the `/ai-instructions/` directory as separate markdown files. When a rule in `/ai-instructions/` conflicts with a rule here, the more specific file wins but you must report the contradiction as soon as you found it.

### Topic-specific instruction files

| File                                                                             | Scope      | Description                                                                              |
| -------------------------------------------------------------------------------- | ---------- | ---------------------------------------------------------------------------------------- |
| [ai-instructions/app-auth.md](ai-instructions/app-auth.md)                       | Full-stack | Authentication: OAuth2/OIDC providers, JWT, route guards, sign-in modal UX               |
| [ai-instructions/app-environment.md](ai-instructions/app-environment.md)         | Full-stack | Environment variables: loading algorithm, LINKSHORTENER_ENV, config validation           |
| [ai-instructions/frontend-components.md](ai-instructions/frontend-components.md) | Frontend   | UI components: shadcn/vue-only policy, no custom components without explicit instruction |
| [ai-instructions/backend-dblayer.md](ai-instructions/backend-dblayer.md)         | Backend    | DB layer: Bun ORM, models, migrations, repositories, mappings, click batching            |
| [ai-instructions/backend-business.md](ai-instructions/backend-business.md)       | Backend    | Business logic: models, interfaces, handlers, viewmodels, mappers, URL shortening rules  |
| [ai-instructions/backend-web.md](ai-instructions/backend-web.md)                 | Backend    | Web layer: Huma registration, viewmodels, mappers, error translation, no business logic  |

!!!CRITICAL "Important"!!!
**All agents must adhere to these instructions without exception.** Whenever you do any development of new features, bug fixes, refactoring, or any other code changes, you MUST check .md files in the `/ai-instructions/` directory for any relevant guidelines. If you find there any rules relevant to the task you are doing, you MUST follow them as well as the general rules in this file. If you find any contradictions between different instruction files, you MUST report them immediately and start a discussion about how to handle the situation. If you are unsure about any rule or how to apply it, you MUST ask for clarification before proceeding.

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
├── docker-compose.yaml     ← Local development environment (PostgreSQL, etc.)
├── Makefile                ← Build and development task shortcuts
├── README.md
├── .github/
│   ├── agents/             ← Custom agent definitions
│   ├── hooks/              ← Logging hook configuration
│   └── prompts/            ← Reusable prompt templates
├── commands/               ← Utility scripts (e.g., logging hooks)
├── openapi/
│   └── LinkShortener.json  ← Auto-generated OpenAPI 3.1 spec (do not edit manually)
├── sql/                    ← SQL files generated from migrations (do not edit manually)
├── backend/                ← Go application (Gin + Huma + Bun + PostgreSQL)
│   ├── go.mod
│   ├── main.go
│   ├── config/             ← Environment variable loading and validation
│   ├── cmd/
│   │   └── gensql/         ← Utility to generate SQL from Go migrations
│   ├── business-logic/
│   │   ├── handlers/       ← Pure business logic handlers (no HTTP/framework dependencies)
│   │   ├── interfaces/     ← Repository & service interfaces
│   │   └── models/         ← Business layer models that might differ from DB schema and view-models used on the web layer. They must be used in handlers and appropriate interfaces, but each level might have mapping to convert between them and DB/view or other types of models.
│   ├── infrastructure/
│   │   └── pg/             ← Bun ORM setup and repository implementations
│   │       ├── migrations/ ← Numbered Go migration functions
│   │       ├── models/     ← Bun model structs (DB schema)
│   │       ├── mappings/   ← ToBusinessModel / ToDbModel conversion functions
│   │       └── seed/       ← Seed data for development
│   └── web/
│       ├── viewmodels/     ← Request/response types used by Huma
│       ├── mappers/        ← Viewmodel ↔ business logic model converters
│       └── routes/         ← Huma operation registrations and web handler functions
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

### 4.4 OpenAPI as the Contract

- All HTTP endpoints must be registered through Huma so they appear in the generated OpenAPI spec at `openapi/LinkShortener.json`. Do not register bare Gin routes that bypass Huma for anything other than the OAuth callback and redirect endpoints (which have special non-JSON response shapes). If in the future we need any other end-points for pure service purposes that are not supposed to be part of the public API, you'll be explicitly instructed what to do. If you find any contradiction about it, report it immediately and start a discussion how to handle the situation.

### 4.5 DRY — Don't Repeat Yourself

- Extract shared logic into helper functions, middleware, or shared packages. Do not copy-paste logic across handlers or components. You still may have some code similarity and even duplication in some places in cases where they cover considerably different functionalities and might be changed independently in the future. If you have doubts about code duplication, report it and start a discussion about whether it should be refactored or left as is.

### 4.6 No Hardcoded Configuration

- All environment-sensitive values (DB connection strings, JWT secret, OAuth client IDs/secrets, quota limits, rate limits, etc.) must be read from environment variables at startup. Provide sensible defaults using the values documented in `app_plan.md`. Never commit secrets. All secrets must be held either in .env.<env_name> or just .env files or in the OS environment variables. env.<env_name> files override .env files if LINKSHORTENER_ENV environment variable is defined on the OS level ad <env_name>. OS environment variables override ones defined in .env* files. It is essential that you DO NOT READ OR OTHERWISE USE any .env* files except initially creating them when you are explicitly instructed to do so. There is NO EXCEPTION to this rule. DO NOT TOUCH any existing files whose names start with .env whatsoever.

---

## 5. Backend Standards

### 5.1 Package Layout

- `business-logic/models/` — business logic models.
- `business-logic/interfaces/` — Go interfaces business layer depends on (repositories, services). These are pure interfaces with no implementation details. Business layer must depend on these interfaces only and NOTHING ELSE.
- `business-logic/handlers/` — business logic handlers. Handlers depend on interfaces, never on concrete implementations directly.
- `infrastructure/pg/` — Bun DB setup and connection pool initialization.
- `infrastructure/pg/repositories/` — Concrete Bun-based implementations of repository interfaces. All database interaction lives here.
- `infrastructure/pg/migrations/` — Database migration functions. Each migration is a separate Go function with a unique sequential number prefix (e.g., `001_initial_schema.go`).
- `infrastructure/pg/models/` — Bun model structs that define the database schema. These may differ from business logic models and viewmodels.
- `infrastructure/pg/mappings/` — Conversion functions between DB models and business logic models. `ToBusinessModel` maps a DB model to a business logic model; `ToDbModel` maps a business logic model to a DB model.
- `web/viewmodels/` — Huma input/output structs (request bodies, response schemas).
- `web/mappers/` — Conversion functions between viewmodels and business logic models.
- `web/routes/` — Huma operation registrations and web handler functions.
- Full web layer conventions are in [ai-instructions/backend-web.md](ai-instructions/backend-web.md).

### 5.2 Logging

- Use `log/slog` with a JSON handler targeting stdout.
- Log levels in production: `ERROR`, `WARN`, `INFO`. No `DEBUG` in production builds.
- Every log entry for a request must include at minimum: `user_id` (if authenticated), relevant entity IDs, and the error value (if one occurred).
- DO NOT log sensitive data: passwords, JWT tokens, OAuth secrets, full IP addresses in error-level logs EVER.

### 5.3 Performance

- All list endpoints must be paginated. Default page size is controlled by the `DEFAULT_PAGE_SIZE` env var (default: 20).

---

## 6. Frontend Standards

### 6.1 Reactive State

- Reactive state lives in `ref` or `reactive`. Avoid direct mutation of Pinia state outside of store actions.

### 6.2 API Client

- Never write raw `fetch` or `axios` calls against the backend API. Use the typed client generated from the OpenAPI spec.
- Regenerate the client whenever the backend OpenAPI spec changes.
- The API client should be the single source of truth for all backend interactions. Do not bypass it with direct HTTP calls.
- Whenever compilation errors occur in the frontend after backend changes, check if the API client needs to be regenerated. If you find any contradiction about it, report it immediately and start a discussion about how to handle the situation.

### 6.3 Component Structure

- shadcn/vue components belong in `src/components/`.
- Page-level components (one per route) belong in `src/views/`.

### 6.4 State Management

- Global shared state (auth token, user info, URL list) lives in Pinia stores under `src/stores/`.
- Do not use `localStorage` directly in components. Wrap all persistence in Pinia store actions.

---

## 7. Testing Standards

### 7.1 Backend

- Minimum 95% code coverage for all packages under `business-logic/` and `infrastructure/`.
- `main()` and auto-generated code are excluded from coverage requirements.
- Tests follow the **Arrange / Act / Assert** pattern.
- Test files live alongside the code they test (e.g., `handlers/urls_test.go`).

### 7.2 Frontend

- Minimum 80% code coverage for component logic (branches, functions, statements, lines).
- Use `@vue/test-utils` for component mounting. Mock all API client calls with `vi.mock()`.
- Test files use the `.spec.ts` suffix and live alongside the component or store they test.
- CSS, animations, and third-party SDK integrations are excluded from coverage requirements.

### 7.3 When to Write Tests

- Every new handler, repository method, utility function, Pinia store action, and Vue component must be accompanied by tests in the same PR/commit.
- Bug fixes must include a regression test that reproduces the bug before the fix, unless they are already covered by existing tests.

---

## 8. Security Rules (must be implemented before any public release)

- **Input validation:** All URLs submitted by users must be validated: `http`/`https` scheme only, no `localhost`, no RFC-1918 private IP ranges (SSRF prevention), max length enforced.
- **Shortcode validation:** User-supplied shortcodes must be alphanumeric + hyphens, exactly 6 characters, and checked against a reserved-words blocklist.
- **Rate limiting:** Per-minute click rate limiting is enforced in middleware. Violations are recorded and may trigger temporary account suspension.
- **Monthly quotas:** Redirect processing enforces the monthly click quota per user before recording the click.
- Never expose stack traces or internal error details in API responses. Log them server-side and return a generic error message to the client.

---

## 9. Git & Commit Conventions

- Commit messages follow the **Conventional Commits** format: `<type>(<scope>): <description>`. Common types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`.
- Each commit should represent one logical, self-contained change.
- ALWAYS MAKE SURE THAT you do not commit secrets, `.env` files, build artifacts, or vendor directories. ESPECIALLY it concerns `.env` files and other secrets. If you have any doubts about whether a particular file you intend to commit contains any secrets, ask explicitly about it before committing. There is NO EXCEPTION to this rule.
- All code must pass `go vet ./...` and `go test ./...` (backend) and `npm run test` (frontend) before being committed.

---

## 10. What Agents Must Not Do

- Do not modify `app_plan.md` unless you are EXPLICITLY INSTRUCTED to update this file and user's instruction contains DIRECT LINK to it. If you find any contradiction between `app_plan.md` and other instructions, report it immediately and start a discussion about how to handle the situation.
- Do not introduce new dependencies or install any packages/libraries without explicit user approval.
- Do not remove or weaken existing input validation or security checks unless you are explicitly told to do so and made sure that the user is aware of potential risks.
- DO NOT hardcode secrets, credentials, or environment-specific URLs.
- Do not bypass Huma registration for API endpoints.
- Do not silently swallow errors in either Go or TypeScript code.
