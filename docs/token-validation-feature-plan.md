# План фичи: Валидация JWT токенов на бэкенде

## Цель
Валидировать JWT на бэкенде для всех защищённых эндпоинтов, возвращая консистентные JSON‑ошибки и прокидывая данные пользователя в `context.Context`.

## Текущее состояние (коротко)
- Уже есть: `AuthService.ValidateToken` (HS256, exp/iat/nbf через `jwt.RegisteredClaims`), `AuthMiddleware` с разбором `Authorization: Bearer`, CORS пропускает `Authorization`.
- Замечания:
  - В `AuthMiddleware` в контекст кладётся `claims.UserID` типа `uuid.UUID`, а `GetUserIDFromContext` ожидает `string` — нужен один формат.
  - Ошибки пишутся через `http.Error` без `Content-Type: application/json` и единообразного формата.
  - Нет проверки `iss`/`aud`, нет допустимого `clock skew`, нет security‑логирования в middleware.
  - Не определён список публичных роутов и поведение для `OPTIONS`.

## Объём работ
- Валидация токена по подписи, `exp/nbf/iat`, `iss`, `aud`.
- Единый формат ошибок и security‑логирование.
- Прокидывание идентичного типа `user_id` в `context`.
- Маршрутизация: публичные/приватные маршруты, исключения для `OPTIONS`.
- Тесты: юнит + интеграционные.

## Дизайн и решения
- Алгоритм: HS256 (как сейчас).
- Источник секрета: env `JWT_SECRET`.
- Доп. проверки: `iss` и `aud` из конфигурации, `clock skew` 1–2 минуты.
- Тип в контексте: `string` для `user_id` и `string` для `email`.
- Единый writer ошибок: `{ "error": { "code": string, "message": string } }` с `Content-Type: application/json`.
- Публичные маршруты: `/health`, `/api/v1/auth/*`, `OPTIONS` для всех.
- Security‑логирование: причина отказа, ip, user‑agent, путь, корреляция по request id.

## План внедрения

### 1. Конфигурация
- Добавить в `config.Config`: `JWT.Issuer`, `JWT.Audience`, `JWT.ClockSkew` (duration).
- Проброс значений из env, описать в `README.md` и `env.example`.

### 2. Service: `ValidateToken`
- Проверять метод подписи HS256 (уже есть).
- Проверять `iss` и `aud`; допустить `ClockSkew` для `exp/nbf`.
- Возвращать осмысленные ошибки: `ErrInvalidSignature`, `ErrExpired`, `ErrNotBefore`, `ErrInvalidAudience`, `ErrInvalidIssuer`.
- Сохранить текущую структуру `Claims`; нормализовать `UserID` в `string` при отдаче наружу, либо сменить тип поля на `string`.

### 3. Middleware
- Поддержать bypass для `OPTIONS`, `/health`, `/api/v1/auth/*`.
- При ошибках всегда возвращать JSON через единый helper (код/сообщение): `UNAUTHORIZED`, `INVALID_TOKEN`, `TOKEN_EXPIRED`, `INVALID_AUDIENCE`, `INVALID_ISSUER`, `BEARER_REQUIRED`, `TOKEN_EMPTY`.
- Прокидывать в контекст: `user_id` (string), `user_email` (string).
- Добавить security‑логирование для всех отказов.

### 4. Роутинг
- Применить `AuthMiddleware` только к защищённым группам.
- Убедиться, что CORS и preflight идут до авторизации.

### 5. Тесты
- Юнит для `ValidateToken`:
  - валидный токен; истёкший; с будущим `nbf`; неверная подпись; неверные `iss/aud`; допустимый `clock skew`.
- Юнит/интеграция для middleware:
  - отсутствие `Authorization`; не `Bearer`; пустой токен; неверный токен; валидный токен (проверка контекста); `OPTIONS` bypass; публичные роуты bypass.
- Команда: `go test ./... -count=1 -race -timeout=60s`.

### 6. Логи и наблюдаемость
- Структурные логи на отказах в middleware.
- Метрики (по возможности): счётчики 401/403, причины отказов.

### 7. Документация
- Обновить `docs/frontend-auth-integration.md`: заголовок `Authorization: Bearer <access_token>`, коды ошибок и ответы, поведение refresh, сроки действия токена.
- Обновить `docs/swagger.*` описанием 401 с JSON‑ошибкой.

### 8. Деплой/безопасность
- Проверить наличие `JWT_SECRET` на окружениях; ротацию секрета описать в документации.
- Рассмотреть будущее: `jti` и deny‑list для отзывов, хранение refresh токенов в БД (хэшировано).

## Критерии приёмки
- Защищённые эндпоинты без валидного `Bearer` возвращают 401 с JSON `{ "error": { "code", "message" } }`.
- Валидный токен даёт 200, в хендлерах доступны `user_id` и `user_email` из `context`.
- Публичные маршруты и `OPTIONS` не требуют токен.
- Все тесты проходят: `go test ./... -count=1 -race -timeout=60s`.
- Логи содержат причины отказов без утечки токенов.

## Риски и компромиссы
- Без БД‑хранилища refresh‑токенов невозможно мгновенно отзывать access‑токены.
- Ротация `JWT_SECRET` потребует контролируемого окна совместимости токенов.

## Краткая сводка
- Добавляем `iss/aud/clock skew` и единый JSON‑формат ошибок.
- Чиним тип `user_id` в контексте.
- Bypass для `OPTIONS` и публичных маршрутов.
- Покрываем валидацию тестами и обновляем документацию.
