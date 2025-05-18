package handlers

import (
	"adbiz_backend/models"
	"fmt"
	"net/http"

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

// GetUser retrieves a user by mobile number
func (h *AuthHandler) GetUser(c *gin.Context) {
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

	// Retrieve user from database by mobile number
	var user models.User
	if result := h.db.Where("mobile_number = ?", mobileNumber).First(&user); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if user is trying to access their own data
	if user.ID != authUserID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only access your own user data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// GetShopByUserMobile retrieves a shop by user's mobile number
func (h *AuthHandler) GetShopByUserMobile(c *gin.Context) {
	// Apply rate limiting
	<-h.rateLimit.C

	fmt.Println("first")
	// Get mobile number from URL parameter
	mobileNumber := c.Param("mobile_number")
	if mobileNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mobile number is required"})
		return
	}
	fmt.Println("first2")

	// Get authenticated user ID from context
	authUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	fmt.Println("first3")

	// First, find the user by mobile number
	var user models.User
	if result := h.db.Where("mobile_number = ?", mobileNumber).First(&user); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	fmt.Println("you are here", user)

	// Check if user is trying to access their own data
	if user.ID != authUserID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only access your own shop data"})
		return
	}

	fmt.Println("now here")

	// Now, find the shop associated with this user
	var shop models.Shop
	if result := h.db.Where("user_id = ?", user.ID).First(&shop); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Shop not found for this user"})
		return
	}
	fmt.Println("this is your shop", shop)

	c.JSON(http.StatusOK, gin.H{
		"shop": shop,
	})
}

// UpdateUserRequest defines the request structure for updating a user
type UpdateUserRequest struct {
	Name         string  `json:"name"`
	Email        *string `json:"email,omitempty"`
	ProfilePhoto *string `json:"profile_photo,omitempty"`
	MobileNumber string  `json:"mobile_number"`
	Role         string  `json:"role"`
}

// UpdateUser updates a user's information
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	// Apply rate limiting
	<-h.rateLimit.C

	// Get mobile number from URL parameter
	mobileNumber := c.Param("mobile_number")
	if mobileNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mobile number is required"})
		return
	}

	// First, find the user by mobile number
	var user models.User
	if result := h.db.Where("mobile_number = ?", mobileNumber).First(&user); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get authenticated user ID from context
	authUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if user is trying to update their own data
	if user.ID != authUserID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own user data"})
		return
	}

	// Parse request body
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update user fields if provided in request
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != nil {
		user.Email = req.Email
	}
	if req.ProfilePhoto != nil {
		user.ProfilePhoto = req.ProfilePhoto
	}
	if req.MobileNumber != "" {
		user.MobileNumber = req.MobileNumber
	}
	if req.Role != "" {
		user.Role = req.Role
	}

	// Save updated user to database
	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

type UpdateShopRequest struct {
	Bio          *string `json:"bio,omitempty"`
	Location     *string `json:"location,omitempty"`
	ShopPhoto    *string `json:"shop_photo,omitempty"`
	ShopUsername string  `json:"shop_username"`
	ShopName     string  `json:"shop_name"`
	ProductType  string  `json:"product_type"`
}

// UpdateUser updates a user's information
func (h *AuthHandler) UpdateShop(c *gin.Context) {
	// Apply rate limiting
	<-h.rateLimit.C

	// Get mobile number from URL parameter
	mobileNumber := c.Param("mobile_number")
	if mobileNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mobile number is required"})
		return
	}

	// First, find the user by mobile number
	var user models.User
	if result := h.db.Where("mobile_number = ?", mobileNumber).First(&user); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get authenticated user ID from context
	authUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if user is trying to update their own data
	if user.ID != authUserID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own shop data"})
		return
	}

	// Parse request body
	var req UpdateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Now, find the shop associated with this user
	var shop models.Shop
	if result := h.db.Where("user_id = ?", user.ID).First(&shop); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Shop not found for this user"})
		return
	}

	if req.Bio != nil {
		shop.Bio = req.Bio
	}
	if req.Location != nil {
		shop.Location = req.Location
	}
	if req.ShopPhoto != nil {
		shop.ShopPhoto = req.ShopPhoto
	}
	if req.ShopUsername != "" {
		shop.ShopUsername = req.ShopUsername
	}
	if req.ShopName != "" {
		shop.ShopName = req.ShopName
	}
	if req.ProductType != "" {
		shop.ProductType = req.ProductType
	}

	// Save updated shop to database
	if err := h.db.Save(&shop).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update shop: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": shop,
	})
}
