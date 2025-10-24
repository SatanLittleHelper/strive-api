package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aleksandr/strive-api/internal/config"
	"github.com/aleksandr/strive-api/internal/logger"
	"github.com/aleksandr/strive-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupWgerTestClient(t *testing.T, handler http.HandlerFunc) (*WgerClient, *httptest.Server) {
	t.Helper()

	server := httptest.NewServer(handler)

	cfg := &config.WgerConfig{
		BaseURL:    server.URL,
		APIKey:     "test-api-key",
		Timeout:    5 * time.Second,
		RetryCount: 2,
	}

	log := logger.New("debug", "json")

	client := NewWgerClient(cfg, log)

	return client, server
}

func TestNewWgerClient(t *testing.T) {
	cfg := &config.WgerConfig{
		BaseURL:    "https://wger.de/api/v2",
		APIKey:     "test-key",
		Timeout:    10 * time.Second,
		RetryCount: 3,
	}

	log := logger.New("info", "json")

	client := NewWgerClient(cfg, log)

	assert.NotNil(t, client)
	assert.Equal(t, cfg.BaseURL, client.baseURL)
	assert.Equal(t, cfg.APIKey, client.apiKey)
	assert.Equal(t, cfg.RetryCount, client.retryCount)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.logger)
}

func TestWgerClient_GetAllExercises(t *testing.T) {
	exercises := []models.WgerExercise{
		{
			ID:          1,
			UUID:        "test-uuid-1",
			Name:        "Bench Press",
			Description: "A chest exercise",
			Category:    models.WgerCategory{ID: 1, Name: "Chest"},
			Muscles:     []models.WgerMuscle{{ID: 1, Name: "Pectorals"}, {ID: 2, Name: "Triceps"}},
			Equipment:   []models.WgerEquipment{{ID: 1, Name: "Barbell"}},
		},
		{
			ID:          2,
			UUID:        "test-uuid-2",
			Name:        "Squat",
			Description: "A leg exercise",
			Category:    models.WgerCategory{ID: 2, Name: "Legs"},
			Muscles:     []models.WgerMuscle{{ID: 3, Name: "Quadriceps"}, {ID: 4, Name: "Glutes"}},
			Equipment:   []models.WgerEquipment{{ID: 2, Name: "Barbell"}},
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/exerciseinfo/", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "limit=20")

		response := models.WgerExerciseListResponse{
			Count:    2,
			Next:     nil,
			Previous: nil,
			Results:  exercises,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})

	client, server := setupWgerTestClient(t, handler)
	defer server.Close()

	ctx := context.Background()
	result, err := client.GetAllExercises(ctx)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Bench Press", result[0].Name)
	assert.Equal(t, "Squat", result[1].Name)
}

func TestWgerClient_GetExerciseByID(t *testing.T) {
	exercise := models.WgerExercise{
		ID:          1,
		UUID:        "test-uuid",
		Name:        "Bench Press",
		Description: "A chest exercise",
		Category:    models.WgerCategory{ID: 1, Name: "Chest"},
		Muscles:     []models.WgerMuscle{{ID: 1, Name: "Pectorals"}, {ID: 2, Name: "Triceps"}},
		Equipment:   []models.WgerEquipment{{ID: 1, Name: "Barbell"}},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/exerciseinfo/1/", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(exercise)
	})

	client, server := setupWgerTestClient(t, handler)
	defer server.Close()

	ctx := context.Background()
	result, err := client.GetExerciseByID(ctx, 1)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.ID)
	assert.Equal(t, "Bench Press", result.Name)
}

func TestWgerClient_GetMuscles(t *testing.T) {
	muscles := []models.WgerMuscle{
		{
			ID:      1,
			Name:    "Pectoralis major",
			NameEn:  "Pectoralis major",
			IsFront: true,
		},
		{
			ID:      2,
			Name:    "Biceps brachii",
			NameEn:  "Biceps brachii",
			IsFront: true,
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/muscle/", r.URL.Path)

		response := models.WgerMuscleListResponse{
			Count:    2,
			Next:     nil,
			Previous: nil,
			Results:  muscles,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})

	client, server := setupWgerTestClient(t, handler)
	defer server.Close()

	ctx := context.Background()
	result, err := client.GetMuscles(ctx)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Pectoralis major", result[0].Name)
	assert.Equal(t, "Biceps brachii", result[1].Name)
}

func TestWgerClient_GetEquipment(t *testing.T) {
	equipment := []models.WgerEquipment{
		{ID: 1, Name: "Barbell"},
		{ID: 2, Name: "Dumbbell"},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/equipment/", r.URL.Path)

		response := models.WgerEquipmentListResponse{
			Count:    2,
			Next:     nil,
			Previous: nil,
			Results:  equipment,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})

	client, server := setupWgerTestClient(t, handler)
	defer server.Close()

	ctx := context.Background()
	result, err := client.GetEquipment(ctx)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Barbell", result[0].Name)
	assert.Equal(t, "Dumbbell", result[1].Name)
}

func TestWgerClient_GetCategories(t *testing.T) {
	categories := []models.WgerCategory{
		{ID: 1, Name: "Arms"},
		{ID: 2, Name: "Legs"},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/exercisecategory/", r.URL.Path)

		response := models.WgerCategoryListResponse{
			Count:    2,
			Next:     nil,
			Previous: nil,
			Results:  categories,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})

	client, server := setupWgerTestClient(t, handler)
	defer server.Close()

	ctx := context.Background()
	result, err := client.GetCategories(ctx)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Arms", result[0].Name)
	assert.Equal(t, "Legs", result[1].Name)
}

func TestWgerClient_HealthCheck(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/exerciseinfo/", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "limit=1")

		response := models.WgerExerciseListResponse{
			Count:    1,
			Next:     nil,
			Previous: nil,
			Results:  []models.WgerExercise{},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})

	client, server := setupWgerTestClient(t, handler)
	defer server.Close()

	ctx := context.Background()
	err := client.HealthCheck(ctx)

	require.NoError(t, err)
}

func TestWgerClient_RetryLogic(t *testing.T) {
	attemptCount := 0

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response := models.WgerExerciseListResponse{
			Count:   0,
			Results: []models.WgerExercise{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})

	client, server := setupWgerTestClient(t, handler)
	defer server.Close()

	ctx := context.Background()
	result, err := client.GetAllExercises(ctx)

	require.NoError(t, err)
	assert.Empty(t, result)
	assert.Equal(t, 2, attemptCount)
}

func TestWgerClient_Pagination(t *testing.T) {
	serverURL := ""

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("offset")

		var response models.WgerExerciseListResponse

		if page == "" || page == "0" {
			nextURL := serverURL + "/exerciseinfo/?limit=100&offset=100"
			response = models.WgerExerciseListResponse{
				Count: 200,
				Next:  &nextURL,
				Results: []models.WgerExercise{
					{ID: 1, Name: "Exercise 1"},
					{ID: 2, Name: "Exercise 2"},
				},
			}
		} else {
			response = models.WgerExerciseListResponse{
				Count:    200,
				Next:     nil,
				Previous: &serverURL,
				Results: []models.WgerExercise{
					{ID: 3, Name: "Exercise 3"},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})

	client, server := setupWgerTestClient(t, handler)
	serverURL = server.URL
	defer server.Close()

	ctx := context.Background()
	result, err := client.GetAllExercises(ctx)

	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "Exercise 1", result[0].Name)
	assert.Equal(t, "Exercise 3", result[2].Name)
}
