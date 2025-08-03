package middleware

import (
	"github.com/elotusteam/microservice-project/shared/config"
	"github.com/gin-gonic/gin"
)

// Security returns a middleware that adds security headers
func Security(cfg *config.Config) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// X-Content-Type-Options
		c.Header("X-Content-Type-Options", "nosniff")

		// X-Frame-Options
		c.Header("X-Frame-Options", "DENY")

		// X-XSS-Protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Strict-Transport-Security (HSTS)
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Content-Security-Policy
		csp := "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' https:; connect-src 'self'"
		c.Header("Content-Security-Policy", csp)

		// Referrer-Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	})
}
