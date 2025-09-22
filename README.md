# Strive API

A modern workout diary API built with Go, featuring user authentication, JWT tokens, and comprehensive testing.

## 🚀 Features

- **User Authentication**: JWT-based authentication with access and refresh tokens
- **User Profile Management**: Complete user profile with theme preferences
- **Theme Support**: Light/Dark theme selection for users
- **Password Security**: bcrypt password hashing with strong validation
- **Database Integration**: PostgreSQL with automatic migrations
- **Clean Architecture**: Separation of concerns with services and repositories
- **Comprehensive Testing**: Unit tests with race detection
- **API Documentation**: OpenAPI/Swagger documentation
- **Containerization**: Docker and Docker Compose support
- **Structured Logging**: JSON/text logging with configurable levels
- **Graceful Shutdown**: Proper server lifecycle management

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
   # Important: Set a strong JWT_SECRET for production!
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
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - User logout

### Protected Endpoints (require JWT token)

- `GET /api/v1/user/me` - Get user profile (includes theme)
- `PUT /api/v1/user/theme` - Update user theme preference

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

**Get user profile:**
```bash
curl -X GET http://localhost:8080/api/v1/user/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Update user theme:**
```bash
curl -X PUT http://localhost:8080/api/v1/user/theme \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"theme":"dark"}'
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
```

## 📁 Project Structure

```
strive-api/
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection and health
│   ├── http/           # HTTP handlers and middleware
│   │   ├── auth_handlers.go    # Authentication endpoints
│   │   ├── user_handlers.go    # User profile and theme endpoints
│   │   ├── middleware.go       # Security and logging middleware
│   │   └── ...
│   ├── logger/         # Structured logging
│   ├── migrate/        # Database migrations
│   ├── models/         # Data models (User, Theme, etc.)
│   ├── repositories/   # Data access layer
│   ├── services/       # Business logic
│   │   ├── auth_service.go     # Authentication logic
│   │   ├── user_service.go     # User profile and theme logic
│   │   └── ...
│   └── validation/     # Input validation
├── docs/               # Generated API documentation
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
JWT_ISSUER=strive-api
JWT_AUDIENCE=strive-app
JWT_CLOCK_SKEW=2m
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