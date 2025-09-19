package services

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aleksandr/strive-api/internal/config"
	"github.com/aleksandr/strive-api/internal/models"
	"github.com/google/uuid"
)

type mockUserRepository struct {
	users map[string]*models.User
}

func (m *mockUserRepository) Create(ctx context.Context, user *models.User) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(user.Email))
	m.users[normalizedEmail] = user
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
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if user, exists := m.users[normalizedEmail]; exists {
		return user, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (m *mockUserRepository) Update(ctx context.Context, user *models.User) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(user.Email))
	m.users[normalizedEmail] = user
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

type mockRefreshTokenRepository struct {
	tokens map[string]*models.RefreshToken
}

func (m *mockRefreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	m.tokens[token.Token] = token
	return nil
}

func (m *mockRefreshTokenRepository) GetByToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	refreshToken, exists := m.tokens[token]
	if !exists {
		return nil, fmt.Errorf("refresh token not found")
	}
	return refreshToken, nil
}

func (m *mockRefreshTokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.RefreshToken, error) {
	var tokens []*models.RefreshToken
	for _, token := range m.tokens {
		if token.UserID == userID {
			tokens = append(tokens, token)
		}
	}
	return tokens, nil
}

func (m *mockRefreshTokenRepository) Delete(ctx context.Context, token string) error {
	delete(m.tokens, token)
	return nil
}

func (m *mockRefreshTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	for token, refreshToken := range m.tokens {
		if refreshToken.UserID == userID {
			delete(m.tokens, token)
		}
	}
	return nil
}

func (m *mockRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	for token, refreshToken := range m.tokens {
		if refreshToken.ExpiresAt.Before(now) {
			delete(m.tokens, token)
		}
	}
	return nil
}

func TestAuthService_Register(t *testing.T) {
	mockRepo := &mockUserRepository{
		users: make(map[string]*models.User),
	}
	mockRefreshRepo := &mockRefreshTokenRepository{
		tokens: make(map[string]*models.RefreshToken),
	}
	jwtConfig := &config.JWTConfig{
		Secret:    "test-secret",
		Issuer:    "test-issuer",
		Audience:  "test-audience",
		ClockSkew: 1 * time.Minute,
	}
	authService := NewAuthService(mockRepo, mockRefreshRepo, jwtConfig)

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
	mockRepo := &mockUserRepository{
		users: make(map[string]*models.User),
	}
	mockRefreshRepo := &mockRefreshTokenRepository{
		tokens: make(map[string]*models.RefreshToken),
	}
	jwtConfig := &config.JWTConfig{
		Secret:    "test-secret",
		Issuer:    "test-issuer",
		Audience:  "test-audience",
		ClockSkew: 1 * time.Minute,
	}
	authService := NewAuthService(mockRepo, mockRefreshRepo, jwtConfig)

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

func TestAuthService_LoginCaseInsensitive(t *testing.T) {
	mockRepo := &mockUserRepository{
		users: make(map[string]*models.User),
	}
	mockRefreshRepo := &mockRefreshTokenRepository{
		tokens: make(map[string]*models.RefreshToken),
	}
	jwtConfig := &config.JWTConfig{
		Secret:    "test-secret",
		Issuer:    "test-issuer",
		Audience:  "test-audience",
		ClockSkew: 1 * time.Minute,
	}
	authService := NewAuthService(mockRepo, mockRefreshRepo, jwtConfig)

	// Register user with lowercase email
	req := &models.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err := authService.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Try to login with different case email
	accessToken, refreshToken, err := authService.Login(context.Background(), "Test@Example.com", req.Password)
	if err != nil {
		t.Fatalf("Expected no error with different case email, got %v", err)
	}

	if accessToken == "" {
		t.Error("Access token should not be empty")
	}

	if refreshToken == "" {
		t.Error("Refresh token should not be empty")
	}
}

func TestAuthService_HashPassword(t *testing.T) {
	mockRepo := &mockUserRepository{
		users: make(map[string]*models.User),
	}
	mockRefreshRepo := &mockRefreshTokenRepository{
		tokens: make(map[string]*models.RefreshToken),
	}
	jwtConfig := &config.JWTConfig{
		Secret:    "test-secret",
		Issuer:    "test-issuer",
		Audience:  "test-audience",
		ClockSkew: 1 * time.Minute,
	}
	authService := NewAuthService(mockRepo, mockRefreshRepo, jwtConfig)

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
