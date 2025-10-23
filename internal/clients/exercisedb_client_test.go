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
)

const exerciseAPIPath = "/api/v2/exercise/"

func TestExerciseDBClient_GetAllExercises(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != exerciseAPIPath {
			t.Errorf("Expected path %s, got %s", exerciseAPIPath, r.URL.Path)
		}

		exercises := []ExerciseDBResponse{
			{
				ID:               1,
				Name:             "Push-up",
				Description:      "Basic push-up exercise",
				Category:         9,
				Muscles:          []int{1, 2},
				MusclesSecondary: []int{3},
				Equipment:        []int{},
			},
			{
				ID:               2,
				Name:             "Squat",
				Description:      "Basic squat exercise",
				Category:         9,
				Muscles:          []int{10},
				MusclesSecondary: []int{8},
				Equipment:        []int{},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(exercises)
	}))
	defer mockServer.Close()

	cfg := &config.ExerciseDBConfig{
		BaseURL:    mockServer.URL,
		Timeout:    5 * time.Second,
		RetryCount: 2,
		Enabled:    true,
	}

	logger := logger.New("DEBUG", "text")
	client := NewExerciseDBClient(cfg, logger)

	ctx := context.Background()
	exercises, err := client.GetAllExercises(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(exercises) != 2 {
		t.Errorf("Expected 2 exercises, got %d", len(exercises))
	}

	if exercises[0].Name != "Push-up" {
		t.Errorf("Expected first exercise to be 'Push-up', got '%s'", exercises[0].Name)
	}
}

func TestExerciseDBClient_GetExercisesByCategory(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v2/exercise/"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		expectedCategory := "9"
		if r.URL.Query().Get("category") != expectedCategory {
			t.Errorf("Expected category %s, got %s", expectedCategory, r.URL.Query().Get("category"))
		}

		exercises := []ExerciseDBResponse{
			{
				ID:       1,
				Name:     "Bench Press",
				Category: 9,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(exercises)
	}))
	defer mockServer.Close()

	cfg := &config.ExerciseDBConfig{
		BaseURL:    mockServer.URL,
		Timeout:    5 * time.Second,
		RetryCount: 2,
		Enabled:    true,
	}

	logger := logger.New("DEBUG", "text")
	client := NewExerciseDBClient(cfg, logger)

	ctx := context.Background()
	exercises, err := client.GetExercisesByCategory(ctx, 9)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(exercises) != 1 {
		t.Errorf("Expected 1 exercise, got %d", len(exercises))
	}

	if exercises[0].Name != "Bench Press" {
		t.Errorf("Expected exercise to be 'Bench Press', got '%s'", exercises[0].Name)
	}
}

func TestExerciseDBClient_RetryLogic(t *testing.T) {
	attemptCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		exercises := []ExerciseDBResponse{
			{
				ID:   1,
				Name: "Success Exercise",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(exercises)
	}))
	defer mockServer.Close()

	cfg := &config.ExerciseDBConfig{
		BaseURL:    mockServer.URL,
		Timeout:    5 * time.Second,
		RetryCount: 3,
		Enabled:    true,
	}

	logger := logger.New("DEBUG", "text")
	client := NewExerciseDBClient(cfg, logger)

	ctx := context.Background()
	exercises, err := client.GetAllExercises(ctx)
	if err != nil {
		t.Fatalf("Expected no error after retries, got %v", err)
	}

	if len(exercises) != 1 {
		t.Errorf("Expected 1 exercise, got %d", len(exercises))
	}

	if exercises[0].Name != "Success Exercise" {
		t.Errorf("Expected exercise to be 'Success Exercise', got '%s'", exercises[0].Name)
	}

	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}
}

func TestExerciseDBClient_HealthCheck(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v2/exercise/"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		expectedLimit := "1"
		if r.URL.Query().Get("limit") != expectedLimit {
			t.Errorf("Expected limit %s, got %s", expectedLimit, r.URL.Query().Get("limit"))
		}

		exercises := []ExerciseDBResponse{
			{
				ID:   1,
				Name: "Health Check Exercise",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(exercises)
	}))
	defer mockServer.Close()

	cfg := &config.ExerciseDBConfig{
		BaseURL:    mockServer.URL,
		Timeout:    5 * time.Second,
		RetryCount: 2,
		Enabled:    true,
	}

	logger := logger.New("DEBUG", "text")
	client := NewExerciseDBClient(cfg, logger)

	ctx := context.Background()
	err := client.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("Expected no error from health check, got %v", err)
	}
}
