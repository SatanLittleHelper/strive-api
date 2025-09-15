## План деплоя Strive API на Render (без карты)

### Почему Render
- Бесплатный веб‑сервис (Free plan), деплой из Git или Dockerfile.
- Не требует привязки карты для старта.

### Что добавлено в репозиторий
- `render.yaml` — blueprint для веб‑сервиса на плане Free.

### Регистрация
- Создать аккаунт: https://dashboard.render.com/register

### Импорт репозитория и деплой
1) Войти в Render и подключить GitHub.
2) New + → Blueprint → выбрать репозиторий.
3) Render найдёт `render.yaml` и предложит создать сервис `strive-api`.
4) Подтвердить деплой.

### Переменные окружения
В `render.yaml` указаны ключи. Заполнить секретные:
- DB_HOST, DB_USER, DB_PASSWORD, DB_NAME, JWT_SECRET
Заданы по умолчанию:
- PORT=8080, LOG_LEVEL=INFO, LOG_FORMAT=json, DB_PORT=5432, DB_SSL_MODE=require

### База данных
- Рекомендуется Neon (без карты): https://console.neon.tech/
- Включить SSL: `require`.

### Health check и маршруты
- Render проверяет `/health`.
- Swagger: `/swagger/`.

### Автодеплой
- Включён `autoDeploy: true` — пуш в default ветку запускает новый деплой.

### Роллбэк
- В Render → сервис → Deploys → выбрать предыдущий деплой и Rollback.


