# Feature Flags Integration Guide

This guide shows how to integrate the feature flag system into existing microservices in the project.

## Overview

The feature flag system is located in `shared/featureflags/` and provides:

- **Backend Integration**: Gin middleware, service integration, background tasks
- **Frontend Integration**: REST API endpoints for JavaScript/React applications
- **Analytics**: Usage tracking and metrics
- **Management**: CRUD operations for flags
- **Multi-environment**: Development, staging, production support

## Quick Integration Steps

### 1. Add Feature Flags to Existing Services

For each microservice (user, file, notification, analytics, search, auth):

#### Step 1: Import the Package

```go
import "your-project/shared/featureflags"
```

#### Step 2: Initialize in Main Function

```go
// In your service's main.go
func main() {
    // Your existing database connection
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Initialize feature flags
    environment := os.Getenv("ENVIRONMENT")
    if environment == "" {
        environment = "development"
    }
    
    featureFlags, err := featureflags.NewQuickStart(db, environment)
    if err != nil {
        log.Fatal("Failed to initialize feature flags:", err)
    }
    defer featureFlags.Stop()
    
    // Your existing Gin setup
    r := gin.Default()
    
    // Add feature flag middleware
    r.Use(featureFlags.Middleware.Handler())
    
    // Add feature flag management endpoints
    flagsAPI := r.Group("/api/v1/flags")
    {
        handler := featureflags.NewFeatureFlagHandler(featureFlags.Manager)
        flagsAPI.GET("", handler.GetFlags)
        flagsAPI.GET("/:id", handler.GetFlag)
        flagsAPI.POST("", handler.CreateFlag)
        flagsAPI.PUT("/:id", handler.UpdateFlag)
        flagsAPI.DELETE("/:id", handler.DeleteFlag)
        flagsAPI.POST("/:id/evaluate", handler.EvaluateFlag)
        flagsAPI.POST("/evaluate", handler.EvaluateFlags)
        flagsAPI.GET("/:id/metrics", handler.GetFlagMetrics)
        flagsAPI.GET("/health", handler.Health)
    }
    
    // Your existing routes...
    setupRoutes(r, featureFlags.Manager)
    
    r.Run(":8080")
}
```

#### Step 3: Update Route Handlers

```go
// Example: User service handler
func getUserHandler(flagManager featureflags.FeatureFlagManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.Param("id")
        
        // Get user context from middleware
        userContext := &featureflags.UserContext{
            UserID: userID,
            IPAddress: c.ClientIP(),
            UserAgent: c.GetHeader("User-Agent"),
            Attributes: map[string]interface{}{
                "service": "user",
            },
        }
        
        // Check feature flags
        newProfileEnabled, _ := flagManager.IsEnabled(c.Request.Context(), "new-user-profile", userContext)
        enhancedDataEnabled, _ := flagManager.IsEnabled(c.Request.Context(), "enhanced-user-data", userContext)
        
        // Get user data based on flags
        var userData map[string]interface{}
        if newProfileEnabled {
            userData = getEnhancedUserProfile(userID)
        } else {
            userData = getBasicUserProfile(userID)
        }
        
        // Add additional data if flag is enabled
        if enhancedDataEnabled {
            userData["analytics"] = getUserAnalytics(userID)
            userData["preferences"] = getUserPreferences(userID)
        }
        
        c.JSON(http.StatusOK, userData)
    }
}
```

### 2. Service-Specific Integration Examples

#### User Service Integration

```go
// services/user/handlers/user_handler.go
package handlers

import (
    "your-project/shared/featureflags"
    "github.com/gin-gonic/gin"
)

type UserHandler struct {
    featureFlags featureflags.FeatureFlagManager
    // ... other dependencies
}

func (h *UserHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")
    
    userContext := &featureflags.UserContext{
        UserID: userID,
        Attributes: map[string]interface{}{
            "service": "user",
        },
    }
    
    // Feature: New user profile format
    newFormat, _ := h.featureFlags.IsEnabled(c.Request.Context(), "user-profile-v2", userContext)
    
    if newFormat {
        // Return new format
        c.JSON(200, gin.H{
            "user": getUserV2(userID),
            "version": "2.0",
        })
    } else {
        // Return legacy format
        c.JSON(200, gin.H{
            "user": getUserV1(userID),
            "version": "1.0",
        })
    }
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    userContext := &featureflags.UserContext{
        IPAddress: c.ClientIP(),
        Attributes: map[string]interface{}{
            "service": "user",
            "action": "create",
        },
    }
    
    // Feature: Enhanced user validation
    enhancedValidation, _ := h.featureFlags.IsEnabled(c.Request.Context(), "enhanced-user-validation", userContext)
    
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    if enhancedValidation {
        if err := validateUserEnhanced(&user); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
    } else {
        if err := validateUserBasic(&user); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
    }
    
    // Create user...
    c.JSON(201, gin.H{"user": user})
}
```

#### File Service Integration

```go
// services/file/handlers/file_handler.go
func (h *FileHandler) UploadFile(c *gin.Context) {
    userID := c.GetHeader("X-User-ID")
    
    userContext := &featureflags.UserContext{
        UserID: userID,
        Attributes: map[string]interface{}{
            "service": "file",
            "action": "upload",
        },
    }
    
    // Feature: Advanced file processing
    advancedProcessing, _ := h.featureFlags.IsEnabled(c.Request.Context(), "advanced-file-processing", userContext)
    
    // Feature: Virus scanning
    virusScanning, _ := h.featureFlags.IsEnabled(c.Request.Context(), "virus-scanning", userContext)
    
    // Feature: Image optimization
    imageOptimization, _ := h.featureFlags.IsEnabled(c.Request.Context(), "image-optimization", userContext)
    
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // Process file based on enabled features
    var processingOptions ProcessingOptions
    processingOptions.AdvancedProcessing = advancedProcessing
    processingOptions.VirusScanning = virusScanning
    processingOptions.ImageOptimization = imageOptimization
    
    result, err := h.processFile(file, processingOptions)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, result)
}
```

#### Analytics Service Integration

```go
// services/analytics/handlers/analytics_handler.go
func (h *AnalyticsHandler) GetDashboard(c *gin.Context) {
    userID := c.GetHeader("X-User-ID")
    
    userContext := &featureflags.UserContext{
        UserID: userID,
        Attributes: map[string]interface{}{
            "service": "analytics",
        },
    }
    
    // Evaluate multiple flags for dashboard features
    flags, _ := h.featureFlags.EvaluateAllFlags(c.Request.Context(), userContext)
    
    dashboard := map[string]interface{}{
        "basic_metrics": getBasicMetrics(userID),
    }
    
    // Conditionally add features based on flags
    if flags["advanced-analytics"] != nil && flags["advanced-analytics"].Enabled {
        dashboard["advanced_metrics"] = getAdvancedMetrics(userID)
    }
    
    if flags["real-time-data"] != nil && flags["real-time-data"].Enabled {
        dashboard["real_time"] = getRealTimeData(userID)
    }
    
    if flags["predictive-analytics"] != nil && flags["predictive-analytics"].Enabled {
        dashboard["predictions"] = getPredictiveAnalytics(userID)
    }
    
    // Include flag status for frontend
    dashboard["_feature_flags"] = formatFlagsForFrontend(flags)
    
    c.JSON(200, dashboard)
}
```

### 3. Frontend Integration

#### JavaScript/React Integration

Create a feature flag service for your frontend:

```javascript
// frontend/src/services/featureFlags.js
class FeatureFlagService {
    constructor(baseURL = '/api/v1/flags') {
        this.baseURL = baseURL;
        this.cache = new Map();
        this.cacheTimeout = 5 * 60 * 1000; // 5 minutes
    }
    
    async evaluateFlag(flagId, userContext) {
        const cacheKey = `${flagId}-${JSON.stringify(userContext)}`;
        const cached = this.cache.get(cacheKey);
        
        if (cached && Date.now() - cached.timestamp < this.cacheTimeout) {
            return cached.result;
        }
        
        try {
            const response = await fetch(`${this.baseURL}/${flagId}/evaluate`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(userContext),
            });
            
            const result = await response.json();
            
            this.cache.set(cacheKey, {
                result: result.enabled,
                timestamp: Date.now(),
            });
            
            return result.enabled;
        } catch (error) {
            console.error('Error evaluating feature flag:', error);
            return false; // Default to disabled
        }
    }
    
    async evaluateAllFlags(userContext) {
        try {
            const response = await fetch(`${this.baseURL}/evaluate`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(userContext),
            });
            
            const results = await response.json();
            return results;
        } catch (error) {
            console.error('Error evaluating feature flags:', error);
            return {};
        }
    }
    
    clearCache() {
        this.cache.clear();
    }
}

export default new FeatureFlagService();
```

#### React Hook for Feature Flags

```javascript
// frontend/src/hooks/useFeatureFlags.js
import { useState, useEffect, useContext } from 'react';
import featureFlagService from '../services/featureFlags';
import { UserContext } from '../contexts/UserContext';

export function useFeatureFlag(flagId) {
    const [enabled, setEnabled] = useState(false);
    const [loading, setLoading] = useState(true);
    const { user } = useContext(UserContext);
    
    useEffect(() => {
        if (!user) return;
        
        const userContext = {
            user_id: user.id,
            email: user.email,
            attributes: {
                plan: user.plan,
                role: user.role,
            },
        };
        
        featureFlagService.evaluateFlag(flagId, userContext)
            .then(setEnabled)
            .finally(() => setLoading(false));
    }, [flagId, user]);
    
    return { enabled, loading };
}

export function useFeatureFlags() {
    const [flags, setFlags] = useState({});
    const [loading, setLoading] = useState(true);
    const { user } = useContext(UserContext);
    
    useEffect(() => {
        if (!user) return;
        
        const userContext = {
            user_id: user.id,
            email: user.email,
            attributes: {
                plan: user.plan,
                role: user.role,
            },
        };
        
        featureFlagService.evaluateAllFlags(userContext)
            .then(setFlags)
            .finally(() => setLoading(false));
    }, [user]);
    
    return { flags, loading };
}
```

#### React Component Usage

```javascript
// frontend/src/components/Dashboard.jsx
import React from 'react';
import { useFeatureFlags } from '../hooks/useFeatureFlags';

function Dashboard() {
    const { flags, loading } = useFeatureFlags();
    
    if (loading) {
        return <div>Loading...</div>;
    }
    
    return (
        <div className="dashboard">
            <h1>Dashboard</h1>
            
            {/* Always show basic widgets */}
            <UserProfileWidget />
            
            {/* Conditionally show widgets based on feature flags */}
            {flags['analytics-widget']?.enabled && (
                <AnalyticsWidget />
            )}
            
            {flags['social-widget']?.enabled && (
                <SocialWidget />
            )}
            
            {flags['beta-features']?.enabled && (
                <BetaFeaturesWidget />
            )}
            
            {/* Show different UI based on flag variant */}
            {flags['new-ui']?.enabled ? (
                <NewUIComponent />
            ) : (
                <LegacyUIComponent />
            )}
        </div>
    );
}

export default Dashboard;
```

### 4. Environment Configuration

#### Development Environment

```bash
# .env.development
ENVIRONMENT=development
FEATURE_FLAGS_ENABLED=true
FEATURE_FLAGS_STORAGE=memory
FEATURE_FLAGS_CACHE_ENABLED=true
FEATURE_FLAGS_DEBUG=true
```

#### Production Environment

```bash
# .env.production
ENVIRONMENT=production
FEATURE_FLAGS_ENABLED=true
FEATURE_FLAGS_STORAGE=database
FEATURE_FLAGS_CACHE_ENABLED=true
FEATURE_FLAGS_DEBUG=false
FEATURE_FLAGS_ANALYTICS_ENABLED=true
```

### 5. Database Migration

Run the feature flags schema:

```bash
# Apply the feature flags schema
psql -d your_database -f shared/featureflags/schema.sql
```

Or use the programmatic setup:

```go
factory := featureflags.NewFactory(config, db)
if err := factory.SetupDatabase(); err != nil {
    log.Fatal("Failed to setup feature flags database:", err)
}
```

### 6. Common Feature Flag Patterns

#### Gradual Rollout

```go
// Create a flag with 10% rollout
flag := &featureflags.FeatureFlag{
    ID:          "new-algorithm",
    Name:        "New Algorithm",
    Description: "Gradually roll out new algorithm",
    Enabled:     true,
    Rollout:     0.1, // 10% of users
    Environment: "production",
    Service:     "analytics",
}
```

#### User Targeting

```go
// Create a flag for premium users only
flag := &featureflags.FeatureFlag{
    ID:          "premium-features",
    Name:        "Premium Features",
    Description: "Features for premium users",
    Enabled:     true,
    Rollout:     1.0,
    Conditions: map[string]interface{}{
        "user_attributes": map[string]interface{}{
            "plan": "premium",
        },
    },
    Environment: "production",
    Service:     "all",
}
```

#### Kill Switch

```go
// Create a kill switch for emergency situations
flag := &featureflags.FeatureFlag{
    ID:          "new-payment-system",
    Name:        "New Payment System",
    Description: "New payment processing system",
    Enabled:     true,
    Rollout:     1.0,
    Environment: "production",
    Service:     "payment",
    Tags:        []string{"payment", "critical"},
}

// In emergency, disable the flag:
// flag.Enabled = false
// manager.UpdateFlag(ctx, flag)
```

### 7. Monitoring and Analytics

#### Track Feature Usage

```go
// Track when a feature is used
event := &featureflags.FeatureFlagEvent{
    FlagID:    "new-feature",
    UserID:    userID,
    Service:   "api",
    EventType: "exposure",
    Result:    true,
    Timestamp: time.Now(),
}

manager.TrackEvent(ctx, event)
```

#### Get Feature Metrics

```go
// Get metrics for the last 7 days
startDate := time.Now().AddDate(0, 0, -7)
endDate := time.Now()

metrics, err := manager.GetMetrics(ctx, "new-feature", startDate, endDate)
if err != nil {
    log.Printf("Error getting metrics: %v", err)
} else {
    log.Printf("Feature usage: %+v", metrics)
}
```

### 8. Best Practices

1. **Naming Convention**: Use kebab-case for flag IDs: `new-user-dashboard`
2. **Environment Separation**: Use different flags for different environments
3. **Gradual Rollouts**: Start with small percentages and increase gradually
4. **Error Handling**: Always handle flag evaluation errors gracefully
5. **Cleanup**: Remove unused flags regularly
6. **Documentation**: Document what each flag does and when it can be removed
7. **Testing**: Test both enabled and disabled states
8. **Monitoring**: Monitor flag usage and performance impact

### 9. Troubleshooting

#### Common Issues

1. **Flag not found**: Check flag ID and environment
2. **Permission denied**: Verify user context and conditions
3. **Cache issues**: Clear cache or reduce TTL
4. **Database connection**: Check database connectivity

#### Debug Mode

```go
config := featureflags.DevelopmentConfig()
config.DebugMode = true
```

#### Health Check

```bash
curl http://localhost:8080/api/v1/flags/health
```

### 10. Migration from Existing Code

To migrate existing conditional code:

#### Before (Environment Variables)

```go
if os.Getenv("ENABLE_NEW_FEATURE") == "true" {
    useNewFeature()
} else {
    useOldFeature()
}
```

#### After (Feature Flags)

```go
if enabled, _ := flagManager.IsEnabled(ctx, "new-feature", userContext); enabled {
    useNewFeature()
} else {
    useOldFeature()
}
```

This provides much more flexibility with user targeting, gradual rollouts, and real-time control.