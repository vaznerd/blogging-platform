# Blogging Platform

A multi-user blogging platform where users write posts in Markdown and publish them under their own profile.

```text
/u/zed
/u/zed/learning-go
/u/alice/linux-notes
```

## Features

- User registration and JWT authentication (access + refresh tokens)
- Email verification and password reset flows
- Create, edit, delete posts in Markdown
- Posts are rendered to HTML on publish
- Comments on posts
- Tagging system for posts
- Category system for posts (many-to-many)
- Role-based access control (guest, user, admin)
- Redis caching layer
- PostgreSQL database with versioned migrations
- Structured logging with slog
- Graceful shutdown

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.26+ |
| HTTP | stdlib `net/http` (no framework) |
| Database | PostgreSQL via pgx/v5 |
| Cache | Redis via go-redis/v9 |
| Auth | JWT via golang-jwt/jwt/v5 |
| Passwords | bcrypt via golang.org/x/crypto |
| Config | koanf (YAML + env overrides) |
| Logging | slog |
| Email | Resend |
| Migrations | golang-migrate |
| Hot Reload | Air |

---

## Quick Start

### Prerequisites

- Go 1.26+
- PostgreSQL
- Redis
- Resend API key

### Setup

```bash
# Clone
git clone <repo-url> && cd blogging-platform

# Environment
cp .env.example .env
./scripts/gen-jwt-secret.sh
# Edit .env — set RESEND_API

# Start infrastructure
make dev-up

# Run migrations
make dev-migrate-up

# Start server (hot reload)
make run-backend
```

Server starts at `http://localhost:8080`. Health check: `GET /health`.

### Environment Variables

```bash
# Required
RESEND_API="re_xxxxx"         # Resend email API key
JWT_SECRET="..."               # 256-bit base64 (run gen-jwt-secret.sh)

# Optional overrides (defaults in configs/config.yaml)
DB__PASSWORD=postgres
DB__HOST__GO=localhost
REDIS__HOST=localhost
```

### Development Commands

```bash
make dev-up               # Start postgres + redis containers
make run-backend          # Start Go API with Air hot reload
make lint-backend         # Run golangci-lint
make format-backend       # Format with gofumpt
make dev-down             # Stop containers
make dev-down-force       # Stop containers and remove volumes
make dev-logs             # Stream container logs
make dev-migrate-up       # Apply pending migrations
make dev-migrate-down     # Rollback last migration
make dev-migrate-version  # Show current migration version
```

---

## Architecture

### Layered Pattern

```
Handler (HTTP) → Service (Business Logic) → Repository (Database)
```

- **Handlers** parse requests and write responses. No business logic.
- **Services** contain all business logic and domain rules.
- **Repositories** only interact with the database via pgx.

### Request Flow

```
Client Request
    ↓
Middleware Stack (Recovery → Logging → CORS)
    ↓
Route Matching (http.ServeMux with "METHOD /path" patterns)
    ↓
Auth Middleware (protected routes only — extracts JWT → context)
    ↓
Handler (parse JSON → call service → write JSON)
    ↓
Service (business logic → call repository)
    ↓
Repository (SQL via pgx → return domain models)
```

### Middleware Stack

```
Global (all routes):
  1. Recovery — catches panics, returns 500
  2. Request Logging — UUID request ID, status-based log levels
  3. CORS — configurable origins, credentials

Route-level (protected routes):
  4. Auth — extracts Bearer token, validates JWT, sets userID/role in context
```

### Domain Package Structure

Each domain follows this structure:

```
internal/<domain>/
├── errors.go        # Sentinel errors (ErrNotFound, ErrForbidden, etc.)
├── routes.go        # Path constants (RouteXxx = "/api/v1/...")
├── router.go        # RegisterRoutes(mux, service, log, mail)
├── handler.go       # HTTP handlers
├── service.go       # Business logic
└── repository.go    # Database access (pgx)
```

### Authentication

JWT tokens use HMAC-SHA256 signing.

**Access Token:**
```json
{
  "sub": "user-uuid",
  "role": "user",
  "exp": 1740000000,
  "iat": 1739999100
}
```

**Refresh Token:**
```json
{
  "sub": "user-uuid",
  "type": "refresh",
  "exp": 1740600000,
  "iat": 1739999100
}
```

**Protected routes** require `Authorization: Bearer <token>` header.

**Context helpers** for extracting user info in handlers:
```go
userID, ok := middleware.GetUserID(r)
role, ok := middleware.GetUserRole(r)
```

### User Roles

| Role | Permissions |
|------|------------|
| Guest | Read posts, read profiles, search |
| User | Create/edit/delete own posts and comments |
| Admin | All user permissions + delete any content, suspend users |

### Ownership Rules

- Posts: `author_id == current_user_id` required for edit/delete
- Comments: `author_id == current_user_id` required for edit/delete
- Admins bypass ownership checks

---

## API Reference

Base path: `/api/v1`

### Authentication

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/auth/register` | No | Register new user |
| POST | `/api/v1/auth/login` | No | Login, returns tokens |
| POST | `/api/v1/auth/logout` | Yes | Revoke refresh token |
| POST | `/api/v1/auth/refresh` | No | Rotate refresh token |
| POST | `/api/v1/auth/verify-email` | No | Verify email with token |
| POST | `/api/v1/auth/resend-verification` | No | Resend verification email |
| POST | `/api/v1/auth/forgot-password` | No | Request password reset |
| POST | `/api/v1/auth/reset-password` | No | Reset password with token |

### Users

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/users/me` | Yes | Get own profile |
| GET | `/api/v1/users/{username}` | No | Get user profile |
| PATCH | `/api/v1/users/me` | Yes | Update own profile |
| DELETE | `/api/v1/users/me` | Yes | Delete own account |

### Posts

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/posts` | No | List posts (paginated) |
| POST | `/api/v1/posts` | Yes | Create post |
| GET | `/api/v1/posts/{postID}` | No | Get post by ID |
| PATCH | `/api/v1/posts/{postID}` | Yes | Update post (owner only) |
| DELETE | `/api/v1/posts/{postID}` | Yes | Delete post (owner only) |

### Comments

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/posts/{postID}/comments` | No | List comments on post |
| POST | `/api/v1/posts/{postID}/comments` | Yes | Add comment to post |
| PATCH | `/api/v1/comments/{id}` | Yes | Update comment (owner only) |
| DELETE | `/api/v1/comments/{id}` | Yes | Delete comment (owner only) |

### Tags

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/tags` | No | List all tags |
| GET | `/api/v1/tags/{name}` | No | Get posts by tag |
| POST | `/api/v1/tags` | Yes | Create tag |
| DELETE | `/api/v1/tags/{name}` | Yes | Delete tag |
| POST | `/api/v1/posts/{postID}/tags` | Yes | Attach tag to post |
| DELETE | `/api/v1/posts/{postID}/tags/{name}` | Yes | Detach tag from post |

### Categories

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/categories` | No | List all categories |
| GET | `/api/v1/categories/{slug}` | No | Get category by slug |
| POST | `/api/v1/categories` | Yes | Create category |
| PUT | `/api/v1/categories/{slug}` | Yes | Update category |
| DELETE | `/api/v1/categories/{slug}` | Yes | Delete category |
| POST | `/api/v1/posts/{postID}/categories` | Yes | Attach category to post |
| DELETE | `/api/v1/posts/{postID}/categories/{slug}` | Yes | Detach category from post |

### Pagination

Listing endpoints support query parameters:

```text
?limit=20&offset=0
```

### Implementation Status

| Domain | Status |
|--------|--------|
| Auth | All endpoints implemented |
| User — Repo methods | CreateUser, GetByEmail, GetByID, MarkEmailVerified, UpdatePassword done |
| User — Me handler | Implemented |
| User — GetUser, UpdateMe, DeleteMe | Stubs (501) |
| Category — repository + service | Implemented |
| Category — handlers | Stubs (501) |
| Tag — repository + service | Implemented |
| Tag — handlers | Stubs (501) |
| Post | Scaffold only (errors + routes) |
| Comment | Scaffold only (errors + routes) |

---

## Database

PostgreSQL is the primary database. Schema is managed via versioned SQL migrations.

### Tables

**users** — user accounts
```sql
id              UUID PRIMARY KEY
username        VARCHAR(255) UNIQUE NOT NULL
email           VARCHAR(255) UNIQUE NOT NULL
password_hash   VARCHAR(255) NOT NULL
bio             TEXT DEFAULT ''
avatar_url      VARCHAR(512) DEFAULT ''
role            VARCHAR(20) DEFAULT 'user'
created_at      TIMESTAMPTZ
updated_at      TIMESTAMPTZ
```

**sessions** — JWT refresh token sessions
```sql
session_id      BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY
user_id         UUID REFERENCES users(id) ON DELETE CASCADE
refresh_token_hash BYTEA UNIQUE NOT NULL
user_agent      TEXT DEFAULT ''
ip_address      INET
expires_at      TIMESTAMPTZ
last_used_at    TIMESTAMPTZ
revoked_at      TIMESTAMPTZ
created_at      TIMESTAMPTZ
```

**posts** — blog posts
```sql
id              UUID PRIMARY KEY
author_id       UUID REFERENCES users(id) ON DELETE CASCADE
title           VARCHAR(500) NOT NULL
slug            VARCHAR(500) NOT NULL
markdown_content TEXT
html_content    TEXT
status          post_status (draft|published|archived)
created_at      TIMESTAMPTZ
updated_at      TIMESTAMPTZ
published_at    TIMESTAMPTZ
UNIQUE(author_id, slug)
```

**comments** — post comments
```sql
id              UUID PRIMARY KEY
post_id         UUID REFERENCES posts(id) ON DELETE CASCADE
author_id       UUID REFERENCES users(id) ON DELETE CASCADE
content         TEXT NOT NULL
created_at      TIMESTAMPTZ
updated_at      TIMESTAMPTZ
```

**tags** — post tags
```sql
id              UUID PRIMARY KEY
name            VARCHAR(255) UNIQUE NOT NULL
```

**post_tags** — many-to-many junction
```sql
post_id         UUID REFERENCES posts(id) ON DELETE CASCADE
tag_id          UUID REFERENCES tags(id) ON DELETE CASCADE
PRIMARY KEY (post_id, tag_id)
```

**categories** — post categories
```sql
id              UUID PRIMARY KEY
name            VARCHAR(255) NOT NULL
slug            VARCHAR(255) UNIQUE NOT NULL
description     TEXT DEFAULT ''
created_at      TIMESTAMPTZ
updated_at      TIMESTAMPTZ
```

**post_categories** — many-to-many junction
```sql
post_id         UUID REFERENCES posts(id) ON DELETE CASCADE
category_id     UUID REFERENCES categories(id) ON DELETE CASCADE
PRIMARY KEY (post_id, category_id)
```

**email_verification_tokens** — email verification
```sql
id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY
user_id         UUID REFERENCES users(id) ON DELETE CASCADE
token_hash      BYTEA UNIQUE NOT NULL
expires_at      TIMESTAMPTZ
created_at      TIMESTAMPTZ
```

**password_reset_tokens** — password reset
```sql
id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY
user_id         UUID REFERENCES users(id) ON DELETE CASCADE
token_hash      BYTEA UNIQUE NOT NULL
expires_at      TIMESTAMPTZ
created_at      TIMESTAMPTZ
```

### Migrations

```bash
make dev-migrate-up       # Apply pending
make dev-migrate-down     # Rollback last
make dev-migrate-version  # Show current version
```

Files live in `backend/migrations/`:
```text
{YYYYMMDDHHMMSS}_{description}.up.sql
{YYYYMMDDHHMMSS}_{description}.down.sql
```

---

## Deployment

### Docker

```bash
# Build
docker build -t blogging-platform -f backend/Dockerfile .

# Run
docker run -p 8080:8080 \
  -e JWT_SECRET="..." \
  -e RESEND_API="..." \
  -e DB__HOST=postgres \
  -e REDIS__HOST=redis \
  blogging-platform
```

### Docker Compose

```bash
# Development
make dev-up

# Production
make prod-up
```

### Environment Configuration

All configuration is loaded from `backend/configs/config.yaml` with environment variable overrides.

**Config hierarchy:**
1. `backend/configs/config.yaml` — defaults
2. Environment variables — override with `__` as separator (e.g., `DB__HOST`)

**Required env vars:**
- `JWT_SECRET` — 256-bit signing secret
- `RESEND_API` — Resend email API key

---

## Project Structure

```text
.
├── .env                          # Environment variables
├── .env.example                  # Template
├── Makefile                      # Dev commands
├── readme.md
├── scripts/
│   └── gen-jwt-secret.sh         # Generate JWT secret
│
└── backend/
    ├── .air.toml                 # Hot-reload config
    ├── .golangci.yml             # Linter config
    ├── go.mod / go.sum
    ├── cmd/
    │   └── server/
    │       └── main.go           # Entrypoint
    ├── configs/
    │   └── config.yaml           # Default config
    ├── migrations/               # SQL migrations (9 versions)
    └── internal/
        ├── auth/                 # Authentication
        │   ├── errors.go
        │   ├── handler.go
        │   ├── repository.go
        │   ├── routes.go
        │   ├── router.go
        │   └── service.go
        ├── category/             # Categories
        │   ├── errors.go
        │   ├── handler.go
        │   ├── repository.go
        │   ├── routes.go
        │   ├── router.go
        │   └── service.go
        ├── user/                 # User profiles
        │   ├── errors.go
        │   ├── handler.go
        │   ├── repository.go
        │   ├── routes.go
        │   ├── router.go
        │   └── service.go
        ├── post/                 # Blog posts
        │   ├── errors.go
        │   └── routes.go
        ├── comment/              # Comments
        │   ├── errors.go
        │   └── routes.go
        ├── tag/                  # Tags
        │   ├── errors.go
        │   ├── handler.go
        │   ├── repository.go
        │   ├── routes.go
        │   ├── router.go
        │   └── service.go
        ├── db/                   # Database utilities
        │   └── db.go
        ├── config/               # Configuration
        │   ├── config.go
        │   └── validator.go
        ├── logger/               # Structured logging
        │   └── logger.go
        ├── middleware/            # HTTP middleware
        │   ├── auth.go
        │   ├── context.go
        │   ├── cors.go
        │   ├── logger.go
        │   ├── middleware.go
        │   ├── recovery.go
        │   └── response_writer.go
        └── server/               # Router
            └── router.go
```

---

## License

See [LICENSE](LICENSE).
