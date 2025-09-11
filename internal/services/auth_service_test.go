package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/aleksandr/strive-api/internal/models"
	"github.com/google/uuid"
)

type mockUserRepository struct {
	users map[string]*models.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*models.User),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *models.User) error {
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if user, exists := m.users[email]; exists {
		return user, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (m *mockUserRepository) Update(ctx context.Context, user *models.User) error {
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	for email, user := range m.users {
		if user.ID == id {
			delete(m.users, email)
			return nil
		}
	}
	return fmt.Errorf("user not found")
}

func TestAuthService_Register(t *testing.T) {
	mockRepo := newMockUserRepository()
	authService := NewAuthService(mockRepo, "test-secret")

	req := &models.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	user, err := authService.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if user.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, user.Email)
	}

	if user.PasswordHash == req.Password {
		t.Error("Password should be hashed")
	}

	if user.ID == uuid.Nil {
		t.Error("User ID should be generated")
	}
}

func TestAuthService_Login(t *testing.T) {
	mockRepo := newMockUserRepository()
	authService := NewAuthService(mockRepo, "test-secret")

	// First register a user
	req := &models.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err := authService.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Now try to login
	accessToken, refreshToken, err := authService.Login(context.Background(), req.Email, req.Password)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if accessToken == "" {
		t.Error("Access token should not be empty")
	}

	if refreshToken == "" {
		t.Error("Refresh token should not be empty")
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	mockRepo := newMockUserRepository()
	authService := NewAuthService(mockRepo, "test-secret")

	// First register a user
	req := &models.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	user, err := authService.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Generate a token by logging in
	accessToken, _, err := authService.Login(context.Background(), user.Email, req.Password)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate the token
	claims, err := authService.ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if claims.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, claims.Email)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, claims.UserID)
	}
}

func TestAuthService_HashPassword(t *testing.T) {
	mockRepo := newMockUserRepository()
	authService := NewAuthService(mockRepo, "test-secret")

	password := "testpassword123"
	hashed, err := authService.HashPassword(password)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if hashed == password {
		t.Error("Hashed password should be different from original")
	}

	// Test password verification
	err = authService.VerifyPassword(hashed, password)
	if err != nil {
		t.Errorf("Password verification should succeed, got %v", err)
	}

	// Test wrong password
	err = authService.VerifyPassword(hashed, "wrongpassword")
	if err == nil {
		t.Error("Password verification should fail for wrong password")
	}
}
