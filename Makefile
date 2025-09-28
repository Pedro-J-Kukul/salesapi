# importing environment variables from .envrc file
include .envrc

#########################################
# Makefile for managing PostgreSQL #
#########################################

# login to postgres as api user
.PHONY: psql/login
psql/login:
	@echo "Logging into PostgreSQL as user 'jam'..."
	@psql ${DB_DSN}

# login to postgres as postgres user
.PHONY: psql
psql:
	@echo "Logging into PostgreSQL as user 'postgres'..."
	@sudo -u postgres psql


#########################################
# Make commands for migrations 			#
#########################################

#  Create a new migration file
.PHONY: migration/new
migration/new:
	@if [ -z "$(name)" ]; then \
		echo "Error: Please provide a name for the migration using 'make migration/create name=your_migration_name'"; \
		exit 1; \
	fi
	@if [ ! -d "./migrations" ]; then mkdir ./migrations; fi
	migrate create -seq -ext=.sql -dir=./migrations $(name)

# Apply all up migrations
.PHONY: migration/up
migration/up:
	migrate -path ./migrations -database "$(DB_DSN)" up 

# Apply all down 1 migrations
.PHONY: migration/down
migration/down:
	migrate -path ./migrations -database "$(DB_DSN)" down 

# fix and reapply the last migration and fix dirty state
.PHONY: migration/fix
migration/fix:
	@echo 'Checking migration status...'
	@migrate -path ./migrations -database "${DB_DSN}" version > /tmp/migrate_version 2>&1
	@cat /tmp/migrate_version
	@if grep -q "dirty" /tmp/migrate_version; then \
		version=$$(grep -o '[0-9]\+' /tmp/migrate_version | head -1); \
		echo "Found dirty migration at version $$version"; \
		echo "Forcing version $$version..."; \
		migrate -path ./migrations -database "${DB_DSN}" force $$version; \
		echo "Running down migration..."; \
		migrate -path ./migrations -database "${DB_DSN}" down 1; \
		echo "Running up migration..."; \
		migrate -path ./migrations -database "${DB_DSN}" up; \
	else \
		echo "No dirty migration found"; \
	fi
	@rm -f /tmp/migrate_version
