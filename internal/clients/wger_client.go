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
	"github.com/aleksandr/strive-api/internal/models"
)

type WgerClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *logger.Logger
	retryCount int
}

func NewWgerClient(cfg *config.WgerConfig, logger *logger.Logger) *WgerClient {
	return &WgerClient{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger:     logger,
		retryCount: cfg.RetryCount,
	}
}

func (c *WgerClient) GetAllExercises(ctx context.Context) ([]models.WgerExercise, error) {
	var allExercises []models.WgerExercise
	url := fmt.Sprintf("%s/exerciseinfo/?limit=100", c.baseURL)

	for url != "" {
		var response models.WgerExerciseListResponse
		err := c.makeRequest(ctx, url, &response)
		if err != nil {
			return nil, err
		}

		allExercises = append(allExercises, response.Results...)

		if response.Next != nil {
			url = *response.Next
		} else {
			url = ""
		}
	}

	return allExercises, nil
}

func (c *WgerClient) GetExerciseByID(ctx context.Context, id int) (*models.WgerExercise, error) {
	url := fmt.Sprintf("%s/exerciseinfo/%d/", c.baseURL, id)
	var result models.WgerExercise
	err := c.makeRequest(ctx, url, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *WgerClient) GetExercisesByCategory(ctx context.Context, category int) ([]models.WgerExercise, error) {
	var allExercises []models.WgerExercise
	url := fmt.Sprintf("%s/exerciseinfo/?category=%d&limit=100", c.baseURL, category)

	for url != "" {
		var response models.WgerExerciseListResponse
		err := c.makeRequest(ctx, url, &response)
		if err != nil {
			return nil, err
		}

		allExercises = append(allExercises, response.Results...)

		if response.Next != nil {
			url = *response.Next
		} else {
			url = ""
		}
	}

	return allExercises, nil
}

func (c *WgerClient) GetExercisesByMuscle(ctx context.Context, muscleID int) ([]models.WgerExercise, error) {
	var allExercises []models.WgerExercise
	url := fmt.Sprintf("%s/exerciseinfo/?muscles=%d&limit=100", c.baseURL, muscleID)

	for url != "" {
		var response models.WgerExerciseListResponse
		err := c.makeRequest(ctx, url, &response)
		if err != nil {
			return nil, err
		}

		allExercises = append(allExercises, response.Results...)

		if response.Next != nil {
			url = *response.Next
		} else {
			url = ""
		}
	}

	return allExercises, nil
}

func (c *WgerClient) GetExercisesByEquipment(ctx context.Context, equipmentID int) ([]models.WgerExercise, error) {
	var allExercises []models.WgerExercise
	url := fmt.Sprintf("%s/exerciseinfo/?equipment=%d&limit=100", c.baseURL, equipmentID)

	for url != "" {
		var response models.WgerExerciseListResponse
		err := c.makeRequest(ctx, url, &response)
		if err != nil {
			return nil, err
		}

		allExercises = append(allExercises, response.Results...)

		if response.Next != nil {
			url = *response.Next
		} else {
			url = ""
		}
	}

	return allExercises, nil
}

func (c *WgerClient) GetMuscles(ctx context.Context) ([]models.WgerMuscle, error) {
	var allMuscles []models.WgerMuscle
	url := fmt.Sprintf("%s/muscle/?limit=100", c.baseURL)

	for url != "" {
		var response models.WgerMuscleListResponse
		err := c.makeRequest(ctx, url, &response)
		if err != nil {
			return nil, err
		}

		allMuscles = append(allMuscles, response.Results...)

		if response.Next != nil {
			url = *response.Next
		} else {
			url = ""
		}
	}

	return allMuscles, nil
}

func (c *WgerClient) GetEquipment(ctx context.Context) ([]models.WgerEquipment, error) {
	var allEquipment []models.WgerEquipment
	url := fmt.Sprintf("%s/equipment/?limit=100", c.baseURL)

	for url != "" {
		var response models.WgerEquipmentListResponse
		err := c.makeRequest(ctx, url, &response)
		if err != nil {
			return nil, err
		}

		allEquipment = append(allEquipment, response.Results...)

		if response.Next != nil {
			url = *response.Next
		} else {
			url = ""
		}
	}

	return allEquipment, nil
}

func (c *WgerClient) GetCategories(ctx context.Context) ([]models.WgerCategory, error) {
	var allCategories []models.WgerCategory
	url := fmt.Sprintf("%s/exercisecategory/?limit=100", c.baseURL)

	for url != "" {
		var response models.WgerCategoryListResponse
		err := c.makeRequest(ctx, url, &response)
		if err != nil {
			return nil, err
		}

		allCategories = append(allCategories, response.Results...)

		if response.Next != nil {
			url = *response.Next
		} else {
			url = ""
		}
	}

	return allCategories, nil
}

func (c *WgerClient) makeRequest(ctx context.Context, url string, result interface{}) error {
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
		if c.apiKey != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.apiKey))
		}

		c.logger.Debug("Making request to wger API", "url", url, "attempt", attempt+1)

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

			c.logger.Debug("Successfully received response from wger API", "url", url, "attempt", attempt+1)
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

func (c *WgerClient) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/exerciseinfo/?limit=1", c.baseURL)
	var result models.WgerExerciseListResponse
	return c.makeRequest(ctx, url, &result)
}
