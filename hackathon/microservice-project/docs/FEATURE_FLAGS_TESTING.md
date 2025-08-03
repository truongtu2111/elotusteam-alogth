# Feature Flags Testing Guide

This guide covers comprehensive testing strategies for the feature flag system, including unit tests, integration tests, and end-to-end testing scenarios.

## Overview

Testing feature flags requires special consideration because:

1. **State Management**: Flags can be enabled/disabled dynamically
2. **User Context**: Different users may see different flag states
3. **Rollout Percentages**: Gradual rollouts need deterministic testing
4. **Cache Behavior**: Caching can affect flag evaluation
5. **Error Handling**: Network failures and service unavailability

## Testing Strategy

### 1. Unit Testing

#### Backend Unit Tests

```go
// tests/featureflags/manager_test.go
package featureflags_test

import (
    "context"
    "testing"
    "time"
    
    "your-project/shared/featureflags"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Mock implementations
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) GetFlag(ctx context.Context, id string) (*featureflags.FeatureFlag, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*featureflags.FeatureFlag), args.Error(1)
}

func (m *MockRepository) GetAllFlags(ctx context.Context) ([]*featureflags.FeatureFlag, error) {
    args := m.Called(ctx)
    return args.Get(0).([]*featureflags.FeatureFlag), args.Error(1)
}

func (m *MockRepository) CreateFlag(ctx context.Context, flag *featureflags.FeatureFlag) error {
    args := m.Called(ctx, flag)
    return args.Error(0)
}

func (m *MockRepository) UpdateFlag(ctx context.Context, flag *featureflags.FeatureFlag) error {
    args := m.Called(ctx, flag)
    return args.Error(0)
}

func (m *MockRepository) DeleteFlag(ctx context.Context, id string) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

type MockCache struct {
    mock.Mock
}

func (m *MockCache) Get(key string) (interface{}, bool) {
    args := m.Called(key)
    return args.Get(0), args.Bool(1)
}

func (m *MockCache) Set(key string, value interface{}, ttl time.Duration) {
    m.Called(key, value, ttl)
}

func (m *MockCache) Delete(key string) {
    m.Called(key)
}

func (m *MockCache) Clear() {
    m.Called()
}

type MockAnalytics struct {
    mock.Mock
}

func (m *MockAnalytics) TrackEvent(ctx context.Context, event *featureflags.FeatureFlagEvent) error {
    args := m.Called(ctx, event)
    return args.Error(0)
}

func (m *MockAnalytics) GetMetrics(ctx context.Context, flagID string, startDate, endDate time.Time) (*featureflags.FlagMetrics, error) {
    args := m.Called(ctx, flagID, startDate, endDate)
    return args.Get(0).(*featureflags.FlagMetrics), args.Error(1)
}

// Test cases
func TestFeatureFlagManager_IsEnabled(t *testing.T) {
    tests := []struct {
        name           string
        flagID         string
        userContext    *featureflags.UserContext
        flag           *featureflags.FeatureFlag
        expectedResult bool
        expectError    bool
    }{
        {
            name:   "enabled flag returns true",
            flagID: "test-flag",
            userContext: &featureflags.UserContext{
                UserID: "user123",
            },
            flag: &featureflags.FeatureFlag{
                ID:      "test-flag",
                Enabled: true,
                Rollout: 1.0,
            },
            expectedResult: true,
            expectError:    false,
        },
        {
            name:   "disabled flag returns false",
            flagID: "test-flag",
            userContext: &featureflags.UserContext{
                UserID: "user123",
            },
            flag: &featureflags.FeatureFlag{
                ID:      "test-flag",
                Enabled: false,
                Rollout: 1.0,
            },
            expectedResult: false,
            expectError:    false,
        },
        {
            name:   "rollout percentage affects result",
            flagID: "test-flag",
            userContext: &featureflags.UserContext{
                UserID: "user123", // This should hash to a value that's excluded by 0.1 rollout
            },
            flag: &featureflags.FeatureFlag{
                ID:      "test-flag",
                Enabled: true,
                Rollout: 0.1, // 10% rollout
            },
            expectedResult: false, // Assuming user123 hashes outside the 10%
            expectError:    false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mocks
            mockRepo := new(MockRepository)
            mockCache := new(MockCache)
            mockAnalytics := new(MockAnalytics)

            // Configure mock expectations
            mockRepo.On("GetFlag", mock.Anything, tt.flagID).Return(tt.flag, nil)
            mockCache.On("Get", mock.Anything).Return(nil, false)
            mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything)
            mockAnalytics.On("TrackEvent", mock.Anything, mock.Anything).Return(nil)

            // Create manager
            config := &featureflags.FeatureFlagConfig{
                Environment: "test",
                CacheEnabled: true,
                CacheTTL: 5 * time.Minute,
            }
            
            manager := featureflags.NewFeatureFlagManager(config, mockRepo, mockCache, mockAnalytics)

            // Execute test
            result, err := manager.IsEnabled(context.Background(), tt.flagID, tt.userContext)

            // Assertions
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expectedResult, result)
            }

            // Verify mock expectations
            mockRepo.AssertExpectations(t)
            mockCache.AssertExpectations(t)
            mockAnalytics.AssertExpectations(t)
        })
    }
}

func TestFeatureFlagManager_EvaluateAllFlags(t *testing.T) {
    mockRepo := new(MockRepository)
    mockCache := new(MockCache)
    mockAnalytics := new(MockAnalytics)

    flags := []*featureflags.FeatureFlag{
        {
            ID:      "flag1",
            Enabled: true,
            Rollout: 1.0,
        },
        {
            ID:      "flag2",
            Enabled: false,
            Rollout: 1.0,
        },
    }

    mockRepo.On("GetAllFlags", mock.Anything).Return(flags, nil)
    mockCache.On("Get", mock.Anything).Return(nil, false)
    mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything)
    mockAnalytics.On("TrackEvent", mock.Anything, mock.Anything).Return(nil)

    config := &featureflags.FeatureFlagConfig{
        Environment: "test",
    }
    
    manager := featureflags.NewFeatureFlagManager(config, mockRepo, mockCache, mockAnalytics)

    userContext := &featureflags.UserContext{
        UserID: "user123",
    }

    results, err := manager.EvaluateAllFlags(context.Background(), userContext)

    assert.NoError(t, err)
    assert.Len(t, results, 2)
    assert.True(t, results["flag1"].Enabled)
    assert.False(t, results["flag2"].Enabled)

    mockRepo.AssertExpectations(t)
}

func TestFeatureFlagEvaluator_RolloutPercentage(t *testing.T) {
    evaluator := featureflags.NewFeatureFlagEvaluator()

    tests := []struct {
        name        string
        userID      string
        rollout     float64
        expected    bool
    }{
        {
            name:     "user in 50% rollout",
            userID:   "user1", // Assuming this hashes to < 0.5
            rollout:  0.5,
            expected: true,
        },
        {
            name:     "user not in 10% rollout",
            userID:   "user999", // Assuming this hashes to > 0.1
            rollout:  0.1,
            expected: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            flag := &featureflags.FeatureFlag{
                ID:      "test-flag",
                Enabled: true,
                Rollout: tt.rollout,
            }

            userContext := &featureflags.UserContext{
                UserID: tt.userID,
            }

            result := evaluator.Evaluate(flag, userContext)
            assert.Equal(t, tt.expected, result.Enabled)
        })
    }
}
```

#### Frontend Unit Tests

```javascript
// frontend/src/services/__tests__/featureFlags.test.js
import featureFlagService from '../featureFlags';

// Mock fetch
global.fetch = jest.fn();

describe('FeatureFlagService', () => {
  beforeEach(() => {
    fetch.mockClear();
    featureFlagService.clearCache();
  });

  describe('evaluateFlag', () => {
    it('should return true for enabled flag', async () => {
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ enabled: true }),
      });

      const userContext = { user_id: 'user123' };
      const result = await featureFlagService.evaluateFlag('test-flag', userContext);

      expect(result).toBe(true);
      expect(fetch).toHaveBeenCalledWith(
        '/api/v1/flags/test-flag/evaluate',
        expect.objectContaining({
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(userContext),
        })
      );
    });

    it('should return false for disabled flag', async () => {
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ enabled: false }),
      });

      const result = await featureFlagService.evaluateFlag('test-flag', { user_id: 'user123' });
      expect(result).toBe(false);
    });

    it('should return false on network error', async () => {
      fetch.mockRejectedValueOnce(new Error('Network error'));

      const result = await featureFlagService.evaluateFlag('test-flag', { user_id: 'user123' });
      expect(result).toBe(false);
    });

    it('should use cache for repeated requests', async () => {
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ enabled: true }),
      });

      const userContext = { user_id: 'user123' };
      
      // First call
      await featureFlagService.evaluateFlag('test-flag', userContext);
      
      // Second call should use cache
      const result = await featureFlagService.evaluateFlag('test-flag', userContext);
      
      expect(result).toBe(true);
      expect(fetch).toHaveBeenCalledTimes(1);
    });
  });

  describe('evaluateAllFlags', () => {
    it('should return all flag results', async () => {
      const mockResults = {
        'flag1': { enabled: true },
        'flag2': { enabled: false },
      };

      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResults,
      });

      const result = await featureFlagService.evaluateAllFlags({ user_id: 'user123' });
      expect(result).toEqual(mockResults);
    });
  });

  describe('healthCheck', () => {
    it('should return true when service is healthy', async () => {
      fetch.mockResolvedValueOnce({ ok: true });

      const result = await featureFlagService.healthCheck();
      expect(result).toBe(true);
    });

    it('should return false when service is unhealthy', async () => {
      fetch.mockResolvedValueOnce({ ok: false });

      const result = await featureFlagService.healthCheck();
      expect(result).toBe(false);
    });
  });
});
```

```javascript
// frontend/src/hooks/__tests__/useFeatureFlags.test.js
import { renderHook, act } from '@testing-library/react-hooks';
import { useFeatureFlag, useFeatureFlags } from '../useFeatureFlags';
import featureFlagService from '../../services/featureFlags';

// Mock the service
jest.mock('../../services/featureFlags');

describe('useFeatureFlag', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should return enabled state', async () => {
    featureFlagService.evaluateFlag.mockResolvedValue(true);

    const { result, waitForNextUpdate } = renderHook(() =>
      useFeatureFlag('test-flag', { user_id: 'user123' })
    );

    expect(result.current.loading).toBe(true);
    expect(result.current.enabled).toBe(false);

    await waitForNextUpdate();

    expect(result.current.loading).toBe(false);
    expect(result.current.enabled).toBe(true);
    expect(result.current.error).toBe(null);
  });

  it('should handle errors gracefully', async () => {
    featureFlagService.evaluateFlag.mockRejectedValue(new Error('Service error'));

    const { result, waitForNextUpdate } = renderHook(() =>
      useFeatureFlag('test-flag', { user_id: 'user123' })
    );

    await waitForNextUpdate();

    expect(result.current.loading).toBe(false);
    expect(result.current.enabled).toBe(false);
    expect(result.current.error).toBe('Service error');
  });

  it('should refresh flag evaluation', async () => {
    featureFlagService.evaluateFlag.mockResolvedValue(true);
    featureFlagService.clearCache = jest.fn();

    const { result, waitForNextUpdate } = renderHook(() =>
      useFeatureFlag('test-flag', { user_id: 'user123' })
    );

    await waitForNextUpdate();

    act(() => {
      result.current.refresh();
    });

    expect(featureFlagService.clearCache).toHaveBeenCalled();
    expect(featureFlagService.evaluateFlag).toHaveBeenCalledTimes(2);
  });
});
```

### 2. Integration Testing

#### Backend Integration Tests

```go
// tests/integration/featureflags_test.go
package integration_test

import (
    "bytes"
    "context"
    "database/sql"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    
    "your-project/shared/featureflags"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    _ "github.com/lib/pq"
)

type FeatureFlagIntegrationSuite struct {
    suite.Suite
    db      *sql.DB
    router  *gin.Engine
    manager featureflags.FeatureFlagManager
}

func (suite *FeatureFlagIntegrationSuite) SetupSuite() {
    // Setup test database
    db, err := sql.Open("postgres", "postgres://test:test@localhost/test_db?sslmode=disable")
    suite.Require().NoError(err)
    
    suite.db = db
    
    // Setup feature flags
    config := featureflags.TestConfig()
    factory := featureflags.NewFactory(config, db)
    
    suite.manager = factory.CreateManager()
    
    // Setup Gin router
    gin.SetMode(gin.TestMode)
    suite.router = gin.New()
    
    // Add middleware
    middleware := factory.CreateMiddleware()
    suite.router.Use(middleware.Handler())
    
    // Add routes
    handler := featureflags.NewFeatureFlagHandler(suite.manager)
    flagsAPI := suite.router.Group("/api/v1/flags")
    {
        flagsAPI.GET("", handler.GetFlags)
        flagsAPI.GET("/:id", handler.GetFlag)
        flagsAPI.POST("", handler.CreateFlag)
        flagsAPI.PUT("/:id", handler.UpdateFlag)
        flagsAPI.DELETE("/:id", handler.DeleteFlag)
        flagsAPI.POST("/:id/evaluate", handler.EvaluateFlag)
        flagsAPI.POST("/evaluate", handler.EvaluateFlags)
    }
}

func (suite *FeatureFlagIntegrationSuite) TearDownSuite() {
    suite.db.Close()
}

func (suite *FeatureFlagIntegrationSuite) SetupTest() {
    // Clean up database before each test
    _, err := suite.db.Exec("TRUNCATE TABLE feature_flags, feature_flag_events CASCADE")
    suite.Require().NoError(err)
}

func (suite *FeatureFlagIntegrationSuite) TestCreateAndEvaluateFlag() {
    // Create a flag
    flag := &featureflags.FeatureFlag{
        ID:          "integration-test-flag",
        Name:        "Integration Test Flag",
        Description: "A flag for integration testing",
        Enabled:     true,
        Rollout:     1.0,
        Environment: "test",
        Service:     "test-service",
    }
    
    flagJSON, _ := json.Marshal(flag)
    
    // Create flag via API
    req := httptest.NewRequest("POST", "/api/v1/flags", bytes.NewBuffer(flagJSON))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    
    suite.router.ServeHTTP(w, req)
    suite.Equal(http.StatusCreated, w.Code)
    
    // Evaluate flag via API
    userContext := &featureflags.UserContext{
        UserID: "test-user",
        Attributes: map[string]interface{}{
            "plan": "premium",
        },
    }
    
    userJSON, _ := json.Marshal(userContext)
    
    req = httptest.NewRequest("POST", "/api/v1/flags/integration-test-flag/evaluate", bytes.NewBuffer(userJSON))
    req.Header.Set("Content-Type", "application/json")
    w = httptest.NewRecorder()
    
    suite.router.ServeHTTP(w, req)
    suite.Equal(http.StatusOK, w.Code)
    
    var result featureflags.EvaluationResult
    err := json.Unmarshal(w.Body.Bytes(), &result)
    suite.NoError(err)
    suite.True(result.Enabled)
}

func (suite *FeatureFlagIntegrationSuite) TestRolloutPercentage() {
    // Create a flag with 50% rollout
    flag := &featureflags.FeatureFlag{
        ID:          "rollout-test-flag",
        Name:        "Rollout Test Flag",
        Enabled:     true,
        Rollout:     0.5, // 50% rollout
        Environment: "test",
        Service:     "test-service",
    }
    
    err := suite.manager.CreateFlag(context.Background(), flag)
    suite.NoError(err)
    
    // Test multiple users to verify rollout percentage
    enabledCount := 0
    totalUsers := 100
    
    for i := 0; i < totalUsers; i++ {
        userContext := &featureflags.UserContext{
            UserID: fmt.Sprintf("user-%d", i),
        }
        
        enabled, err := suite.manager.IsEnabled(context.Background(), "rollout-test-flag", userContext)
        suite.NoError(err)
        
        if enabled {
            enabledCount++
        }
    }
    
    // Allow for some variance in the rollout percentage
    suite.InDelta(50, enabledCount, 10, "Rollout percentage should be approximately 50%")
}

func (suite *FeatureFlagIntegrationSuite) TestCacheInvalidation() {
    // Create a flag
    flag := &featureflags.FeatureFlag{
        ID:          "cache-test-flag",
        Name:        "Cache Test Flag",
        Enabled:     true,
        Rollout:     1.0,
        Environment: "test",
        Service:     "test-service",
    }
    
    err := suite.manager.CreateFlag(context.Background(), flag)
    suite.NoError(err)
    
    userContext := &featureflags.UserContext{
        UserID: "test-user",
    }
    
    // First evaluation should cache the result
    enabled, err := suite.manager.IsEnabled(context.Background(), "cache-test-flag", userContext)
    suite.NoError(err)
    suite.True(enabled)
    
    // Update the flag to disabled
    flag.Enabled = false
    err = suite.manager.UpdateFlag(context.Background(), flag)
    suite.NoError(err)
    
    // Evaluation should now return false (cache should be invalidated)
    enabled, err = suite.manager.IsEnabled(context.Background(), "cache-test-flag", userContext)
    suite.NoError(err)
    suite.False(enabled)
}

func TestFeatureFlagIntegrationSuite(t *testing.T) {
    suite.Run(t, new(FeatureFlagIntegrationSuite))
}
```

### 3. End-to-End Testing

#### Cypress E2E Tests

```javascript
// cypress/integration/feature-flags.spec.js
describe('Feature Flags E2E', () => {
  beforeEach(() => {
    // Setup test data
    cy.task('db:seed');
    cy.visit('/dashboard');
  });

  it('should show different UI based on feature flags', () => {
    // Login as a user with specific flags enabled
    cy.login('premium-user@example.com');
    
    // Check that premium features are visible
    cy.get('[data-testid="premium-features"]').should('be.visible');
    cy.get('[data-testid="analytics-widget"]').should('be.visible');
    
    // Login as a basic user
    cy.login('basic-user@example.com');
    
    // Check that premium features are not visible
    cy.get('[data-testid="premium-features"]').should('not.exist');
    cy.get('[data-testid="upgrade-prompt"]').should('be.visible');
  });

  it('should handle A/B test variants', () => {
    // Setup A/B test flag
    cy.task('flags:create', {
      id: 'dashboard-layout',
      enabled: true,
      variants: ['A', 'B'],
    });
    
    // Test variant A
    cy.task('flags:setUserVariant', {
      userId: 'test-user-1',
      flagId: 'dashboard-layout',
      variant: 'A',
    });
    
    cy.loginAs('test-user-1');
    cy.get('[data-testid="modern-layout"]').should('be.visible');
    
    // Test variant B
    cy.task('flags:setUserVariant', {
      userId: 'test-user-2',
      flagId: 'dashboard-layout',
      variant: 'B',
    });
    
    cy.loginAs('test-user-2');
    cy.get('[data-testid="compact-layout"]').should('be.visible');
  });

  it('should handle flag updates in real-time', () => {
    cy.login('test-user@example.com');
    
    // Initially, beta features should not be visible
    cy.get('[data-testid="beta-features"]').should('not.exist');
    
    // Enable beta features flag
    cy.task('flags:update', {
      id: 'beta-features',
      enabled: true,
    });
    
    // Refresh feature flags
    cy.get('[data-testid="refresh-features"]').click();
    
    // Beta features should now be visible
    cy.get('[data-testid="beta-features"]').should('be.visible');
  });

  it('should gracefully handle feature flag service outage', () => {
    // Simulate service outage
    cy.intercept('POST', '/api/v1/flags/evaluate', {
      statusCode: 500,
      body: { error: 'Service unavailable' },
    }).as('flagsError');
    
    cy.login('test-user@example.com');
    
    // App should still load with default behavior
    cy.get('[data-testid="dashboard"]').should('be.visible');
    
    // Premium features should not be visible (default to disabled)
    cy.get('[data-testid="premium-features"]').should('not.exist');
    
    // Error should be handled gracefully (no error messages shown to user)
    cy.get('[data-testid="error-message"]').should('not.exist');
  });
});
```

#### Cypress Commands

```javascript
// cypress/support/commands.js
Cypress.Commands.add('login', (email) => {
  cy.request({
    method: 'POST',
    url: '/api/auth/login',
    body: { email, password: 'test-password' },
  }).then((response) => {
    window.localStorage.setItem('authToken', response.body.token);
  });
});

Cypress.Commands.add('loginAs', (userId) => {
  cy.request({
    method: 'POST',
    url: '/api/auth/test-login',
    body: { userId },
  }).then((response) => {
    window.localStorage.setItem('authToken', response.body.token);
  });
});
```

#### Cypress Tasks

```javascript
// cypress/plugins/index.js
module.exports = (on, config) => {
  on('task', {
    'db:seed': () => {
      // Seed test database with feature flags
      return seedDatabase();
    },
    
    'flags:create': (flag) => {
      // Create a feature flag for testing
      return createTestFlag(flag);
    },
    
    'flags:update': (flag) => {
      // Update a feature flag
      return updateTestFlag(flag);
    },
    
    'flags:setUserVariant': ({ userId, flagId, variant }) => {
      // Set a specific variant for a user
      return setUserVariant(userId, flagId, variant);
    },
  });
};

async function seedDatabase() {
  // Implementation to seed test database
  const flags = [
    {
      id: 'premium-features',
      enabled: true,
      conditions: {
        user_attributes: { plan: 'premium' },
      },
    },
    {
      id: 'beta-features',
      enabled: false,
    },
    {
      id: 'analytics-widget',
      enabled: true,
      rollout: 0.5,
    },
  ];
  
  // Insert flags into database
  for (const flag of flags) {
    await insertFlag(flag);
  }
  
  return null;
}
```

### 4. Performance Testing

#### Load Testing with Artillery

```yaml
# artillery/feature-flags-load-test.yml
config:
  target: 'http://localhost:8080'
  phases:
    - duration: 60
      arrivalRate: 10
    - duration: 120
      arrivalRate: 50
    - duration: 60
      arrivalRate: 100
  variables:
    userIds:
      - "user1"
      - "user2"
      - "user3"
      - "user4"
      - "user5"
    flagIds:
      - "feature-a"
      - "feature-b"
      - "feature-c"

scenarios:
  - name: "Evaluate single flag"
    weight: 60
    flow:
      - post:
          url: "/api/v1/flags/{{ $randomString() }}/evaluate"
          json:
            user_id: "{{ $randomString() }}"
            attributes:
              plan: "premium"
          capture:
            - json: "$.enabled"
              as: "flagEnabled"
      - think: 1

  - name: "Evaluate all flags"
    weight: 30
    flow:
      - post:
          url: "/api/v1/flags/evaluate"
          json:
            user_id: "{{ $randomString() }}"
            attributes:
              plan: "basic"
          capture:
            - json: "$"
              as: "allFlags"
      - think: 2

  - name: "Get flag details"
    weight: 10
    flow:
      - get:
          url: "/api/v1/flags/{{ flagIds }}"
      - think: 0.5
```

#### Performance Benchmarks

```go
// tests/benchmark/featureflags_bench_test.go
package benchmark_test

import (
    "context"
    "testing"
    
    "your-project/shared/featureflags"
)

func BenchmarkFeatureFlagEvaluation(b *testing.B) {
    // Setup
    config := featureflags.TestConfig()
    manager := setupTestManager(config)
    
    userContext := &featureflags.UserContext{
        UserID: "benchmark-user",
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := manager.IsEnabled(context.Background(), "test-flag", userContext)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkFeatureFlagEvaluationWithCache(b *testing.B) {
    // Setup with cache enabled
    config := featureflags.TestConfig()
    config.CacheEnabled = true
    manager := setupTestManager(config)
    
    userContext := &featureflags.UserContext{
        UserID: "benchmark-user",
    }
    
    // Warm up cache
    manager.IsEnabled(context.Background(), "test-flag", userContext)
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := manager.IsEnabled(context.Background(), "test-flag", userContext)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkEvaluateAllFlags(b *testing.B) {
    config := featureflags.TestConfig()
    manager := setupTestManager(config)
    
    userContext := &featureflags.UserContext{
        UserID: "benchmark-user",
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := manager.EvaluateAllFlags(context.Background(), userContext)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### 5. Test Data Management

#### Test Fixtures

```go
// tests/fixtures/feature_flags.go
package fixtures

import (
    "time"
    "your-project/shared/featureflags"
)

func BasicEnabledFlag() *featureflags.FeatureFlag {
    return &featureflags.FeatureFlag{
        ID:          "basic-enabled-flag",
        Name:        "Basic Enabled Flag",
        Description: "A basic flag that is enabled",
        Enabled:     true,
        Rollout:     1.0,
        Environment: "test",
        Service:     "test-service",
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
}

func RolloutFlag(percentage float64) *featureflags.FeatureFlag {
    return &featureflags.FeatureFlag{
        ID:          "rollout-flag",
        Name:        "Rollout Flag",
        Description: "A flag with percentage rollout",
        Enabled:     true,
        Rollout:     percentage,
        Environment: "test",
        Service:     "test-service",
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
}

func ConditionalFlag() *featureflags.FeatureFlag {
    return &featureflags.FeatureFlag{
        ID:          "conditional-flag",
        Name:        "Conditional Flag",
        Description: "A flag with user conditions",
        Enabled:     true,
        Rollout:     1.0,
        Conditions: map[string]interface{}{
            "user_attributes": map[string]interface{}{
                "plan": "premium",
            },
        },
        Environment: "test",
        Service:     "test-service",
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
}

func PremiumUserContext() *featureflags.UserContext {
    return &featureflags.UserContext{
        UserID: "premium-user",
        Email:  "premium@example.com",
        Attributes: map[string]interface{}{
            "plan": "premium",
            "role": "user",
        },
    }
}

func BasicUserContext() *featureflags.UserContext {
    return &featureflags.UserContext{
        UserID: "basic-user",
        Email:  "basic@example.com",
        Attributes: map[string]interface{}{
            "plan": "basic",
            "role": "user",
        },
    }
}
```

### 6. Testing Best Practices

#### Test Organization

1. **Separate test types**: Unit, integration, and E2E tests in different directories
2. **Use test fixtures**: Create reusable test data
3. **Mock external dependencies**: Database, cache, analytics
4. **Test error scenarios**: Network failures, service outages
5. **Performance testing**: Benchmark critical paths

#### Test Coverage

1. **Flag evaluation logic**: All rollout percentages and conditions
2. **Cache behavior**: Hit, miss, invalidation
3. **Error handling**: Graceful degradation
4. **User targeting**: Different user attributes and conditions
5. **API endpoints**: All CRUD operations

#### Continuous Integration

```yaml
# .github/workflows/feature-flags-test.yml
name: Feature Flags Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:13
        env:
          POSTGRES_PASSWORD: test
          POSTGRES_DB: test_db
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
    - uses: actions/checkout@v2
    
    - name: Set up Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.19
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run unit tests
      run: go test -v ./shared/featureflags/...
    
    - name: Run integration tests
      run: go test -v ./tests/integration/...
      env:
        DATABASE_URL: postgres://postgres:test@localhost/test_db?sslmode=disable
    
    - name: Run benchmarks
      run: go test -bench=. ./tests/benchmark/...
    
    - name: Generate coverage report
      run: |
        go test -coverprofile=coverage.out ./shared/featureflags/...
        go tool cover -html=coverage.out -o coverage.html
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v1
      with:
        file: ./coverage.out

  e2e:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v2
    
    - name: Set up Node.js
      uses: actions/setup-node@v2
      with:
        node-version: '16'
    
    - name: Install dependencies
      run: npm ci
      working-directory: ./frontend
    
    - name: Start application
      run: |
        docker-compose up -d
        sleep 30  # Wait for services to start
    
    - name: Run Cypress tests
      run: npm run cypress:run
      working-directory: ./frontend
    
    - name: Upload test artifacts
      uses: actions/upload-artifact@v2
      if: failure()
      with:
        name: cypress-screenshots
        path: frontend/cypress/screenshots
```

This comprehensive testing guide ensures that your feature flag system is robust, reliable, and performs well under various conditions.