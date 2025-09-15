## Calorie Service — Caching Feature Plan (Draft; requires deeper design)

This plan outlines an initial approach to introduce caching for calorie calculation results. It intentionally avoids deep analysis and MUST be refined before implementation in production.

### 1. Scope and Goals
- Reduce latency and CPU for repeated identical calculation requests.
- Keep API stateless and avoid persistent storage for calculations.
- Remain compatible with current architecture: Go stdlib `net/http`, dependency injection, no globals.

### 2. Non‑Goals
- No cross‑service distributed cache for v1 (consider later).
- No user history beyond the latest result per user; no cache warmup or precomputation.

### 3. Cached Endpoints
- POST `/api/calories/calculate`: cacheable; same input should return identical output.
- GET `/api/calories/last`: served from DB; optional short‑circuit from cache if key is available.

### 4. Correctness and Consistency
- Functional correctness has priority over hit ratio.
- Cache is best‑effort; stale within TTL is acceptable.

### 5. Cache Key Strategy
- Base key on normalized request payload (`CalorieCalculationData`).
- Include user identity when authenticated (UUID) to tie cache to user context.
- Optional (configurable) inclusion of client fingerprint for anonymous requests (IP + User‑Agent) to reduce cross‑user sharing if needed.
- Key format: `hash( canonical_json(payload) + "|uid:" + userID + "|fp:" + fingerprint )` using SHA‑256; canonicalization must produce stable field ordering.

### 6. Eviction and TTL
- TTL: default 10 minutes (configurable).
- Capacity: LRU with max items to cap memory (configurable).
- Eviction policy: LRU on insert when over capacity or on explicit TTL expiry.

### 7. Concurrency and Safety
- Single‑process in‑memory cache.
- Sharded map (e.g., 16–64 shards) with RWMutex per shard to reduce contention.
- Values immutable after insert to avoid races.

### 8. Observability
- Metrics: hits, misses, evictions, set errors; latency buckets for lookup/set.
- Logs: structured JSON for cache events at DEBUG level; no sensitive data in logs.

### 9. Configuration (env)
- `CACHE_ENABLED` (bool, default: false)
- `CACHE_TTL` (duration, default: `10m`)
- `CACHE_MAX_ITEMS` (int, default: `10000`)
- `CACHE_INCLUDE_USER_IN_KEY` (bool, default: true)
- `CACHE_INCLUDE_FINGERPRINT_IN_KEY` (bool, default: false)

### 10. Interfaces and Package Structure
- Package: `internal/cache`
- Interface:
  - `type Cache interface { Get(ctx context.Context, key string) (any, bool); Set(ctx context.Context, key string, value any, ttl time.Duration) bool; Delete(ctx context.Context, key string); Stats() Stats }`
  - `type Stats struct { Hits uint64; Misses uint64; Evictions uint64 }`
- Implementation: `lru_ttl.go` with sharded LRU.

### 11. Integration Strategy
- Create `CalorieCalculator` service (pure function style) if not yet present.
- Add caching decorator `CachedCalorieCalculator` implementing the same interface and injected into handlers when `CACHE_ENABLED=true`.
- Persist latest result per authenticated user in DB via repository.
- Handler logic:
  1) Build key from normalized request and context (user id, optional fingerprint).
  2) Attempt cache Get → on hit return value.
  3) On miss compute via underlying service.
  4) If user authenticated → upsert latest result into DB (single row per user).
  5) Set cache with TTL and return.

### 11.1 DB Interplay
- After successful calculation, if the request is authenticated, persist the latest result for that user.
- Data model: single row per user, overwritten on each calculation.
- GET `/api/calories/last` reads from DB; may consult cache first if a fresh entry exists.

### 12. Testing
- Unit tests: key canonicalization, TTL expiry, LRU eviction, concurrency (with `-race`).
- Integration tests: handler returns cached result on repeated requests; ensure headers/content consistent.

### 13. Risks and Open Questions (to be designed in detail)
- Canonical JSON: choose stable encoder or implement a canonicalization to avoid key drift.
- Memory footprint: estimate object sizes and set sane defaults; expose metrics to monitor.
- Shard count and lock contention under load.
- Anonymous fingerprinting: trade‑off between privacy and hit ratio; may keep disabled by default.
- Horizontal scaling: for multi‑instance deployments, consider Redis or in‑memory per‑pod acceptable duplication.
- DB schema and migrations alignment with service plan; upsert strategy and conflict handling.

### 14. Future Enhancements
- Pluggable backends: Redis implementation behind the same `Cache` interface.
- Negative caching for invalid requests (short TTL) to reduce abuse.
- Fine‑grained cache invalidation hooks if calculation rules change (versioned key prefix).

Note: This draft requires deeper design decisions before implementation, especially around key canonicalization, memory limits, and concurrency under production traffic.


