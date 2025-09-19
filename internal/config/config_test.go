package config

import (
	"os"
	"strings"
	"testing"
)

func TestJWTSecretValidation(t *testing.T) {
	tests := []struct {
		name        string
		jwtSecret   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty secret should fail",
			jwtSecret:   "",
			expectError: true,
			errorMsg:    "JWT_SECRET is required",
		},
		{
			name:        "short secret should fail",
			jwtSecret:   "short",
			expectError: true,
			errorMsg:    "JWT_SECRET must be at least 32 characters long",
		},
		{
			name:        "valid secret should pass",
			jwtSecret:   "this-is-a-very-long-and-secure-secret-key-for-jwt-tokens",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.Setenv("JWT_SECRET", tt.jwtSecret); err != nil {
				t.Fatalf("Failed to set JWT_SECRET: %v", err)
			}
			defer func() {
				if err := os.Unsetenv("JWT_SECRET"); err != nil {
					t.Errorf("Failed to unset JWT_SECRET: %v", err)
				}
			}()

			config, err := Load()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if config != nil && config.JWT.Secret != tt.jwtSecret {
					t.Errorf("Expected JWT secret '%s', got '%s'", tt.jwtSecret, config.JWT.Secret)
				}
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	if err := os.Setenv("JWT_SECRET", "this-is-a-very-long-and-secure-secret-key-for-jwt-tokens"); err != nil {
		t.Fatalf("Failed to set JWT_SECRET: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("JWT_SECRET"); err != nil {
			t.Errorf("Failed to unset JWT_SECRET: %v", err)
		}
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if err := config.Validate(); err != nil {
		t.Errorf("Config validation failed: %v", err)
	}
}
