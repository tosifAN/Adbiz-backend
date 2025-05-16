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

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// User routes
			protected.GET("/users/:id", authHandler.GetUser)
			protected.PUT("/users/:id", authHandler.UpdateUser)
		}
	}
	return r
}
