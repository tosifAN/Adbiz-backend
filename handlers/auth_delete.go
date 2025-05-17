package handlers

import (
	"adbiz_backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DeleteUser handles the soft deletion of a user account
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	// Apply rate limiting
	<-h.rateLimit.C

	// Get mobile number from URL parameter
	mobileNumber := c.Param("mobile_number")
	if mobileNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mobile number is required"})
		return
	}

	// Get authenticated user ID from context
	authUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Begin transaction
	tx := h.db.Begin()

	// Find the user by mobile number
	var user models.User
	if result := tx.Where("mobile_number = ? AND deleted_at IS NULL", mobileNumber).First(&user); result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if user is trying to delete their own data
	if user.ID != authUserID.(uint) {
		tx.Rollback()
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own account"})
		return
	}

	// Soft delete the user
	now := time.Now()
	if err := tx.Model(&user).Update("deleted_at", &now).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user: " + err.Error()})
		return
	}

	// If user is a seller, soft delete their shop as well
	if user.Role == "seller" {
		var shop models.Shop
		if result := tx.Where("user_id = ? AND deleted_at IS NULL", user.ID).First(&shop); result.Error == nil {
			if err := tx.Model(&shop).Update("deleted_at", &now).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete associated shop: " + err.Error()})
				return
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User account deleted successfully",
	})
}

// DeleteShop handles the soft deletion of a shop
func (h *AuthHandler) DeleteShop(c *gin.Context) {
	// Apply rate limiting
	<-h.rateLimit.C

	// Get mobile number from URL parameter
	mobileNumber := c.Param("mobile_number")
	if mobileNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mobile number is required"})
		return
	}

	// Get authenticated user ID from context
	authUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Begin transaction
	tx := h.db.Begin()

	// First, find the user by mobile number
	var user models.User
	if result := tx.Where("mobile_number = ? AND deleted_at IS NULL", mobileNumber).First(&user); result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if user is trying to delete their own shop
	if user.ID != authUserID.(uint) {
		tx.Rollback()
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own shop"})
		return
	}

	// Check if user is a seller
	if user.Role != "seller" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only sellers can have shops"})
		return
	}

	// Find the shop associated with this user
	var shop models.Shop
	if result := tx.Where("user_id = ? AND deleted_at IS NULL", user.ID).First(&shop); result.Error != nil {
		tx.Rollback()
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Shop not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find shop: " + result.Error.Error()})
		}
		return
	}

	// Soft delete the shop
	now := time.Now()
	if err := tx.Model(&shop).Update("deleted_at", &now).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete shop: " + err.Error()})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Shop deleted successfully",
	})
}
