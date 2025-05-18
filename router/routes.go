package router

import (
	"adbiz_backend/config"
	"adbiz_backend/handlers"
	"adbiz_backend/middleware"

	"github.com/gin-gonic/gin"
)

// @BasePath /api/v1
func SetupRouter() *gin.Engine {
	r := config.SetupServer()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(config.Db)
	favHandler := handlers.NewFevHandler(config.Db)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes - Mobile verification and registration flow
		v1.POST("/verify-mobile", authHandler.VerifyMobile)            // Step 1: Verify if mobile exists
		v1.POST("/register-basic", authHandler.RegisterBasicInfo)      // Step 2: Register basic info
		v1.POST("/register-seller", authHandler.RegisterSellerDetails) // Step 3: Register seller details

		// Legacy routes (can be kept for backward compatibility)
		v1.POST("/register", authHandler.Register)
		v1.POST("/login", authHandler.Login)

		// Favorite routes
		v1.POST("/fav", favHandler.HandleFav) // Handle user favorites

		v1.POST("/user/reactivate/:mobile_number", authHandler.ReactivateUser)
		v1.POST("/user/shop/reactivate/:mobile_number", authHandler.ReactivateShop)

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// User routes
			protected.GET("/user/:mobile_number", authHandler.GetUser)
			protected.PUT("/user/:mobile_number", authHandler.UpdateUser)
			protected.GET("/user/shop/:mobile_number", authHandler.GetShopByUserMobile)
			protected.PUT("/user/shop/:mobile_number", authHandler.UpdateShop)
			protected.DELETE("/user/:mobile_number", authHandler.DeleteUser)
			protected.DELETE("/user/shop/:mobile_number", authHandler.DeleteShop)
			protected.GET("/user/favs/:mobile_number", authHandler.GetFavs)
		}
	}
	return r
}
