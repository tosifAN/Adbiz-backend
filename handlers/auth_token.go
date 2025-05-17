package handlers

import (
	"adbiz_backend/models"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// GenerateToken creates a JWT token for the given user or user ID
func GenerateToken(userOrID interface{}) (string, error) {
	expiryDays, _ := strconv.Atoi(os.Getenv("JWT_EXPIRY_DAYS"))
	if expiryDays <= 0 {
		expiryDays = 7 // Default value
	}

	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 24 * time.Duration(expiryDays)).Unix(),
	}

	switch v := userOrID.(type) {
	case uint:
		claims["user_id"] = v
	case *models.User:
		claims["user_id"] = v.ID
		claims["role"] = v.Role
	default:
		return "", fmt.Errorf("invalid argument type for generateToken")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "Tosif@123" // Default value
	}

	return token.SignedString([]byte(secretKey))
}

// VerifyToken validates a JWT token and returns the claims if valid
func VerifyToken(tokenString string) (jwt.MapClaims, error) {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "Tosif@123" // Default value
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
