package featureflags

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GinMiddleware implements FeatureFlagMiddleware for Gin framework
type GinMiddleware struct {
	manager FeatureFlagManager
	config  *MiddlewareConfig
}

// MiddlewareConfig holds configuration for the middleware
type MiddlewareConfig struct {
	// DefaultUserIDHeader is the header name to extract user ID from
	DefaultUserIDHeader string
	// DefaultServiceName is the service name to use when not specified
	DefaultServiceName string
	// EnableLogging enables request logging for feature flag evaluations
	EnableLogging bool
	// EnableMetrics enables metrics collection
	EnableMetrics bool
	// SkipPaths are paths to skip feature flag evaluation
	SkipPaths []string
	// RequiredFlags are flags that must be enabled for the request to proceed
	RequiredFlags []string
	// HeaderPrefix is the prefix for feature flag headers
	HeaderPrefix string
}

// DefaultMiddlewareConfig returns a default middleware configuration
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		DefaultUserIDHeader: "X-User-ID",
		DefaultServiceName:  "unknown",
		EnableLogging:       true,
		EnableMetrics:       true,
		SkipPaths:           []string{"/health", "/metrics", "/ping"},
		RequiredFlags:       []string{},
		HeaderPrefix:        "X-Feature-Flag-",
	}
}

// NewGinMiddleware creates a new Gin middleware for feature flags
func NewGinMiddleware(manager FeatureFlagManager, config *MiddlewareConfig) *GinMiddleware {
	if config == nil {
		config = DefaultMiddlewareConfig()
	}
	return &GinMiddleware{
		manager: manager,
		config:  config,
	}
}

// Handler returns a Gin middleware handler
func (m *GinMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip certain paths
		for _, skipPath := range m.config.SkipPaths {
			if strings.HasPrefix(c.Request.URL.Path, skipPath) {
				c.Next()
				return
			}
		}

		// Extract user context
		userContext := m.extractUserContext(c)

		// Check required flags
		for _, flagID := range m.config.RequiredFlags {
			enabled, _ := m.manager.IsEnabled(c.Request.Context(), flagID, userContext)
			if !enabled {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "Feature not available",
					"flag_id": flagID,
				})
				c.Abort()
				return
			}
		}

		// Add feature flag context to Gin context
		c.Set("feature_flags_manager", m.manager)
		c.Set("feature_flags_user_context", userContext)

		// Add feature flag helper functions
		c.Set("IsFeatureEnabled", func(flagID string) bool {
			enabled, _ := m.manager.IsEnabled(c.Request.Context(), flagID, userContext)
			return enabled
		})

		c.Set("GetFeatureVariant", func(flagID string) string {
			variant, _ := m.manager.GetVariant(c.Request.Context(), flagID, userContext)
			return variant
		})

		c.Set("GetFeatureValue", func(flagID string, defaultValue interface{}) interface{} {
			value, _ := m.manager.GetValue(c.Request.Context(), flagID, userContext, defaultValue)
			return value
		})

		// Set feature flag headers for downstream services
		m.setFeatureFlagHeaders(c, userContext)

		c.Next()
	}
}

// RequireFlag returns a middleware that requires a specific flag to be enabled
func (m *GinMiddleware) RequireFlag(flagID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext := m.extractUserContext(c)
		enabled, _ := m.manager.IsEnabled(c.Request.Context(), flagID, userContext)
		if !enabled {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Feature not available",
				"flag_id": flagID,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAnyFlag returns a middleware that requires at least one of the specified flags to be enabled
func (m *GinMiddleware) RequireAnyFlag(flagIDs ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext := m.extractUserContext(c)
		for _, flagID := range flagIDs {
			enabled, _ := m.manager.IsEnabled(c.Request.Context(), flagID, userContext)
			if enabled {
				c.Next()
				return
			}
		}
		c.JSON(http.StatusForbidden, gin.H{
			"error":    "Feature not available",
			"flag_ids": flagIDs,
		})
		c.Abort()
	}
}

// RequireAllFlags returns a middleware that requires all specified flags to be enabled
func (m *GinMiddleware) RequireAllFlags(flagIDs ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext := m.extractUserContext(c)
		for _, flagID := range flagIDs {
			enabled, _ := m.manager.IsEnabled(c.Request.Context(), flagID, userContext)
			if !enabled {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "Feature not available",
					"flag_id": flagID,
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// ConditionalHandler returns different handlers based on feature flag status
func (m *GinMiddleware) ConditionalHandler(flagID string, enabledHandler, disabledHandler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext := m.extractUserContext(c)
		enabled, _ := m.manager.IsEnabled(c.Request.Context(), flagID, userContext)
		if enabled {
			if enabledHandler != nil {
				enabledHandler(c)
			} else {
				c.Next()
			}
		} else {
			if disabledHandler != nil {
				disabledHandler(c)
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": "Feature not available"})
			}
		}
	}
}

// VariantHandler returns different handlers based on feature flag variant
func (m *GinMiddleware) VariantHandler(flagID string, handlers map[string]gin.HandlerFunc, defaultHandler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext := m.extractUserContext(c)
		variant, _ := m.manager.GetVariant(c.Request.Context(), flagID, userContext)

		if handler, exists := handlers[variant]; exists {
			handler(c)
		} else if defaultHandler != nil {
			defaultHandler(c)
		} else {
			c.Next()
		}
	}
}

// extractUserContext extracts user context from the Gin context
func (m *GinMiddleware) extractUserContext(c *gin.Context) *UserContext {
	userContext := &UserContext{
		Attributes: make(map[string]interface{}),
	}

	// Extract user ID from header
	if userID := c.GetHeader(m.config.DefaultUserIDHeader); userID != "" {
		userContext.UserID = userID
	}

	// Extract user ID from JWT token if available
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			userContext.UserID = uid
		}
	}

	// Extract session ID
	if sessionID := c.GetHeader("X-Session-ID"); sessionID != "" {
		userContext.Attributes["session_id"] = sessionID
	}

	// Extract IP address
	userContext.IPAddress = c.ClientIP()

	// Extract user agent
	userContext.UserAgent = c.GetHeader("User-Agent")

	// Extract custom attributes from headers
	for key, values := range c.Request.Header {
		if strings.HasPrefix(key, "X-User-") {
			attrKey := strings.ToLower(strings.TrimPrefix(key, "X-User-"))
			if len(values) > 0 {
				userContext.Attributes[attrKey] = values[0]
			}
		}
	}

	// Extract query parameters as attributes
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			userContext.Attributes["query_"+key] = values[0]
		}
	}

	// Add request context
	userContext.Attributes["method"] = c.Request.Method
	userContext.Attributes["path"] = c.Request.URL.Path
	userContext.Attributes["service"] = m.config.DefaultServiceName

	return userContext
}

// setFeatureFlagHeaders sets feature flag information in response headers
func (m *GinMiddleware) setFeatureFlagHeaders(c *gin.Context, userContext *UserContext) {
	// Get all flags for the user
	flags, _ := m.manager.EvaluateAllFlags(c.Request.Context(), userContext)

	// Set individual flag headers
	for flagID, result := range flags {
		headerName := m.config.HeaderPrefix + flagID
		c.Header(headerName+"-Enabled", strconv.FormatBool(result.Enabled))
		if result.Variant != "" {
			c.Header(headerName+"-Variant", result.Variant)
		}
		if result.Value != nil {
			if valueStr, err := json.Marshal(result.Value); err == nil {
				c.Header(headerName+"-Value", string(valueStr))
			}
		}
	}

	// Set summary header
	enabledFlags := make([]string, 0)
	for flagID, result := range flags {
		if result.Enabled {
			enabledFlags = append(enabledFlags, flagID)
		}
	}
	c.Header(m.config.HeaderPrefix+"Enabled", strings.Join(enabledFlags, ","))
}

// FeatureFlagHandler provides HTTP endpoints for feature flag management
type FeatureFlagHandler struct {
	manager FeatureFlagManager
}

// NewFeatureFlagHandler creates a new feature flag HTTP handler
func NewFeatureFlagHandler(manager FeatureFlagManager) *FeatureFlagHandler {
	return &FeatureFlagHandler{manager: manager}
}

// RegisterRoutes registers feature flag routes with a Gin router
func (h *FeatureFlagHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1/feature-flags")
	{
		// Flag management
		api.GET("/", h.GetFlags)
		api.GET("/:id", h.GetFlag)
		api.POST("/", h.CreateFlag)
		api.PUT("/:id", h.UpdateFlag)
		api.DELETE("/:id", h.DeleteFlag)

		// Flag evaluation
		api.POST("/evaluate", h.EvaluateFlags)
		api.POST("/evaluate/:id", h.EvaluateFlag)

		// Analytics
		api.GET("/:id/metrics", h.GetFlagMetrics)
		api.GET("/:id/usage", h.GetFlagUsage)

		// System
		api.GET("/health", h.Health)
		api.POST("/refresh", h.RefreshCache)
	}
}

// GetFlags returns all feature flags
func (h *FeatureFlagHandler) GetFlags(c *gin.Context) {
	flags, err := h.manager.GetAllFlags(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"flags": flags})
}

// GetFlag returns a specific feature flag
func (h *FeatureFlagHandler) GetFlag(c *gin.Context) {
	flagID := c.Param("id")
	flag, err := h.manager.GetFlag(c.Request.Context(), flagID)
	if err != nil {
		if err == ErrFlagNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Flag not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"flag": flag})
}

// CreateFlag creates a new feature flag
func (h *FeatureFlagHandler) CreateFlag(c *gin.Context) {
	var flag FeatureFlag
	if err := c.ShouldBindJSON(&flag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.manager.CreateFlag(c.Request.Context(), &flag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"flag": flag})
}

// UpdateFlag updates an existing feature flag
func (h *FeatureFlagHandler) UpdateFlag(c *gin.Context) {
	flagID := c.Param("id")
	var flag FeatureFlag
	if err := c.ShouldBindJSON(&flag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	flag.ID = flagID
	if err := h.manager.UpdateFlag(c.Request.Context(), &flag); err != nil {
		if err == ErrFlagNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Flag not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"flag": flag})
}

// DeleteFlag deletes a feature flag
func (h *FeatureFlagHandler) DeleteFlag(c *gin.Context) {
	flagID := c.Param("id")
	if err := h.manager.DeleteFlag(c.Request.Context(), flagID); err != nil {
		if err == ErrFlagNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Flag not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Flag deleted successfully"})
}

// EvaluateFlags evaluates all flags for a user
func (h *FeatureFlagHandler) EvaluateFlags(c *gin.Context) {
	var userContext UserContext
	if err := c.ShouldBindJSON(&userContext); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.manager.EvaluateAllFlags(c.Request.Context(), &userContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}

// EvaluateFlag evaluates a specific flag for a user
func (h *FeatureFlagHandler) EvaluateFlag(c *gin.Context) {
	flagID := c.Param("id")
	var userContext UserContext
	if err := c.ShouldBindJSON(&userContext); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.manager.EvaluateFlag(c.Request.Context(), flagID, &userContext)
	if err != nil {
		if err == ErrFlagNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Flag not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}

// GetFlagMetrics returns metrics for a specific flag
func (h *FeatureFlagHandler) GetFlagMetrics(c *gin.Context) {
	flagID := c.Param("id")
	startDate := time.Now().AddDate(0, 0, -7) // Default to last 7 days
	endDate := time.Now()

	if start := c.Query("start_date"); start != "" {
		if parsed, err := time.Parse("2006-01-02", start); err == nil {
			startDate = parsed
		}
	}

	if end := c.Query("end_date"); end != "" {
		if parsed, err := time.Parse("2006-01-02", end); err == nil {
			endDate = parsed
		}
	}

	// For now, return basic metrics info
	c.JSON(http.StatusOK, gin.H{
		"flag_id":    flagID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
		"message":    "Metrics endpoint - implementation depends on analytics backend",
	})
}

// GetFlagUsage returns usage statistics for a specific flag
func (h *FeatureFlagHandler) GetFlagUsage(c *gin.Context) {
	flagID := c.Param("id")
	startDate := time.Now().AddDate(0, 0, -7) // Default to last 7 days
	endDate := time.Now()

	if start := c.Query("start_date"); start != "" {
		if parsed, err := time.Parse("2006-01-02", start); err == nil {
			startDate = parsed
		}
	}

	if end := c.Query("end_date"); end != "" {
		if parsed, err := time.Parse("2006-01-02", end); err == nil {
			endDate = parsed
		}
	}

	// For now, return basic usage info
	c.JSON(http.StatusOK, gin.H{
		"flag_id":    flagID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
		"message":    "Usage endpoint - implementation depends on analytics backend",
	})
}

// Health returns the health status of the feature flag system
func (h *FeatureFlagHandler) Health(c *gin.Context) {
	// Basic health check - can be extended with actual health checks
	status := "healthy"
	statusCode := http.StatusOK

	c.JSON(statusCode, gin.H{
		"status":    status,
		"timestamp": time.Now().UTC(),
		"service":   "feature-flags",
	})
}

// RefreshCache refreshes the feature flag cache
func (h *FeatureFlagHandler) RefreshCache(c *gin.Context) {
	if err := h.manager.RefreshCache(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cache refreshed successfully"})
}

// Helper functions for Gin context

// IsFeatureEnabled checks if a feature flag is enabled for the current request
func IsFeatureEnabled(c *gin.Context, flagID string) bool {
	if fn, exists := c.Get("IsFeatureEnabled"); exists {
		if checkFn, ok := fn.(func(string) bool); ok {
			return checkFn(flagID)
		}
	}
	return false
}

// GetFeatureVariant gets the variant for a feature flag in the current request
func GetFeatureVariant(c *gin.Context, flagID string) string {
	if fn, exists := c.Get("GetFeatureVariant"); exists {
		if getFn, ok := fn.(func(string) string); ok {
			return getFn(flagID)
		}
	}
	return ""
}

// GetFeatureValue gets the value for a feature flag in the current request
func GetFeatureValue(c *gin.Context, flagID string, defaultValue interface{}) interface{} {
	if fn, exists := c.Get("GetFeatureValue"); exists {
		if getFn, ok := fn.(func(string, interface{}) interface{}); ok {
			return getFn(flagID, defaultValue)
		}
	}
	return defaultValue
}
