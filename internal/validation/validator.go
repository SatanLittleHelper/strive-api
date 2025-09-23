package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/aleksandr/strive-api/internal/models"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors []ValidationError

type Validator struct{}

func (v *Validator) Validate(data interface{}) error {
	var errors ValidationErrors

	if d, ok := data.(models.CalorieCalculationData); ok {
		if err := ValidateGender(d.Gender); err != nil {
			errors = append(errors, ValidationError{Field: "gender", Message: err.Error()})
		}
		if err := ValidateAge(d.Age); err != nil {
			errors = append(errors, ValidationError{Field: "age", Message: err.Error()})
		}
		if err := ValidateHeight(d.Height); err != nil {
			errors = append(errors, ValidationError{Field: "height", Message: err.Error()})
		}
		if err := ValidateWeight(d.Weight); err != nil {
			errors = append(errors, ValidationError{Field: "weight", Message: err.Error()})
		}
		if err := ValidateActivityLevel(d.ActivityLevel); err != nil {
			errors = append(errors, ValidationError{Field: "activityLevel", Message: err.Error()})
		}
		if err := ValidateGoal(d.Goal); err != nil {
			errors = append(errors, ValidationError{Field: "goal", Message: err.Error()})
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

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

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case strings.ContainsRune(specialChars, char):
			hasSpecial = true
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
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
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

func ValidateGender(gender string) error {
	if gender == "" {
		return fmt.Errorf("gender is required")
	}
	for _, validGender := range models.Genders() {
		if gender == validGender {
			return nil
		}
	}
	return fmt.Errorf("gender must be 'male' or 'female'")
}

func ValidateAge(age int) error {
	if age < 15 {
		return fmt.Errorf("age must be at least 15 years")
	}
	if age > 120 {
		return fmt.Errorf("age must be at most 120 years")
	}
	return nil
}

func ValidateHeight(height float64) error {
	if height < 100 {
		return fmt.Errorf("height must be at least 100 cm")
	}
	if height > 250 {
		return fmt.Errorf("height must be at most 250 cm")
	}
	return nil
}

func ValidateWeight(weight float64) error {
	if weight < 30 {
		return fmt.Errorf("weight must be at least 30 kg")
	}
	if weight > 300 {
		return fmt.Errorf("weight must be at most 300 kg")
	}
	return nil
}

func ValidateActivityLevel(activityLevel string) error {
	for _, level := range models.ActivityLevels() {
		if activityLevel == level {
			return nil
		}
	}
	return fmt.Errorf("activity level must be one of: sedentary, lightly_active, moderately_active, very_active, extremely_active")
}

func ValidateGoal(goal string) error {
	for _, validGoal := range models.Goals() {
		if goal == validGoal {
			return nil
		}
	}
	return fmt.Errorf("goal must be one of: lose_weight, maintain_weight, gain_weight")
}
