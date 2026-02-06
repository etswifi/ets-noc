package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) SetupRouter() *gin.Engine {
	router := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	// Public routes
	router.GET("/health", s.handleHealth)
	router.POST("/api/v1/auth/login", s.handleLogin)
	router.GET("/api/v1/auth/google", s.handleGoogleLogin)
	router.GET("/api/v1/auth/google/callback", s.handleGoogleCallback)

	// Protected routes
	api := router.Group("/api/v1")
	api.Use(AuthMiddleware(s.postgres))
	{
		// Auth
		api.GET("/auth/me", s.handleGetMe)

		// Dashboard
		api.GET("/dashboard", s.handleDashboard)

		// Properties
		api.GET("/properties", s.handleListProperties)
		api.POST("/properties", s.handleCreateProperty)
		api.GET("/properties/:id", s.handleGetProperty)
		api.PUT("/properties/:id", s.handleUpdateProperty)
		api.DELETE("/properties/:id", s.handleDeleteProperty)
		api.GET("/properties/:id/status", s.handleGetPropertyStatus)
		api.GET("/properties/:id/devices", s.handleGetPropertyDevices)
		api.POST("/properties/:id/sync-devices", s.handleSyncDevicesFromPfSense)

		// Contacts
		api.GET("/properties/:id/contacts", s.handleListContactsForProperty)
		api.POST("/properties/:id/contacts", s.handleCreateContact)
		api.GET("/contacts/:id", s.handleGetContact)
		api.PUT("/contacts/:id", s.handleUpdateContact)
		api.DELETE("/contacts/:id", s.handleDeleteContact)

		// Attachments
		api.GET("/properties/:id/attachments", s.handleListAttachmentsForProperty)
		api.POST("/properties/:id/attachments", s.handleUploadAttachment)
		api.GET("/attachments/:id/download", s.handleDownloadAttachment)
		api.DELETE("/attachments/:id", s.handleDeleteAttachment)

		// Devices
		api.GET("/devices", s.handleListDevices)
		api.POST("/devices", s.handleCreateDevice)
		api.GET("/devices/:id", s.handleGetDevice)
		api.PUT("/devices/:id", s.handleUpdateDevice)
		api.DELETE("/devices/:id", s.handleDeleteDevice)
		api.GET("/devices/:id/status", s.handleGetDeviceStatus)
		api.GET("/devices/:id/history", s.handleGetDeviceHistory)
		api.GET("/devices/:id/errors", s.handleGetDeviceErrors)

		// Admin-only routes
		admin := api.Group("")
		admin.Use(AdminOnlyMiddleware())
		{
			// Users
			admin.GET("/users", s.handleListUsers)
			admin.POST("/users", s.handleCreateUser)
			admin.PUT("/users/:id", s.handleUpdateUser)
			admin.DELETE("/users/:id", s.handleDeleteUser)

			// Settings
			admin.GET("/settings", s.handleGetSettings)
			admin.PUT("/settings", s.handleUpdateSettings)
		}
	}

	return router
}
