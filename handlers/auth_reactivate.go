package handlers

import (
	"adbiz_backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ReactivateUser handles the reactivation of a soft-deleted user account
func (h *AuthHandler) ReactivateUser(c *gin.Context) {
	// Apply rate limiting
	<-h.rateLimit.C

	// Get mobile number from URL parameter
	mobileNumber := c.Param("mobile_number")
	if mobileNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mobile number is required"})
		return
	}

	// Begin transaction
	tx := h.db.Begin()

	// Find the user by mobile number (including soft-deleted users)
	var user models.User
	if result := tx.Unscoped().Where("mobile_number = ?", mobileNumber).First(&user); result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if the account is actually deleted
	if user.DeletedAt.Time.IsZero() {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account is already active"})
		return
	}

	// Reactivate the user by setting deleted_at to null
	if err := tx.Unscoped().Model(&user).Update("deleted_at", nil).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reactivate user: " + err.Error()})
		return
	}

	// If user is a seller, reactivate their shop as well
	if user.Role == "seller" {
		var shop models.Shop
		if result := tx.Unscoped().Where("user_id = ?", user.ID).First(&shop); result.Error == nil {
			if err := tx.Unscoped().Model(&shop).Update("deleted_at", nil).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reactivate associated shop: " + err.Error()})
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
		"message": "User account reactivated successfully",
	})
}

// ReactivateShop handles the reactivation of a soft-deleted shop
func (h *AuthHandler) ReactivateShop(c *gin.Context) {
	// Apply rate limiting
	<-h.rateLimit.C

	// Get mobile number from URL parameter
	mobileNumber := c.Param("mobile_number")
	if mobileNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mobile number is required"})
		return
	}
	// Begin transaction
	tx := h.db.Begin()

	// First, find the user by mobile number
	var user models.User
	if result := tx.Where("mobile_number = ?", mobileNumber).First(&user); result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if user is a seller
	if user.Role != "seller" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only sellers can have shops"})
		return
	}

	// Find the shop associated with this user (including soft-deleted)
	var shop models.Shop
	if result := tx.Unscoped().Where("user_id = ?", user.ID).First(&shop); result.Error != nil {
		tx.Rollback()
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Shop not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find shop: " + result.Error.Error()})
		}
		return
	}

	// Check if the shop is actually deleted
	if shop.DeletedAt.Time.IsZero() {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Shop is already active"})
		return
	}

	// Reactivate the shop
	if err := tx.Unscoped().Model(&shop).Update("deleted_at", nil).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reactivate shop: " + err.Error()})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Shop reactivated successfully",
	})
}
