Проведи code review выделенного кода или указанного файла в Go проекте:

**Проверки:**

- ✅ Все функции имеют явные типы возвращаемых значений
- ✅ Используется современный Go синтаксис (Go 1.22+)
- ✅ В коде отсутствуют комментарии
- ✅ Naming следует Go conventions (camelCase для переменных, PascalCase для экспортируемых)
- ✅ Соблюдены принципы Clean Architecture
- ✅ Правильная типизация Go
- ✅ golangci-lint проходит без ошибок
- ✅ go vet проходит без ошибок
- ✅ Тесты проходят без ошибок

**Перед началом review:**

1. **ОБЯЗАТЕЛЬНО:** Изучи текущую архитектуру проекта:
   - Прочитай файл `.cursor/rules/project-architecture.mdc`
   - Пойми структуру проекта и принципы Clean Architecture
   - Определи какие компоненты затронуты изменениями

2. Запусти команду "FixLinter" для автоматического исправления ошибок
3. Убедись, что все проверки проходят:
   ```bash
   golangci-lint run
   go vet ./...
   go test ./... -count=1 -race -timeout=60s
   ```

**Проверки архитектуры:**

- ✅ **Clean Architecture** - правильное разделение слоев:
  - `internal/models/` - доменные модели
  - `internal/services/` - бизнес-логика
  - `internal/repositories/` - слой доступа к данным
  - `internal/http/` - HTTP слой
- ✅ **Dependency Injection** - зависимости инъектируются через конструкторы
- ✅ **Context.Context** - используется для request-scoped данных и отмены операций
- ✅ **Error Handling** - правильная обработка ошибок
- ✅ **No Global State** - избегание глобального состояния

**Проверки безопасности:**

- ✅ **JWT токены** - правильная валидация и обработка
- ✅ **Пароли** - использование bcrypt для хеширования
- ✅ **CORS** - правильная настройка CORS middleware
- ✅ **Rate Limiting** - защита от брутфорс атак
- ✅ **Security Headers** - HTTP заголовки безопасности
- ✅ **Input Validation** - валидация всех входных данных

**Проверки кода:**

- ✅ **Форматирование** - gofumpt форматирование
- ✅ **Импорты** - goimports организация импортов
- ✅ **Naming** - Go naming conventions
- ✅ **Типизация** - правильное использование типов Go
- ✅ **Error Handling** - обработка ошибок без panic
- ✅ **Context Usage** - правильное использование context.Context
- ✅ **Memory Management** - избегание утечек памяти
- ✅ **Concurrency** - правильное использование горутин и каналов

**Предоставь:**

1. **Список найденных проблем** с указанием файла и строки
2. **Конкретные рекомендации по исправлению** с примерами кода
3. **Результаты проверки линтера и тестов**
4. **Оценку соответствия Clean Architecture**
5. **Проверку безопасности кода**

**Примеры правильного кода:**

```go
// ✅ Правильно: явный тип возвращаемого значения
func (s *AuthService) Register(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
    // implementation
}

// ✅ Правильно: использование context.Context
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
    // implementation
}

// ✅ Правильно: dependency injection
func NewAuthService(repo UserRepository, jwtSecret string) *AuthService {
    return &AuthService{
        repo:      repo,
        jwtSecret: jwtSecret,
    }
}
```

**Примеры неправильного кода:**

```go
// ❌ Неправильно: отсутствует тип возвращаемого значения
func (s *AuthService) Register(ctx context.Context, req *models.CreateUserRequest) {
    // implementation
}

// ❌ Неправильно: глобальное состояние
var globalDB *sql.DB

// ❌ Неправильно: игнорирование ошибок
user, _ := repo.GetByID(ctx, id)
```