package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration from environment
	host := getEnv("SERVER_HOST", "localhost")
	port := getEnvAsInt("SERVER_PORT", 8083)

	// Set Gin mode to debug for development
	gin.SetMode(gin.DebugMode)

	// Setup router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "user",
			"time":    time.Now().UTC(),
		})
	})

	// Basic API routes (placeholder for now)
	api := router.Group("/api/v1")
	{
		// User routes
		users := api.Group("/users")
		{
			users.POST("/", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Create user endpoint - implementation pending"})
			})
			users.GET("/:id", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Get user endpoint - implementation pending"})
			})
			users.PUT("/:id", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Update user endpoint - implementation pending"})
			})
			users.DELETE("/:id", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Delete user endpoint - implementation pending"})
			})
			users.GET("/", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "List users endpoint - implementation pending"})
			})
		}

		// Profile routes
		profile := api.Group("/profile")
		{
			profile.GET("/", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Get profile endpoint - implementation pending"})
			})
			profile.PUT("/", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Update profile endpoint - implementation pending"})
			})
		}
	}

	// Start server
	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", host, port),
		Handler:           router,
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("User service starting on %s:%d", host, port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
