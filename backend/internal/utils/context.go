package utils

import (
	"context"
	"errors"
)

type contextKey string

const userIDKey contextKey = "userID"

func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return "", errors.New("user ID not found in context")
	}
	return userID, nil
}
