# Exercise API Documentation

## Overview

The Exercise API provides access to a comprehensive database of exercises with filtering, search, and caching capabilities. The API integrates with ExerciseDB to provide up-to-date exercise information.

## Base URL

```
https://strive-api-zjtl.onrender.com/api/v1/exercises
```

## Authentication

**Authentication required** - Exercise API endpoints require valid JWT token.

Include the JWT token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

## Endpoints

### 1. Get Exercises List

**GET** `/api/v1/exercises`

Returns a paginated list of exercises with optional filtering.

#### Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `muscle_group_id` | UUID | No | Filter by muscle group ID |
| `equipment_id` | UUID | No | Filter by equipment ID |
| `category` | Integer | No | Filter by category (9=strength, 10=bodyweight, 11=weighted) |
| `search` | String | No | Search by exercise name |
| `page` | Integer | No | Page number (default: 1) |
| `limit` | Integer | No | Items per page (default: 20, max: 100) |

#### Example Request

```bash
GET /api/v1/exercises?muscle_group_id=123e4567-e89b-12d3-a456-426614174000&category=9&page=1&limit=20
Authorization: Bearer <your-jwt-token>
```

#### Response

```json
{
  "exercises": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "Push-ups",
      "description": "A basic bodyweight exercise for chest and triceps",
      "instructions": "Start in plank position...",
      "tips": "Keep your core tight...",
      "category": 10,
      "muscle_groups": [
        {
          "id": "456e7890-e89b-12d3-a456-426614174001",
          "name": "Chest"
        }
      ],
      "equipment": [],
      "alternatives": [
        {
          "id": "789e0123-e89b-12d3-a456-426614174002",
          "name": "Incline Push-ups"
        }
      ],
      "created_at": "2024-01-15T10:30:00Z",
      "cached_at": "2024-01-15T10:30:00Z",
      "expires_at": "2024-01-16T10:30:00Z"
    }
  ],
  "total": 150,
  "page": 1,
  "limit": 20
}
```

### 2. Get Exercise by ID

**GET** `/api/v1/exercises/{id}`

Returns detailed information about a specific exercise.

#### Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | UUID | Yes | Exercise ID |

#### Example Request

```bash
GET /api/v1/exercises/123e4567-e89b-12d3-a456-426614174000
Authorization: Bearer <your-jwt-token>
```

#### Response

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "Push-ups",
  "description": "A basic bodyweight exercise for chest and triceps",
  "instructions": "Start in plank position with hands slightly wider than shoulders...",
  "tips": "Keep your core tight and maintain a straight line from head to heels",
  "category": 10,
  "language": 7,
  "license": 2,
  "license_author": "ExerciseDB",
  "status": "2",
  "name_original": "",
  "creation_date": "2015-10-22",
  "uuid": "583281c7-2362-48e7-95d5-8fd6c455e0fb",
  "muscle_groups": [
    {
      "id": "456e7890-e89b-12d3-a456-426614174001",
      "name": "Chest",
      "created_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": "789e0123-e89b-12d3-a456-426614174002",
      "name": "Triceps",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "equipment": [],
  "alternatives": [
    {
      "id": "012e3456-e89b-12d3-a456-426614174003",
      "name": "Incline Push-ups",
      "description": "Easier variation of push-ups",
      "category": 10
    }
  ],
  "created_at": "2024-01-15T10:30:00Z",
  "cached_at": "2024-01-15T10:30:00Z",
  "expires_at": "2024-01-16T10:30:00Z"
}
```

### 3. Get Muscle Groups

**GET** `/api/v1/exercises/muscle-groups`

Returns a list of all available muscle groups.

#### Response

```json
[
  {
    "id": "456e7890-e89b-12d3-a456-426614174001",
    "name": "Chest",
    "created_at": "2024-01-15T10:30:00Z"
  },
  {
    "id": "789e0123-e89b-12d3-a456-426614174002",
    "name": "Triceps",
    "created_at": "2024-01-15T10:30:00Z"
  },
  {
    "id": "012e3456-e89b-12d3-a456-426614174003",
    "name": "Biceps",
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

### 4. Get Equipment

**GET** `/api/v1/exercises/equipment`

Returns a list of all available equipment.

#### Response

```json
[
  {
    "id": "111e2222-e89b-12d3-a456-426614174001",
    "name": "Dumbbells",
    "created_at": "2024-01-15T10:30:00Z"
  },
  {
    "id": "333e4444-e89b-12d3-a456-426614174002",
    "name": "Barbell",
    "created_at": "2024-01-15T10:30:00Z"
  },
  {
    "id": "555e6666-e89b-12d3-a456-426614174003",
    "name": "Bodyweight",
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

### 5. Get Cache Status

**GET** `/api/v1/exercises/cache/status`

Returns information about the exercise cache status.

#### Response

```json
{
  "last_updated": "2024-01-15T10:30:00Z",
  "total_exercises": 1250,
  "total_muscles": 14,
  "total_equipment": 8,
  "is_valid": true,
  "expires_at": "2024-01-16T10:30:00Z"
}
```

### 6. Refresh Cache

**POST** `/api/v1/exercises/cache/refresh`

Manually refreshes the exercise cache from ExerciseDB.

#### Response

```json
{
  "message": "Cache refreshed successfully"
}
```

## Data Models

### Exercise

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique exercise identifier |
| `name` | String | Exercise name |
| `description` | String? | Exercise description |
| `instructions` | String? | How to perform the exercise |
| `tips` | String? | Tips for proper form |
| `category` | Integer? | Exercise category (9=strength, 10=bodyweight, 11=weighted) |
| `muscle_groups` | Array | Primary and secondary muscle groups |
| `equipment` | Array | Required equipment |
| `alternatives` | Array | Alternative exercises |
| `created_at` | DateTime | When the exercise was cached |
| `cached_at` | DateTime | Cache timestamp |
| `expires_at` | DateTime | Cache expiration time |

### MuscleGroup

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique muscle group identifier |
| `name` | String | Muscle group name |
| `created_at` | DateTime | Creation timestamp |

### Equipment

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique equipment identifier |
| `name` | String | Equipment name |
| `created_at` | DateTime | Creation timestamp |

## Error Responses

All endpoints return consistent error responses:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message"
  }
}
```

### Common Error Codes

| Code | Description |
|------|-------------|
| `UNAUTHORIZED` | Missing or invalid JWT token |
| `INVALID_PARAMETER` | Invalid request parameters |
| `EXERCISE_NOT_FOUND` | Exercise not found |
| `INTERNAL_ERROR` | Internal server error |


## Caching

The API uses intelligent caching with the following characteristics:

- **TTL**: 24 hours for exercises
- **Auto-refresh**: Cache is automatically refreshed when expired
- **Fallback**: Works with stale data if ExerciseDB is unavailable
- **Manual refresh**: Available via `/cache/refresh` endpoint

## Rate Limiting

The API respects the same rate limiting as other endpoints:

- **General**: 60 requests per minute
- **Burst**: 10 requests per burst

## Performance Tips

1. **Use pagination**: Always specify `limit` parameter for large datasets
2. **Filter early**: Use `muscle_group_id` and `equipment_id` filters to reduce results
3. **Cache responses**: Consider caching API responses on the frontend
4. **Use search**: Implement search functionality for better UX

## Support

For questions or issues with the Exercise API, please refer to the main API documentation or contact the development team.
