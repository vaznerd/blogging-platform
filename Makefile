include .env
export

MIGRATE_DSN = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
.DEFAULT_GOAL := help

help:
	@echo "=== Dev Environment ==="
	@echo "  make dev-up               - Start dev containers"
	@echo "  make run-backend          - Start the go api server"
	@echo "  make lint-backend         - Lint go backend"
	@echo "  make format-backend       - Format go backend"
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
