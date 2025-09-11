# Этап 5: API v1

## Цель этапа
Реализовать основную функциональность API для работы с упражнениями, тренировками и сетами.

## 5.1 Пользовательские endpoints
- [ ] GET `/api/v1/user/profile` - получение профиля
- [ ] PUT `/api/v1/user/profile` - обновление профиля
- [ ] PUT `/api/v1/user/password` - смена пароля

### Детали реализации
- Все endpoints требуют аутентификации
- Профиль содержит email и дату регистрации
- При смене пароля требуется текущий пароль

## 5.2 Упражнения (Exercises)
- [ ] POST `/api/v1/exercises` - создание упражнения
- [ ] GET `/api/v1/exercises` - список упражнений пользователя
- [ ] GET `/api/v1/exercises/{id}` - получение упражнения
- [ ] PUT `/api/v1/exercises/{id}` - обновление упражнения
- [ ] DELETE `/api/v1/exercises/{id}` - удаление упражнения

### Примеры запросов
```json
// POST /api/v1/exercises
{
  "name": "Bench Press",
  "description": "Chest exercise"
}

// GET /api/v1/exercises - ответ
{
  "exercises": [
    {
      "id": "uuid",
      "name": "Bench Press",
      "description": "Chest exercise",
      "created_at": "2024-01-01T10:00:00Z",
      "updated_at": "2024-01-01T10:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20
}
```

## 5.3 Тренировки (Workouts)
- [ ] POST `/api/v1/workouts` - создание тренировки
- [ ] GET `/api/v1/workouts` - список тренировок с пагинацией
- [ ] GET `/api/v1/workouts/{id}` - получение тренировки с сетами
- [ ] PUT `/api/v1/workouts/{id}` - обновление тренировки
- [ ] DELETE `/api/v1/workouts/{id}` - удаление тренировки

### Примеры запросов
```json
// POST /api/v1/workouts
{
  "name": "Upper Body Workout",
  "date": "2024-01-01",
  "notes": "Great workout today"
}

// GET /api/v1/workouts/{id} - ответ
{
  "id": "uuid",
  "name": "Upper Body Workout",
  "date": "2024-01-01",
  "notes": "Great workout today",
  "sets": [
    {
      "id": "uuid",
      "exercise_id": "uuid",
      "exercise_name": "Bench Press",
      "reps": 10,
      "weight": 80.5,
      "rest_time": 120
    }
  ],
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

## 5.4 Сеты (Sets)
- [ ] POST `/api/v1/workouts/{workout_id}/sets` - добавление сета
- [ ] PUT `/api/v1/sets/{id}` - обновление сета
- [ ] DELETE `/api/v1/sets/{id}` - удаление сета

### Примеры запросов
```json
// POST /api/v1/workouts/{workout_id}/sets
{
  "exercise_id": "uuid",
  "reps": 10,
  "weight": 80.5,
  "rest_time": 120
}
```

## 5.5 Валидация и обработка ошибок
- [ ] Создать единый формат ошибок API
- [ ] Добавить валидацию входных данных для всех endpoints
- [ ] Реализовать обработку ошибок базы данных

### Формат ошибок
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": {
      "field": "email",
      "reason": "Invalid email format"
    }
  }
}
```

## Критерии готовности
- [ ] Все CRUD операции работают корректно
- [ ] Валидация входных данных работает
- [ ] Пагинация реализована для списков
- [ ] Обработка ошибок единообразна
- [ ] Все endpoints защищены аутентификацией

## Время выполнения
Ориентировочно: 3-4 дня

## Предыдущий этап
[Этап 4: Аутентификация и авторизация](./04-auth.md)

## Следующий этап
[Этап 6: Тестирование](./06-testing.md)
