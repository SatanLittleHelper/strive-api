.PHONY: run run-dev test lint format clean build db-up db-down db-reset

run:
	go run ./cmd/server

run-dev:
	@echo "Starting server with development environment variables..."
	PORT=8080 \
	LOG_LEVEL=INFO \
	LOG_FORMAT=json \
	DB_HOST=localhost \
	DB_PORT=5432 \
	DB_USER=postgres \
	DB_PASSWORD=password \
	DB_NAME=strive \
	DB_SSL_MODE=disable \
	JWT_SECRET=dev-secret-key-12345 \
	go run ./cmd/server

db-up:
	@echo "Starting PostgreSQL database..."
	docker compose up -d postgres

db-down:
	@echo "Stopping PostgreSQL database..."
	docker compose down

db-reset:
	@echo "Resetting PostgreSQL database..."
	docker compose down -v
	docker compose up -d postgres

test:
	go test ./... -count=1 -race -timeout=60s

test-unit:
	@echo "Running unit tests..."
	go test ./internal/services ./internal/http -count=1 -race -timeout=60s

test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -count=1 -race -timeout=60s -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint:
	golangci-lint run

format:
	gofumpt -l -w .
	goimports -w .

build:
	go build -o bin/server ./cmd/server

clean:
	rm -rf bin/

deps:
	go mod download
	go mod tidy

install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install mvdan.cc/gofumpt@latest
	go install golang.org/x/tools/cmd/goimports@latest
