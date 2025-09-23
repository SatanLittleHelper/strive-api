# План реализации калькулятора калорий для Strive API

## Обзор
План реализации API калькулятора калорий на основе технической спецификации. Реализация включает модели данных, сервисы, репозитории, HTTP handlers и интеграцию в основное приложение.

## Структура реализации

### 1. Модели данных (`internal/models/calorie.go`)

#### CalorieCalculationData
```go
type CalorieCalculationData struct {
    Gender        string  `json:"gender" validate:"required,oneof=male female"`
    Age           int     `json:"age" validate:"required,min=15,max=120"`
    Height        float64 `json:"height" validate:"required,min=100,max=250"`
    Weight        float64 `json:"weight" validate:"required,min=30,max=300"`
    ActivityLevel string  `json:"activityLevel" validate:"required,oneof=sedentary lightly_active moderately_active very_active extremely_active"`
    Goal          string  `json:"goal" validate:"required,oneof=lose_weight maintain_weight gain_weight"`
}
```

#### Macronutrients
```go
type Macronutrients struct {
    ProteinGrams      int     `json:"proteinGrams"`
    ProteinPercentage float64 `json:"proteinPercentage"`
    FatGrams          int     `json:"fatGrams"`
    FatPercentage     float64 `json:"fatPercentage"`
    CarbsGrams        int     `json:"carbsGrams"`
    CarbsPercentage   float64 `json:"carbsPercentage"`
}
```

#### CalorieResults
```go
type CalorieResults struct {
    BMR            int           `json:"bmr"`
    TDEE           int           `json:"tdee"`
    TargetCalories int           `json:"targetCalories"`
    Formula        string        `json:"formula"`
    Macros         Macronutrients `json:"macros"`
}
```

#### CalorieCalculation (для БД)
```go
type CalorieCalculation struct {
    ID              uuid.UUID      `json:"id" db:"id"`
    UserID          uuid.UUID      `json:"user_id" db:"user_id"`
    Gender          string         `json:"gender" db:"gender"`
    Age             int            `json:"age" db:"age"`
    Height          float64        `json:"height" db:"height"`
    Weight          float64        `json:"weight" db:"weight"`
    ActivityLevel   string         `json:"activityLevel" db:"activity_level"`
    Goal            string         `json:"goal" db:"goal"`
    BMR             int            `json:"bmr" db:"bmr"`
    TDEE            int            `json:"tdee" db:"tdee"`
    TargetCalories  int            `json:"targetCalories" db:"target_calories"`
    Formula         string         `json:"formula" db:"formula"`
    ProteinGrams    int            `json:"proteinGrams" db:"protein_grams"`
    ProteinPercentage float64     `json:"proteinPercentage" db:"protein_percentage"`
    FatGrams        int            `json:"fatGrams" db:"fat_grams"`
    FatPercentage   float64        `json:"fatPercentage" db:"fat_percentage"`
    CarbsGrams      int            `json:"carbsGrams" db:"carbs_grams"`
    CarbsPercentage float64        `json:"carbsPercentage" db:"carbs_percentage"`
    CreatedAt       time.Time      `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
}
```

#### CalorieCalculationResponse
```go
type CalorieCalculationResponse struct {
    Data      CalorieCalculationData `json:"data"`
    Results   CalorieResults         `json:"results"`
    Timestamp time.Time              `json:"timestamp"`
}
```

### 2. Сервис (`internal/services/calorie_service.go`)

#### CalorieService Interface
```go
type CalorieService interface {
    CalculateCalories(ctx context.Context, userID uuid.UUID, data *models.CalorieCalculationData) (*models.CalorieResults, error)
    GetLastCalculation(ctx context.Context, userID uuid.UUID) (*models.CalorieCalculationResponse, error)
}
```

#### Алгоритмы расчета

**BMR (Basal Metabolic Rate) - Формула Миффлина-Сан Жеора:**
```go
func (s *calorieService) calculateBMRMifflin(data models.CalorieCalculationData) float64 {
    const (
        weightMultiplier = 10.0
        heightMultiplier = 6.25
        ageMultiplier    = 5.0
        maleOffset       = 5.0
        femaleOffset     = -161.0
    )

    base := weightMultiplier*data.Weight + heightMultiplier*data.Height - ageMultiplier*float64(data.Age)

    if data.Gender == "male" {
        return base + maleOffset
    }
    return base + femaleOffset
}
```

**TDEE (Total Daily Energy Expenditure):**
```go
func (s *calorieService) calculateTDEE(bmr float64, activityLevel string) int {
    multipliers := map[string]float64{
        "sedentary":         1.2,
        "lightly_active":    1.375,
        "moderately_active": 1.55,
        "very_active":       1.725,
        "extremely_active":  1.9,
    }

    return int(math.Round(bmr * multipliers[activityLevel]))
}
```

**Target Calories:**
```go
func (s *calorieService) calculateTargetCalories(tdee int, goal string) int {
    modifiers := map[string]float64{
        "lose_weight":   -0.20,
        "maintain_weight": 0.0,
        "gain_weight":    0.15,
    }

    modifier := modifiers[goal]
    return int(math.Round(float64(tdee) * (1 + modifier)))
}
```

**Macronutrients Calculation:**
- Белок: базовое значение по активности + корректировка по цели
- Жиры: базовое значение по цели
- Углеводы: остаток калорий после белка и жиров
- Проверка на отрицательные углеводы с корректировкой

### 3. Репозиторий (`internal/repositories/calorie_repository.go`)

#### CalorieRepository Interface
```go
type CalorieRepository interface {
    SaveOrUpdate(ctx context.Context, calculation *models.CalorieCalculation) error
    GetByUserID(ctx context.Context, userID uuid.UUID) (*models.CalorieCalculation, error)
}
```

#### Особенности реализации:
- `SaveOrUpdate` использует `ON CONFLICT (user_id) DO UPDATE` для обновления существующего расчета
- `GetByUserID` возвращает последний расчет пользователя
- Подготовленные SQL запросы для безопасности
- Обработка ошибок БД

### 4. HTTP Handlers (`internal/http/calorie_handlers.go`)

#### CalorieHandlers
```go
type CalorieHandlers struct {
    calorieService services.CalorieService
    logger         *logger.Logger
    validator      *validation.Validator
}
```

#### Endpoints:
- `POST /api/v1/calories/calculate` - расчет калорий
- `GET /api/v1/calories/last` - получение последнего расчета

#### Особенности:
- Валидация входных данных
- Структурированное логирование
- Обработка ошибок с JSON ответами
- Аутентификация через middleware

### 5. Валидация (`internal/validation/validator.go`)

#### Новые функции валидации:
```go
func ValidateGender(gender string) error
func ValidateAge(age int) error
func ValidateHeight(height float64) error
func ValidateWeight(weight float64) error
func ValidateActivityLevel(activityLevel string) error
func ValidateGoal(goal string) error
```

#### Validator структура:
```go
type Validator struct{}

func (v *Validator) Validate(data interface{}) error
```

### 6. Миграции (`migrations/000004_calorie_calculations.up.sql`)

#### Таблица calorie_calculations:
```sql
CREATE TABLE calorie_calculations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    gender VARCHAR(10) NOT NULL CHECK (gender IN ('male', 'female')),
    age INTEGER NOT NULL CHECK (age >= 15 AND age <= 120),
    height DECIMAL(5,2) NOT NULL CHECK (height >= 100 AND height <= 250),
    weight DECIMAL(5,2) NOT NULL CHECK (weight >= 30 AND weight <= 300),
    activity_level VARCHAR(20) NOT NULL CHECK (activity_level IN ('sedentary', 'lightly_active', 'moderately_active', 'very_active', 'extremely_active')),
    goal VARCHAR(20) NOT NULL CHECK (goal IN ('lose_weight', 'maintain_weight', 'gain_weight')),
    bmr INTEGER NOT NULL,
    tdee INTEGER NOT NULL,
    target_calories INTEGER NOT NULL,
    formula VARCHAR(10) NOT NULL DEFAULT 'mifflin',
    protein_grams INTEGER NOT NULL,
    protein_percentage DECIMAL(5,2) NOT NULL,
    fat_grams INTEGER NOT NULL,
    fat_percentage DECIMAL(5,2) NOT NULL,
    carbs_grams INTEGER NOT NULL,
    carbs_percentage DECIMAL(5,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### Ограничения:
- Foreign key на users(id) с CASCADE
- Unique constraint на user_id (только один расчет на пользователя)
- Индексы для производительности

### 7. Интеграция в main.go

#### Обновление структур:
```go
type Services struct {
    Auth    services.AuthService
    User    services.UserService
    Calorie services.CalorieService  // НОВОЕ
}

type Handlers struct {
    Auth    *httphandler.AuthHandlers
    User    *httphandler.UserHandlers
    Health  *httphandler.DetailedHealthHandler
    Calorie *httphandler.CalorieHandlers  // НОВОЕ
}
```

#### Настройка сервисов:
```go
func setupServices(db *database.Database, cfg *config.Config) *Services {
    userRepo := repositories.NewUserRepository(db.Pool())
    refreshTokenRepo := repositories.NewRefreshTokenRepository(db.Pool())
    calorieRepo := repositories.NewCalorieRepository(db.Pool())  // НОВОЕ
    authService := services.NewAuthService(userRepo, refreshTokenRepo, &cfg.JWT)
    userService := services.NewUserService(userRepo)
    calorieService := services.NewCalorieService(calorieRepo)  // НОВОЕ
    return &Services{
        Auth:    authService,
        User:    userService,
        Calorie: calorieService,  // НОВОЕ
    }
}
```

#### Защищенные маршруты:
```go
func setupProtectedRoutes(mux *http.ServeMux, authService services.AuthService, logger *logger.Logger, handlers *Handlers) {
    // User protected routes
    userProtectedMux := http.NewServeMux()
    userProtectedMux.HandleFunc("/me", handlers.User.Me)
    userProtectedMux.HandleFunc("/theme", handlers.User.UpdateTheme)
    userProtectedHandler := httphandler.AuthMiddleware(authService, logger)(userProtectedMux)
    mux.Handle("/api/v1/user/", http.StripPrefix("/api/v1/user", userProtectedHandler))

    // Calorie protected routes  // НОВОЕ
    calorieProtectedMux := http.NewServeMux()
    calorieProtectedMux.HandleFunc("/calculate", handlers.Calorie.CalculateCalories)
    calorieProtectedMux.HandleFunc("/last", handlers.Calorie.GetLastCalculation)
    calorieProtectedHandler := httphandler.AuthMiddleware(authService, logger)(calorieProtectedMux)
    mux.Handle("/api/v1/calories/", http.StripPrefix("/api/v1/calories", calorieProtectedHandler))
}
```

## API Endpoints

### POST /api/v1/calories/calculate
**Назначение**: Расчет калорий и макронутриентов

**Request Body**:
```json
{
  "gender": "male",
  "age": 30,
  "height": 180,
  "weight": 80,
  "activityLevel": "moderately_active",
  "goal": "maintain_weight"
}
```

**Response (200 OK)**:
```json
{
  "bmr": 1800,
  "tdee": 2790,
  "targetCalories": 2790,
  "formula": "mifflin",
  "macros": {
    "proteinGrams": 152,
    "proteinPercentage": 21.8,
    "fatGrams": 80,
    "fatPercentage": 25.8,
    "carbsGrams": 279,
    "carbsPercentage": 40.0
  }
}
```

### GET /api/v1/calories/last
**Назначение**: Получение последнего расчета

**Response (200 OK)**:
```json
{
  "data": {
    "gender": "male",
    "age": 30,
    "height": 180,
    "weight": 80,
    "activityLevel": "moderately_active",
    "goal": "maintain_weight"
  },
  "results": {
    "bmr": 1800,
    "tdee": 2790,
    "targetCalories": 2790,
    "formula": "mifflin",
    "macros": {
      "proteinGrams": 152,
      "proteinPercentage": 21.8,
      "fatGrams": 80,
      "fatPercentage": 25.8,
      "carbsGrams": 279,
      "carbsPercentage": 40.0
    }
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Порядок реализации

1. **Модели данных** - создать `internal/models/calorie.go`
2. **Валидация** - добавить функции валидации в `internal/validation/validator.go`
3. **Репозиторий** - создать `internal/repositories/calorie_repository.go`
4. **Сервис** - создать `internal/services/calorie_service.go`
5. **HTTP Handlers** - создать `internal/http/calorie_handlers.go`
6. **Миграции** - создать `migrations/000004_calorie_calculations.up.sql`
7. **Интеграция** - обновить `cmd/server/main.go`
8. **Тестирование** - написать unit и integration тесты

## Тестирование

### Unit Tests
- Тестирование алгоритмов расчета BMR, TDEE, макронутриентов
- Граничные значения (минимальный/максимальный возраст, вес, рост)
- Различные комбинации пола, активности и целей
- Корректность округления и математических операций

### Integration Tests
- Тестирование полного flow от запроса до сохранения в БД
- Тестирование валидации входных данных
- Тестирование обработки ошибок
- Тестирование аутентификации

## Безопасность

- Валидация всех входных данных
- Аутентификация через JWT middleware
- Rate limiting для защиты от злоупотреблений
- Структурированное логирование
- Подготовленные SQL запросы

## Мониторинг

- Логирование расчетов с метриками
- Отслеживание времени выполнения
- Мониторинг ошибок валидации
- Популярные комбинации параметров

## Заключение

Данный план содержит все необходимые компоненты для реализации API калькулятора калорий. Алгоритмы расчета точно соответствуют спецификации, обеспечивая консистентность результатов между клиентом и сервером.
