# TODO

All remaining work for the blogging platform. See `readme.md` for architecture and API reference.

---

## Phase 1 — Auth Domain (Foundation)

Everything depends on auth. Build this first.

### Repository Layer
- [ ] Define `RefreshTokenRepository` methods: `Save`, `FindByToken`, `DeleteByToken`, `DeleteByUserID`, `DeleteExpired`
- [ ] Implement `Save` — insert refresh token row with expiry
- [ ] Implement `FindByToken` — lookup by token hash, check expiry/revocation
- [ ] Implement `DeleteByToken` — soft revoke (set `revoked_at`)
- [ ] Implement `DeleteByUserID` — revoke all user tokens (used in logout-all)
- [ ] Implement `DeleteExpired` — cleanup job

### Service Layer
- [ ] `Register(email, username, password)` — validate input, hash password, create user, generate tokens, send verification email
- [ ] `Login(email, password)` — verify credentials, check email verified, generate tokens
- [ ] `Logout(userID, refreshToken)` — revoke refresh token
- [ ] `Refresh(refreshToken)` — validate token, rotate (revoke old, issue new pair)
- [ ] `Me(userID)` — fetch user by ID, return profile
- [ ] `VerifyEmail(token)` — validate verification token, mark email verified
- [ ] `ResendVerification(email)` — generate new verification token, send email
- [ ] `ForgotPassword(email)` — generate reset token, send email
- [ ] `ResetPassword(token, newPassword)` — validate reset token, update password

### Handler Layer (replace all stubs)
- [ ] `Register` — parse JSON body → call service → return user + tokens
- [ ] `Login` — parse JSON body → call service → return tokens
- [ ] `Logout` — extract userID from context → call service → 204
- [ ] `Refresh` — parse refresh token from body → call service → return new tokens
- [ ] `Me` — extract userID from context → call service → return user
- [ ] `VerifyEmail` — parse token from query → call service → 200
- [ ] `ResendVerification` — parse email from body → call service → 200
- [ ] `ForgotPassword` — parse email from body → call service → 200
- [ ] `ResetPassword` — parse token + password from body → call service → 200

---

## Phase 2 — User Domain

### Repository Layer
- [ ] Define `UserRepository` methods: `Create`, `GetByID`, `GetByEmail`, `GetByUsername`, `Update`, `Delete`
- [ ] Implement `Create` — insert user row, return created user
- [ ] Implement `GetByID` — lookup by UUID
- [ ] Implement `GetByEmail` — lookup by email
- [ ] Implement `GetByUsername` — lookup by username
- [ ] Implement `Update` — update bio, avatar_url, username
- [ ] Implement `Delete` — hard delete user (cascades to posts/comments via FK)

### Service Layer
- [ ] `GetByUsername(username)` — fetch user profile
- [ ] `UpdateProfile(userID, input)` — validate input, check uniqueness, update
- [ ] `DeleteAccount(userID, password)` — verify password, delete user

### Handler Layer (replace all stubs)
- [ ] `GetUser` — parse `{username}` from path → call service → return profile
- [ ] `UpdateMe` — extract userID from context → parse JSON body → call service → return updated user
- [ ] `DeleteMe` — extract userID from context → parse password from body → call service → 204

---

## Phase 3 — Post Domain (only errors.go + routes.go exist)

### Create Full Package
- [ ] Create `internal/post/repository.go` — interface + struct + constructor
- [ ] Create `internal/post/service.go` — struct + constructor
- [ ] Create `internal/post/handler.go` — struct + constructor + all handlers
- [ ] Create `internal/post/router.go` — `RegisterRoutes` with auth middleware

### Repository Layer
- [ ] `Create(userID, input)` — insert post, generate slug, render markdown→HTML
- [ ] `GetByID(postID)` — fetch single post with author info
- [ ] `GetByUserSlug(username, slug)` — fetch by username + slug (public URL)
- [ ] `List(page, limit)` — paginated list of published posts
- [ ] `ListByUser(username, page, limit)` — paginated list by author
- [ ] `Update(postID, input)` — update title/content/status, re-render HTML
- [ ] `Delete(postID)` — hard delete post (cascades to comments/tags)

### Service Layer
- [ ] `CreatePost(userID, input)` — validate, generate slug, call repo
- [ ] `GetPost(postID)` — fetch, verify status=published for non-owners
- [ ] `ListPosts(page, limit)` — paginated listing
- [ ] `UpdatePost(postID, userID, input)` — ownership check, update
- [ ] `DeletePost(postID, userID)` — ownership check, delete

### Handler Layer
- [ ] `ListPosts` — parse query params → call service → return paginated posts
- [ ] `CreatePost` — extract userID → parse JSON → call service → 201
- [ ] `GetPost` — parse `{id}` → call service → return post
- [ ] `UpdatePost` — extract userID → parse `{id}` + JSON → call service → return updated
- [ ] `DeletePost` — extract userID → parse `{id}` → call service → 204

---

## Phase 4 — Comment Domain (only errors.go + routes.go exist)

### Create Full Package
- [ ] Create `internal/comment/repository.go`
- [ ] Create `internal/comment/service.go`
- [ ] Create `internal/comment/handler.go`
- [ ] Create `internal/comment/router.go`

### Repository Layer
- [ ] `Create(postID, authorID, content)` — insert comment
- [ ] `ListByPost(postID, page, limit)` — paginated comments for a post
- [ ] `GetByID(commentID)` — fetch single comment
- [ ] `Update(commentID, content)` — update comment content
- [ ] `Delete(commentID)` — hard delete comment

### Service Layer
- [ ] `CreateComment(postID, userID, content)` — validate post exists, create
- [ ] `ListComments(postID, page, limit)` — paginated listing
- [ ] `UpdateComment(commentID, userID, content)` — ownership check, update
- [ ] `DeleteComment(commentID, userID)` — ownership check, delete

### Handler Layer
- [ ] `ListComments` — parse `{id}` + query params → call service
- [ ] `CreateComment` — extract userID → parse `{id}` + JSON → call service → 201
- [ ] `UpdateComment` — extract userID → parse `{id}` + JSON → call service
- [ ] `DeleteComment` — extract userID → parse `{id}` → call service → 204

---

## Phase 5 — Tag Domain (only errors.go + routes.go exist)

### Create Full Package
- [ ] Create `internal/tag/repository.go`
- [ ] Create `internal/tag/service.go`
- [ ] Create `internal/tag/handler.go`
- [ ] Create `internal/tag/router.go`

### Repository Layer
- [ ] `Create(name)` — insert tag, handle unique constraint
- [ ] `List()` — list all tags with post counts
- [ ] `GetByName(name)` — fetch tag by name
- [ ] `GetByPost(postID)` — fetch tags for a post
- [ ] `AttachToPost(postID, tagID)` — insert into post_tags
- [ ] `DetachFromPost(postID, tagID)` — remove from post_tags

### Service Layer
- [ ] `ListTags()` — list all tags
- [ ] `GetTag(name)` — get tag with post count
- [ ] `AttachTags(postID, tagNames)` — find/create tags, attach to post

### Handler Layer
- [ ] `ListTags` — call service → return tags
- [ ] `GetTag` — parse `{name}` → call service → return tag with posts

---

## Phase 6 — Wiring & Infrastructure

### Server Registration
- [ ] Register post domain in `internal/server/router.go`
- [ ] Register comment domain in `internal/server/router.go`
- [ ] Register tag domain in `internal/server/router.go`
- [ ] Wire post/comment/tag services + repos in `cmd/server/main.go`

### Makefile
- [ ] Implement `dev-up` target — `docker compose -f docker-compose.dev.yml up -d`
- [ ] Implement `run-backend` target — `cd backend && air`
- [ ] Implement `lint-backend` target — `cd backend && golangci-lint run`
- [ ] Implement `format-backend` target — `cd backend && gofumpt -w .`
- [ ] Implement `dev-down` target — `docker compose -f docker-compose.dev.yml down`
- [ ] Implement `dev-down-force` target — `docker compose -f docker-compose.dev.yml down -v`
- [ ] Implement `dev-logs` target — `docker compose -f docker-compose.dev.yml logs -f`
- [ ] Implement `dev-migrate-up` target — `migrate -path backend/migrations -database "$(MIGRATE_DSN)" up`
- [ ] Implement `dev-migrate-down` target — `migrate -path backend/migrations -database "$(MIGRATE_DSN)" down 1`
- [ ] Implement `dev-migrate-version` target — `migrate -path backend/migrations -database "$(MIGRATE_DSN)" version`

### Docker
- [ ] Create `docker-compose.dev.yml` — postgres + redis containers
- [ ] Create `docker-compose.prod.yml` — postgres + redis + app containers
- [ ] Create `backend/Dockerfile` — multi-stage build (build + runtime)

---

## Phase 7 — Quality

### Testing
- [ ] Auth service tests — JWT generation/validation, bcrypt hash/compare
- [ ] Auth handler tests — register, login, logout, refresh flows
- [ ] User repository tests — CRUD operations
- [ ] User handler tests — get profile, update, delete
- [ ] Post repository tests — CRUD, slug uniqueness
- [ ] Post handler tests — all endpoints
- [ ] Comment handler tests — all endpoints
- [ ] Tag handler tests — all endpoints
- [ ] Middleware tests — auth extraction, CORS, recovery

### Code Quality
- [ ] Run `golangci-lint` and fix all warnings
- [ ] Run `gofumpt` on all files
- [ ] Remove `backend/build/` binary from git tracking
- [ ] Verify `.gitignore` excludes `backend/build/`
- [ ] Verify `.env` is gitignored (currently contains live secrets)

---

## Future Features (not now)

- Rate limiting middleware (token-bucket per-IP)
- Standardized API error response envelope
- Health check endpoints (`/live`, `/ready`)
- CI pipeline (lint, test, build, vulncheck)
- Recommendations service
- Testing with testcontainers
- API docs with Scalar (OpenAPI 3.1 + oapi-codegen)
- Likes, bookmarks, follow users
- Notifications, reading history
- Scheduled publishing, RSS feeds
- View counters, newsletter subscriptions
- User themes, dark mode
