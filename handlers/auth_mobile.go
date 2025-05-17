package handlers

import (
	"adbiz_backend/cache"
	"adbiz_backend/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MobileVerificationRequest struct {
	MobileNumber string `json:"mobile_number" binding:"required"`
}

type UserExistsResponse struct {
	Exists bool         `json:"exists"`
	User   *models.User `json:"user,omitempty"`
	Token  string       `json:"token,omitempty"`
}

// VerifyMobile checks if a user with the given mobile number exists
// This is the first step in the authentication flow
func (h *AuthHandler) VerifyMobile(c *gin.Context) {
	// Apply rate limiting
	<-h.rateLimit.C

	var req MobileVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user with this mobile number already exists
	var user models.User
	result := h.db.Where("mobile_number = ?", req.MobileNumber).First(&user)

	if result.Error == nil {
		// User exists, return user data
		// Cache user data
		if err := cache.CacheUser(c.Request.Context(), &user); err != nil {
			log.Printf("Failed to cache user data: %v", err)
		}

		// Generate JWT token
		token, err := GenerateToken(&user)
		if err != nil {
			log.Printf("Failed to generate token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
			return
		}

		c.JSON(http.StatusOK, UserExistsResponse{
			Exists: true,
			User:   &user,
			Token:  token,
		})
	} else {
		// User doesn't exist
		c.JSON(http.StatusOK, UserExistsResponse{
			Exists: false,
		})
	}
}

type UserRegistrationRequest struct {
	MobileNumber string `json:"mobile_number" binding:"required"`
	Name         string `json:"name" binding:"required"`
	Role         string `json:"role" binding:"required,oneof=buyer seller"` // "buyer" or "seller"
}

type SellerDetailsRequest struct {
	MobileNumber string `json:"mobile_number" binding:"required"`
	ShopName     string `json:"shop_name" binding:"required"`
	ProductType  string `json:"product_type" binding:"required"` // "food", "clothes", "beauty", "healthcare"
	ShopUsername string `json:"shop_username" binding:"required"`
}

// RegisterBasicInfo registers basic user information after mobile verification
// This is the second step in the registration flow
func (h *AuthHandler) RegisterBasicInfo(c *gin.Context) {
	<-h.rateLimit.C

	var req UserRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User
	result := h.db.Where("mobile_number = ?", req.MobileNumber).First(&existingUser)
	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this mobile number already exists"})
		return
	}

	user := models.User{
		MobileNumber: req.MobileNumber,
		Name:         req.Name,
		Role:         req.Role,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
		return
	}

	if err := cache.CacheUser(c.Request.Context(), &user); err != nil {
		log.Printf("Failed to cache user data: %v", err)
	}

	// Generate JWT token
	token, err := GenerateToken(&user)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) RegisterSellerDetails(c *gin.Context) {
	<-h.rateLimit.C

	var req SellerDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch the user
	var user models.User
	if err := h.db.Where("mobile_number = ?", req.MobileNumber).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found. Please register first."})
		return
	}

	if user.Role != "seller" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not registered as a seller"})
		return
	}

	// Check if a shop already exists for this user
	var existingShop models.Shop
	if err := h.db.Where("user_id = ?", user.ID).First(&existingShop).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Shop already registered for this seller"})
		return
	}

	// Begin transaction
	tx := h.db.Begin()

	shopID := req.ProductType + req.MobileNumber + req.ShopUsername + req.ShopName

	shop := models.Shop{
		ShopID:       shopID,
		ShopName:     req.ShopName,
		ProductType:  req.ProductType,
		ShopUsername: req.ShopUsername,
		UserID:       user.ID,
	}

	if err := tx.Create(&shop).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create shop: " + err.Error()})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction: " + err.Error()})
		return
	}

	// Optional: refresh user from DB if needed
	if err := cache.CacheUser(c.Request.Context(), &user); err != nil {
		log.Printf("Failed to cache user data: %v", err)
	}

	// Generate JWT token
	token, err := GenerateToken(&user)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Shop registered successfully",
		"shop":    shop,
		"token":   token,
	})
}
