# Deployment Guide

## Локальная разработка

### 1. Запуск локального окружения

```bash
# Запуск базы данных
make db-up

# Запуск приложения в dev режиме
make run-dev
```

### 2. Запуск через Docker (dev окружение)

```bash
# Запуск dev окружения
make dev-up

# Пересборка и запуск
make dev-up-build

# Просмотр логов
make dev-logs

# Остановка
make dev-down
```

### 3. Запуск через Docker (production окружение)

```bash
# Запуск production окружения
make prod-up

# Пересборка и запуск
make prod-up-build

# Просмотр логов
make prod-logs

# Остановка
make prod-down
```

## Деплой на Render.com

### 1. Настройка переменных окружения

В Render Dashboard настройте следующие переменные:

#### Обязательные переменные:
- `JWT_SECRET` - секретный ключ (минимум 32 символа)
- `DB_HOST` - хост базы данных
- `DB_USER` - пользователь БД
- `DB_PASSWORD` - пароль БД
- `DB_NAME` - имя БД

#### CORS настройки:
- `CORS_ALLOWED_ORIGINS` - разрешенные домены (через запятую)
- `CORS_ALLOW_CREDENTIALS=true`

#### Cookie настройки:
- `COOKIE_SECURE=true` (для HTTPS)
- `COOKIE_SAMESITE=Strict` (для production)
- `COOKIE_DOMAIN=` (пустое для текущего домена)

### 2. Деплой через Render

#### Development стенд:
1. Подключите репозиторий к Render
2. Используйте `render.dev.yaml` для dev стенда
3. Настройте переменные окружения
4. Деплой произойдет автоматически при push в main

#### Production стенд:
1. Используйте `render.yaml` для production
2. Настройте все переменные окружения
3. Убедитесь, что `CORS_ALLOWED_ORIGINS` содержит ваш production домен
4. Деплой произойдет автоматически при push в main

### 3. Проверка деплоя

```bash
# Проверка здоровья
curl https://your-app.onrender.com/health

# Проверка детального статуса
curl https://your-app.onrender.com/health/detailed

# Проверка CORS
curl -H "Origin: https://your-frontend.com" \
     -H "Access-Control-Request-Method: POST" \
     -H "Access-Control-Request-Headers: Content-Type" \
     -X OPTIONS \
     https://your-app.onrender.com/api/v1/auth/login
```

## Настройка для разных окружений

### Development (локальная разработка)
- `COOKIE_SECURE=false`
- `COOKIE_SAMESITE=Lax`
- `CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:4200`

### Staging (dev стенд)
- `COOKIE_SECURE=true`
- `COOKIE_SAMESITE=None`
- `CORS_ALLOWED_ORIGINS=https://your-dev-frontend.com`

### Production
- `COOKIE_SECURE=true`
- `COOKIE_SAMESITE=Strict`
- `CORS_ALLOWED_ORIGINS=https://your-production-frontend.com`

## Troubleshooting

### Проблемы с куки
1. Убедитесь, что `CORS_ALLOW_CREDENTIALS=true`
2. Проверьте настройки `COOKIE_SECURE` и `COOKIE_SAMESITE`
3. Для cross-domain куки нужен `SameSite=None` и `Secure=true`

### Проблемы с CORS
1. Проверьте `CORS_ALLOWED_ORIGINS`
2. Убедитесь, что домен точно совпадает
3. Проверьте, что `CORS_ALLOW_CREDENTIALS=true`

### Проблемы с базой данных
1. Проверьте подключение к БД
2. Убедитесь, что миграции выполнены
3. Проверьте SSL настройки (`DB_SSL_MODE`)
