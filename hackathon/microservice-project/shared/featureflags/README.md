Get # Feature Flags System

A comprehensive feature flag system for microservices that supports backend and frontend integration with real-time evaluation, analytics, and management capabilities.

## Features

- **Real-time Flag Evaluation**: Instant flag evaluation with user context
- **Multiple Storage Backends**: In-memory, PostgreSQL, Redis support
- **Caching Layer**: Configurable caching with TTL support
- **Analytics & Metrics**: Track flag usage and performance
- **User Targeting**: Advanced user segmentation and targeting
- **Rollout Control**: Gradual rollouts with percentage-based targeting
- **Environment Support**: Multi-environment flag management
- **HTTP/Gin Middleware**: Easy integration with web services
- **RESTful API**: Complete CRUD operations for flag management

## Quick Start

### 1. Basic Setup

```go
package main

import (
    "database/sql"
    "log"
    
    "your-project/shared/featureflags"
    _ "github.com/lib/pq"
)

func main() {
    // Connect to database (optional)
    db, err := sql.Open("postgres", "your-connection-string")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Create feature flag system
    qs, err := featureflags.NewQuickStart(db, "development")
    if err != nil {
        log.Fatal(err)
    }
    defer qs.Stop()
    
    // Create sample flags
    if err := qs.CreateSampleFlags(); err != nil {
        log.Printf("Warning: %v", err)
    }
    
    // Use the manager
    userContext := &featureflags.UserContext{
        UserID: "user123",
        Email:  "user@example.com",
        Attributes: map[string]interface{}{
            "plan": "premium",
        },
    }
    
    enabled, err := qs.Manager.IsEnabled(nil, "new-ui", userContext)
    if err != nil {
        log.Printf("Error: %v", err)
    } else {
        log.Printf("New UI enabled: %v", enabled)
    }
}
```

### 2. Gin Middleware Integration

```go
package main

import (
    "github.com/gin-gonic/gin"
    "your-project/shared/featureflags"
)

func main() {
    // Setup feature flags
    qs, err := featureflags.NewQuickStart(nil, "development")
    if err != nil {
        panic(err)
    }
    defer qs.Stop()
    
    // Setup Gin router
    r := gin.Default()
    
    // Add feature flag middleware
    r.Use(qs.Middleware.GinMiddleware())
    
    // Add feature flag management routes
    api := r.Group("/api/v1")
    qs.Handler.RegisterRoutes(api)
    
    // Your application routes
    r.GET("/", func(c *gin.Context) {
        // Feature flags are available in context
        if flags, exists := c.Get("feature_flags"); exists {
            c.JSON(200, gin.H{
                "message": "Hello World",
                "flags":   flags,
            })
        } else {
            c.JSON(200, gin.H{"message": "Hello World"})
        }
    })
    
    r.Run(":8080")
}
```

## Configuration

### Environment-based Configuration

```go
// Development configuration
config := featureflags.DevelopmentConfig()

// Production configuration
config := featureflags.ProductionConfig()

// Test configuration
config := featureflags.TestConfig()

// Custom configuration
config := &featureflags.FeatureFlagConfig{
    Enabled:          true,
    StorageType:      "database", // "memory", "database"
    CacheEnabled:     true,
    CacheTTL:         5 * time.Minute,
    RefreshInterval:  30 * time.Second,
    AnalyticsEnabled: true,
    Environment:      "production",
    Service:          "api",
    MetricsEnabled:   true,
    DebugMode:        false,
}
```

## Creating Feature Flags

### 1. Simple Boolean Flags

```go
// Create a simple on/off flag
flag := &featureflags.FeatureFlag{
    ID:          "new-feature",
    Name:        "New Feature",
    Description: "Enable the new feature",
    Enabled:     true,
    Rollout:     1.0, // 100% rollout
    Environment: "production",
    Service:     "api",
    CreatedBy:   "admin",
}

err := manager.CreateFlag(ctx, flag)
```

### 2. Gradual Rollout Flags

```go
// Create a flag with gradual rollout
flag := &featureflags.FeatureFlag{
    ID:          "beta-feature",
    Name:        "Beta Feature",
    Description: "Gradually roll out beta feature",
    Enabled:     true,
    Rollout:     0.1, // 10% of users
    Environment: "production",
    Service:     "api",
    CreatedBy:   "admin",
}

err := manager.CreateFlag(ctx, flag)
```

### 3. User Targeting Flags

```go
// Create a flag with user targeting
flag := &featureflags.FeatureFlag{
    ID:          "premium-feature",
    Name:        "Premium Feature",
    Description: "Feature for premium users only",
    Enabled:     true,
    Rollout:     1.0,
    Environment: "production",
    Service:     "api",
    Conditions: map[string]interface{}{
        "user_attributes": map[string]interface{}{
            "plan": "premium",
        },
    },
    CreatedBy: "admin",
}

err := manager.CreateFlag(ctx, flag)
```

### 4. Configuration Flags

```go
// Create a flag that returns configuration values
flag := &featureflags.FeatureFlag{
    ID:          "api-timeout",
    Name:        "API Timeout",
    Description: "Configure API timeout value",
    Enabled:     true,
    Rollout:     1.0,
    Environment: "production",
    Service:     "api",
    Metadata: map[string]interface{}{
        "value": 30, // 30 seconds
    },
    CreatedBy: "admin",
}

err := manager.CreateFlag(ctx, flag)
```

## Evaluating Feature Flags

### 1. Simple Boolean Check

```go
userContext := &featureflags.UserContext{
    UserID: "user123",
    Email:  "user@example.com",
    Attributes: map[string]interface{}{
        "plan": "premium",
    },
}

enabled, err := manager.IsEnabled(ctx, "new-feature", userContext)
if err != nil {
    log.Printf("Error checking flag: %v", err)
    // Use default behavior
} else if enabled {
    // New feature is enabled for this user
    useNewFeature()
} else {
    // Use old behavior
    useOldFeature()
}
```

### 2. Get Configuration Values

```go
value, err := manager.GetValue(ctx, "api-timeout", userContext, 15)
if err != nil {
    log.Printf("Error getting flag value: %v", err)
    timeout = 15 // default value
} else {
    timeout = value.(int)
}
```

### 3. Detailed Evaluation

```go
result, err := manager.EvaluateFlag(ctx, "new-feature", userContext)
if err != nil {
    log.Printf("Error evaluating flag: %v", err)
} else {
    log.Printf("Flag: %s, Enabled: %v, Reason: %s", 
        result.FlagID, result.Enabled, result.Reason)
}
```

### 4. Evaluate All Flags

```go
results, err := manager.EvaluateAllFlags(ctx, userContext)
if err != nil {
    log.Printf("Error evaluating flags: %v", err)
} else {
    for flagID, result := range results {
        log.Printf("Flag %s: %v", flagID, result.Enabled)
    }
}
```

## Frontend Integration

### 1. REST API Endpoints

The system provides RESTful endpoints for frontend integration:

```
GET    /flags                    # Get all flags
GET    /flags/:id               # Get specific flag
POST   /flags                   # Create new flag
PUT    /flags/:id               # Update flag
DELETE /flags/:id               # Delete flag
POST   /flags/:id/evaluate      # Evaluate flag for user
POST   /flags/evaluate          # Evaluate all flags for user
GET    /flags/:id/metrics       # Get flag metrics
GET    /health                  # Health check
```

### 2. JavaScript/Frontend Usage

```javascript
// Evaluate a single flag
async function checkFeatureFlag(flagId, userContext) {
    try {
        const response = await fetch(`/api/v1/flags/${flagId}/evaluate`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                user_id: userContext.userId,
                email: userContext.email,
                attributes: userContext.attributes,
            }),
        });
        
        const result = await response.json();
        return result.enabled;
    } catch (error) {
        console.error('Error checking feature flag:', error);
        return false; // Default to disabled
    }
}

// Evaluate all flags
async function getAllFeatureFlags(userContext) {
    try {
        const response = await fetch('/api/v1/flags/evaluate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(userContext),
        });
        
        const results = await response.json();
        return results;
    } catch (error) {
        console.error('Error getting feature flags:', error);
        return {};
    }
}

// Usage in React component
function MyComponent() {
    const [featureFlags, setFeatureFlags] = useState({});
    
    useEffect(() => {
        const userContext = {
            userId: 'user123',
            email: 'user@example.com',
            attributes: {
                plan: 'premium',
            },
        };
        
        getAllFeatureFlags(userContext).then(setFeatureFlags);
    }, []);
    
    return (
        <div>
            {featureFlags['new-ui']?.enabled && (
                <NewUIComponent />
            )}
            {!featureFlags['new-ui']?.enabled && (
                <OldUIComponent />
            )}
        </div>
    );
}
```

## Database Setup

If using PostgreSQL storage, run the schema file:

```sql
-- Run the schema.sql file to create required tables
\i shared/featureflags/schema.sql
```

Or use the factory setup method:

```go
factory := featureflags.NewFactory(config, db)
err := factory.SetupDatabase()
```

## Analytics and Metrics

### Track Events

```go
// Track flag evaluation
event := &featureflags.FeatureFlagEvent{
    FlagID:    "new-feature",
    UserID:    "user123",
    Service:   "api",
    EventType: "evaluation",
    Result:    true,
    Timestamp: time.Now(),
}

err := manager.TrackEvent(ctx, event)
```

### Get Metrics

```go
startDate := time.Now().AddDate(0, 0, -7) // Last 7 days
endDate := time.Now()

metrics, err := manager.GetMetrics(ctx, "new-feature", startDate, endDate)
if err != nil {
    log.Printf("Error getting metrics: %v", err)
} else {
    log.Printf("Metrics: %+v", metrics)
}
```

## Best Practices

### 1. Flag Naming
- Use kebab-case for flag IDs: `new-user-dashboard`
- Use descriptive names: `Enable New User Dashboard`
- Include the feature area: `user-dashboard-redesign`

### 2. Environment Management
- Use different flag configurations per environment
- Test flags in development before production
- Use gradual rollouts in production

### 3. User Context
- Always provide user context for evaluation
- Include relevant user attributes for targeting
- Consider privacy and data protection

### 4. Error Handling
- Always handle flag evaluation errors gracefully
- Provide sensible defaults when flags fail
- Log errors for monitoring

### 5. Performance
- Use caching to reduce database load
- Evaluate flags once per request when possible
- Monitor flag evaluation performance

### 6. Cleanup
- Remove unused flags regularly
- Set expiration dates for temporary flags
- Archive old flags instead of deleting

## Troubleshooting

### Common Issues

1. **Flag not found**: Ensure the flag exists and is in the correct environment
2. **Permission denied**: Check user context and flag conditions
3. **Cache issues**: Clear cache or reduce TTL for testing
4. **Database connection**: Verify database connectivity and schema

### Debug Mode

Enable debug mode for detailed logging:

```go
config := featureflags.DevelopmentConfig()
config.DebugMode = true
```

### Health Check

```go
err := manager.HealthCheck(ctx)
if err != nil {
    log.Printf("Feature flag system unhealthy: %v", err)
}
```

## API Reference

For detailed API documentation, see the interface definitions in `config.go`:

- `FeatureFlagManager`: Main management interface
- `FeatureFlagRepository`: Storage interface
- `FeatureFlagCache`: Caching interface
- `FeatureFlagAnalytics`: Analytics interface
- `FeatureFlagEvaluator`: Evaluation logic interface
- `FeatureFlagMiddleware`: HTTP middleware interface

## Contributing

When adding new features:

1. Update the relevant interfaces
2. Implement in all storage backends
3. Add tests
4. Update documentation
5. Consider backward compatibility

## License

This feature flag system is part of the microservice project.