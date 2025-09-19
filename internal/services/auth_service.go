package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aleksandr/strive-api/internal/config"
	"github.com/aleksandr/strive-api/internal/models"
	"github.com/aleksandr/strive-api/internal/repositories"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidSignature = errors.New("invalid token signature")
	ErrTokenExpired     = errors.New("token has expired")
	ErrTokenNotBefore   = errors.New("token used before valid")
	ErrInvalidAudience  = errors.New("invalid token audience")
	ErrInvalidIssuer    = errors.New("invalid token issuer")
)

type AuthService interface {
	Register(ctx context.Context, req *models.CreateUserRequest) (*models.User, error)
	Login(ctx context.Context, email, password string) (string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	ValidateToken(tokenString string) (*Claims, error)
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) error
	Logout(ctx context.Context, refreshToken string) error
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

type authService struct {
	userRepo         repositories.UserRepository
	refreshTokenRepo repositories.RefreshTokenRepository
	config           *config.JWTConfig
	accessTTL        time.Duration
	refreshTTL       time.Duration
}

func NewAuthService(
	userRepo repositories.UserRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
	jwtConfig *config.JWTConfig,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		config:           jwtConfig,
		accessTTL:        15 * time.Minute,
		refreshTTL:       7 * 24 * time.Hour,
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *authService) Register(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	normalizedEmail := normalizeEmail(req.Email)
	existingUser, err := s.userRepo.GetByEmail(ctx, normalizedEmail)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	hashedPassword, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        normalizedEmail,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, string, error) {
	normalizedEmail := normalizeEmail(email)
	user, err := s.userRepo.GetByEmail(ctx, normalizedEmail)
	if err != nil {
		s.addLoginDelay()
		return "", "", fmt.Errorf("invalid credentials")
	}

	if err := s.VerifyPassword(user.PasswordHash, password); err != nil {
		s.addLoginDelay()
		return "", "", fmt.Errorf("invalid credentials")
	}

	accessToken, err := s.generateToken(user, s.accessTTL)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	refreshTokenModel := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(s.refreshTTL),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshTokenModel); err != nil {
		return "", "", fmt.Errorf("failed to save refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	refreshTokenModel, err := s.refreshTokenRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token")
	}

	user, err := s.userRepo.GetByID(ctx, refreshTokenModel.UserID)
	if err != nil {
		return "", "", fmt.Errorf("user not found")
	}

	accessToken, err := s.generateToken(user, s.accessTTL)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.generateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err := s.refreshTokenRepo.Delete(ctx, refreshToken); err != nil {
		return "", "", fmt.Errorf("failed to delete old refresh token: %w", err)
	}

	newRefreshTokenModel := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(s.refreshTTL),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.refreshTokenRepo.Create(ctx, newRefreshTokenModel); err != nil {
		return "", "", fmt.Errorf("failed to save new refresh token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

func (s *authService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignature
		}
		return []byte(s.config.Secret), nil
	}, jwt.WithoutClaimsValidation())
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	now := time.Now()
	clockSkew := s.config.ClockSkew

	if claims.ExpiresAt != nil && now.After(claims.ExpiresAt.Time.Add(clockSkew)) {
		return nil, ErrTokenExpired
	}

	if claims.NotBefore != nil && now.Before(claims.NotBefore.Time.Add(-clockSkew)) {
		return nil, ErrTokenNotBefore
	}

	if claims.Issuer != s.config.Issuer {
		return nil, ErrInvalidIssuer
	}

	audience, err := claims.RegisteredClaims.GetAudience()
	if err != nil || len(audience) == 0 {
		return nil, ErrInvalidAudience
	}

	found := false
	for _, aud := range audience {
		if aud == s.config.Audience {
			found = true
			break
		}
	}
	if !found {
		return nil, ErrInvalidAudience
	}

	return claims, nil
}

func (s *authService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

func (s *authService) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (s *authService) generateToken(user *models.User, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Audience:  jwt.ClaimStrings{s.config.Audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.Secret))
}

func (s *authService) generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	if err := s.refreshTokenRepo.Delete(ctx, refreshToken); err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}

func (s *authService) addLoginDelay() {
	time.Sleep(500 * time.Millisecond)
}
