package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aleksandr/strive-api/internal/config"
	"github.com/aleksandr/strive-api/internal/logger"
)

type ExerciseDBClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logger.Logger
	retryCount int
}

type ExerciseDBResponse struct {
	ID               int    `json:"id"`
	LicenseAuthor    string `json:"license_author"`
	Status           string `json:"status"`
	Description      string `json:"description"`
	Name             string `json:"name"`
	NameOriginal     string `json:"name_original"`
	CreationDate     string `json:"creation_date"`
	UUID             string `json:"uuid"`
	License          int    `json:"license"`
	Category         int    `json:"category"`
	Language         int    `json:"language"`
	Muscles          []int  `json:"muscles"`
	MusclesSecondary []int  `json:"muscles_secondary"`
	Equipment        []int  `json:"equipment"`
}

type MuscleGroupResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type EquipmentResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TargetResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func NewExerciseDBClient(cfg *config.ExerciseDBConfig, logger *logger.Logger) *ExerciseDBClient {
	return &ExerciseDBClient{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger:     logger,
		retryCount: cfg.RetryCount,
	}
}

func (c *ExerciseDBClient) GetAllExercises(ctx context.Context) ([]ExerciseDBResponse, error) {
	url := fmt.Sprintf("%s/api/v2/exercise/", c.baseURL)
	var result []ExerciseDBResponse
	err := c.makeRequest(ctx, url, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *ExerciseDBClient) GetExercisesByCategory(ctx context.Context, category int) ([]ExerciseDBResponse, error) {
	url := fmt.Sprintf("%s/api/v2/exercise/?category=%d", c.baseURL, category)
	var result []ExerciseDBResponse
	err := c.makeRequest(ctx, url, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *ExerciseDBClient) GetExercisesByMuscle(ctx context.Context, muscleID int) ([]ExerciseDBResponse, error) {
	url := fmt.Sprintf("%s/api/v2/exercise/?muscles=%d", c.baseURL, muscleID)
	var result []ExerciseDBResponse
	err := c.makeRequest(ctx, url, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *ExerciseDBClient) GetExercisesByEquipment(ctx context.Context, equipmentID int) ([]ExerciseDBResponse, error) {
	url := fmt.Sprintf("%s/api/v2/exercise/?equipment=%d", c.baseURL, equipmentID)
	var result []ExerciseDBResponse
	err := c.makeRequest(ctx, url, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *ExerciseDBClient) GetExerciseByID(ctx context.Context, id int) (*ExerciseDBResponse, error) {
	url := fmt.Sprintf("%s/api/v2/exercise/%d", c.baseURL, id)
	var result []ExerciseDBResponse
	err := c.makeRequest(ctx, url, &result)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("exercise with id %d not found", id)
	}
	return &result[0], nil
}

func (c *ExerciseDBClient) GetMuscleGroups(ctx context.Context) ([]MuscleGroupResponse, error) {
	url := fmt.Sprintf("%s/api/v2/bodypart/", c.baseURL)
	var result []MuscleGroupResponse
	err := c.makeRequest(ctx, url, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *ExerciseDBClient) GetEquipment(ctx context.Context) ([]EquipmentResponse, error) {
	url := fmt.Sprintf("%s/api/v2/equipment/", c.baseURL)
	var result []EquipmentResponse
	err := c.makeRequest(ctx, url, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *ExerciseDBClient) GetTargets(ctx context.Context) ([]TargetResponse, error) {
	url := fmt.Sprintf("%s/api/v2/target/", c.baseURL)
	var result []TargetResponse
	err := c.makeRequest(ctx, url, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *ExerciseDBClient) makeRequest(ctx context.Context, url string, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.retryCount; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * time.Second
			c.logger.Debug("Retrying request", "attempt", attempt, "backoff", backoff, "url", url)
			time.Sleep(backoff)
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "Strive-API/1.0")

		c.logger.Debug("Making request to ExerciseDB", "url", url, "attempt", attempt+1)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to make request: %w", err)
			c.logger.Warn("Request failed, will retry", "error", err, "url", url, "attempt", attempt+1)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			_ = resp.Body.Close()

			if err != nil {
				lastErr = fmt.Errorf("failed to read response body: %w", err)
				c.logger.Warn("Failed to read response body, will retry", "error", err, "url", url, "attempt", attempt+1)
				continue
			}

			if err := json.Unmarshal(body, &result); err != nil {
				lastErr = fmt.Errorf("failed to unmarshal response: %w", err)
				c.logger.Warn("Failed to unmarshal response, will retry", "error", err, "url", url, "attempt", attempt+1)
				continue
			}

			c.logger.Debug("Successfully received response from ExerciseDB", "url", url, "attempt", attempt+1)
			return nil
		}

		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("API returned server error %d: %s", resp.StatusCode, string(body))
			c.logger.Warn("Server error, will retry", "status", resp.StatusCode, "url", url, "attempt", attempt+1)
			continue
		}

		lastErr = fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
		c.logger.Error("Client error, will not retry", "status", resp.StatusCode, "body", string(body), "url", url)
		break
	}

	c.logger.Error("All retry attempts failed", "error", lastErr, "url", url, "retry_count", c.retryCount)
	return lastErr
}

func (c *ExerciseDBClient) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v2/exercise/?limit=1", c.baseURL)
	var result []ExerciseDBResponse
	return c.makeRequest(ctx, url, &result)
}
