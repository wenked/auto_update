#include .env
# Simple Makefile for a Go project

# Build the application
all: build


build:
	@echo "Building..."
	@go build -o /main cmd/api/main.go
	@ls -la
# Run the application
run:
	@templ generate
	@go run cmd/api/main.go

db-stats:
	@echo "Getting DB stats..."
	@GOOSE_DRIVER=turso GOOSE_DBSTRING=${DB_URL} goose status
# Create DB container
migrate:
	@echo "Migrating..."
	@GOOSE_DRIVER=turso GOOSE_DBSTRING=${DB_URL} goose -dir "${MIGRATION_DIR}" up
revert:
	@echo "Reverting..."
	@GOOSE_DRIVER=turso GOOSE_DBSTRING=${DB_URL} goose down
seed:
	@echo "Seeding..."
	@GOOSE_DRIVER=turso GOOSE_DBSTRING=${DB_URL} GOOSE_MIGRATION_DIR=${MIGRATION_PATH} goose up

create-migration:
	@echo "Creating migration..."
	@GOOSE_DRIVER=turso GOOSE_DBSTRING=${DB_URL} GOOSE_MIGRATION_DIR=${MIGRATION_PATH} goose create $(name) sql

docker-run:
	@if docker compose up 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@go test ./tests -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
	    air; \
	    echo "Watching...";\
	else \
	    read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
	    if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
	        go install github.com/cosmtrek/air@latest; \
	        air; \
	        echo "Watching...";\
	    else \
	        echo "You chose not to install air. Exiting..."; \
	        exit 1; \
	    fi; \
	fi

.PHONY: all build run test clean
start: build
	   air

## css: build tailwindcss
.PHONY: css
css:
	./tailwindcss -i css/input.css -o css/output.css --minify

## css-watch: watch build tailwindcss
.PHONY: css-watch
css-watch:
	./tailwindcss -i css/input.css -o css/output.css --watch
