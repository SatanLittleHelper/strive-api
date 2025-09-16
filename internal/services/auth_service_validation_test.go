package services

import (
	"testing"
	"time"

	"github.com/aleksandr/strive-api/internal/config"
	"github.com/aleksandr/strive-api/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthService_ValidateToken(t *testing.T) {
	jwtConfig := &config.JWTConfig{
		Secret:    "test-secret-key",
		Issuer:    "test-issuer",
		Audience:  "test-audience",
		ClockSkew: 1 * time.Minute,
	}

	service := &authService{
		config:     jwtConfig,
		accessTTL:  15 * time.Minute,
		refreshTTL: 7 * 24 * time.Hour,
	}

	testUser := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	t.Run("ValidToken", func(t *testing.T) {
		token, err := service.generateToken(testUser, 15*time.Minute)
		require.NoError(t, err)

		claims, err := service.ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, testUser.ID, claims.UserID)
		assert.Equal(t, testUser.Email, claims.Email)
		assert.Equal(t, jwtConfig.Issuer, claims.Issuer)
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		now := time.Now()
		claims := &Claims{
			UserID: testUser.ID,
			Email:  testUser.Email,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    jwtConfig.Issuer,
				Audience:  jwt.ClaimStrings{jwtConfig.Audience},
				ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
				NotBefore: jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(jwtConfig.Secret))
		require.NoError(t, err)

		_, err = service.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("ExpiredTokenWithinClockSkew", func(t *testing.T) {
		now := time.Now()
		claims := &Claims{
			UserID: testUser.ID,
			Email:  testUser.Email,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    jwtConfig.Issuer,
				Audience:  jwt.ClaimStrings{jwtConfig.Audience},
				ExpiresAt: jwt.NewNumericDate(now.Add(-30 * time.Second)),
				IssuedAt:  jwt.NewNumericDate(now.Add(-1 * time.Hour)),
				NotBefore: jwt.NewNumericDate(now.Add(-1 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(jwtConfig.Secret))
		require.NoError(t, err)

		validatedClaims, err := service.ValidateToken(tokenString)
		assert.NoError(t, err)
		assert.Equal(t, testUser.ID, validatedClaims.UserID)
	})

	t.Run("NotBeforeToken", func(t *testing.T) {
		now := time.Now()
		claims := &Claims{
			UserID: testUser.ID,
			Email:  testUser.Email,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    jwtConfig.Issuer,
				Audience:  jwt.ClaimStrings{jwtConfig.Audience},
				ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(now),
				NotBefore: jwt.NewNumericDate(now.Add(2 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(jwtConfig.Secret))
		require.NoError(t, err)

		_, err = service.ValidateToken(tokenString)
		assert.ErrorIs(t, err, ErrTokenNotBefore)
	})

	t.Run("NotBeforeTokenWithinClockSkew", func(t *testing.T) {
		now := time.Now()
		claims := &Claims{
			UserID: testUser.ID,
			Email:  testUser.Email,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    jwtConfig.Issuer,
				Audience:  jwt.ClaimStrings{jwtConfig.Audience},
				ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(now),
				NotBefore: jwt.NewNumericDate(now.Add(30 * time.Second)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(jwtConfig.Secret))
		require.NoError(t, err)

		validatedClaims, err := service.ValidateToken(tokenString)
		assert.NoError(t, err)
		assert.Equal(t, testUser.ID, validatedClaims.UserID)
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		now := time.Now()
		claims := &Claims{
			UserID: testUser.ID,
			Email:  testUser.Email,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    jwtConfig.Issuer,
				Audience:  jwt.ClaimStrings{jwtConfig.Audience},
				ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(now),
				NotBefore: jwt.NewNumericDate(now),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("wrong-secret"))
		require.NoError(t, err)

		_, err = service.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signature")
	})

	t.Run("InvalidIssuer", func(t *testing.T) {
		now := time.Now()
		claims := &Claims{
			UserID: testUser.ID,
			Email:  testUser.Email,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "wrong-issuer",
				Audience:  jwt.ClaimStrings{jwtConfig.Audience},
				ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(now),
				NotBefore: jwt.NewNumericDate(now),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(jwtConfig.Secret))
		require.NoError(t, err)

		_, err = service.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "issuer")
	})

	t.Run("InvalidAudience", func(t *testing.T) {
		now := time.Now()
		claims := &Claims{
			UserID: testUser.ID,
			Email:  testUser.Email,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    jwtConfig.Issuer,
				Audience:  jwt.ClaimStrings{"wrong-audience"},
				ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(now),
				NotBefore: jwt.NewNumericDate(now),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(jwtConfig.Secret))
		require.NoError(t, err)

		_, err = service.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "audience")
	})

	t.Run("MalformedToken", func(t *testing.T) {
		_, err := service.ValidateToken("invalid.jwt.token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parse")
	})

	t.Run("EmptyToken", func(t *testing.T) {
		_, err := service.ValidateToken("")
		assert.Error(t, err)
	})
}
