.PHONY: dev test lint migrate-up migrate-down migrate-status migrate-create build docker-build docker-run clean e2e-start e2e-stop e2e-test e2e-full

# Variables
GO := go
MIGRATE := migrate
DATABASE_URL ?= postgres://postgres:password@localhost:5432/localvisibilityos?sslmode=disable

# Development
dev:
	@echo "Starting development server..."
	$(GO) run cmd/server/main.go

# Testing
test:
	@echo "Running tests..."
	$(GO) test ./...

test-verbose:
	@echo "Running tests with verbose output..."
	$(GO) test -v ./...

# Linting
lint:
	@echo "Running linters..."
	$(GO) vet ./...
	@echo "Checking formatting..."
	@if [ $$($(GO) fmt -l . | wc -l) -ne 0 ]; then \
		echo "Code is not formatted. Run 'go fmt ./...'"; \
		exit 1; \
	fi

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Database migrations
migrate-up:
	@echo "Running migrations up..."
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	@echo "Running migrations down..."
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" down 1

migrate-status:
	@echo "Checking migration status..."
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" version

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration: $(NAME)"
	$(MIGRATE) create -ext sql -dir migrations -seq $(NAME)

# Build
build:
	@echo "Building application..."
	$(GO) build -o bin/farohq-core-app cmd/server/main.go

# Docker
docker-build:
	@echo "Building Docker image..."
	docker build -t farohq-core-app:latest .

docker-run:
	@echo "Running Docker container..."
	docker run --rm -p 8080:8080 \
		-e PORT=8080 \
		-e DATABASE_URL="$(DATABASE_URL)" \
		-e CLERK_JWKS_URL="$(CLERK_JWKS_URL)" \
		farohq-core-app:latest

# Clean
clean:
	@echo "Cleaning up..."
	$(GO) clean
	rm -rf bin/
	rm -f *.test

# Install dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# Install migrate tool (if not already installed)
install-migrate:
	@echo "Installing golang-migrate..."
	$(GO) install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# E2E Testing
e2e-start:
	@echo "Starting E2E environment..."
	@./scripts/start-e2e.sh

e2e-stop:
	@echo "Stopping E2E environment..."
	@./scripts/stop-e2e.sh

e2e-test:
	@echo "Running E2E tests..."
	@./scripts/run-e2e-tests.sh

e2e-full: e2e-start
	@echo "E2E environment ready. Run 'make e2e-test' to run tests, 'make e2e-stop' to stop."
