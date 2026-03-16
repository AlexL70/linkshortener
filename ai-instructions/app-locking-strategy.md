# Optimistic Locking Strategy — Agent Instructions

> **Scope:** Full-stack
> These coding standards and conventions **must be strictly followed** whenever you work on any code that involves editing or deleting mutable entities, regardless of the nature of the work (new features, bug fixes, refactoring, or any other changes). If you find any contradiction between the rules in this file and those in any other instruction file (including `AGENTS.md`), **report it immediately and start a discussion** before making any changes.

---

## 1. Overview

This application uses **optimistic concurrency control (OCC)** to prevent lost updates when multiple users (or browser tabs) concurrently edit or delete the same entity. Rather than holding locks while a user fills in a form, the system records _when_ the entity was last modified and rejects any write whose version precondition is stale.

The strategy applies to all mutable entities (currently `ShortenedUrl`). Apply the same pattern to every new mutable entity added in the future.

---

## 2. Business Model — `LastUpdated` Field

- Business models for mutable entities **must not** expose two separate audit fields (`CreatedAt` / `UpdatedAt`).
- Instead, every mutable business model must have exactly one version field:

  ```go
  LastUpdated time.Time
  ```

- `LastUpdated` maps to the `updated_at` column in the corresponding DB model. `CreatedAt` is required in DB models (per [backend-dblayer.md](backend-dblayer.md)) but **must not** be present in business models unless a feature explicitly requires it.
- The mapping function in `infrastructure/pg/mappings/` must assign `dbModel.UpdatedAt → bizModel.LastUpdated`.

---

## 3. Read Flow — Load for Edit or Delete

Whenever a user initiates an edit or a delete, the **latest version** of the entity must be fetched fresh from the database at that moment — never rely on a version already loaded for display purposes.

1. The business handler calls the repository read method (e.g., `GetByID`) immediately before returning the entity to the caller.
2. The returned entity — including `LastUpdated` — is passed through the web layer to the frontend via the response viewmodel.
3. The frontend stores the received `last_updated` value and includes it verbatim in every subsequent save or delete request for that entity.

---

## 4. Viewmodels — Carrying `LastUpdated`

- All **update and delete request** viewmodels **must** include a `last_updated` field:

  ```go
  LastUpdated time.Time `json:"last_updated" validate:"required"`
  ```

- All **response** viewmodels for OCC-protected entities **must** include `last_updated` so the frontend always has the current value after a successful write.
- Mappers must propagate `LastUpdated` in both directions (viewmodel ↔ business model).

---

## 5. Repository — Atomic Optimistic Lock Check

The lock check must be **atomic** to prevent TOCTOU (time-of-check / time-of-use) races. Use a conditional `WHERE` clause rather than a separate pre-read followed by a write.

### 5.1 Update

Include the OCC condition in the `WHERE` clause and check rows affected:

```go
result, err := db.NewUpdate().
    Model(&dbModel).
    Where("id = ? AND updated_at = ?", id, lastUpdated).
    Exec(ctx)
if err != nil {
    return fmt.Errorf("UrlRepository.Update: %w", err)
}

n, err := result.RowsAffected()
if err != nil {
    return fmt.Errorf("UrlRepository.Update rows-affected: %w", err)
}
if n == 0 {
    exists, checkErr := r.exists(ctx, id)
    if checkErr != nil {
        return fmt.Errorf("UrlRepository.Update exists-check: %w", checkErr)
    }
    if !exists {
        return businesslogic.ErrNotFound
    }
    return businesslogic.ErrVersionConflict
}
```

### 5.2 Delete

Apply the same pattern:

```go
result, err := db.NewDelete().
    Model((*models.ShortenedUrl)(nil)).
    Where("id = ? AND updated_at = ?", id, lastUpdated).
    Exec(ctx)
// ... identical rows-affected check: ErrNotFound if gone, ErrVersionConflict if modified
```

### 5.3 No Separate Pre-Read

Do **not** add a separate `SELECT` before every `UPDATE`/`DELETE` solely for the version check. The conditional `WHERE` clause achieves the same result atomically and with fewer round-trips.

---

## 6. Sentinel Error

Add a new sentinel error to `business-logic/errors.go`:

```go
// ErrVersionConflict is returned when an entity was modified by another user
// since it was last read. The caller should ask the user to refresh and retry.
ErrVersionConflict = errors.New("version conflict")
```

Do **not** reuse `ErrConflict` for this purpose. `ErrConflict` is reserved for uniqueness constraint violations (e.g., shortcode collisions). Keeping them distinct allows the web layer and frontend to give users specific, actionable messages.

---

## 7. Web Layer Error Mapping

Add `ErrVersionConflict` to the shared `mapError` function in `web/routes/`:

| Business Error       | HTTP Status    | User-facing message                                                |
| -------------------- | -------------- | ------------------------------------------------------------------ |
| `ErrVersionConflict` | `409 Conflict` | `"This item was modified by another user. Refresh and try again."` |

All other error-mapping rules remain as defined in [backend-web.md](backend-web.md).

---

## 8. Frontend — Handling Conflict Responses

- When the API returns `409 Conflict` with the OCC message, display a visible, user-friendly toast or inline alert: _"This item was recently changed by someone else. Please refresh and try again."_
- After showing the conflict error, **reload the entity from the server** so the user sees the latest version before attempting another edit.
- Never silently discard a `409` response. The user must be informed so they can make an informed decision about whether to re-apply their changes.

---

## 9. Testing Requirements

### Backend

- Every repository update/delete method with OCC must have tests covering:
  1. Successful write when `LastUpdated` matches the current DB value.
  2. `ErrVersionConflict` returned when `LastUpdated` does not match.
  3. `ErrNotFound` returned when the entity no longer exists at all.
- Business handler tests must mock the repository and verify that `ErrVersionConflict` is propagated without modification.

### Frontend

- Component tests for edit and delete flows must verify that a `409` response causes the conflict message to be displayed and triggers a reload of the entity.
