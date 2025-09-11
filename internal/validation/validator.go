package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

func (ve ValidationErrors) ToJSON() map[string]interface{} {
	errors := make(map[string]interface{})
	for _, err := range ve {
		errors[err.Field] = err.Message
	}
	return map[string]interface{}{
		"error": map[string]interface{}{
			"code":    "VALIDATION_ERROR",
			"message": "Validation failed",
			"details": errors,
		},
	}
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	if len(email) > 255 {
		return fmt.Errorf("email too long (max 255 characters)")
	}
	return nil
}

func ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	if len(password) > 128 {
		return fmt.Errorf("password too long (max 128 characters)")
	}

	var hasUpper, hasLower, hasDigit bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	return nil
}

func ValidateString(value, fieldName string, minLen, maxLen int) error {
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	if len(value) < minLen {
		return fmt.Errorf("%s must be at least %d characters long", fieldName, minLen)
	}
	if len(value) > maxLen {
		return fmt.Errorf("%s too long (max %d characters)", fieldName, maxLen)
	}
	return nil
}
