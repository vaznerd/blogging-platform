# Blogging Platform — AI-Friendly Development Guide

**Purpose**: Universal AI assistant guidelines for the Blogging Platform project.

---

## Project Overview

A multi-user blogging platform where users write posts in Markdown and publish them under their own profile. Written in Go with standard library `net/http` (no framework), pgx for PostgreSQL, and Redis for caching.

### Technology Stack

- **Go 1.26+** — check with `go version`
- **stdlib `net/http`** — HTTP server and routing (Go 1.22+ method-based patterns)
- **pgx/v5** — PostgreSQL driver
- **Redis (go-redis/v9)** — caching layer (connected but not yet used in business logic)
- **golang-jwt/jwt/v5** — JWT authentication (HS256)
- **golang-migrate/migrate** — database migrations
- **Resend (resend-go/v3)** — transactional email
- **Air** — hot-reload development
- **golangci-lint** — code quality
- **koanf** — configuration management (YAML + env overrides)
- **slog** — structured logging
- **bcrypt** — password hashing (`golang.org/x/crypto`)

### Module

```
module codeberg.org/vaznerd/blogging-platform
```

### Documentation

- **Source**: `readme.md` (project root)
- **Work tracking**: `TODO.md` (project root)

---

## Architecture

### Layered Pattern

```
Handler (HTTP) → Service (Business Logic) → Repository (Database)
```

### Domain Package Structure

```
internal/<domain>/
├── errors.go           # Sentinel errors (ErrNotFound, ErrForbidden, etc.)
├── routes.go           # Path constants (RouteXxx = "/api/v1/...")
├── router.go           # RegisterRoutes(mux, service, log, mail, authMW)
├── handler.go          # HTTP handlers
├── service.go          # Business logic
└── repository.go       # Repository interface + concrete pgx implementation
```

**Key Rules**:
- Handlers parse requests and write responses only
- Services contain all business logic
- Repositories only interact with database
- No skipping layers or crossing boundaries
- No framework — use `http.ServeMux` with `"METHOD /path"` patterns

### Route Constants Pattern

Each domain defines its own path constants in `routes.go`:

```go
package auth

const (
    RouteRegister = "/api/v1/auth/register"
    RouteLogin    = "/api/v1/auth/login"
)
```

Used in `router.go`:

```go
mux.HandleFunc("POST "+RouteRegister, h.Register)
```

---

## Implementation Status

| Domain | errors | routes | router | handler | service | repo | Status |
|--------|--------|--------|--------|---------|---------|------|--------|
| auth | done | done | done | done | done | done | **FULLY IMPLEMENTED** |
| user | done | done | done | done | done | done | **PARTIAL** — `Me` works; `GetUser`, `UpdateMe`, `DeleteMe` are stubs returning 501 |
| category | done | done | done | stub | done | done | **HANDLERS STUB** — repo+service done |
| tag | done | done | done | stub | done | done | **HANDLERS STUB** — repo+service done |
| post | done | done | — | — | — | — | **SCAFFOLD ONLY** — just errors + routes |
| comment | done | done | — | — | — | — | **SCAFFOLD ONLY** — just errors + routes |

---

## Development Workflow

### Available Commands

```bash
make dev-up            # Start dev containers (postgres, redis)
make run-backend       # Start Go API server with Air hot reload
make lint-backend      # Run go vet
make format-backend    # Format code with gofumpt
make dev-down          # Stop dev containers
make dev-down-force    # Stop and remove volumes
make dev-logs          # View container logs
make dev-migrate-up    # Apply pending migrations
make dev-migrate-down  # Rollback last migration
make dev-migrate-version # Show current migration version
```

### Pre-Commit Checklist

```bash
make format-backend
make lint-backend
go build ./...
```

---

## Common Tasks

### Adding a New Domain

1. Create `internal/<domain>/` directory
2. Create `errors.go` — sentinel errors (`ErrNotFound`, `ErrForbidden`, etc.)
3. Create `routes.go` — define path constants
4. Create `handler.go` — handler struct + constructor (stub methods)
5. Create `service.go` — service struct + constructor (stub methods)
6. Create `repository.go` — repository interface + struct + constructor
7. Create `router.go` — `RegisterRoutes(mux, service, log, mail)`
8. Register in `internal/server/router.go`
9. Wire dependencies in `cmd/server/main.go`

### Adding a New Handler Method

```go
func (h *Handler) CreateWidget(w http.ResponseWriter, r *http.Request) {
    // ...
}
```

```go
mux.HandleFunc("POST "+RouteWidgets, h.CreateWidget)
```

If protected, wrap with auth middleware:

```go
authMW := middleware.Auth(authService.ValidateToken, log)
mux.Handle("POST "+RouteWidgets, authMW(http.HandlerFunc(h.CreateWidget)))
```

### Database Migrations

Migrations use golang-migrate with sequential numbering:

```bash
make dev-migrate-up        # Apply pending
make dev-migrate-down      # Rollback last
make dev-migrate-version   # Show current version
```

Migration files live in `backend/migrations/`:

```
{NNN}_{description}.up.sql
{NNN}_{description}.down.sql
```

Current migrations: 001–009 (users, sessions, posts, comments, tags, post_tags, categories, email_verification, password_reset_tokens).

Best practices:
- Wrap in `BEGIN;` / `COMMIT;`
- Use `IF NOT EXISTS`
- Create indexes for foreign keys
- Always write corresponding `.down.sql`

---

## Authentication & Authorization

### Session Management

The auth domain uses a hybrid JWT + database sessions approach:

- **Access tokens** — pure JWT (HS256), validated by signature only (no DB lookup). Contains `sub`, `role`, `exp`, `iat`.
- **Refresh tokens** — JWT signed, SHA-256 hashed and stored in the `sessions` table. Contains `sub`, `type:"refresh"`, `exp`, `iat`.
- **Session validation** — hash the refresh token, look up in DB, check `revoked_at` and `expires_at`.
- **Email verification** — `email_verification_tokens` table, 24h expiry, SHA-256 hashed tokens.
- **Password reset** — `password_reset_tokens` table, 1h expiry, SHA-256 hashed tokens.

Token defaults: access = 15m, refresh = 168h (7 days). Configurable in `config.yaml`.

```go
hash := auth.HashRefreshToken(refreshToken)

err := authService.CreateSession(ctx, userID, hash, userAgent, ipAddress)

session, err := authService.GetSessionByRefreshTokenHash(ctx, hash)
```

### Token Flow

1. **Register:** create user → generate access+refresh → SHA-256 hash refresh → store session → send verification email (goroutine)
2. **Login:** find user by email → bcrypt compare → generate tokens → store session
3. **Refresh:** parse refresh token → hash → lookup session → check not revoked/expired → generate new tokens → create session → revoke old
4. **Logout:** auth middleware (access token) → parse refresh body → hash → lookup → verify ownership → revoke
5. **Verify email:** parse token → hash → lookup → check expiry → mark user verified → delete tokens
6. **Forgot password:** parse email → find user → generate token → store → email reset link (goroutine)
7. **Reset password:** parse token+password → hash → lookup → delete tokens → hash new password → update user → revoke all sessions

**Note:** Login has timing-attack mitigation — on user-not-found, it still calls `ComparePassword` against a dummy hash.

### Context Helpers

```go
import "codeberg.org/vaznerd/blogging-platform/internal/middleware"

userID, ok := middleware.GetUserID(r)
role, ok := middleware.GetUserRole(r)
```

### Protecting Routes

```go
authMW := middleware.Auth(authService.ValidateToken, log)

mux.HandleFunc("GET /api/v1/posts", h.ListPosts)  // public

mux.Handle("POST /api/v1/posts", authMW(http.HandlerFunc(h.CreatePost)))  // protected
```

### Ownership Checks (in Service Layer)

```go
func (s *Service) UpdatePost(ctx context.Context, postID, userID string) error {
    post, err := s.repo.GetByID(ctx, postID)
    if err != nil {
        return err
    }
    if post.AuthorID != userID {
        return ErrForbidden
    }
    return nil
}
```

---

## Middleware Stack

Applied in `internal/server/router.go`, outermost to innermost:

1. **RecoveryMiddleware** — panic recovery, logs stack trace, returns 500
2. **LoggingMiddleware** — UUID request ID, status-based log levels (Info/Warn/Error), injects `X-Request-ID` header
3. **CorsMiddleware** — configurable origin, credentials, max-age 300s (wraps `go-chi/cors`)

Auth middleware is applied per-route, not globally.

---

## Error Handling

Use sentinel errors in the domain package:

```go
package domain

var (
    ErrNotFound   = errors.New("not found")
    ErrForbidden  = errors.New("forbidden")
    ErrConflict   = errors.New("conflict")
)
```

Map to HTTP status codes in handlers using `errors.Is()`:

```go
if errors.Is(err, domain.ErrNotFound) {
    http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
    return
}
if errors.Is(err, domain.ErrForbidden) {
    http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
    return
}
```

---

## Configuration

Uses `koanf` with YAML file + env overrides:

- Config file: `backend/configs/config.yaml`
- Env overrides: `DB__HOST`, `DB__PORT`, etc. (`__` separator)
- Secrets from env only: `JWT_SECRET`, `RESEND_API` (never in YAML)
- `.env` file auto-loaded by `godotenv/autoload`
- Validation in `config.go` includes production-specific checks (debug must be false, DB password required, SSL mode cannot be "disable")

---

## Testing

- Use Go's standard `testing` package
- Table-driven tests preferred
- Repository tests use a test PostgreSQL instance
- Handler tests use `httptest.NewRecorder()` + `httptest.NewRequest()`
- No tests exist yet — 8 test suites tracked in `TODO.md`

```bash
go test ./...
```

---

## Best Practices for AI Assistants

1. **No framework** — use `http.ServeMux` with `"METHOD /path"` patterns
2. **pgx, not GORM** — use `pgx/v5` for all database access, never import GORM
3. **Reference existing code** — check `internal/auth/` for full patterns; `internal/category/` and `internal/tag/` follow the same single-file pattern with stub handlers
4. **Follow layered architecture** — never skip Handler → Service → Repository
5. **Route constants** — always define paths in `routes.go`, never hardcode in `router.go`
6. **No comments** — write self-documenting code, never add comments unless explaining non-obvious WHY
7. **Error handling** — use sentinel errors, map in handlers with `errors.Is()`
8. **JWT auth** — Bearer token in `Authorization` header, validate via `auth.Service`
9. **Configuration** — use koanf (YAML + env overrides), never hardcode secrets
10. **Logging** — use `slog` with structured attributes, never `log.Printf`
11. **Check Makefile** — all development commands in `make help`
12. **No ORM** — write raw SQL with pgx, never use an ORM
13. **Reference url-shortener** — always read `../url-shortener` (at `/home/piyush/Projects/url-shortner`) when creating new features to check how similar features were implemented there; it uses the same tech stack (Go, pgx, Redis, resend) with a flat package structure
14. **Don't write code unless explicitly asked** — explain how to do things, don't create files or write code unless the user explicitly says "create it" or "write it"
15. **TODO cleanup** — when a task is completed, remove the line from `TODO.md` entirely; never mark with `[x]`
16. **Build artifacts** — `backend/build/` contains a compiled binary that shouldn't be in git; check `.gitignore`
