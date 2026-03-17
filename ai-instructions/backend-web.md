# Backend Web Layer — Agent Instructions

> **Scope:** Backend
> These coding standards and conventions **must be strictly followed** whenever you work on any web-layer code, regardless of the nature of the work (new features, bug fixes, refactoring, or any other changes). If you find any contradiction between the rules in this file and those in any other instruction file (including `AGENTS.md`), **report it immediately and start a discussion** before making any changes.

---

## 1. Core Principle — Strict Separation of Concerns

**The web layer must contain zero business logic.** Its only responsibilities are:

1. Declare the HTTP contract (Huma registration, viewmodels, validation tags).
2. Validate and deserialize the incoming request via Huma.
3. Map the validated viewmodel to a business logic model using a mapper function.
4. Call the appropriate business logic handler method.
5. Map the handler's result back to a viewmodel (or translate an error into an HTTP error response).
6. Return the HTTP response.

Any logic that goes beyond these six steps belongs in `business-logic/handlers/`. If you are unsure whether something is "business logic", **do not make assumptions — discuss it with the user and act according to their response.**

---

## 2. Directory Layout

All web layer code lives under `backend/web/`:

| Path              | Purpose                                                           |
| ----------------- | ----------------------------------------------------------------- |
| `web/viewmodels/` | Huma input/output structs (request bodies, response schemas)      |
| `web/mappers/`    | Conversion functions between viewmodels and business logic models |
| `web/routes/`     | Huma operation registrations and web handler functions            |

Do not place viewmodels, mappers, or route registrations anywhere outside these directories.

---

## 3. Viewmodels

- Viewmodels are plain Go structs with `json` and `validate` struct tags — and **nothing else**.
- Never embed Bun model structs or business logic model structs inside viewmodels.
- Separate request and response types even when their fields overlap. Name them clearly: `CreateUrlRequest`, `CreateUrlResponse`, `UrlListResponse`, etc.
- Use Huma's supported validation tags (`validate:"required"`, `validate:"url"`, `minLength`, `maxLength`, `pattern`, etc.) to express all constraints that can be described declaratively. This keeps validation out of handler bodies.
- Input structs must embed `huma.Operation` inputs as required by the Huma API (i.e. declare them as `Body *YourRequestBody`).

Example:

```go
type CreateUrlRequest struct {
    Body *CreateUrlRequestBody
}

type CreateUrlRequestBody struct {
    LongUrl   string  `json:"long_url"   validate:"required,url" maxLength:"2048"`
    Shortcode *string `json:"shortcode,omitempty" minLength:"6" maxLength:"6"`
    ExpiresAt *string `json:"expires_at,omitempty"`
}

type CreateUrlResponse struct {
    Body *CreateUrlResponseBody
}

type CreateUrlResponseBody struct {
    Shortcode string `json:"shortcode"`
    LongUrl   string `json:"long_url"`
    ShortUrl  string `json:"short_url"`
}
```

---

## 4. Mappers

- Every viewmodel ↔ business logic model conversion must be implemented as an explicit function in `web/mappers/`.
- Naming convention:
  - `ToBusinessModel(vm *viewmodels.Foo) *bizmodels.Foo` — viewmodel to business logic model.
  - `ToViewModel(m *bizmodels.Foo) *viewmodels.FooResponse` — business logic model to response viewmodel.
- Mapper functions must be pure (no side effects, no error returns unless type conversion can meaningfully fail).
- Business logic handlers must never import anything from `web/`. Mappers are called exclusively by web layer route handlers.

---

## 5. Huma Route Handlers

### 5.1 Registration

- Register every endpoint as a Huma operation. Do **not** register bare Gin routes for API endpoints. The only permitted exceptions are the OAuth callback and redirect endpoints (see [app-auth.md](app-auth.md)).
- Every registration must include:
  - A unique, descriptive `OperationID` (e.g., `"create-shortened-url"`)
  - A human-readable `Summary`
  - The correct HTTP method, path, and expected status code(s)

Example:

```go
huma.Register(api, huma.Operation{
    OperationID: "create-shortened-url",
    Method:      http.MethodPost,
    Path:        "/urls",
    Summary:     "Create a new shortened URL",
}, createUrlHandler)
```

Account deletion example (returns `204 No Content` with an empty body on success):

```go
huma.Register(api, huma.Operation{
    OperationID:  "delete-account",
    Method:       http.MethodDelete,
    Path:         "/user/account",
    Summary:      "Permanently delete the authenticated user's account and all associated data",
    DefaultStatus: http.StatusNoContent,
}, deleteAccountHandler)
```

### 5.2 Handler Function Structure

Every web handler function must follow this exact pattern — no deviations:

```go
func createUrlHandler(ctx context.Context, input *viewmodels.CreateUrlRequest) (*viewmodels.CreateUrlResponse, error) {
    // 1. Map viewmodel → business model
    bizInput := mappers.ToBusinessModel(input.Body)

    // 2. Call business logic handler
    result, err := h.urlHandler.CreateUrl(ctx, bizInput)
    if err != nil {
        return nil, mapError(err)
    }

    // 3. Map business model → response viewmodel
    return &viewmodels.CreateUrlResponse{Body: mappers.ToViewModel(result)}, nil
}
```

- Do **not** add conditional logic, computations, or data transformations outside of these three steps.
- Do **not** call repository or infrastructure code directly from a web handler. All calls go through business logic handlers.

### 5.3 Authentication & Authorization

- All protected endpoints must extract the authenticated user from the JWT middleware context and pass the `user_id` to the business logic handler. Never perform access checks in the web handler itself — delegate to the business handler.

---

## 6. Error Mapping

- Translate business logic sentinel errors to Huma HTTP errors in a single, shared `mapError` function located in `web/routes/`. Do not inline this mapping in individual handlers.
- The mapping table must cover at minimum:

| Business Error    | HTTP Response                                                            |
| ----------------- | ------------------------------------------------------------------------ |
| `ErrNotFound`     | `huma.Error404NotFound`                                                  |
| `ErrConflict`     | `huma.Error409Conflict`                                                  |
| `ErrValidation`   | `huma.Error400BadRequest`                                                |
| `ErrExpired`      | `huma.Error410Gone`                                                      |
| `ErrUnauthorized` | `huma.Error403Forbidden`                                                 |
| Any other error   | `huma.Error500InternalServerError` (generic message; log the real error) |

- Never expose internal error messages, stack traces, or infrastructure details in HTTP responses. Log them server-side at `ERROR` level; return a generic user-facing message.

---

## 7. Logging

- Log at the web layer only for cross-cutting concerns (e.g., unhandled/unexpected errors mapped to 500). Business-level events are logged inside handlers.
- Every error log entry must include `user_id` (when authenticated) and relevant entity IDs.
- Never log JWT tokens, OAuth secrets, request bodies containing sensitive fields, or full IP addresses at `ERROR` level.

---

## 8. Pagination

- All list endpoints must accept `page` and `page_size` query parameters.
- `page_size` defaults to the `DEFAULT_PAGE_SIZE` environment variable (default: 20) when not supplied by the caller.
- Pass pagination parameters to the business handler as part of the business model input; do not apply pagination in the web layer.

---

## 9. Testing

- Minimum **95% code coverage** for all packages under `web/`.
- Tests follow the **Arrange / Act / Assert** pattern and live alongside the code they test.
- Mock all business logic handler dependencies using `mockgen`-generated mocks.
- Do not test business logic rules (URL validation, shortcode format, SSRF checks, etc.) in web layer tests — those belong in `business-logic/` tests. Web layer tests verify request/response mapping, Huma registration, and error translation only.
