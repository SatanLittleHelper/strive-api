# Этап 7: Документация и развертывание

## Цель этапа
Создать документацию API и настроить контейнеризацию для развертывания.

## 7.1 OpenAPI документация
- [ ] Создать OpenAPI спецификацию
- [ ] Добавить аннотации к хендлерам
- [ ] Настроить генерацию документации

### Детали реализации
- Использовать `swaggo/swag` для генерации OpenAPI из комментариев
- Создать полную спецификацию для всех endpoints
- Добавить примеры запросов и ответов
- Настроить Swagger UI для интерактивной документации

### Пример аннотаций
```go
// @Summary Login user
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // ...
}
```

## 7.2 Контейнеризация
- [ ] Создать Dockerfile для приложения
- [ ] Создать docker-compose.yml для разработки
- [ ] Добавить PostgreSQL в docker-compose

### Dockerfile
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
CMD ["./server"]
```

### docker-compose.yml
```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=strive
      - DB_USER=strive
      - DB_PASSWORD=strive
    depends_on:
      - postgres

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=strive
      - POSTGRES_USER=strive
      - POSTGRES_PASSWORD=strive
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

## 7.3 Настройка окружения
- [ ] Создать .env.example файл
- [ ] Добавить docker-compose для продакшена
- [ ] Настроить переменные окружения для разных сред

## 7.4 Документация
- [ ] Создать README.md с инструкциями по запуску
- [ ] Добавить документацию по API
- [ ] Создать руководство по развертыванию

### Структура README.md
- Описание проекта
- Требования
- Установка и запуск
- API документация
- Развертывание
- Контрибьюция

## Критерии готовности
- [ ] OpenAPI спецификация полная и корректная
- [ ] Swagger UI доступен и работает
- [ ] Dockerfile создает рабочий образ
- [ ] docker-compose запускает полное окружение
- [ ] README.md содержит все необходимые инструкции

## Время выполнения
Ориентировочно: 1-2 дня

## Предыдущий этап
[Этап 6: Тестирование](./06-testing.md)

## Следующий этап
[Этап 8: Дополнительные возможности](./08-additional.md)
