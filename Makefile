.PHONY: run run-dev test lint format clean build db-up db-down db-reset docker-up docker-down docker-restart docker-logs

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
	JWT_SECRET=dev-secret-key-12345-very-long-for-security \
	JWT_ISSUER=strive-api \
	JWT_AUDIENCE=strive-app \
	JWT_CLOCK_SKEW=2m \
	CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001,http://localhost:4200,http://127.0.0.1:3000,http://127.0.0.1:4200 \
	CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS \
	CORS_ALLOWED_HEADERS=Accept,Authorization,Content-Type,X-Request-ID \
	CORS_EXPOSED_HEADERS=X-Request-ID \
	CORS_ALLOW_CREDENTIALS=true \
	CORS_MAX_AGE=86400 \
	COOKIE_SECURE=false \
	COOKIE_SAMESITE=Lax \
	COOKIE_DOMAIN= \
	RATE_LIMIT_ENABLED=true \
	RATE_LIMIT_AUTH_PER_MINUTE=5 \
	RATE_LIMIT_GENERAL_PER_MINUTE=60 \
	RATE_LIMIT_BURST_SIZE=10 \
	EXERCISEDB_ENABLED=true \
	EXERCISEDB_BASE_URL=https://exercise.hellogym.io \
	EXERCISEDB_TIMEOUT=30s \
	EXERCISEDB_RETRY_COUNT=3 \
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

docs:
	@echo "Generating API documentation..."
	swag init -g cmd/server/main.go
	@echo "Documentation generated in docs/"

docs-serve:
	@echo "Serving API documentation..."
	@echo "Visit http://localhost:8080/swagger/ after starting the server"

docker-build:
	@echo "Building Docker image..."
	docker build -t strive-api .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 strive-api

docker-up:
	@echo "Starting application with Docker Compose..."
	docker compose up -d

docker-up-build:
	@echo "Building and starting application with Docker Compose..."
	docker compose up --build -d

docker-down:
	@echo "Stopping Docker Compose services..."
	docker compose down

docker-restart:
	@echo "Restarting application with rebuild..."
	docker compose down
	docker compose up --build -d

docker-logs:
	@echo "Showing application logs..."
	docker compose logs -f app

docker-logs-all:
	@echo "Showing all logs..."
	docker compose logs -f

# Development deployment commands
dev-up:
	@echo "Starting development environment..."
	docker compose -f docker-compose.dev.yml up -d

dev-up-build:
	@echo "Building and starting development environment..."
	docker compose -f docker-compose.dev.yml up --build -d

dev-down:
	@echo "Stopping development environment..."
	docker compose -f docker-compose.dev.yml down

dev-restart:
	@echo "Restarting development environment..."
	docker compose -f docker-compose.dev.yml down
	docker compose -f docker-compose.dev.yml up --build -d

dev-logs:
	@echo "Showing development logs..."
	docker compose -f docker-compose.dev.yml logs -f app

dev-logs-all:
	@echo "Showing all development logs..."
	docker compose -f docker-compose.dev.yml logs -f

# Production deployment commands
prod-up:
	@echo "Starting production environment..."
	docker compose up -d

prod-up-build:
	@echo "Building and starting production environment..."
	docker compose up --build -d

prod-down:
	@echo "Stopping production environment..."
	docker compose down

prod-restart:
	@echo "Restarting production environment..."
	docker compose down
	docker compose up --build -d

prod-logs:
	@echo "Showing production logs..."
	docker compose logs -f app

prod-logs-all:
	@echo "Showing all production logs..."
	docker compose logs -f

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
