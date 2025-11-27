# Makefile for managing database migrations and seeds using golang-migrate
# Importing environment variables from .env file
include .env

# PHONY targets
.PHONY: migrate/create migrate/up migrate/down migrate/fix seed/create seed/up seed/down seed/fix run test test/integration test/live db/setup db/reset build clean help

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
	go run ./cmd/api

build:
	@echo 'Building application...'
	go build -o ./bin/api ./cmd/api
	@echo 'Build complete!'

clean:
	@echo 'Cleaning build artifacts...'
	rm -rf ./bin
	@echo 'Clean complete!'

# --- Testing ---
test:
	@echo 'Running all tests...'
	go test -v -race -buildvcs ./...

test/coverage:
	@echo 'Running tests with coverage...'
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

test/integration:
	@echo 'Running integration tests...'
	@echo 'NOTE: Make sure test database is running and accessible'
	go test -v ./cmd/api -run Integration

test/live:
	@echo 'Running live API tests...'
	@echo 'NOTE: Make sure the API server is running'
	./scripts/test_commands.sh

# --- Database Management ---
db/setup:
	@echo 'Setting up database...'
	@echo 'Creating sales database and user...'
	psql -U postgres -f scripts/database_setup.sql
	@echo 'Database setup complete!'

db/reset:
	@echo 'Resetting test database...'
	@echo 'WARNING: This will delete all data in the sales database!'
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		psql -U sales -d sales -c "TRUNCATE TABLE sales, users_permissions, tokens, products, permissions, users CASCADE;"; \
		psql -U sales -d sales -c "ALTER SEQUENCE users_id_seq RESTART WITH 1;"; \
		psql -U sales -d sales -c "ALTER SEQUENCE permissions_id_seq RESTART WITH 1;"; \
		psql -U sales -d sales -c "ALTER SEQUENCE tokens_id_seq RESTART WITH 1;"; \
		psql -U sales -d sales -c "ALTER SEQUENCE products_id_seq RESTART WITH 1;"; \
		psql -U sales -d sales -c "ALTER SEQUENCE sales_id_seq RESTART WITH 1;"; \
		echo 'Database reset complete!'; \
	else \
		echo 'Database reset cancelled.'; \
	fi

db/schema:
	@echo 'Initializing database schema...'
	psql -U sales -d sales -f scripts/schema.sql
	@echo 'Schema initialization complete!'

# --- Help ---
help:
	@echo 'Usage:'
	@echo '  make run              - Run the application'
	@echo '  make build            - Build the application binary'
	@echo '  make clean            - Remove build artifacts'
	@echo '  make test             - Run all tests'
	@echo '  make test/coverage    - Run tests with coverage report'
	@echo '  make test/integration - Run integration tests only'
	@echo '  make test/live        - Run live API tests with curl'
	@echo '  make db/setup         - Create database and user'
	@echo '  make db/schema        - Initialize database schema'
	@echo '  make db/reset         - Reset test database (WARNING: deletes all data)'
	@echo '  make migrate/up       - Run database migrations'
	@echo '  make migrate/down     - Rollback database migrations'
	@echo '  make migrate/create   - Create new migration (name=migration_name)'
	@echo '  make help             - Show this help message'