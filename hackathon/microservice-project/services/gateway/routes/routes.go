package routes

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/elotusteam/microservice-project/shared/config"
)

// SetupRoutes configures all routes for the API gateway
func SetupRoutes(router *gin.Engine, cfg *config.Config) {
	// API version group
	v1 := router.Group("/api/v1")
	{
		// User service routes
		userGroup := v1.Group("/users")
		{
			userGroup.Any("/*path", proxyToService(cfg.Services.User.BaseURL))
		}

		// File service routes
		fileGroup := v1.Group("/files")
		{
			fileGroup.Any("/*path", proxyToService(cfg.Services.File.BaseURL))
		}

		// Notification service routes
		notificationGroup := v1.Group("/notifications")
		{
			notificationGroup.Any("/*path", proxyToService(cfg.Services.Notification.BaseURL))
		}

		// Analytics service routes
		analyticsGroup := v1.Group("/analytics")
		{
			analyticsGroup.Any("/*path", proxyToService(cfg.Services.Analytics.BaseURL))
		}

		// Search service routes
		searchGroup := v1.Group("/search")
		{
			searchGroup.Any("/*path", proxyToService(cfg.Services.Search.BaseURL))
		}
	}

	// Admin routes
	admin := router.Group("/admin")
	{
		// Admin user management
		admin.Any("/users/*path", proxyToService(cfg.Services.User.BaseURL))
		
		// Admin file management
		admin.Any("/files/*path", proxyToService(cfg.Services.File.BaseURL))
		
		// Admin analytics
		admin.Any("/analytics/*path", proxyToService(cfg.Services.Analytics.BaseURL))
	}
}

// proxyToService creates a reverse proxy handler for a service
func proxyToService(serviceURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse the target service URL
		target, err := url.Parse(serviceURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid service URL",
			})
			return
		}

		// Create reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(target)

		// Modify the request
		proxy.Director = func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = c.Request.URL.Path
			req.URL.RawQuery = c.Request.URL.RawQuery
			req.Header = c.Request.Header
			
			// Add gateway headers
			req.Header.Set("X-Gateway-Request-ID", c.GetHeader("X-Request-ID"))
			req.Header.Set("X-Forwarded-For", c.ClientIP())
			req.Header.Set("X-Forwarded-Proto", c.Request.Header.Get("X-Forwarded-Proto"))
			if c.Request.TLS != nil {
				req.Header.Set("X-Forwarded-Proto", "https")
			} else {
				req.Header.Set("X-Forwarded-Proto", "http")
			}
		}

		// Handle errors
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			c.JSON(http.StatusBadGateway, gin.H{
				"error": "Service unavailable",
				"message": err.Error(),
			})
		}

		// Serve the request
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}