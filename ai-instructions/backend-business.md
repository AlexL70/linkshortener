# Backend Business Logic — Agent Instructions

> **Scope:** Backend
> These coding standards and conventions **must be strictly followed** whenever you work on any business logic code, regardless of the nature of the work (new features, bug fixes, refactoring, or any other changes). If you find any contradiction between the rules in this file and those in any other instruction file (including `AGENTS.md`), **report it immediately and start a discussion** before making any changes.

---

## 1. Directory Layout

Business logic code lives under `backend/business-logic/`:

| Path                         | Purpose                                                                        |
| ---------------------------- | ------------------------------------------------------------------------------ |
| `business-logic/models/`     | Pure Go domain model structs (no ORM or HTTP tags)                             |
| `business-logic/interfaces/` | Pure Go interfaces: repository contracts, service contracts                    |
| `business-logic/handlers/`   | Pure Go business logic handlers; no HTTP or third-party framework dependencies |

Auth-specific handler conventions are governed by [app-auth.md](app-auth.md). DB layer conventions are governed by [backend-dblayer.md](backend-dblayer.md). Web layer conventions (viewmodels, mappers) are governed by [backend-web.md](backend-web.md).

---

## 2. Business Logic Models

- Models in `business-logic/models/` are **plain Go structs** — no `bun`, `json`, or `validate` struct tags.
- They represent domain concepts only. A model struct must not know about database storage, HTTP serialization, or any external framework.
- They are the single shared vocabulary between handlers and repository/service interfaces. All interface method signatures must use these types.
- When a domain concept requires different shapes at the DB or HTTP layer, create a separate DB model or viewmodel and add a mapping function. Do not add tags or fields to the business model to accommodate those layers.

---

## 3. Interfaces

- All interfaces live in `business-logic/interfaces/`. They must be pure Go interfaces with no imports from `infrastructure/` or `web/`.
- Define one interface per logical concern (e.g., `UrlRepository`, `UserRepository`, `ClickWriter`, `ShortcodeGenerator`).
- Interface methods must only accept and return types from `business-logic/models/`, standard library types (`context.Context`, `error`, primitives), or other interface types. They must never reference Bun model structs, Huma types, or viewmodels.
- Keep interfaces narrow: prefer several small interfaces over one large one.

---

## 4. Handlers

- Handlers in `business-logic/handlers/` are **pure Go service structs** — they apply business logic, orchestrate repository and service calls, and return business models or errors. They have no knowledge of HTTP, Huma, Gin, or any web framework.
- Handlers receive their dependencies (repository interfaces, service interfaces) by constructor injection — **never** by global variables or `init()`:

  ```go
  type UrlHandler struct {
      urls    interfaces.UrlRepository
      clicks  interfaces.ClickWriter
  }

  func NewUrlHandler(urls interfaces.UrlRepository, clicks interfaces.ClickWriter) *UrlHandler {
      return &UrlHandler{urls: urls, clicks: clicks}
  }
  ```

- Return standard Go `error` values. Declare **sentinel errors** or typed error structs in `business-logic/` for predictable domain failures (e.g., `ErrNotFound`, `ErrExpired`, `ErrConflict`, `ErrValidation`). The web layer is responsible for translating these into HTTP responses.
- All domain-level validation (URL format, shortcode rules, SSRF checks, expiration) must be performed in the handler or in a dedicated validation helper within `business-logic/`.
- Every log entry must include `user_id` (when authenticated) and all relevant entity IDs. Never log JWT tokens, OAuth secrets, or any sensitive user data.

---

## 5. URL Shortening Logic

### 5.1 Shortcode Generation

- Shortcodes are generated using **Base62 encoding** (`[0-9A-Za-z]`, exactly 6 characters).
- If a user supplies a custom shortcode it must be validated: alphanumeric + hyphens only, exactly 6 characters, checked against a reserved-words blocklist. Return `ErrValidation` on failure.
- On shortcode collision (auto-generated), retry generation up to a reasonable limit; if all retries fail, return a wrapped error and log the event.
- Shortcode generation must be encapsulated behind an interface in `business-logic/interfaces/` so it can be mocked in tests.

### 5.2 Long URL Validation

Every long URL submitted by a user must be validated before storing:

1. Scheme must be `http` or `https`. Return `ErrValidation` otherwise.
2. Host must not be `localhost`, `127.0.0.1`, `::1`, or any loopback address.
3. Host must not resolve to an RFC-1918 private IP range (`10.x.x.x`, `172.16–31.x.x`, `192.168.x.x`) — SSRF prevention.
4. URL length must not exceed 2048 characters.

Report any exception requests to this rule to the user before implementing them.

### 5.3 Expiration

- `expires_at` is optional. When set, the handler must check it and return `ErrExpired` for expired links. The web layer maps `ErrExpired` to `410 Gone`.
- Do not silently process expired links. Log the expiration event at `INFO` level.

---

## 6. Dependency Injection & Wiring

- All interface implementations are wired at application startup (`main.go` or a dedicated `wire.go`). No handler or business logic code constructs its own concrete dependencies.
- Use constructor functions for every handler and service, accepting their interface dependencies as parameters.
- Do not use global variables to hold repository or service instances. Pass them through constructors.

---

## 7. Testing

- Minimum **95% code coverage** for all packages under `business-logic/`.
- Generate mock implementations with `mockgen` for all interfaces in `business-logic/interfaces/`. Place generated mocks in a `mocks/` sub-package alongside the interfaces (e.g., `business-logic/interfaces/mocks/`).
- Handler tests must use these mocks — never real repository implementations.
- Tests follow the **Arrange / Act / Assert** pattern.
- Test files live alongside the code they test (e.g., `business-logic/handlers/url_handler_test.go`).
- `main()` and auto-generated mock files are excluded from the 95% coverage requirement.

---

## 8. What Agents Must Not Do

- Do not import any third-party package (Huma, Gin, Bun, etc.) from `business-logic/`. The only permitted imports are the Go standard library and other `business-logic/` packages.
- Do not import `infrastructure/pg` (or any sub-package) from `business-logic/`. Always wire through interfaces.
- Do not add ORM tags (`bun:"..."`) or HTTP tags (`json:"..."`, `validate:"..."`) to business model structs.
- Do not return HTTP status codes or HTTP-specific error types from handlers; return sentinel or typed errors and let the web layer handle translation.
- Do not process expired links silently — always return `ErrExpired` and log the event.
- Do not construct concrete infrastructure dependencies inside handlers or business logic; always inject via interfaces.
