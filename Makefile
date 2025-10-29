.PHONY: help up down migrate-up migrate-down test test-unit test-integration health logs clean

# Load environment variables from .env if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Detect container backend: Podman > Arion > Docker
DOCKER_COMPOSE := $(shell command -v podman-compose >/dev/null 2>&1 && echo "podman-compose" || (command -v arion >/dev/null 2>&1 && echo "arion" || echo "docker-compose"))

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''
	@echo 'Docker backend: $(DOCKER_COMPOSE)'

up: ## Start all services (PostgreSQL, Redis)
	@echo "Using container backend: $(DOCKER_COMPOSE)"
	$(DOCKER_COMPOSE) up -d
	@echo "Waiting for services to be healthy..."
	@sleep 5
	$(DOCKER_COMPOSE) ps

down: ## Stop all services
	$(DOCKER_COMPOSE) down

migrate-up: ## Run database migrations (up)
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Copy .env.example to .env first."; \
		exit 1; \
	fi
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down: ## Rollback last migration (down)
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Copy .env.example to .env first."; \
		exit 1; \
	fi
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1

test: test-unit test-integration ## Run all tests (unit + integration)

test-unit: ## Run unit tests
	go test ./pkg/... -v -cover

test-integration: ## Run integration tests (requires Docker)
	go test ./tests/integration/... -v

health: ## Run health check script
	@if [ -f ./scripts/health-check.sh ]; then \
		./scripts/health-check.sh; \
	else \
		echo "Health check script not found at ./scripts/health-check.sh"; \
		echo "Checking services manually..."; \
		$(DOCKER_COMPOSE) ps; \
	fi

logs: ## View logs from all services
	$(DOCKER_COMPOSE) logs -f

clean: ## Stop services and remove volumes
	$(DOCKER_COMPOSE) down -v
	@echo "Cleaned up services and volumes"

.DEFAULT_GOAL := help
