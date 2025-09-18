# План улучшения безопасности Strive API

## Обзор

Данный документ содержит детальный план по устранению выявленных проблем безопасности в проекте Strive API. План структурирован по приоритетам и включает конкретные шаги реализации.

## 🔴 Критические проблемы (Приоритет 1)

### 1. Слабый JWT Secret по умолчанию

**Проблема**: Используется предсказуемый секрет по умолчанию `"your-secret-key-change-in-production"`

**Файл**: `internal/config/config.go:70`

**Риск**: Компрометация всех JWT токенов, возможность подделки токенов

**Решение**:
- [x] Удалить значение по умолчанию для `JWT_SECRET`
- [x] Добавить валидацию силы секрета (минимум 32 символа)
- [x] Добавить проверку на production окружение
- [x] Обновить документацию с требованиями к секрету

**Статус**: ✅ **ВЫПОЛНЕНО** - Коммит: `b02f90c`, Ветка: `security/fix-jwt-secret-validation`

**Код для реализации**:
```go
func (c *Config) Validate() error {
    // ... существующие проверки
    
    if c.JWT.Secret == "" {
        return fmt.Errorf("JWT_SECRET is required")
    }
    
    if len(c.JWT.Secret) < 32 {
        return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
    }
    
    // Проверка на слабые секреты
    weakSecrets := []string{
        "your-secret-key-change-in-production",
        "secret",
        "password",
        "1234567890",
    }
    
    for _, weak := range weakSecrets {
        if c.JWT.Secret == weak {
            return fmt.Errorf("JWT_SECRET is too weak, use a strong random secret")
        }
    }
    
    return nil
}
```

### 2. Отсутствие Rate Limiting

**Проблема**: Нет защиты от брутфорс атак на эндпоинты аутентификации

**Риск**: Атаки на подбор паролей, DoS атаки

**Решение**:
- [x] Создать rate limiting middleware
- [x] Настроить разные лимиты для разных эндпоинтов
- [x] Добавить конфигурацию через ENV переменные
- [x] Интегрировать с SecurityLogger

**Статус**: ✅ **ВЫПОЛНЕНО** - Коммит: `7395654`, Ветка: `security/add-rate-limiting-middleware`

### 4. Добавление HTTP Security Headers

**Проблема**: Отсутствие заголовков безопасности

**Риск**: XSS, clickjacking, MITM атаки

**Решение**:
- [x] Создать security headers middleware
- [x] Добавить HSTS, CSP, X-Frame-Options
- [x] Настроить X-Content-Type-Options
- [x] Добавить Referrer-Policy

**Статус**: ✅ **ВЫПОЛНЕНО** - Коммит: `security/implement-security-headers`, Ветка: `security/implement-security-headers`

**Файлы для создания**:
- `internal/http/rate_limit_middleware.go`
- Обновить `internal/config/config.go`

**Конфигурация**:
```go
type RateLimitConfig struct {
    AuthRequestsPerMinute int
    GeneralRequestsPerMinute int
    BurstSize int
}
```

### 3. Небезопасная конфигурация CORS

**Проблема**: Хардкод доменов в коде, включая подозрительный домен

**Файл**: `internal/http/cors.go:18`

**Риск**: CORS атаки, несанкционированный доступ

**Решение**:
- [x] Вынести CORS конфигурацию в ENV переменные
- [x] Добавить валидацию доменов
- [x] Удалить подозрительный домен
- [x] Добавить поддержку wildcard для development

**Новая конфигурация**:
```go
type CORSConfig struct {
    AllowedOrigins []string
    AllowedMethods []string
    AllowedHeaders []string
    AllowCredentials bool
    MaxAge int
}
```

## 🟠 Серьезные проблемы (Приоритет 2)

### 4. Отсутствие защиты от CSRF

**Проблема**: Нет защиты от Cross-Site Request Forgery

**Риск**: Выполнение несанкционированных действий от имени пользователя

**Решение**:
- [ ] Добавить CSRF middleware
- [ ] Использовать Double Submit Cookie pattern
- [ ] Настроить SameSite атрибуты для cookies
- [ ] Добавить проверку Referer header

### 5. Недостаточная валидация паролей

**Проблема**: Отсутствует проверка на распространенные пароли, нет требования специальных символов

**Файл**: `internal/validation/validator.go:54-88`

**Риск**: Слабые пароли пользователей

**Решение**:
- [x] Добавить требование специальных символов
- [x] Интегрировать проверку на распространенные пароли
- [x] Добавить проверку на последовательности (123456, qwerty)
- [x] Добавить проверку на повторяющиеся символы

**Статус**: ✅ **ВЫПОЛНЕНО** - Коммит: `security/improve-password-validation`, Ветка: `security/improve-password-validation`

**Улучшенная валидация**:
```go
func ValidatePassword(password string) error {
    // ... существующие проверки
    
    // Проверка специальных символов
    hasSpecial := false
    specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
    for _, char := range password {
        if strings.ContainsRune(specialChars, char) {
            hasSpecial = true
            break
        }
    }
    
    if !hasSpecial {
        return fmt.Errorf("password must contain at least one special character")
    }
    
    // Проверка на распространенные пароли
    if isCommonPassword(password) {
        return fmt.Errorf("password is too common, choose a stronger password")
    }
    
    return nil
}
```

### 6. Отсутствие HTTP Security Headers

**Проблема**: Нет заголовков безопасности

**Риск**: XSS, clickjacking, MITM атаки

**Решение**:
- [ ] Создать security headers middleware
- [ ] Добавить HSTS, CSP, X-Frame-Options
- [ ] Настроить X-Content-Type-Options
- [ ] Добавить Referrer-Policy

**Заголовки для добавления**:
```
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Referrer-Policy: strict-origin-when-cross-origin
```

### 7. Информативные сообщения об ошибках

**Проблема**: Одинаковые сообщения для разных ошибок

**Файл**: `internal/services/auth_service.go:85`

**Риск**: Information disclosure

**Решение**:
- [x] Унифицировать сообщения об ошибках
- [x] Не различать "пользователь не найден" и "неверный пароль"
- [x] Добавить задержку для неудачных попыток входа
- [x] Логировать детали только на сервере

**Статус**: ✅ **ВЫПОЛНЕНО** - Коммит: `security/unify-error-messages`, Ветка: `security/unify-error-messages`

## 🟡 Средние проблемы (Приоритет 3)

### 8. Отсутствие логирования безопасности

**Проблема**: SecurityLogger создан, но не используется в критических местах

**Риск**: Сложность обнаружения атак

**Решение**:
- [ ] Интегрировать SecurityLogger во все auth операции
- [ ] Добавить логирование подозрительной активности
- [ ] Настроить алерты на множественные неудачные попытки
- [ ] Добавить метрики безопасности

### 9. Недостаточная защита от timing атак

**Проблема**: Время ответа может различаться для существующих/несуществующих пользователей

**Риск**: Enumeration атаки

**Решение**:
- [ ] Добавить фиксированную задержку для неудачных попыток входа
- [ ] Использовать constant-time сравнение для паролей
- [ ] Добавить jitter для времени ответа

### 10. Отсутствие валидации размера запросов

**Проблема**: Нет ограничений на размер тела запроса

**Риск**: DoS атаки через большие запросы

**Решение**:
- [ ] Добавить лимиты на размер запросов
- [ ] Настроить разные лимиты для разных эндпоинтов
- [ ] Добавить middleware для проверки размера

## 🟢 Положительные аспекты (сохранить)

- ✅ Хорошая архитектура JWT с правильной валидацией
- ✅ Безопасное хеширование паролей с bcrypt
- ✅ Валидация входных данных
- ✅ Структурированное логирование
- ✅ Docker security (non-root пользователь)
- ✅ Graceful shutdown

## План реализации

### Фаза 1 (Критические проблемы) - 1-2 недели
1. Исправление JWT secret
2. Добавление rate limiting
3. Исправление CORS конфигурации
4. Добавление security headers

### Фаза 2 (Серьезные проблемы) - 2-3 недели
1. Реализация CSRF защиты
2. Усиление валидации паролей
3. Унификация сообщений об ошибках
4. Интеграция SecurityLogger

### Фаза 3 (Средние проблемы) - 1-2 недели
1. Защита от timing атак
2. Валидация размера запросов
3. Улучшение мониторинга безопасности

## Тестирование безопасности

### Автоматизированные тесты
- [ ] Добавить тесты на rate limiting
- [ ] Тесты на валидацию паролей
- [ ] Тесты на CORS политики
- [ ] Тесты на security headers

### Ручное тестирование
- [ ] Penetration testing
- [ ] Проверка на OWASP Top 10
- [ ] Тестирование на timing атаки
- [ ] Проверка на CSRF уязвимости

## Мониторинг и алерты

### Метрики для отслеживания
- [ ] Количество неудачных попыток входа
- [ ] Количество запросов превышающих rate limit
- [ ] Подозрительная активность
- [ ] Ошибки валидации

### Алерты
- [ ] Множественные неудачные попытки входа
- [ ] Превышение rate limit
- [ ] Подозрительные паттерны запросов
- [ ] Ошибки аутентификации

## Документация

### Обновить документацию
- [ ] README с требованиями безопасности
- [ ] Руководство по развертыванию
- [ ] Политика безопасности
- [ ] Процедуры реагирования на инциденты

## Заключение

Данный план обеспечивает комплексный подход к улучшению безопасности Strive API. Реализация должна проводиться поэтапно, начиная с критических проблем. После каждой фазы рекомендуется проводить тестирование безопасности и обновлять документацию.

**Общее время реализации**: 4-7 недель
**Критический путь**: Фаза 1 (JWT secret, rate limiting, CORS, security headers)
