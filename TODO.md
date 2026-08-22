# TODO

## Category Domain

- Implement 7 handler methods (repo + service are done, service is an interface):
  - `CreateCategory` — parse JSON body, call service
  - `GetCategory` — parse `{slug}` from path, return category
  - `ListCategories` — list all categories with optional pagination
  - `UpdateCategory` — parse `{slug}` from path, parse JSON body, call service
  - `DeleteCategory` — parse `{slug}` from path, call service
  - `AttachCategoryToPost` — parse `{postID}` + `{slug}` from path, call service
  - `DetachCategoryFromPost` — parse `{postID}` + `{slug}` from path, call service

## Tag Domain

- Implement 6 handler methods (repo + service are done, service is an interface):
  - `CreateTag` — parse JSON body, call service
  - `GetTag` — parse `{name}` from path, return tag
  - `ListTags` — list all tags with optional pagination
  - `DeleteTag` — parse `{name}` from path, call service
  - `AttachTagToPost` — parse `{postID}` + `{name}` from path, call service
  - `DetachTagFromPost` — parse `{postID}` + `{name}` from path, call service

## Post Domain

- Implement 6 handler methods (repo + service are done, service is an interface):
  - `CreatePost` — extract userID from context, parse JSON body, call service
  - `GetPost` — parse `{postID}` from path, return post
  - `ListPosts` — list posts with optional pagination
  - `ListPostsByAuthor` — parse `{username}` from path, list their posts
  - `UpdatePost` — parse `{postID}` from path, parse JSON body, verify ownership, call service
  - `DeletePost` — parse `{postID}` from path, verify ownership, call service

## Comment Domain

- Implement 5 handler methods (repo + service are done, service is an interface):
  - `CreateComment` — extract userID from context, parse `{postID}` from path, parse JSON body, call service
  - `GetComment` — parse `{id}` from path, return comment
  - `ListComments` — parse `{postID}` from path, list comments for the post
  - `UpdateComment` — parse `{id}` from path, parse JSON body, verify ownership, call service
  - `DeleteComment` — parse `{id}` from path, verify ownership, call service

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

