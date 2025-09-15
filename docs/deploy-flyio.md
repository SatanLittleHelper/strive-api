## План деплоя Strive API на Fly.io (без карты)

### Зачем Fly.io
- Бесплатный план без привязки карты для небольших сервисов.
- Деплой контейнеров из Dockerfile, глобальный Anycast, HTTPS.

### Что подготовлено в репозитории
- `fly.toml` с настройкой сервиса (порт 8080, health `/health`).
- GitHub Actions: `.github/workflows/deploy-fly.yml` (деплой по push в `main` или вручную).

### Регистрация и установка
1) Создать аккаунт: https://fly.io/
2) Установить CLI: https://fly.io/docs/hands-on/install-flyctl/
3) Войти: `flyctl auth login`

### Создание приложения и секрета
1) В корне репозитория: `flyctl apps create strive-api` (или другое имя) и подставить его в `fly.toml` в `app`.
2) Секреты и переменные окружения:
   - `flyctl secrets set PORT=8080 LOG_LEVEL=INFO LOG_FORMAT=json \
     DB_HOST=... DB_PORT=5432 DB_USER=... DB_PASSWORD=... DB_NAME=... DB_SSL_MODE=require \
     JWT_SECRET=...`

### База данных
- Рекомендуется Neon (бесплатно): https://console.neon.tech/
- Использовать SSL: `require`.

### Деплой
- Из локали: `flyctl deploy --remote-only`
- Через GitHub Actions: добавить секрет `FLY_API_TOKEN` в Settings → Secrets → Actions, затем push в `main`.

### Проверка
- Открыть `https://<app>.fly.dev/health` → ожидается 200 и `{ "status": "ok" }`.
- Документация: `https://<app>.fly.dev/swagger/`.

### Роллбэк
- `flyctl releases` и `flyctl releases info <version>` → `flyctl deploy --image <предыдущий-образ>`.


