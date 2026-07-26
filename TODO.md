# TODO

## User Domain

- Implement `GetByUsername` — repo method, service method, handler (parse `{username}` from path, return profile)
- Implement `Update` — repo method (bio, avatar_url, username), service method (`UpdateProfile`), handler (`UpdateMe` — extract userID from context, parse JSON body)
- Implement `Delete` — repo method, service method (`DeleteAccount`), handler (`DeleteMe` — extract userID, verify password, delete account)

## Category Domain

- Implement 7 handler methods (repo + service are done, service is an interface):
  - `CreateCategory` — parse JSON body, call service
  - `GetCategory` — parse `{slug}` from path, return category
  - `ListCategories` — list all categories with optional pagination
  - `UpdateCategory` — parse `{slug}` from path, parse JSON body, call service
  - `DeleteCategory` — parse `{slug}` from path, call service
  - `AttachCategoryToPost` — parse `{slug}` + `{id}` from path, call service
  - `DetachCategoryFromPost` — parse `{slug}` + `{id}` from path, call service

## Tag Domain

- Implement 6 handler methods (repo + service are done, service is an interface):
  - `CreateTag` — parse JSON body, call service
  - `GetTag` — parse `{name}` from path, return tag
  - `ListTags` — list all tags with optional pagination
  - `DeleteTag` — parse `{name}` from path, call service
  - `AttachTagToPost` — parse `{name}` + `{id}` from path, call service
  - `DetachTagFromPost` — parse `{name}` + `{id}` from path, call service

## Post Domain

- Create `repository.go` — interface + pgx implementation (CRUD, list with pagination, by author, by slug)
- Create `service.go` — business logic (create, get by ID, update, delete, list, ownership checks)
- Create `handler.go` — HTTP handlers (CreatePost, GetPost, UpdatePost, DeletePost, ListPosts, ListPostsByAuthor)
- Create `router.go` — `RegisterRoutes` with public + protected routes
- Register in `internal/server/router.go`
- Wire dependencies in `cmd/server/main.go`

## Comment Domain

- Create `repository.go` — interface + pgx implementation (CRUD, list by post)
- Create `service.go` — business logic (create, get, update, delete, list by post, ownership checks)
- Create `handler.go` — HTTP handlers (CreateComment, GetComment, UpdateComment, DeleteComment, ListComments)
- Create `router.go` — `RegisterRoutes` with public + protected routes
- Register in `internal/server/router.go`
- Wire dependencies in `cmd/server/main.go`

## Infrastructure

- Create `docker-compose.dev.yml` — postgres + redis (referenced by Makefile `dev-up` / `dev-down` targets)
- Create `docker-compose.prod.yml` — postgres + redis + app (referenced by Makefile `prod-up` / `prod-down` targets)
- Create `backend/Dockerfile` — multi-stage build (golang alpine → distroless)
- Create `frontend/` directory — app (referenced by Makefile `run-frontend` / `lint-frontend` / `format-frontend`)
- Integrate Redis caching (currently connected but unused in business logic)
- Add rate limiting middleware to auth endpoints (login, register, forgot-password, reset-password, resend-verification)

## Testing

- Auth domain tests — service + handler
- User domain tests — repository + handler
- Category domain tests — repository + handler
- Tag domain tests — repository + handler
- Post domain tests — repository + handler
- Comment domain tests — repository + handler
- Middleware tests — auth extraction, CORS, recovery

