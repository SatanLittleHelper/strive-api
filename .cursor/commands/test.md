Создай тесты для указанного компонента или функции в Go проекте:

**Требования:**

- Современный синтаксис Go тестирования
- Запуск тестов только в headless режиме
- Все функции с явными типами возвращаемых значений
- Без комментариев в коде
- Все строки на английском языке
- Покрыть основные сценарии использования
- Тесты на edge cases
- Моки и стабы где необходимо

**Структура тестов:**

1. **ОБЯЗАТЕЛЬНО:** Изучи текущую архитектуру проекта:
   - Прочитай файл `.cursor/rules/project-architecture.mdc`
   - Пойми структуру проекта и принципы Clean Architecture
   - Определи какие компоненты нужно тестировать

2. **Unit тесты** - для каждого слоя архитектуры:
   - `internal/services/` - бизнес-логика
   - `internal/repositories/` - слой доступа к данным
   - `internal/http/` - HTTP handlers
   - `internal/validation/` - валидация

3. **Integration тесты** - для API endpoints:
   - HTTP handlers с реальными зависимостями
   - Тестирование middleware
   - Тестирование полного flow

**Покрытие:**

1. **Happy path** - успешные сценарии
2. **Error handling** - обработка ошибок
3. **Edge cases** - граничные случаи
4. **Validation** - валидация входных данных
5. **Security** - проверка безопасности

**Команды для запуска тестов:**

```bash
# Все тесты с race detection
go test ./... -count=1 -race -timeout=60s

# Тесты с покрытием
go test ./... -count=1 -race -timeout=60s -cover

# Тесты конкретного пакета
go test ./internal/services -count=1 -race -timeout=60s

# Тесты с verbose выводом
go test ./... -count=1 -race -timeout=60s -v
```

**Примеры тестов:**

**Unit тест для сервиса:**
```go
func TestAuthService_Register(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{}
    service := NewAuthService(mockRepo, "test-secret")
    
    // Act
    user, err := service.Register(context.Background(), &models.CreateUserRequest{
        Email:    "test@example.com",
        Password: "password123",
    })
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "test@example.com", user.Email)
}
```

**Integration тест для HTTP handler:**
```go
func TestAuthHandlers_Register(t *testing.T) {
    // Arrange
    server := setupTestServer(t)
    defer server.Close()
    
    payload := `{"email":"test@example.com","password":"password123"}`
    
    // Act
    resp, err := http.Post(server.URL+"/api/v1/auth/register", 
        "application/json", strings.NewReader(payload))
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
}
```

**Правила написания тестов:**

- Используй `testify/assert` и `testify/require`
- Создавай моки для внешних зависимостей
- Тестируй каждый слой архитектуры отдельно
- Используй table-driven tests для множественных случаев
- Проверяй как успешные, так и ошибочные сценарии
- Тестируй валидацию входных данных
- Проверяй безопасность (JWT, пароли, CORS)

**Создай тестовый файл и напиши команду для запуска в headless режиме.**