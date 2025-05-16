package handlers

import (
	"adbiz_backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Register is a legacy handler for user registration
// It's kept for backward compatibility
func (h *AuthHandler) Register(c *gin.Context) {
	// Apply rate limiting
	<-h.rateLimit.C

	// This is a legacy endpoint, redirect to the new registration flow
	c.JSON(http.StatusOK, gin.H{
		"message": "This endpoint is deprecated. Please use the new registration flow: /verify-mobile, /register-basic, and /register-seller.",
	})
}

// GetUser retrieves a user by ID
func (h *AuthHandler) GetUser(c *gin.Context) {
	// Apply rate limiting
	<-h.rateLimit.C

	// Get user ID from URL parameter
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get authenticated user ID from context
	authUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if user is trying to access their own data
	if uint(userID) != authUserID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only access your own user data"})
		return
	}

	// Retrieve user from database
	var user models.User
	if result := h.db.First(&user, userID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// UpdateUserRequest defines the request structure for updating a user
type UpdateUserRequest struct {
	Name string `json:"name" binding:"required"`
}

// UpdateUser updates a user's information
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	// Apply rate limiting
	<-h.rateLimit.C

	// Get user ID from URL parameter
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get authenticated user ID from context
	authUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if user is trying to update their own data
	if uint(userID) != authUserID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own user data"})
		return
	}

	// Parse request body
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retrieve user from database
	var user models.User
	if result := h.db.First(&user, userID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update user fields
	user.Name = req.Name

	// Save updated user to database
	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
