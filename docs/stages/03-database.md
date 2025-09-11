# Этап 3: База данных

## Цель этапа
Настроить подключение к PostgreSQL, создать модели данных и репозитории для работы с базой.

## 3.1 Подключение к PostgreSQL
- [ ] Добавить зависимость `github.com/lib/pq` или `github.com/jackc/pgx`
- [ ] Создать connection pool для базы данных
- [ ] Добавить health check для базы данных

### Детали реализации
- Настроить connection pool с оптимальными параметрами
- Добавить retry логику для подключения
- Health check должен проверять доступность БД

## 3.2 Миграции
- [ ] Установить `golang-migrate/migrate`
- [ ] Создать директорию `migrations/`
- [ ] Настроить автоматическое применение миграций при старте

### Детали реализации
- Создать миграции для всех таблиц
- Добавить команду для отката миграций
- Автоматически применять новые миграции при старте

## 3.3 Модели данных
- [ ] Создать модель User (id, email, password_hash, created_at, updated_at)
- [ ] Создать модель Exercise (id, user_id, name, description, created_at, updated_at)
- [ ] Создать модель Workout (id, user_id, name, date, notes, created_at, updated_at)
- [ ] Создать модель Set (id, workout_id, exercise_id, reps, weight, rest_time, created_at)

### Детали моделей
```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Exercises table
CREATE TABLE exercises (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Workouts table
CREATE TABLE workouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    date DATE NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Sets table
CREATE TABLE sets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_id UUID NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    reps INTEGER NOT NULL,
    weight DECIMAL(5,2),
    rest_time INTEGER, -- в секундах
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## 3.4 Репозитории
- [ ] Создать интерфейс UserRepository
- [ ] Создать интерфейс ExerciseRepository
- [ ] Создать интерфейс WorkoutRepository
- [ ] Создать интерфейс SetRepository
- [ ] Реализовать PostgreSQL версии всех репозиториев

### Методы репозиториев
- UserRepository: Create, GetByID, GetByEmail, Update, Delete
- ExerciseRepository: Create, GetByID, GetByUserID, Update, Delete
- WorkoutRepository: Create, GetByID, GetByUserID, Update, Delete, GetWithSets
- SetRepository: Create, GetByID, GetByWorkoutID, Update, Delete

## Критерии готовности
- [ ] Все таблицы созданы в БД
- [ ] Репозитории реализованы и протестированы
- [ ] Connection pool работает стабильно
- [ ] Health check возвращает статус БД

## Время выполнения
Ориентировочно: 2-3 дня

## Предыдущий этап
[Этап 2: Конфигурация и наблюдаемость](./02-config-observability.md)

## Следующий этап
[Этап 4: Аутентификация и авторизация](./04-auth.md)
