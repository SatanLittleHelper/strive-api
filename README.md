# Strive API

API для дневника тренировок с аутентификацией и управлением пользователями.

## Быстрый старт

### 1. Установка зависимостей

```bash
# Go модули
go mod download

# Инструменты разработки (опционально)
make install-tools
```

### 2. Настройка базы данных

Создайте базу данных PostgreSQL:

```sql
CREATE DATABASE strive;
CREATE USER strive_user WITH PASSWORD 'password';
GRANT ALL PRIVILEGES ON DATABASE strive TO strive_user;
```

### 3. Запуск приложения

#### Для разработки (с предустановленными переменными окружения):
```bash
make run-dev
```

#### С кастомными переменными окружения:
```bash
PORT=8080 \
DB_HOST=localhost \
DB_PORT=5432 \
DB_USER=postgres \
DB_PASSWORD=password \
DB_NAME=strive \
JWT_SECRET=your-secret-key \
go run ./cmd/server
```

## Переменные окружения

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `PORT` | Порт сервера | `8080` |
| `LOG_LEVEL` | Уровень логирования | `INFO` |
| `LOG_FORMAT` | Формат логов | `json` |
| `DB_HOST` | Хост базы данных | `localhost` |
| `DB_PORT` | Порт PostgreSQL | `5432` |
| `DB_USER` | Пользователь БД | `postgres` |
| `DB_PASSWORD` | Пароль БД | `password` |
| `DB_NAME` | Имя базы данных | `strive` |
| `DB_SSL_MODE` | SSL режим | `disable` |
| `JWT_SECRET` | Секрет для JWT токенов | `your-secret-key-change-in-production` |

## API Endpoints

### Публичные endpoints

- `GET /health` - Health check
- `POST /api/v1/auth/register` - Регистрация пользователя
- `POST /api/v1/auth/login` - Вход в систему

### Защищенные endpoints (требуют JWT токен)

- `GET /api/v1/user/profile` - Профиль пользователя

## Примеры использования

### Регистрация пользователя

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123"
  }'
```

### Вход в систему

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123"
  }'
```

### Доступ к защищенному endpoint

```bash
curl -X GET http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Команды разработки

```bash
make run-dev     # Запуск с переменными разработки
make test        # Запуск тестов
make lint        # Проверка кода линтером
make format      # Форматирование кода
make build       # Сборка бинарника
```

## Структура проекта

```
strive-api/
├── cmd/server/              # Точка входа приложения
├── internal/
│   ├── config/             # Конфигурация
│   ├── database/           # Подключение к БД
│   ├── http/               # HTTP хендлеры и middleware
│   ├── logger/             # Логирование
│   ├── migrate/            # Миграции БД
│   ├── models/             # Модели данных
│   ├── repositories/       # Репозитории для работы с БД
│   └── services/           # Бизнес-логика
├── migrations/             # SQL миграции
├── docs/                   # Документация
└── Makefile               # Команды разработки
```

## Безопасность

- Пароли хешируются с помощью bcrypt
- JWT токены подписываются HMAC SHA256
- Access токены действительны 15 минут
- Refresh токены действительны 7 дней

## Разработка

1. Клонируйте репозиторий
2. Установите PostgreSQL
3. Создайте базу данных
4. Запустите `make run-dev`
5. API будет доступен на `http://localhost:8080`
