package cache

import (
	"adbiz_backend/config"
	"adbiz_backend/models"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

var (
	PostCachePrefix     = os.Getenv("REDIS_POST_CACHE_PREFIX")
	UserCachePrefix     = os.Getenv("REDIS_USER_CACHE_PREFIX")
	TempUserInfoPrefix  = "temp:user:"
	DefaultExpiration   = time.Duration(func() int {
		exp, err := strconv.Atoi(os.Getenv("REDIS_CACHE_EXPIRATION"))
		if err != nil || exp <= 0 {
			return 30 // Default to 30 minutes if not set or invalid
		}
		return exp
	}()) * time.Minute
	TempDataExpiration = 15 * time.Minute // Temporary data expires after 15 minutes
)

// CacheUser stores a user in Redis cache
func CacheUser(ctx context.Context, user *models.User) error {
	if user == nil {
		return fmt.Errorf("cannot cache nil user")
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	key := fmt.Sprintf("%s%d", UserCachePrefix, user.ID)
	return config.RedisClient.Set(ctx, key, userJSON, DefaultExpiration).Err()
}

// GetCachedUser retrieves a user from Redis cache by ID
func GetCachedUser(ctx context.Context, userID uint) (*models.User, error) {
	key := fmt.Sprintf("%s%d", UserCachePrefix, userID)
	userJSON, err := config.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// InvalidateUserCache removes a user from Redis cache
func InvalidateUserCache(ctx context.Context, userID uint) error {
	key := fmt.Sprintf("%s%d", UserCachePrefix, userID)
	return config.RedisClient.Del(ctx, key).Err()
}

// CacheTempUserInfo stores temporary user information during the registration process
func CacheTempUserInfo(ctx context.Context, mobileNumber string, user models.User) error {
	userJSON, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal temporary user info: %w", err)
	}

	key := fmt.Sprintf("%s%s", TempUserInfoPrefix, mobileNumber)
	return config.RedisClient.Set(ctx, key, userJSON, TempDataExpiration).Err()
}

// GetTempUserInfo retrieves temporary user information during the registration process
func GetTempUserInfo(ctx context.Context, mobileNumber string) (models.User, error) {
	key := fmt.Sprintf("%s%s", TempUserInfoPrefix, mobileNumber)
	userJSON, err := config.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return models.User{}, fmt.Errorf("temporary user info not found: %w", err)
	}

	var user models.User
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return models.User{}, fmt.Errorf("failed to unmarshal temporary user info: %w", err)
	}

	return user, nil
}

// RemoveTempUserInfo removes temporary user information after registration is complete
func RemoveTempUserInfo(ctx context.Context, mobileNumber string) error {
	key := fmt.Sprintf("%s%s", TempUserInfoPrefix, mobileNumber)
	return config.RedisClient.Del(ctx, key).Err()
}
