package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elotusteam/microservice-project/services/notification/domain"
	"github.com/elotusteam/microservice-project/services/notification/usecases"
	"github.com/elotusteam/microservice-project/shared/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create Gin router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "notification",
			"timestamp": time.Now().UTC(),
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Notification endpoints
		v1.POST("/notifications", sendNotification)
		v1.POST("/notifications/bulk", sendBulkNotifications)
		v1.GET("/notifications", getNotifications)
		v1.GET("/notifications/:id", getNotificationByID)
		v1.PUT("/notifications/read", markAsRead)
		v1.DELETE("/notifications/:id", deleteNotification)
		v1.GET("/notifications/unread/count", getUnreadCount)

		// Preference endpoints
		v1.GET("/preferences", getPreferences)
		v1.PUT("/preferences", updatePreferences)

		// Template endpoints (admin)
		v1.POST("/templates", createTemplate)
		v1.GET("/templates/:id", getTemplate)
		v1.PUT("/templates/:id", updateTemplate)
		v1.DELETE("/templates/:id", deleteTemplate)
	}

	// Create HTTP server
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           router,
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Notification service starting on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down notification service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Notification service stopped")
}

// Handler functions (simplified implementations)
func sendNotification(c *gin.Context) {
	var req usecases.SendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mock response for now
	response := &usecases.SendNotificationResponse{
		NotificationID: uuid.New(),
		Status:         "sent",
		Message:        "Notification sent successfully",
	}

	c.JSON(http.StatusOK, response)
}

func sendBulkNotifications(c *gin.Context) {
	var req struct {
		UserIDs      []uuid.UUID                      `json:"user_ids"`
		Notification usecases.SendNotificationRequest `json:"notification"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bulk notifications sent successfully"})
}

func getNotifications(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Mock response
	response := &usecases.GetNotificationsResponse{
		Notifications: []*domain.Notification{},
		Total:         0,
		UnreadCount:   0,
	}

	c.JSON(http.StatusOK, response)
}

func getNotificationByID(c *gin.Context) {
	notificationID := c.Param("id")
	userID := c.Query("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Mock response
	c.JSON(http.StatusOK, gin.H{
		"id":      notificationID,
		"user_id": userID,
		"title":   "Sample Notification",
		"message": "This is a sample notification",
		"status":  "sent",
	})
}

func markAsRead(c *gin.Context) {
	var req usecases.MarkAsReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notifications marked as read"})
}

func deleteNotification(c *gin.Context) {
	notificationID := c.Param("id")
	userID := c.Query("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Notification %s deleted", notificationID)})
}

func getUnreadCount(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": 0})
}

func getPreferences(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"preferences": []*domain.NotificationPreference{}})
}

func updatePreferences(c *gin.Context) {
	var req usecases.UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Preferences updated successfully"})
}

func createTemplate(c *gin.Context) {
	var req usecases.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      uuid.New(),
		"name":    req.Name,
		"type":    req.Type,
		"message": "Template created successfully",
	})
}

func getTemplate(c *gin.Context) {
	templateID := c.Param("id")

	c.JSON(http.StatusOK, gin.H{
		"id":   templateID,
		"name": "Sample Template",
		"type": "email",
	})
}

func updateTemplate(c *gin.Context) {
	templateID := c.Param("id")

	c.JSON(http.StatusOK, gin.H{
		"id":      templateID,
		"message": "Template updated successfully",
	})
}

func deleteTemplate(c *gin.Context) {
	templateID := c.Param("id")

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Template %s deleted", templateID),
	})
}
