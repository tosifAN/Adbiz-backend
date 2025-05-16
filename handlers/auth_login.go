package handlers

import (
	"adbiz_backend/cache"
	"adbiz_backend/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	MobileNumber string `json:"mobile_number" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	// Apply rate limiting
	<-h.rateLimit.C

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if result := h.db.Where("mobile_number = ?", req.MobileNumber).First(&user); result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found with this mobile number"})
		return
	}

	// Cache user data after successful login
	if err := cache.CacheUser(c.Request.Context(), &user); err != nil {
		// Log the error but don't fail the request
		log.Printf("Failed to cache user data: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
