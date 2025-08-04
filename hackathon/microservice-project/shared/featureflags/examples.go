package featureflags

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ExampleUserService demonstrates how to integrate feature flags into a user service
type ExampleUserService struct {
	featureFlags FeatureFlagManager
	db           *sql.DB
}

// NewExampleUserService creates a new user service with feature flags
func NewExampleUserService(db *sql.DB) (*ExampleUserService, error) {
	// Initialize feature flags
	qs, err := NewQuickStart(db, "development")
	if err != nil {
		return nil, err
	}

	// Create some example flags for the user service
	if err := qs.CreateSampleFlags(); err != nil {
		log.Printf("Warning: Failed to create sample flags: %v", err)
	}

	return &ExampleUserService{
		featureFlags: qs.Manager,
		db:           db,
	}, nil
}

// GetUser demonstrates feature flag usage in a service method
func (s *ExampleUserService) GetUser(ctx context.Context, userID string) (map[string]interface{}, error) {
	// Create user context for feature flag evaluation
	userContext := &UserContext{
		UserID: userID,
		Attributes: map[string]interface{}{
			"service": "user",
		},
	}

	// Check if new user profile feature is enabled
	newProfileEnabled, err := s.featureFlags.IsEnabled(ctx, "new-user-profile", userContext)
	if err != nil {
		log.Printf("Error checking new-user-profile flag: %v", err)
		newProfileEnabled = false // Default to old behavior
	}

	var user map[string]interface{}

	if newProfileEnabled {
		// Use new user profile logic
		user = s.getUserWithNewProfile(ctx, userID)
		log.Printf("Using new user profile for user %s", userID)
	} else {
		// Use legacy user profile logic
		user = s.getUserWithLegacyProfile(ctx, userID)
		log.Printf("Using legacy user profile for user %s", userID)
	}

	// Track the feature flag usage
	event := &FeatureFlagEvent{
		FlagID:    "new-user-profile",
		UserID:    userID,
		Service:   "user",
		EventType: "evaluation",
		Result:    newProfileEnabled,
		Timestamp: time.Now(),
	}

	if err := s.featureFlags.TrackEvent(ctx, event); err != nil {
		log.Printf("Error tracking feature flag event: %v", err)
	}

	return user, nil
}

// getUserWithNewProfile simulates new user profile logic
func (s *ExampleUserService) getUserWithNewProfile(ctx context.Context, userID string) map[string]interface{} {
	return map[string]interface{}{
		"id":       userID,
		"name":     "John Doe",
		"email":    "john@example.com",
		"profile":  "enhanced",
		"features": []string{"social", "analytics", "recommendations"},
		"version":  "2.0",
	}
}

// getUserWithLegacyProfile simulates legacy user profile logic
func (s *ExampleUserService) getUserWithLegacyProfile(ctx context.Context, userID string) map[string]interface{} {
	return map[string]interface{}{
		"id":      userID,
		"name":    "John Doe",
		"email":   "john@example.com",
		"profile": "basic",
		"version": "1.0",
	}
}

// ExampleAPIHandler demonstrates feature flag integration in HTTP handlers
type ExampleAPIHandler struct {
	featureFlags FeatureFlagManager
	userService  *ExampleUserService
}

// NewExampleAPIHandler creates a new API handler with feature flags
func NewExampleAPIHandler(userService *ExampleUserService, featureFlags FeatureFlagManager) *ExampleAPIHandler {
	return &ExampleAPIHandler{
		featureFlags: featureFlags,
		userService:  userService,
	}
}

// SetupRoutes demonstrates how to setup routes with feature flag middleware
func (h *ExampleAPIHandler) SetupRoutes(r *gin.Engine) {
	// Create feature flag middleware
	config := DefaultMiddlewareConfig()
	middleware := NewGinMiddleware(h.featureFlags, config)

	// Apply middleware to all routes
	r.Use(middleware.Handler())

	// API routes
	api := r.Group("/api/v1")
	{
		api.GET("/users/:id", h.GetUserHandler)
		api.GET("/dashboard", h.GetDashboardHandler)
		api.POST("/analytics", h.PostAnalyticsHandler)
	}

	// Feature flag management routes
	flags := r.Group("/api/v1/flags")
	{
		handler := NewFeatureFlagHandler(h.featureFlags)
		// Register individual routes
		flags.GET("", handler.GetFlags)
		flags.GET("/:id", handler.GetFlag)
		flags.POST("", handler.CreateFlag)
		flags.PUT("/:id", handler.UpdateFlag)
		flags.DELETE("/:id", handler.DeleteFlag)
		flags.POST("/:id/evaluate", handler.EvaluateFlag)
		flags.POST("/evaluate", handler.EvaluateFlags)
		flags.GET("/:id/metrics", handler.GetFlagMetrics)
		flags.GET("/health", handler.Health)
	}
}

// GetUserHandler demonstrates feature flag usage in HTTP handlers
func (h *ExampleAPIHandler) GetUserHandler(c *gin.Context) {
	userID := c.Param("id")

	// Get user context from middleware
	userContext, exists := c.Get("user_context")
	if !exists {
		// Fallback: create user context from request
		userContext = &UserContext{
			UserID:    userID,
			IPAddress: c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
		}
	}

	// Get user data
	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if we should include additional data
	ctx := c.Request.Context()
	uc := userContext.(*UserContext)

	// Check for enhanced response feature
	enhancedResponse, err := h.featureFlags.IsEnabled(ctx, "enhanced-api-response", uc)
	if err != nil {
		log.Printf("Error checking enhanced-api-response flag: %v", err)
		enhancedResponse = false
	}

	response := gin.H{"user": user}

	if enhancedResponse {
		// Add additional metadata
		response["metadata"] = gin.H{
			"timestamp":   time.Now(),
			"api_version": "2.0",
			"features":    []string{"enhanced"},
		}
		response["_flags"] = gin.H{
			"enhanced_response": true,
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetDashboardHandler demonstrates conditional feature rendering
func (h *ExampleAPIHandler) GetDashboardHandler(c *gin.Context) {
	// Get user context
	userContext, _ := c.Get("user_context")
	uc := userContext.(*UserContext)
	ctx := c.Request.Context()

	// Evaluate multiple flags for dashboard features
	flags, err := h.featureFlags.EvaluateAllFlags(ctx, uc)
	if err != nil {
		log.Printf("Error evaluating flags: %v", err)
		flags = make(map[string]*EvaluationResult)
	}

	// Build dashboard response based on flags
	dashboard := gin.H{
		"title":   "User Dashboard",
		"widgets": []gin.H{},
	}

	// Always include basic widgets
	dashboard["widgets"] = append(dashboard["widgets"].([]gin.H), gin.H{
		"type": "profile",
		"name": "User Profile",
	})

	// Conditionally add widgets based on feature flags
	if flags["analytics-widget"] != nil && flags["analytics-widget"].Enabled {
		dashboard["widgets"] = append(dashboard["widgets"].([]gin.H), gin.H{
			"type": "analytics",
			"name": "Analytics Dashboard",
			"data": h.getAnalyticsData(ctx, uc),
		})
	}

	if flags["social-widget"] != nil && flags["social-widget"].Enabled {
		dashboard["widgets"] = append(dashboard["widgets"].([]gin.H), gin.H{
			"type": "social",
			"name": "Social Feed",
			"data": h.getSocialData(ctx, uc),
		})
	}

	if flags["beta-features"] != nil && flags["beta-features"].Enabled {
		dashboard["widgets"] = append(dashboard["widgets"].([]gin.H), gin.H{
			"type": "beta",
			"name": "Beta Features",
			"data": h.getBetaFeatures(ctx, uc),
		})
	}

	// Include flag information for frontend
	dashboard["_feature_flags"] = h.formatFlagsForFrontend(flags)

	c.JSON(http.StatusOK, dashboard)
}

// PostAnalyticsHandler demonstrates feature gating for new endpoints
func (h *ExampleAPIHandler) PostAnalyticsHandler(c *gin.Context) {
	userContext, _ := c.Get("user_context")
	uc := userContext.(*UserContext)
	ctx := c.Request.Context()

	// Check if analytics endpoint is enabled
	analyticsEnabled, err := h.featureFlags.IsEnabled(ctx, "analytics-endpoint", uc)
	if err != nil {
		log.Printf("Error checking analytics-endpoint flag: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Feature check failed"})
		return
	}

	if !analyticsEnabled {
		c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint not available"})
		return
	}

	// Process analytics data
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Track analytics event
	event := &FeatureFlagEvent{
		FlagID:    "analytics-endpoint",
		UserID:    uc.UserID,
		Service:   "api",
		EventType: "exposure",
		Result:    true,
		Metadata:  data,
		Timestamp: time.Now(),
	}

	if err := h.featureFlags.TrackEvent(ctx, event); err != nil {
		log.Printf("Error tracking analytics event: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Analytics data received",
		"status":  "processed",
	})
}

// Helper methods

func (h *ExampleAPIHandler) getAnalyticsData(ctx context.Context, uc *UserContext) map[string]interface{} {
	return map[string]interface{}{
		"page_views": 1234,
		"sessions":   56,
		"duration":   "2h 30m",
	}
}

func (h *ExampleAPIHandler) getSocialData(ctx context.Context, uc *UserContext) map[string]interface{} {
	return map[string]interface{}{
		"posts":     []string{"Hello world!", "Feature flags are awesome!"},
		"followers": 123,
		"following": 45,
	}
}

func (h *ExampleAPIHandler) getBetaFeatures(ctx context.Context, uc *UserContext) map[string]interface{} {
	return map[string]interface{}{
		"features": []string{"ai-assistant", "advanced-search", "real-time-collab"},
		"feedback": "https://feedback.example.com",
	}
}

func (h *ExampleAPIHandler) formatFlagsForFrontend(flags map[string]*EvaluationResult) map[string]interface{} {
	result := make(map[string]interface{})
	for flagID, evaluation := range flags {
		result[flagID] = gin.H{
			"enabled": evaluation.Enabled,
			"variant": evaluation.Variant,
			"reason":  evaluation.Reason,
		}
	}
	return result
}

// ExampleBackgroundService demonstrates feature flags in background services
type ExampleBackgroundService struct {
	featureFlags FeatureFlagManager
	ticker       *time.Ticker
	done         chan bool
}

// NewExampleBackgroundService creates a new background service with feature flags
func NewExampleBackgroundService(featureFlags FeatureFlagManager) *ExampleBackgroundService {
	return &ExampleBackgroundService{
		featureFlags: featureFlags,
		ticker:       time.NewTicker(1 * time.Minute),
		done:         make(chan bool),
	}
}

// Start begins the background service
func (s *ExampleBackgroundService) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-s.done:
				return
			case <-s.ticker.C:
				s.processTask(ctx)
			}
		}
	}()
}

// Stop stops the background service
func (s *ExampleBackgroundService) Stop() {
	s.ticker.Stop()
	s.done <- true
}

// processTask demonstrates feature flag usage in background tasks
func (s *ExampleBackgroundService) processTask(ctx context.Context) {
	// Create system user context for background tasks
	systemContext := &UserContext{
		UserID: "system",
		Attributes: map[string]interface{}{
			"service": "background",
			"type":    "system",
		},
	}

	// Check if new processing algorithm is enabled
	newAlgorithm, err := s.featureFlags.IsEnabled(ctx, "new-processing-algorithm", systemContext)
	if err != nil {
		log.Printf("Error checking new-processing-algorithm flag: %v", err)
		newAlgorithm = false
	}

	if newAlgorithm {
		log.Println("Using new processing algorithm")
		s.processWithNewAlgorithm(ctx)
	} else {
		log.Println("Using legacy processing algorithm")
		s.processWithLegacyAlgorithm(ctx)
	}

	// Check if enhanced logging is enabled
	enhancedLogging, err := s.featureFlags.IsEnabled(ctx, "enhanced-logging", systemContext)
	if err != nil {
		log.Printf("Error checking enhanced-logging flag: %v", err)
		enhancedLogging = false
	}

	if enhancedLogging {
		log.Printf("Task completed at %v with algorithm: %s", time.Now(), map[bool]string{true: "new", false: "legacy"}[newAlgorithm])
	}
}

func (s *ExampleBackgroundService) processWithNewAlgorithm(ctx context.Context) {
	// Simulate new algorithm processing
	time.Sleep(100 * time.Millisecond)
	log.Println("New algorithm processing completed")
}

func (s *ExampleBackgroundService) processWithLegacyAlgorithm(ctx context.Context) {
	// Simulate legacy algorithm processing
	time.Sleep(200 * time.Millisecond)
	log.Println("Legacy algorithm processing completed")
}

// ExampleMain demonstrates a complete application setup
func ExampleMain() {
	// Database connection (optional)
	var db *sql.DB
	// db, err := sql.Open("postgres", "your-connection-string")
	// if err != nil {
	//     log.Fatal(err)
	// }
	// defer db.Close()

	// Create user service with feature flags
	userService, err := NewExampleUserService(db)
	if err != nil {
		log.Fatal("Failed to create user service:", err)
	}

	// Create feature flag system
	qs, err := NewQuickStart(db, "development")
	if err != nil {
		log.Fatal("Failed to create feature flag system:", err)
	}
	defer qs.Stop()

	// Create API handler
	apiHandler := NewExampleAPIHandler(userService, qs.Manager)

	// Create background service
	backgroundService := NewExampleBackgroundService(qs.Manager)
	backgroundService.Start(context.Background())
	defer backgroundService.Stop()

	// Setup Gin router
	r := gin.Default()
	apiHandler.SetupRoutes(r)

	// Add health check endpoint
	r.GET("/health", func(c *gin.Context) {
		ctx := c.Request.Context()
		if err := qs.Manager.HealthCheck(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now(),
		})
	})

	// Start server
	log.Println("Starting server on :8080")
	log.Fatal(r.Run(":8080"))
}
