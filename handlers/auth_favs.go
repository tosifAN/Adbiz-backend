package handlers

import (
	"adbiz_backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) GetFavs(c *gin.Context) {
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

	// Fetch the user's favorites
	var favs models.Fav1
	if result := tx.Where("user_id = ?", user.ID).First(&favs); result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Favorites not found"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User account favs fetched successfully",
		"favs":    favs,
	})
}
