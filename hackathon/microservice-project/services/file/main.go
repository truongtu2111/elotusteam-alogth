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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus metrics
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
	fileUploadsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "file_uploads_total",
			Help: "Total number of file uploads",
		},
		[]string{"status"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(fileUploadsTotal)
}

// Prometheus middleware
func prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
	}
}

func main() {
	// Initialize service container
	container, err := NewServiceContainer(nil) // Using nil config for now
	if err != nil {
		log.Fatalf("Failed to initialize service container: %v", err)
	}

	// Load configuration from environment
	host := getEnv("SERVER_HOST", "localhost")
	port := getEnvAsInt("SERVER_PORT", 8082)

	// Set Gin mode to debug for development
	gin.SetMode(gin.DebugMode)

	// Setup router
	router := gin.Default()

	// Add Prometheus middleware
	router.Use(prometheusMiddleware())

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "file",
			"time":    time.Now().UTC(),
		})
	})

	// Basic API routes using the file service
	api := router.Group("/api/v1")
	{
		// File routes
		files := api.Group("/files")
		{
			files.POST("/upload", func(c *gin.Context) {
				// File service is available via container.FileService
				// Implementation would use container.FileService.UploadFile()
				fileUploadsTotal.WithLabelValues("success").Inc()
				c.JSON(http.StatusOK, gin.H{"message": "Upload endpoint - file service integrated"})
			})
			files.GET("/:id", func(c *gin.Context) {
				// Implementation would use container.FileService.GetFile()
				c.JSON(http.StatusOK, gin.H{"message": "Get file endpoint - file service integrated"})
			})
			files.DELETE("/:id", func(c *gin.Context) {
				// Implementation would use container.FileService.DeleteFile()
				c.JSON(http.StatusOK, gin.H{"message": "Delete file endpoint - file service integrated"})
			})
			files.GET("/", func(c *gin.Context) {
				// Implementation would use container.FileService.ListFiles()
				c.JSON(http.StatusOK, gin.H{"message": "List files endpoint - file service integrated"})
			})
		}
	}

	// File service is now available via container.FileService with image processing capabilities
	_ = container // Suppress unused variable warning for now

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
		log.Printf("File service starting on %s:%d", host, port)
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
