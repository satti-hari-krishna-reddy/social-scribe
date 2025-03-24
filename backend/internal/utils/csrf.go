package utils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"social-scribe/backend/internal/models"
	repo "social-scribe/backend/internal/repositories"
)

func GetOrCreateCsrfToken(userID string) (string, error) {
	cacheKey := fmt.Sprintf("CSRF_%s", userID)
	tokenData, exists := repo.GetCache(cacheKey)
	var csrfToken string

	if exists {
		tokenInfo, ok := tokenData.(models.CacheItem)
		if ok && tokenInfo.ExpiresAt.After(time.Now()) {
			// Use the token if it's valid.
			token, ok := tokenInfo.Value.(string)
			if ok {
				csrfToken = token
			} else {
				// Invalid type: delete it and generate a new one.
				_ = repo.DeleteCache(cacheKey)
			}
		} else {
			// Expired or invalid: remove it.
			_ = repo.DeleteCache(cacheKey)
		}
	}

	// Generate a new token if none exists.
	if csrfToken == "" {
		csrfToken = uuid.New().String()
		err := repo.SetCache(cacheKey, csrfToken, 10*time.Minute)
		if err != nil {
			return "", err
		}
	}

	return csrfToken, nil
}
