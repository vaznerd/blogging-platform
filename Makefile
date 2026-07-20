include .env
export

MIGRATE_DSN = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
.DEFAULT_GOAL := help

help:
	@echo "=== Dev Environment ==="
	@echo "  make dev-up               - Start dev containers"
	@echo "  make run-backend          - Start the go api server"
	@echo "  make run-frontend         - Start the frontend"
	@echo "  make lint-backend         - Lint go backend"
	@echo "  make lint-frontend        - Lint the frontend"
	@echo "  make format-backend       - Format go backend"
	@echo "  make format-frontend      - Format the backend"
	@echo "  make dev-down             - Stop dev containers"
	@echo "  make dev-down-force       - Stop dev containers and remove volumes"
	@echo "  make dev-logs             - Stream dev container logs"
	@echo "  make dev-migrate-up       - Run pending migrations"
	@echo "  make dev-migrate-down     - Rollback last migration"
	@echo "  make dev-migrate-version  - Show current migration version"
	@echo ""
	@echo "=== Prod Environment ==="
	@echo "  make prod-up              - Start prod containers"
	@echo "  make prod-down            - Stop prod containers"
	@echo "  make prod-down-force      - Stop prod containers and remove volumes"
	@echo "  make prod-logs            - Stream prod container logs"
	@echo "  make prod-migrate-up      - Run pending migrations"
	@echo "  make prod-migrate-down    - Rollback last migration"
	@echo "  make prod-migrate-version - Show current migration version"

run-backend:
	cd backend && go run ./cmd/server

run-frontend:
	cd frontend && npm run dev

lint-backend:
	cd backend && golangci-lint run ./...

format-backend:
	cd backend && gofumpt -w .

lint-frontend:
	cd frontend && npm run lint

format-frontend:
	cd frontend && npm run format

dev-up:
	docker compose -f docker-compose.dev.yml up -d

dev-down:
	docker compose -f docker-compose.dev.yml down

dev-down-force:
	docker compose -f docker-compose.dev.yml down -v

dev-logs:
	docker compose -f docker-compose.dev.yml logs -f

dev-migrate-up:
	migrate -path backend/migrations -database "$(MIGRATE_DSN)" up

dev-migrate-down:
	migrate -path backend/migrations -database "$(MIGRATE_DSN)" down 1

dev-migrate-version:
	migrate -path backend/migrations -database "$(MIGRATE_DSN)" version

prod-up:
	docker compose -f docker-compose.prod.yml up -d

prod-down:
	docker compose -f docker-compose.prod.yml down

prod-down-force:
	docker compose -f docker-compose.prod.yml down -v

prod-logs:
	docker compose -f docker-compose.prod.yml logs -f

prod-migrate-up:
	migrate -path backend/migrations -database "$(PROD_MIGRATE_DSN)" up

prod-migrate-down:
	migrate -path backend/migrations -database "$(PROD_MIGRATE_DSN)" down 1

prod-migrate-version:
	migrate -path backend/migrations -database "$(PROD_MIGRATE_DSN)" version
