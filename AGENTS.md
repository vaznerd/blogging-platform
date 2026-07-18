# Blogging Platform — AI-Friendly Development Guide

**Purpose**: Universal AI assistant guidelines for the Blogging Platform project.

---

## Project Overview

A multi-user blogging platform where users write posts in Markdown and publish them under their own profile. Written in Go with standard library `net/http` (no framework), pgx for PostgreSQL, and Redis for caching.

### Technology Stack

- **Go 1.26+** — check with `go version`
- **stdlib `net/http`** — HTTP server and routing (Go 1.22+ method-based patterns)
- **pgx/v5** — PostgreSQL driver
- **Redis (go-redis/v9)** — caching layer
- **golang-jwt/jwt/v5** — JWT authentication
- **golang-migrate/migrate** — database migrations
- **Air** — hot-reload development
- **golangci-lint** — code quality
- **koanf** — configuration management
- **slog** — structured logging

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
├── errors.go        # Sentinel errors (ErrNotFound, ErrForbidden, etc.)
├── routes.go        # Path constants (RouteXxx = "/api/v1/...")
├── router.go        # RegisterRoutes(mux, service, log, mail)
├── handler.go       # HTTP handlers
├── service.go       # Business logic
└── repository.go    # Database access (pgx)
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
    // ...
)
```

Used in `router.go`:

```go
mux.HandleFunc("POST "+RouteRegister, h.Register)
```

---

## Development Workflow

### Available Commands

```bash
make dev-up            # Start dev containers (postgres, redis)
make run-backend       # Start Go API server with Air hot reload
make lint-backend      # Run golangci-lint
make format-backend    # Format code with gofumpt
make dev-down          # Stop dev containers
make dev-down-force    # Stop and remove volumes
make dev-logs          # View container logs
make dev-migrate-up    # Apply pending migrations
make dev-migrate-down  # Rollback last migration
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
// handler.go
func (h *Handler) CreateWidget(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request body
    // 2. Call service
    // 3. Write response
}
```

```go
// router.go
mux.HandleFunc("POST "+RouteWidgets, h.CreateWidget)
```

If protected, wrap with auth middleware:

```go
authMW := middleware.Auth(authService.ValidateToken)
mux.Handle("POST "+RouteWidgets, authMW(http.HandlerFunc(h.CreateWidget)))
```

### Database Migrations

Migrations use golang-migrate with timestamped filenames:

```bash
make dev-migrate-up        # Apply pending
make dev-migrate-down      # Rollback last
```

Migration files live in `backend/migrations/`:

```
{YYYYMMDDHHMMSS}_{description}.up.sql
{YYYYMMDDHHMMSS}_{description}.down.sql
```

Best practices:
- Wrap in `BEGIN;` / `COMMIT;`
- Use `IF NOT EXISTS`
- Create indexes for foreign keys
- Always write corresponding `.down.sql`

---

## Authentication & Authorization

### Context Helpers

```go
import "codeberg.org/vaznerd/blogging-platform/internal/middleware"

userID, ok := middleware.GetUserID(r)
role, ok := middleware.GetUserRole(r)
```

### Protecting Routes

```go
authMW := middleware.Auth(authService.ValidateToken)

// Public
mux.HandleFunc("GET /api/v1/posts", h.ListPosts)

// Protected
mux.Handle("POST /api/v1/posts", authMW(http.HandlerFunc(h.CreatePost)))
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
    // ...
}
```

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

## Testing

- Use Go's standard `testing` package
- Table-driven tests preferred
- Repository tests use a test PostgreSQL instance
- Handler tests use `httptest.NewRecorder()` + `httptest.NewRequest()`

```bash
go test ./...
```

---

## Best Practices for AI Assistants

1. **No framework** — use `http.ServeMux` with `"METHOD /path"` patterns, never import `gin`, `chi`, or `echo`
2. **pgx, not GORM** — use `pgx/v5` for all database access, never import GORM
3. **Reference existing code** — check `internal/auth/` and `internal/user/` for patterns before creating new domains
4. **Follow layered architecture** — never skip Handler → Service → Repository
5. **Route constants** — always define paths in `routes.go`, never hardcode in `router.go`
6. **Minimal comments** — write self-documenting code, comment WHY not WHAT
7. **Error handling** — use sentinel errors, map in handlers with `errors.Is()`
8. **JWT auth** — Bearer token in `Authorization` header, validate via `auth.Service`
9. **Configuration** — use koanf (YAML + env overrides), never hardcode secrets
10. **Logging** — use `slog` with structured attributes, never `log.Printf`
11. **Check Makefile** — all development commands in `make help`
12. **No ORM** — write raw SQL with pgx, never use an ORM
