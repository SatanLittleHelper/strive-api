# План MVP дневника калорий для Strive API

## Обзор

Данный документ содержит план реализации MVP версии дневника калорий с интеграцией Open Food Facts API для сканирования штрихкодов продуктов.

**Связанные планы:**
- [План разработки API дневника тренировок](./plan.md) - основная архитектура
- [Будущие улучшения](./stages/09-future.md) - расширенная функциональность
- [План улучшения безопасности](./security-improvement-plan.md) - меры безопасности

## 🎯 Цели MVP

- Отслеживание потребления продуктов через сканирование штрихкодов
- Интеграция с Open Food Facts API для получения данных о продуктах
- Ведение дневника питания с расчетом калорий и БЖУ
- Базовая аналитика потребления

## 🏗️ Архитектурные решения

**Расширение существующей архитектуры:**
- Использование текущей JWT аутентификации
- Расширение схемы БД новыми таблицами
- Добавление новых сервисов в существующую структуру
- Сохранение принципов Clean Architecture

## 📊 Схема базы данных

```sql
-- Таблица продуктов
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    barcode VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    brand VARCHAR(255),
    category VARCHAR(100),
    image_url TEXT,
    nutrition_per_100g JSONB NOT NULL, -- калории, белки, жиры, углеводы
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Таблица дневника питания
CREATE TABLE food_diary_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity_grams DECIMAL(10,2) NOT NULL,
    consumed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы для производительности
CREATE INDEX idx_food_diary_user_date ON food_diary_entries(user_id, consumed_at);
CREATE INDEX idx_products_barcode ON products(barcode);
```

## 🔌 API Endpoints

**Новые эндпоинты для дневника калорий:**

```
# Поиск продуктов
GET  /api/v1/products/search?query=название
GET  /api/v1/products/barcode/{barcode}

# Дневник питания
GET  /api/v1/diary/entries?date=2024-01-01
POST /api/v1/diary/entries
PUT  /api/v1/diary/entries/{id}
DELETE /api/v1/diary/entries/{id}

# Аналитика
GET  /api/v1/diary/summary?date=2024-01-01
GET  /api/v1/diary/summary?from=2024-01-01&to=2024-01-07
```

## 🏛️ Структура кода

**Новые компоненты в существующей архитектуре:**

```
internal/
├── models/
│   ├── product.go          # Модели продуктов
│   ├── food_diary.go       # Модели дневника питания
│   └── nutrition.go        # Модели питательных веществ
├── services/
│   ├── product_service.go  # Бизнес-логика продуктов
│   ├── food_diary_service.go # Бизнес-логика дневника
│   └── nutrition_service.go  # Расчеты БЖУ
├── repositories/
│   ├── product_repository.go
│   └── food_diary_repository.go
├── http/
│   ├── product_handlers.go
│   └── food_diary_handlers.go
└── external/
    └── openfoodfacts_client.go # Клиент для Open Food Facts API
```

## 🔌 Интеграция с Open Food Facts API

**Основные функции:**
- Поиск продукта по штрихкоду
- Получение данных о питательности
- Кэширование результатов для оптимизации
- Обработка ошибок API

```go
type OpenFoodFactsClient interface {
    GetProductByBarcode(ctx context.Context, barcode string) (*Product, error)
    SearchProducts(ctx context.Context, query string) ([]*Product, error)
}
```

## 📋 Модели данных

```go
// Продукт
type Product struct {
    ID           uuid.UUID     `json:"id" db:"id"`
    Barcode      string        `json:"barcode" db:"barcode"`
    Name         string        `json:"name" db:"name"`
    Brand        string        `json:"brand" db:"brand"`
    Category     string        `json:"category" db:"category"`
    ImageURL     string        `json:"image_url" db:"image_url"`
    Nutrition    *Nutrition    `json:"nutrition" db:"nutrition_per_100g"`
    CreatedAt    time.Time     `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time     `json:"updated_at" db:"updated_at"`
}

// Питательные вещества
type Nutrition struct {
    Calories    float64 `json:"calories"`
    Protein     float64 `json:"protein"`
    Fat         float64 `json:"fat"`
    Carbs       float64 `json:"carbs"`
    Fiber       float64 `json:"fiber"`
    Sugar       float64 `json:"sugar"`
}

// Запись в дневнике
type FoodDiaryEntry struct {
    ID           uuid.UUID `json:"id" db:"id"`
    UserID       uuid.UUID `json:"user_id" db:"user_id"`
    ProductID    uuid.UUID `json:"product_id" db:"product_id"`
    Product      *Product  `json:"product,omitempty"`
    QuantityGrams float64 `json:"quantity_grams" db:"quantity_grams"`
    ConsumedAt   time.Time `json:"consumed_at" db:"consumed_at"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
```

## 🚀 Этапы реализации

### **Этап 1: Подготовка инфраструктуры (1-2 дня)**
1. **Миграции БД**
   - Создать таблицы `products` и `food_diary_entries`
   - Добавить индексы для производительности
   - Создать миграции up/down

2. **Модели данных**
   - `Product`, `FoodDiaryEntry`, `Nutrition`
   - Request/Response модели
   - Валидация данных

### **Этап 2: Интеграция с Open Food Facts (2-3 дня)**
1. **HTTP клиент**
   - Реализация `OpenFoodFactsClient`
   - Обработка ошибок API
   - Rate limiting и retry логика

2. **Сервис продуктов**
   - `ProductService` с методами поиска
   - Кэширование результатов
   - Обработка отсутствующих продуктов

### **Этап 3: Репозитории и сервисы (2-3 дня)**
1. **Репозитории**
   - `ProductRepository` - CRUD операции
   - `FoodDiaryRepository` - управление дневником

2. **Бизнес-логика**
   - `FoodDiaryService` - расчеты калорий
   - `NutritionService` - расчеты БЖУ
   - Валидация данных

### **Этап 4: HTTP слой (2-3 дня)**
1. **Handlers**
   - `ProductHandlers` - поиск продуктов
   - `FoodDiaryHandlers` - управление дневником

2. **Middleware**
   - Валидация входных данных
   - Обработка ошибок

### **Этап 5: Тестирование (2-3 дня)**
1. **Unit тесты**
   - Тесты сервисов и репозиториев
   - Моки для внешних API
   - Покрытие >70%

2. **Integration тесты**
   - Тесты API endpoints
   - Тесты с реальной БД

### **Этап 6: Документация и развертывание (1-2 дня)**
1. **Swagger документация**
   - Описание новых endpoints
   - Примеры запросов/ответов

2. **Конфигурация**
   - ENV переменные для Open Food Facts
   - Настройки кэширования

## ⚙️ Технические детали

### **Конфигурация**
```bash
# Open Food Facts API
OPENFOODFACTS_API_URL=https://world.openfoodfacts.org/api/v0
OPENFOODFACTS_USER_AGENT=StriveAPI/1.0
OPENFOODFACTS_TIMEOUT=10s
OPENFOODFACTS_RETRY_ATTEMPTS=3

# Кэширование продуктов
PRODUCT_CACHE_TTL=24h
PRODUCT_CACHE_SIZE=1000
```

### **Обработка ошибок**
- Продукт не найден в Open Food Facts
- Неверный формат штрихкода
- Ошибки сети при обращении к API
- Валидация данных пользователя

### **Производительность**
- Кэширование продуктов в памяти
- Индексы БД для быстрого поиска
- Пагинация для больших списков
- Оптимизация SQL запросов

## 📊 Аналитика и отчеты

### **Сводка по дням**
```json
{
  "date": "2024-01-01",
  "total_calories": 2500,
  "total_protein": 120.5,
  "total_fat": 85.2,
  "total_carbs": 300.8,
  "entries_count": 5,
  "entries": [...]
}
```

### **Статистика по периодам**
- Дневная сводка
- Недельная статистика
- Месячные отчеты
- Тренды потребления

## ✅ Критерии готовности MVP

- ✅ Пользователь может найти продукт по штрихкоду
- ✅ Пользователь может добавить продукт в дневник
- ✅ Система рассчитывает калории и БЖУ
- ✅ Пользователь видит сводку за день
- ✅ Все операции покрыты тестами
- ✅ API документирован в Swagger

## 🚀 Следующие шаги

1. **Создать ветку для фичи**: `git checkout -b feature/calorie-diary`
2. **Начать с миграций БД** - создать схему таблиц
3. **Реализовать модели данных** - Product, FoodDiaryEntry
4. **Интегрировать Open Food Facts API** - HTTP клиент
5. **Добавить бизнес-логику** - сервисы и репозитории
6. **Создать API endpoints** - handlers и middleware
7. **Написать тесты** - unit и integration
8. **Обновить документацию** - Swagger и README

## 📈 Время выполнения

**Общее время**: 10-15 дней
- Этап 1: 1-2 дня
- Этап 2: 2-3 дня  
- Этап 3: 2-3 дня
- Этап 4: 2-3 дня
- Этап 5: 2-3 дня
- Этап 6: 1-2 дня

## 🎯 Приоритеты

1. **Высокий**: Базовая функциональность дневника, интеграция с Open Food Facts
2. **Средний**: Аналитика и отчеты, кэширование
3. **Низкий**: Расширенная аналитика, оптимизация производительности

## 🔗 Связанные задачи

### Дополнительные возможности (будущие этапы):
- **Расширенная аналитика** - тренды, рекомендации по питанию
- **Экспорт данных** - выгрузка дневника в CSV/JSON
- **Уведомления** - напоминания о приеме пищи
- **Социальные функции** - обмен рецептами, достижения

*Эти задачи планируются в рамках будущих этапов развития*
