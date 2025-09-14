# Strive API

![Pull Request](https://github.com/aleksandr/strive-api/workflows/Pull%20Request/badge.svg)

A modern workout diary API built with Go, featuring user authentication, JWT tokens, and comprehensive testing.

## 🚀 Features

- **User Authentication**: JWT-based authentication with access and refresh tokens
- **Password Security**: bcrypt password hashing
- **Database Integration**: PostgreSQL with automatic migrations
- **Comprehensive Testing**: 17 unit tests with 73%+ code coverage
- **API Documentation**: OpenAPI/Swagger documentation
- **Containerization**: Docker and Docker Compose support
- **Structured Logging**: JSON/text logging with configurable levels
- **Graceful Shutdown**: Proper server lifecycle management
- **CI/CD Pipeline**: Automated testing and deployment with GitHub Actions

## 📋 Requirements

- Go 1.22+
- PostgreSQL 15+
- Docker & Docker Compose (optional)
- Make (optional, for convenience commands)

## 🛠️ Installation

### Option 1: Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd strive-api
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   ```bash
   cp env.example .env
   # Edit .env with your configuration
   ```

4. **Start PostgreSQL**
   ```bash
   make db-up
   ```

5. **Run the application**
   ```bash
   make run-dev
   ```

### Option 2: Docker Compose

1. **Clone and start**
   ```bash
   git clone <repository-url>
   cd strive-api
   docker compose up --build
   ```

## 🔧 Configuration

The application uses environment variables for configuration. Copy `env.example` to `.env` and customize:

```env
PORT=8080
LOG_LEVEL=INFO
LOG_FORMAT=json
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=strive
DB_SSL_MODE=disable
JWT_SECRET=your-secret-key-change-in-production
```

## 📚 API Documentation

Once the server is running, visit:
- **Swagger UI**: http://localhost:8080/swagger/
- **OpenAPI JSON**: http://localhost:8080/swagger/doc.json

## 🔌 API Endpoints

### Public Endpoints

- `GET /health` - Health check
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login

### Protected Endpoints (require JWT token)

- `GET /api/v1/user/profile` - Get user profile

### Example Usage

**Register a new user:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

**Access protected endpoint:**
```bash
curl -X GET http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 🧪 Testing

Run the test suite:

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run tests with coverage report
make test-coverage

# Run specific test package
go test ./internal/services -v
go test ./internal/http -v
```

## 🐳 Docker Commands

```bash
# Build and start all services
docker compose up --build

# Start only database
docker compose up postgres

# Stop all services
docker compose down

# Reset database
docker compose down -v
docker compose up postgres
```

## 🛠️ Development Commands

```bash
# Format code
make format

# Run linter
make lint

# Build binary
make build

# Run with development settings
make run-dev

# Start database
make db-up

# Stop database
make db-down

# Reset database
make db-reset

# Run database migrations
make migrate-up

# Rollback database migrations
make migrate-down
```

## 🔄 CI/CD

Проект использует GitHub Actions для автоматического тестирования:

### Pull Request Workflow
Запускается при создании PR в любую ветку и выполняет:

- 🎨 **Форматирование кода** - проверка gofumpt и goimports
- 🔍 **Линтинг** - golangci-lint проверки
- 🚀 **Миграции БД** - автоматический запуск миграций
- 🧪 **Тестирование** - полный набор unit тестов
- 📊 **Покрытие кода** - генерация отчетов покрытия
- 🔨 **Сборка** - проверка компиляции приложения и инструментов
- 🐳 **Docker** - сборка Docker образа
- 🔒 **Безопасность** - базовые проверки безопасности

Подробная документация: [docs/CI_CD.md](docs/CI_CD.md)

## 📁 Project Structure

```
strive-api/
├── cmd/
│   ├── server/          # Main application entry point
│   └── migrate/         # Database migration tool
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection and health
│   ├── http/           # HTTP handlers and middleware
│   ├── logger/         # Structured logging
│   ├── migrate/        # Database migration logic
│   ├── models/         # Data models
│   ├── repositories/   # Data access layer
│   ├── services/       # Business logic
│   └── validation/     # Input validation
├── .github/
│   ├── workflows/      # GitHub Actions CI/CD
│   └── dependabot.yml # Dependency updates
├── docs/               # API documentation and guides
├── migrations/         # Database migration files
├── docker-compose.yml  # Docker Compose configuration
├── Dockerfile         # Docker image definition
├── Makefile          # Development commands
└── README.md         # This file
```

## 🔐 Security Features

- **Password Hashing**: bcrypt with configurable cost
- **JWT Tokens**: HMAC SHA256 signed tokens
- **Token Expiration**: Access tokens (15 min), Refresh tokens (7 days)
- **Input Validation**: Request validation and sanitization
- **Graceful Error Handling**: No sensitive data leakage

## 📊 Testing Coverage

- **HTTP Handlers**: 73% coverage
- **Services**: 72.5% coverage
- **Total Tests**: 17 unit tests
- **Test Types**: AuthService, HTTP handlers, middleware

## 🚀 Deployment

### Production Deployment

1. **Set production environment variables**
2. **Build Docker image**
   ```bash
   docker build -t strive-api .
   ```
3. **Run with production database**
   ```bash
   docker run -d \
     -p 8080:8080 \
     -e DB_HOST=your-db-host \
     -e JWT_SECRET=your-production-secret \
     strive-api
   ```

### Environment Variables for Production

```env
PORT=8080
LOG_LEVEL=INFO
LOG_FORMAT=json
DB_HOST=your-postgres-host
DB_PORT=5432
DB_USER=your-db-user
DB_PASSWORD=your-secure-password
DB_NAME=strive
DB_SSL_MODE=require
JWT_SECRET=your-very-secure-jwt-secret
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support, email support@example.com or create an issue in the repository.

## 📈 Roadmap

- [ ] Integration tests with testcontainers
- [ ] Rate limiting
- [ ] Metrics and monitoring (Prometheus)
- [ ] CI/CD pipeline
- [ ] Additional business logic (exercises, workouts, sets)
- [ ] File upload support
- [ ] Email notifications