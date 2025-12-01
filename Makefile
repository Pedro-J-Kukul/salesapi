# ==================================================================================== #
# MAKEFILE FOR SALES API
# ==================================================================================== #

# Import environment variables from .envrc file
-include .envrc
export

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run: run the application locally
.PHONY: run
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
		-smtp-sender=$(SMTP_SENDER) \
		-github-token=$(GITHUB_TOKEN)

## build: build the application binary
.PHONY: build
build:
	@echo 'Building salesapi binary...'
	@go build -o=./bin/salesapi ./cmd/api
	@echo 'Binary created at ./bin/salesapi'

# ==================================================================================== #
# DATABASE MIGRATIONS
# ==================================================================================== #

## migrate/create: create a new migration file (usage: make migrate/create name=migration_name)
.PHONY: migrate/create
migrate/create:
	@if [ -z "$(name)" ]; then \
		echo "Error: usage 'make migrate/create name=your_migration_name'"; \
		exit 1; \
	fi
	@mkdir -p ./migrations
	@migrate create -seq -ext=.sql -dir=./migrations $(name)

## migrate/up: apply all pending migrations
.PHONY: migrate/up
migrate/up:
	@echo 'Running migrations...'
	@migrate -path ./migrations -database "$(DB_DSN)" up

## migrate/down: rollback the last migration
.PHONY: migrate/down
migrate/down: confirm
	@echo 'Rolling back last migration...'
	@migrate -path ./migrations -database "$(DB_DSN)" down 1

## migrate/goto: migrate to a specific version (usage: make migrate/goto version=5)
.PHONY: migrate/goto
migrate/goto:
	@echo 'Migrating to version $(version)...'
	@migrate -path ./migrations -database "$(DB_DSN)" goto $(version)

## migrate/force: force migration version (usage: make migrate/force version=5)
.PHONY: migrate/force
migrate/force:
	@echo 'Forcing migration to version $(version)...'
	@migrate -path ./migrations -database "$(DB_DSN)" force $(version)

## migrate/version: check current migration version
.PHONY: migrate/version
migrate/version:
	@migrate -path ./migrations -database "$(DB_DSN)" version

## migrate/fix: fix dirty migration state
.PHONY: migrate/fix
migrate/fix:
	@echo 'Checking migration status...'
	@migrate -path ./migrations -database "$(DB_DSN)" version > /tmp/migrate_version 2>&1
	@cat /tmp/migrate_version
	@if grep -q "dirty" /tmp/migrate_version; then \
		version=$$(grep -o '[0-9]\+' /tmp/migrate_version | head -1); \
		echo "Found dirty migration at version $$version"; \
		echo "Forcing version $$version..."; \
		migrate -path ./migrations -database "$(DB_DSN)" force $$version; \
		echo "Re-running migration..."; \
		migrate -path ./migrations -database "$(DB_DSN)" down 1; \
		migrate -path ./migrations -database "$(DB_DSN)" up 1; \
	else \
		echo "No dirty migration found"; \
	fi
	@rm -f /tmp/migrate_version

# ==================================================================================== #
# DOCKER
# ==================================================================================== #

## docker/build: build Docker image
.PHONY: docker/build
docker/build:
	@echo 'Building Docker image...'
	@docker build -t salesapi:latest .

## docker/up: start Docker containers
.PHONY: docker/up
docker/up:
	@echo 'Starting Docker containers...'
	@docker-compose up -d

## docker/down: stop Docker containers
.PHONY: docker/down
docker/down:
	@echo 'Stopping Docker containers...'
	@docker-compose down

## docker/logs: view Docker container logs
.PHONY: docker/logs
docker/logs:
	@docker-compose logs -f

## docker/logs/api: view API container logs
.PHONY: docker/logs/api
docker/logs/api:
	@docker-compose logs -f api

## docker/logs/db: view database container logs
.PHONY: docker/logs/db
docker/logs/db:
	@docker-compose logs -f postgres

## docker/restart: restart Docker containers
.PHONY: docker/restart
docker/restart:
	@echo 'Restarting Docker containers...'
	@docker-compose restart

## docker/rebuild: rebuild and restart containers
.PHONY: docker/rebuild
docker/rebuild:
	@echo 'Rebuilding and restarting containers...'
	@docker-compose down
	@docker-compose build --no-cache
	@docker-compose up -d

## docker/ps: list running containers
.PHONY: docker/ps
docker/ps:
	@docker-compose ps

## docker/exec/api: execute shell in API container
.PHONY: docker/exec/api
docker/exec/api:
	@docker-compose exec api sh

## docker/exec/db: execute psql in database container
.PHONY: docker/exec/db
docker/exec/db:
	@docker-compose exec postgres psql -U sales -d sales

## docker/migrate/up: run migrations in Docker
.PHONY: docker/migrate/up
docker/migrate/up:
	@echo 'Running migrations in Docker...'
	@docker-compose run --rm migrate up

## docker/migrate/down: rollback last migration in Docker
.PHONY: docker/migrate/down
docker/migrate/down: confirm
	@echo 'Rolling back migration in Docker...'
	@docker-compose run --rm migrate down 1

## docker/clean: remove containers, volumes, and images
.PHONY: docker/clean
docker/clean: confirm
	@echo 'Cleaning up Docker resources...'
	@docker-compose down -v
	@docker rmi salesapi:latest 2>/dev/null || true

# ==================================================================================== #
# API REQUESTS & SEEDING
# ==================================================================================== #

## seed/admin: create and activate an admin user (admin@example.com / SecurePass123!)
.PHONY: seed/admin
seed/admin:
	@echo 'Creating admin user...'
	@curl -s -X POST http://localhost:4000/v1/users \
		-H "Content-Type: application/json" \
		-d '{"first_name":"Admin","last_name":"User","email":"admin@example.com","password":"SecurePass123!","role":"admin"}' > /dev/null
	@echo 'Activating admin user in database...'
	@docker-compose exec postgres psql -U sales -d sales -c "UPDATE users SET is_active = true WHERE email = 'admin@example.com';"
	@echo 'Admin user created and activated!'
	@echo 'Credentials: admin@example.com / SecurePass123!'

## req/register: register a new guest user (alice@example.com / SecurePass123!)
.PHONY: req/register
req/register:
	@echo 'Registering user...'
	@curl -i -X POST http://localhost:4000/v1/users \
		-H "Content-Type: application/json" \
		-d '{"first_name":"Alice","last_name":"Smith","email":"alice@example.com","password":"SecurePass123!"}'

## req/login: login as admin user
.PHONY: req/login
req/login:
	@echo 'Logging in as admin...'
	@curl -i -X POST http://localhost:4000/v1/tokens/authentication \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@example.com","password":"SecurePass123!"}'

## req/product: create a product (requires token=...)
.PHONY: req/product
req/product:
	@if [ -z "$(token)" ]; then echo "Error: token is required. Usage: make req/product token=YOUR_TOKEN"; exit 1; fi
	@echo 'Creating product...'
	@curl -i -X POST http://localhost:4000/v1/products \
		-H "Authorization: Bearer $(token)" \
		-H "Content-Type: application/json" \
		-d '{"name":"Laptop Pro","price":1299.99}'

## req/sale: create a sale (requires token=...)
.PHONY: req/sale
req/sale:
	@if [ -z "$(token)" ]; then echo "Error: token is required. Usage: make req/sale token=YOUR_TOKEN"; exit 1; fi
	@echo 'Creating sale...'
	@curl -i -X POST http://localhost:4000/v1/sales \
		-H "Authorization: Bearer $(token)" \
		-H "Content-Type: application/json" \
		-d '{"user_id":1,"product_id":1,"quantity":1}'

## req/chatbot: query the chatbot (requires token=...)
.PHONY: req/chatbot
req/chatbot:
	@if [ -z "$(token)" ]; then echo "Error: token is required. Usage: make req/chatbot token=YOUR_TOKEN"; exit 1; fi
	@echo 'Querying chatbot...'
	@curl -i -X POST http://localhost:4000/v1/chatbot \
		-H "Authorization: Bearer $(token)" \
		-H "Content-Type: application/json" \
		-d '{"message":"What are our top selling products?"}'

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## test: run all tests
.PHONY: test
test:
	@echo 'Running tests...'
	@go test -v -race -buildvcs ./...

## test/cover: run tests with coverage
.PHONY: test/cover
test/cover:
	@echo 'Running tests with coverage...'
	@go test -v -race -buildvcs -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo 'Coverage report generated at coverage.html'

## audit: tidy dependencies and format, vet, and test code
.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies...'
	@go mod tidy
	@go mod verify
	@echo 'Formatting code...'
	@go fmt ./...
	@echo 'Vetting code...'
	@go vet ./...
	@staticcheck ./...
	@echo 'Running tests...'
	@go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	@go mod tidy
	@go mod verify
	@echo 'Vendoring dependencies...'
	@go mod vendor

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/api: build the application for current OS
.PHONY: build/api
build/api:
	@echo 'Building salesapi...'
	@go build -ldflags='-s -w' -o=./bin/salesapi ./cmd/api
	@echo 'Binary created at ./bin/salesapi'

## build/linux: build for Linux
.PHONY: build/linux
build/linux:
	@echo 'Building salesapi for Linux...'
	@GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o=./bin/linux_amd64/salesapi ./cmd/api

## build/windows: build for Windows
.PHONY: build/windows
build/windows:
	@echo 'Building salesapi for Windows...'
	@GOOS=windows GOARCH=amd64 go build -ldflags='-s -w' -o=./bin/windows_amd64/salesapi.exe ./cmd/api

## build/mac: build for macOS
.PHONY: build/mac
build/mac:
	@echo 'Building salesapi for macOS...'
	@GOOS=darwin GOARCH=amd64 go build -ldflags='-s -w' -o=./bin/darwin_amd64/salesapi ./cmd/api

# ==================================================================================== #
# PRODUCTION
# ==================================================================================== #

## production/deploy: deploy to production (customize as needed)
.PHONY: production/deploy
production/deploy: confirm
	@echo 'Deploying to production...'
	@echo 'This is a placeholder - customize for your deployment strategy'

.PHONY: all
.DEFAULT_GOAL := help