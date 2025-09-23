package http

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
)

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID format: %w", err)
	}
	return userUUID, nil
}

func GetUserEmailFromContext(ctx context.Context) (string, error) {
	email, ok := ctx.Value(UserEmailKey).(string)
	if !ok {
		return "", fmt.Errorf("user email not found in context")
	}
	return email, nil
}
