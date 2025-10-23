Создай commit message для текущих изменений в Go проекте:

**Формат:**

```
[branch-name]: <short-description>
```

**Требования:**

- В начале всегда указывать название текущей ветки
- Краткое описание на английском (до 72 символов)
- Понятное описание WHAT, а не HOW
- **ВАЖНО:** Коммит должен содержать ТОЛЬКО одну строку с кратким описанием
- **НЕ ДОБАВЛЯТЬ** детальное описание изменений, списки файлов или bullet points
- Commit message должен быть максимально лаконичным

**Обязательные проверки перед коммитом:**

1. **Удаление комментариев** - удалить все комментарии из только что добавленного кода
2. **Линтер** - `golangci-lint run` (должен завершиться с кодом 0)
3. **Тесты** - `go test ./... -count=1 -race -timeout=60s` (все тесты должны проходить)

**Шаги:**

1. **ОБЯЗАТЕЛЬНО:** Изучи текущую архитектуру проекта:
   - Прочитай файл `.cursor/rules/project-architecture.mdc`
   - Пойми какие компоненты затронуты изменениями
   - Проверь соответствие изменений принципам Clean Architecture

2. **ОБЯЗАТЕЛЬНО:** Запусти команду "Review" для проверки качества кода
3. **ОБЯЗАТЕЛЬНО:** Проверь текущую ветку:
   ```bash
   git branch --show-current
   ```
4. Если находимся в ветке `main` - создай новую ветку:
   ```bash
   git checkout -b <feature-type>/<feature-name>
   ```
5. Посмотри git status и git diff
6. Удали все комментарии из добавленного кода
7. Запусти `golangci-lint run`
8. Запусти `go test ./... -count=1 -race -timeout=60s`
9. Если проверки прошли успешно - сформируй commit message
10. **ОБЯЗАТЕЛЬНО:** Обнови файл архитектуры если изменения затрагивают:
    - Новые компоненты или слои
    - Изменения в API endpoints
    - Новые middleware или handlers
    - Изменения в структуре проекта
    - Новые конфигурационные параметры
11. Предложи команду для коммита

**Названия веток:**

- **security/** - для изменений безопасности
- **auth/** - для аутентификации и авторизации
- **api/** - для новых API endpoints
- **config/** - для конфигурации
- **test/** - для тестов
- **docs/** - для документации
- **refactor/** - для рефакторинга

**Примеры правильных коммитов:**

```
security/add-rate-limiting: fix linter errors in tests
auth/jwt-validation: add secret strength validation
config/cors: move domains to env variables
api/user-endpoints: add profile update handler
```

**Примеры НЕПРАВИЛЬНЫХ коммитов (слишком детальные):**

```
security/add-rate-limiting: fix linter errors in tests

- Fix unused imports in auth_middleware_test.go
- Remove commented code in rate_limit_middleware.go
- Update test assertions for better coverage
```

**Исключения:**

**НЕТ ИСКЛЮЧЕНИЙ** - правила применяются ко всем коммитам без исключений.