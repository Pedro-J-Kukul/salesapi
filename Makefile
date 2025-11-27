# Makefile for managing database migrations and seeds using golang-migrate
# Importing environment variables from .env file
include .env

# PHONY targets
.PHONY: migrate/create migrate/up migrate/down migrate/fix seed/create seed/up seed/down seed/fix run

# --- Migrations ---
migrate/create:
	@if [ -z "$(name)" ]; then \
		echo "Error: usage 'make migrate/create name=your_migration_name'"; \
		exit 1; \
	fi
	@mkdir -p ./migrations
	migrate create -seq -ext=.sql -dir=$(MIGRATIONS_PATH) $(name)

migrate/up:
	migrate -path "$(MIGRATIONS_PATH)" -database "$(DB_DSN)" up 

migrate/down:
	migrate -path "$(MIGRATIONS_PATH)" -database "$(DB_DSN)" down

migrate/fix:
	@echo 'Checking migration status...'
	@migrate -path "$(MIGRATIONS_PATH)" -database "$(DB_DSN)" version > /tmp/migrate_version 2>&1
	@cat /tmp/migrate_version
	@if grep -q "dirty" /tmp/migrate_version; then \
		version=$$(grep -o '[0-9]\+' /tmp/migrate_version | head -1); \
		echo "Found dirty migration at version $$version"; \
		echo "Forcing version $$version..."; \
		migrate -path "$(MIGRATIONS_PATH)" -database "$(DB_DSN)" force $$version; \
		migrate -path "$(MIGRATIONS_PATH)" -database "$(DB_DSN)" down 1; \
		migrate -path "$(MIGRATIONS_PATH)" -database "$(DB_DSN)" up 1; \
	else \
		echo "No dirty migration found"; \
	fi
	@rm -f /tmp/migrate_version

# --- Seeding ---
seed/create:
	@if [ -z "$(name)" ]; then \
		echo "Error: usage 'make seed/create name=seed_name'"; \
		exit 1; \
	fi
	@mkdir -p $(SEEDS_PATH)
	migrate create -seq -ext=.sql -dir=$(SEEDS_PATH) $(name)

seed/up:
	migrate -path "$(SEEDS_PATH)" -database "$(DB_DSN)" up

seed/down:
	migrate -path "$(SEEDS_PATH)" -database "$(DB_DSN)" down

# --- App ---
run:
	@echo "Starting API server on port $(PORT) in $(ENVIRONMENT) mode..."
	@go run ./cmd/api \
		-port=$(PORT) \
		-env=$(ENVIRONMENT) \
		-db-dsn=$(DB_DSN) \
		-db-max-open-conns=$(DB_MAX_OPEN_CONNS) \
		-db-max-idle-conns=$(DB_MAX_IDLE_CONNS) \
		-db-max-idle-time=$(DB_MAX_IDLE_TIME) \
		-cors-trusted-origins=$(CORS_ALLOWED_ORIGINS) \
		-limiter-enabled=$(RATE_LIMITER_ENABLED) \
		-limiter-rps=$(RATE_LIMITER_RPS) \
		-limiter-burst=$(RATE_LIMITER_BURST) \
		-smtp-host=$(SMTP_HOST) \
		-smtp-port=$(SMTP_PORT) \
		-smtp-username=$(SMTP_USERNAME) \
		-smtp-password=$(SMTP_PASSWORD) \
		-smtp-sender=$(SMTP_SENDER)