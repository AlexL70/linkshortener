# Authentication — Agent Instructions

> **Scope:** Full-stack (backend + frontend)
> These coding standards and conventions **must be strictly followed** whenever you work on any authentication-related code, regardless of the nature of the work (new features, bug fixes, refactoring, or any other changes). If you find any contradiction between the rules in this file and those in any other instruction file (including `AGENTS.md`), **report it immediately and start a discussion** before making any changes.

---

## 1. Auth Method — Non-Negotiable Constraints

- **Third-party OAuth2/OIDC providers are the only permitted authentication method.** No exceptions.
- **Never implement or accept passwords**, magic links, email/OTP codes, API-key-based user auth, or any other form of credential-based authentication. If a task description implies adding such a mechanism, refuse and flag it to the user.
- No password fields may be added to the database schema, viewmodels, or any API contracts.
- The MVP providers are **Google**, **Microsoft**, and **Facebook**. The architecture must remain open to adding further providers (GitHub, Apple, LinkedIn, etc.) without schema changes.

---

## 2. Backend Auth Architecture

### 2.1 Libraries

| Concern                         | Technology                                          |
| ------------------------------- | --------------------------------------------------- |
| OAuth2/OIDC provider management | `github.com/markbates/goth` + `gothic` sub-package  |
| Transient OAuth state           | `github.com/gorilla/sessions` (signed cookie store) |
| Post-auth session tokens        | `github.com/golang-jwt/jwt/v5`                      |

Do not replace or supplement any of these with alternative libraries without explicit user approval.

### 2.2 OAuth2 Flow

1. Frontend user initiates login → backend `GET /auth/login/{provider}` starts the OAuth2 flow via `gothic.BeginAuthHandler`.
2. Provider redirects back to `GET /auth/callback` with the authorization code.
3. Backend exchanges the code for user info, looks up or creates the user in the `UserProviders` + `Users` tables, and issues a JWT.
4. The `gorilla/sessions` cookie store is used **only** to hold the transient OAuth state between the redirect and the callback. It must be cleared immediately after the callback completes. **It must not be used for any other purpose.**

### 2.3 JWT Tokens

- JWT tokens are **stateless** — do not store server-side session state after the OAuth callback completes.
- Required claims: `user_id`, `user_name`, `provider_email`, `sub`, `iat`, `exp`.
  - `provider_email` is the email address returned by the OAuth provider at the time of login. It is included so the frontend `/auth/me` endpoint can return it without a separate DB look-up.
- **JWT tokens are delivered exclusively as `HttpOnly` cookies** — never in the response body, URL fragments, or `Authorization` headers.
- All protected endpoints must be guarded by JWT middleware that reads the session cookie (`linkshortener_session`) from the incoming request. The middleware must **not** accept `Authorization: Bearer` headers.
- JWT secret must be read from the `JWT_SECRET` environment variable at startup. Never hardcode it.

### 2.4 OAuth Endpoints

- `GET /auth/login/{provider}` — public; initiates the OAuth2 flow.
- `GET /auth/callback` — public; completes the OAuth2 flow, sets the session cookie, then redirects to the frontend `/auth/callback` page (no `#token=` fragment). For new users, redirects with `#pre_registration_token=...&suggested_user_name=...`.
- `POST /auth/register` — public (pre-registration token required); creates user, sets session cookie, returns 204.
- `POST /auth/logout` — session-cookie-protected; blacklists the JTI and clears the session cookie (Max-Age=0).
- `GET /auth/me` — session-cookie-protected; returns current user info (user_id, user_name, provider_email). Used by the frontend to restore auth state on page load.
- These endpoints have non-standard response shapes and **may** be registered as bare Gin routes (bypassing Huma), as permitted by `AGENTS.md` §4.6, **except for `GET /auth/me`**, `POST /auth/register`, and `POST /auth/logout`, which have standard JSON response shapes and must be registered as Huma operations.

### 2.5 Account Management Endpoints

- `DELETE /user/account` — JWT-protected; permanently deletes the authenticated user's account and all associated data. Must be registered as a Huma operation (standard JSON response shape). The business handler must enforce that the super-admin account (identified by `SUPER_ADMIN_EMAIL`) cannot be deleted via this endpoint — return `ErrUnauthorized` (maps to `403 Forbidden`) if the requesting user is the super-admin.

### 2.5 New-User Registration

- When a callback completes for a user not yet in the database, redirect to the frontend registration form with provider info pre-filled (username suggestion from OAuth profile).
- The backend `POST /auth/register` endpoint creates the `Users` and `UserProviders` records and returns a JWT.
- Username must be unique in the `Users` table. If the provider-supplied name is taken, the user must be prompted to choose a different one.

---

## 3. Frontend Auth Architecture

### 3.1 Auth Store (`src/stores/auth.ts`)

- All auth state (logged-in user info) must live in the Pinia auth store.
- **JWT tokens must never be stored in `localStorage`, `sessionStorage`, or any JavaScript-readable location.** The token lives exclusively in the `HttpOnly` session cookie managed by the browser and server.
- Expose actions: `login(provider)`, `logout()`, `fetchMe()`, `deleteAccount()`, `isAuthenticated` (computed getter).
- `fetchMe()` calls `GET /auth/me` to restore user state from the session cookie. It must be called on application startup (before mounting) and after successful registration.
- The store must expose `user.email` (returned by `/auth/me` as `provider_email`) as a readable property. This value is used by the account deletion confirmation dialog.

### 3.2 Sign-In / Sign-Up UX

- **Sign-in and sign-up are the same user-facing action** — a new user is registered automatically on first login.
- Triggering either action **must always open a modal dialogue** — never navigate to a dedicated `/login` or `/signup` page.
- The modal must present all supported OAuth2 provider buttons (Google, Microsoft, Facebook). Clicking a provider button redirects the browser to `GET /auth/login/{provider}`.
- Do not render an inline login form on the home page or any other page outside the modal.

### 3.3 Route Guards

| Route                             | Rule                                                                                              |
| --------------------------------- | ------------------------------------------------------------------------------------------------- |
| `/dashboard` (and all sub-routes) | **Protected** — requires a valid JWT. Unauthenticated visitors are redirected to `/` (home page). |
| `/` (home page)                   | **Public** — but if the user **is already authenticated**, redirect immediately to `/dashboard`.  |
| All other public routes           | No authentication check required unless otherwise specified.                                      |

Guards must be implemented using Vue Router's `beforeEach` global navigation guard (or per-route `beforeEnter`) reading state from the auth store. Do not duplicate guard logic across routes.

### 3.4 Post-Auth Redirect

- After a successful OAuth2 callback (new or returning user), the frontend must redirect the user to `/dashboard`.
- After logout, redirect the user to `/` (home page).
- After a successful account deletion (`deleteAccount()` action), clear all auth store state (JWT and user info), then redirect the user to `/` (home page). A brief informational message acknowledging the deletion must be displayed (e.g., a toast notification).

---

## 4. Security Checklist

- OAuth state cookie must be `HttpOnly` and `SameSite=Lax` (or `Strict`).
- **Session cookie containing the JWT must be `HttpOnly`, `SameSite=Strict`, `Secure` (in non-development environments), `Path=/`, and `MaxAge=86400`.** The cookie name is `linkshortener_session`.
- The frontend **must** send all API requests with `credentials: 'include'` so that the session cookie is transmitted cross-origin. The generated API client `OpenAPI.WITH_CREDENTIALS` must be `true`.
- Never log JWT tokens, OAuth secrets, or raw authorization codes — not even at DEBUG level.
- Never expose internal auth errors (e.g., OAuth exchange failures) in API responses. Log server-side, return a generic error to the client.
- Provider client IDs and secrets must be read from environment variables; never hardcode them.
- The super-admin account must never be self-deletable. `DELETE /user/account` must return `403 Forbidden` when called by the user whose `provider_email` matches `SUPER_ADMIN_EMAIL`. This check belongs in the business handler, not the web layer.
