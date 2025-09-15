## Calorie Service Backend Plan (Go)

### Scope
- Move calorie calculation logic from frontend to backend.
- Provide stable HTTP API for calculation and retrieval.
- No localStorage. Persist server-side (DB). Support PWA-friendly flows.

### Endpoints
- POST `/api/calories/calculate`
  - Request: `CalorieCalculationData` { gender, age, height, weight, activityLevel, goal }
  - Response: `CalorieResults` payload { bmr, tdee, targetCalories, formula: "mifflin", macros }
  - Side effect: persist latest calculation for authenticated user (single record per user)
- GET `/api/calories/last`
  - Response: { data: CalorieCalculationData, results: CalorieResults } for the authenticated user or 404 if none

### Models (Go)
- Gender: `"male" | "female"`
- ActivityLevel: keys matching frontend (`sedentary`, `lightly_active`, `moderately_active`, `very_active`, `extremely_active`)
- Goal: (`lose_weight`, `maintain_weight`, `gain_weight`)
- CalorieCalculationData: fields above
- CalorieResults: { bmr: int, tdee: int, targetCalories: int, formula: "mifflin", macros }
- Macronutrients: { proteinGrams, fatGrams, carbsGrams, proteinPercentage, fatPercentage, carbsPercentage }

### Business Rules (parity with frontend)
- BMR (Mifflin-St Jeor):
  - male: 10*weight + 6.25*height - 5*age + 5
  - female: 10*weight + 6.25*height - 5*age - 161
- TDEE: round(bmr * activity.multiplier) from ActivityLevelOptions
- Target calories: round(tdee + tdee * goal.percentageModifier/100)
- Macros:
  - protein: base by activity + goal adjustment, min 1.4 g/kg
  - fat: base by goal (g/kg), min 20% of calories
  - carbs: remaining calories, g = kcal/4, not negative
  - Round grams to integers at the end; percentages with one decimal

### Validation
- age: 14–100, height: 120–250 cm, weight: 30–300 kg
- enums: strict; unknown values → 400
- typed errors: `{ code, message }`

### Architecture (Go)
- Go 1.22, framework: chi or gin
- Layers: http (handlers) → domain/service (pure logic) → repository (DB)
- Package layout:
  - `/cmd/server/main.go`
  - `/internal/http` (handlers, DTOs)
  - `/internal/domain` (calculation service, models, constants)
  - `/internal/repo` (sqlite/postgres)
  - `/internal/auth` (Telegram initData/JWT, optional)

### Persistence
- DB: Postgres
- Table `user_calorie_calculation` (single latest record per user):
  - user_id (uuid) PRIMARY KEY, payload_json (jsonb), result_json (jsonb), updated_at (timestamptz)
  - Upsert on `user_id` to always keep the latest calculation
  - Index: PRIMARY KEY (user_id)

### AuthN/AuthZ (optional)
- Telegram WebApp initData validation; extract userId
- Alt: JWT from gateway

### Observability & Ops
- Logging with request ids, input validation failures, calc timings
- Metrics: request count, latency, errors; health endpoint `/healthz`
- Tracing (OTel) optional

### API Contract
- OpenAPI 3.0 specification kept in repo `/api/openapi.yaml`
- CI check to validate handlers vs spec (oapi-codegen or kin-openapi)

### Testing
- Unit tests for domain calculation with golden vectors (match TS results)
- Handler tests with httptest (200/400/500)
- Migration tests (if using DB)

### Security
- Rate limit per IP/user
- CORS: allow frontend origin, `GET, POST`
- Input sanitization and strict JSON decoding

### Deployment
- Dockerfile (distroless/base), healthchecks
- Envs: `PORT`, `DATABASE_URL`, `ALLOW_ORIGIN`, `LOG_LEVEL`
- K8s or simple VM service; add migration command

### Timeline
- Day 1–2: Domain logic port + golden tests
- Day 3: HTTP handlers + OpenAPI
- Day 4: Persistence + last endpoint
- Day 5: CI/CD, observability, hardening


