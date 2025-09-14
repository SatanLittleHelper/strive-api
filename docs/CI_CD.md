# CI/CD Configuration

Этот документ описывает настройку CI/CD pipeline для проекта Strive API.

## Обзор

В проекте настроен GitHub Actions workflow:

### Pull Request Workflow (`.github/workflows/pull-request.yml`)

Запускается при создании или обновлении pull request к **любой** ветке.

**Выполняемые проверки:**
- 🎨 **Форматирование кода** - проверка `gofumpt` и `goimports`
- 🔍 **Линтинг кода** - `golangci-lint`
- 🚀 **Миграции БД** - запуск миграций через отдельный инструмент
- 🧪 **Тестирование** - выполнение всех unit тестов
- 📊 **Покрытие кода** - генерация отчетов и загрузка в Codecov
- 🔨 **Сборка приложения** - компиляция основного приложения
- 🔨 **Сборка инструментов** - компиляция инструмента миграций
- 🐳 **Docker сборка** - создание Docker образа
- 🔒 **Проверки безопасности** - базовая проверка зависимостей


## Требования для проходжения CI

Для успешного прохождения CI/CD pipeline ваш код должен:

1. **Форматирование**: Код должен быть отформатирован с помощью `gofumpt` и `goimports`
   ```bash
   make format
   ```

2. **Линтинг**: Не должно быть ошибок линтера
   ```bash
   make lint
   ```

3. **Тесты**: Все тесты должны проходить
   ```bash
   make test
   ```

4. **Сборка**: Код должен компилироваться без ошибок
   ```bash
   go build ./cmd/server
   ```

## Инструмент для миграций

Создан отдельный инструмент `cmd/migrate` для управления миграциями базы данных:

```bash
# Применить миграции
go run ./cmd/migrate -direction=up
# или через Makefile
make migrate-up

# Откатить миграции
go run ./cmd/migrate -direction=down
# или через Makefile
make migrate-down
```

Это решение отделяет логику миграций от основного приложения, что лучше для:
- Чистоты архитектуры
- Тестирования в CI/CD
- Управления инфраструктурой

## Локальная разработка

Перед отправкой pull request рекомендуется выполнить локально:

```bash
# Форматирование кода
make format

# Проверка линтером
make lint

# Запуск тестов
make test

# Проверка сборки
go build ./cmd/server

# Запуск миграций (если нужно)
make migrate-up
```

## Тестовая среда

Pull Request workflow использует следующую конфигурацию:

- **Go версия**: 1.22
- **PostgreSQL**: 17
- **База данных**: `strive_test`
- **Переменные окружения**:
  - `DB_HOST=localhost`
  - `DB_PORT=5432`
  - `DB_USER=postgres`
  - `DB_PASSWORD=password`
  - `DB_NAME=strive_test`
  - `DB_SSL_MODE=disable`
  - `JWT_SECRET=test-secret-key-12345`

## Уведомления и отчеты

Workflow создает подробный отчет с результатами всех проверок в виде таблицы в GitHub Summary, что позволяет быстро оценить статус PR.

## Автоматические обновления

### Dependabot

Настроен Dependabot для автоматического обновления:
- Go модулей (еженедельно)
- Docker образов (еженедельно)
- GitHub Actions (еженедельно)

## Branch Protection Rules (рекомендуется настроить в GitHub)

Для ветки `main` рекомендуется настроить:

1. **Require pull request reviews before merging**
   - Require approvals: 1
   - Require review from code owners: ✅

2. **Require status checks to pass before merging**
   - Require branches to be up to date before merging: ✅
   - Required status checks:
     - `test` (Pull Request workflow)

3. **Require conversation resolution before merging**: ✅

4. **Restrict pushes that create files**: ✅

## Мониторинг и отчеты

- **Test Coverage**: Отчеты загружаются в Codecov
- **Security**: Базовые проверки безопасности включены
- **Dependencies**: Dependabot создает PR для обновлений


## Troubleshooting

### Проблемы с форматированием

```bash
# Исправить форматирование
make format

# Проверить изменения
git diff
```

### Проблемы с тестами

```bash
# Запустить тесты локально с подробным выводом
go test ./... -v

# Запустить тесты с покрытием
make test-coverage
```

### Проблемы с миграциями

```bash
# Запустить миграции локально
make db-up
make run-dev
```
