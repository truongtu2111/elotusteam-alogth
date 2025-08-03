// React Component Examples for Feature Flags Integration
// JavaScript implementation with comprehensive usage patterns

import React, { useState, useEffect, useContext, useCallback } from 'react';
import featureFlagService from '../services/featureFlags';

// Example User Context (you would create this in your app)
const UserContext = React.createContext({ user: null });

// Custom hook for a single feature flag
export function useFeatureFlag(flagId, userContext) {
  const [enabled, setEnabled] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const evaluateFlag = useCallback(async () => {
    if (!userContext) {
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError(null);
      
      const isEnabled = await featureFlagService.evaluateFlag(flagId, userContext);
      setEnabled(isEnabled);
    } catch (err) {
      setError(err.message);
      setEnabled(false); // Default to disabled on error
    } finally {
      setLoading(false);
    }
  }, [flagId, userContext]);

  useEffect(() => {
    evaluateFlag();
  }, [evaluateFlag]);

  const refresh = useCallback(() => {
    featureFlagService.clearCache();
    evaluateFlag();
  }, [evaluateFlag]);

  return { enabled, loading, error, refresh };
}

// Custom hook for multiple feature flags
export function useFeatureFlags(userContext) {
  const [flags, setFlags] = useState({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const evaluateFlags = useCallback(async () => {
    if (!userContext) {
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError(null);
      
      const results = await featureFlagService.evaluateAllFlags(userContext);
      setFlags(results);
    } catch (err) {
      setError(err.message);
      setFlags({});
    } finally {
      setLoading(false);
    }
  }, [userContext]);

  useEffect(() => {
    evaluateFlags();
  }, [evaluateFlags]);

  const refresh = useCallback(() => {
    featureFlagService.clearCache();
    evaluateFlags();
  }, [evaluateFlags]);

  const isEnabled = useCallback((flagId) => {
    return flags[flagId]?.enabled || false;
  }, [flags]);

  return { flags, loading, error, refresh, isEnabled };
}

// Feature Flag Provider Component
export function FeatureFlagProvider({ children }) {
  const { user } = useContext(UserContext);
  
  const userContext = user ? {
    user_id: user.id,
    email: user.email,
    attributes: {
      plan: user.plan,
      role: user.role,
      ...user.attributes,
    },
  } : null;

  const { flags, loading, error, refresh, isEnabled } = useFeatureFlags(userContext);

  const contextValue = {
    flags,
    loading,
    error,
    refresh,
    isEnabled,
    userContext,
  };

  return (
    <FeatureFlagContext.Provider value={contextValue}>
      {children}
    </FeatureFlagContext.Provider>
  );
}

const FeatureFlagContext = React.createContext({});

// Hook to use feature flag context
export function useFeatureFlagContext() {
  return useContext(FeatureFlagContext);
}

// Conditional Component for Feature Flags
export function FeatureFlag({ flagId, children, fallback = null }) {
  const { isEnabled, loading } = useFeatureFlagContext();

  if (loading) {
    return fallback;
  }

  return isEnabled(flagId) ? children : fallback;
}

// A/B Test Component
export function ABTest({ flagId, variantA, variantB, control }) {
  const { flags, loading } = useFeatureFlagContext();

  if (loading) {
    return control || null;
  }

  const flag = flags[flagId];
  if (!flag?.enabled) {
    return control || null;
  }

  const variant = flag.variant || 'control';
  
  switch (variant) {
    case 'A':
      return variantA;
    case 'B':
      return variantB;
    default:
      return control || null;
  }
}

// Example Dashboard Component
export function Dashboard() {
  const { user } = useContext(UserContext);
  const { flags, loading, error, refresh } = useFeatureFlagContext();

  if (loading) {
    return (
      <div className="dashboard-loading">
        <div className="spinner">Loading dashboard...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="dashboard-error">
        <p>Error loading feature flags: {error}</p>
        <button onClick={refresh}>Retry</button>
      </div>
    );
  }

  return (
    <div className="dashboard">
      <header className="dashboard-header">
        <h1>Welcome, {user?.name || 'User'}!</h1>
        <button onClick={refresh} className="refresh-btn">
          Refresh Features
        </button>
      </header>

      <main className="dashboard-content">
        {/* Always show basic widgets */}
        <UserProfileWidget />
        <NotificationsWidget />

        {/* Conditionally show widgets based on feature flags */}
        <FeatureFlag flagId="analytics-widget">
          <AnalyticsWidget />
        </FeatureFlag>

        <FeatureFlag flagId="social-widget">
          <SocialWidget />
        </FeatureFlag>

        <FeatureFlag flagId="beta-features">
          <BetaFeaturesWidget />
        </FeatureFlag>

        {/* A/B test for different UI layouts */}
        <ABTest
          flagId="dashboard-layout"
          variantA={<ModernLayoutWidget />}
          variantB={<CompactLayoutWidget />}
          control={<DefaultLayoutWidget />}
        />

        {/* Conditional premium features */}
        <FeatureFlag 
          flagId="premium-features"
          fallback={
            <div className="upgrade-prompt">
              <h3>Upgrade to Premium</h3>
              <p>Unlock advanced features with our premium plan!</p>
              <button className="upgrade-btn">Upgrade Now</button>
            </div>
          }
        >
          <PremiumFeaturesWidget />
        </FeatureFlag>
      </main>

      {/* Debug panel for development */}
      <FeatureFlag flagId="debug-panel">
        <DebugPanel flags={flags} />
      </FeatureFlag>
    </div>
  );
}

// Example User Profile Widget with Feature Flags
export function UserProfileWidget() {
  const { user } = useContext(UserContext);
  const { isEnabled } = useFeatureFlagContext();

  const showEnhancedProfile = isEnabled('enhanced-user-profile');
  const showSocialLinks = isEnabled('social-links');
  const showActivityFeed = isEnabled('activity-feed');

  return (
    <div className="user-profile-widget">
      <div className="profile-header">
        <img 
          src={user?.avatar || '/default-avatar.png'} 
          alt="Profile" 
          className="profile-avatar"
        />
        <div className="profile-info">
          <h2>{user?.name}</h2>
          <p>{user?.email}</p>
          {showEnhancedProfile && (
            <div className="enhanced-info">
              <p>Member since: {user?.joinDate}</p>
              <p>Plan: {user?.plan}</p>
              <div className="badges">
                {user?.badges?.map(badge => (
                  <span key={badge} className="badge">{badge}</span>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>

      {showSocialLinks && (
        <div className="social-links">
          <h3>Connect</h3>
          <div className="social-buttons">
            <button className="social-btn twitter">Twitter</button>
            <button className="social-btn linkedin">LinkedIn</button>
            <button className="social-btn github">GitHub</button>
          </div>
        </div>
      )}

      {showActivityFeed && (
        <div className="activity-feed">
          <h3>Recent Activity</h3>
          <div className="activity-list">
            {user?.recentActivity?.map((activity, index) => (
              <div key={index} className="activity-item">
                <span className="activity-time">{activity.time}</span>
                <span className="activity-text">{activity.text}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// Example Analytics Widget
export function AnalyticsWidget() {
  const { isEnabled } = useFeatureFlagContext();
  const [analytics, setAnalytics] = useState(null);
  const [loading, setLoading] = useState(true);

  const showAdvancedMetrics = isEnabled('advanced-analytics');
  const showRealTimeData = isEnabled('real-time-analytics');

  useEffect(() => {
    // Simulate loading analytics data
    setTimeout(() => {
      setAnalytics({
        pageViews: 1234,
        uniqueVisitors: 567,
        bounceRate: 0.23,
        avgSessionDuration: '2m 34s',
        conversionRate: 0.045,
      });
      setLoading(false);
    }, 1000);
  }, []);

  if (loading) {
    return (
      <div className="analytics-widget loading">
        <h3>Analytics</h3>
        <div className="loading-spinner">Loading...</div>
      </div>
    );
  }

  return (
    <div className="analytics-widget">
      <h3>Analytics Dashboard</h3>
      
      <div className="metrics-grid">
        <div className="metric">
          <span className="metric-value">{analytics.pageViews}</span>
          <span className="metric-label">Page Views</span>
        </div>
        
        <div className="metric">
          <span className="metric-value">{analytics.uniqueVisitors}</span>
          <span className="metric-label">Unique Visitors</span>
        </div>
        
        {showAdvancedMetrics && (
          <>
            <div className="metric">
              <span className="metric-value">{(analytics.bounceRate * 100).toFixed(1)}%</span>
              <span className="metric-label">Bounce Rate</span>
            </div>
            
            <div className="metric">
              <span className="metric-value">{analytics.avgSessionDuration}</span>
              <span className="metric-label">Avg Session</span>
            </div>
            
            <div className="metric">
              <span className="metric-value">{(analytics.conversionRate * 100).toFixed(1)}%</span>
              <span className="metric-label">Conversion Rate</span>
            </div>
          </>
        )}
      </div>

      {showRealTimeData && (
        <div className="real-time-section">
          <h4>Real-time Data</h4>
          <div className="real-time-metrics">
            <div className="real-time-metric">
              <span className="value">23</span>
              <span className="label">Active Users</span>
            </div>
            <div className="real-time-metric">
              <span className="value">5</span>
              <span className="label">Active Sessions</span>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// Debug Panel Component
export function DebugPanel({ flags }) {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <div className="debug-panel">
      <button 
        className="debug-toggle"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        🐛 Debug Panel {isExpanded ? '▼' : '▶'}
      </button>
      
      {isExpanded && (
        <div className="debug-content">
          <h4>Active Feature Flags</h4>
          <div className="flags-list">
            {Object.entries(flags).map(([flagId, flag]) => (
              <div key={flagId} className="flag-item">
                <span className={`flag-status ${flag.enabled ? 'enabled' : 'disabled'}`}>
                  {flag.enabled ? '✅' : '❌'}
                </span>
                <span className="flag-id">{flagId}</span>
                {flag.variant && (
                  <span className="flag-variant">({flag.variant})</span>
                )}
              </div>
            ))}
          </div>
          
          <button 
            className="clear-cache-btn"
            onClick={() => {
              featureFlagService.clearCache();
              window.location.reload();
            }}
          >
            Clear Cache & Reload
          </button>
        </div>
      )}
    </div>
  );
}

// Example App Component with Feature Flag Provider
export function App() {
  const [user, setUser] = useState(null);

  useEffect(() => {
    // Simulate user authentication
    setUser({
      id: '123',
      name: 'John Doe',
      email: 'john@example.com',
      plan: 'premium',
      role: 'user',
      avatar: '/avatars/john.jpg',
      joinDate: '2023-01-15',
      badges: ['Early Adopter', 'Power User'],
      recentActivity: [
        { time: '2h ago', text: 'Updated profile' },
        { time: '1d ago', text: 'Completed tutorial' },
        { time: '3d ago', text: 'Joined the platform' },
      ],
    });
  }, []);

  return (
    <UserContext.Provider value={{ user }}>
      <FeatureFlagProvider>
        <div className="app">
          <Dashboard />
        </div>
      </FeatureFlagProvider>
    </UserContext.Provider>
  );
}

// Placeholder components
function NotificationsWidget() {
  return <div className="widget notifications-widget">Notifications</div>;
}

function SocialWidget() {
  return <div className="widget social-widget">Social Feed</div>;
}

function BetaFeaturesWidget() {
  return <div className="widget beta-features-widget">Beta Features</div>;
}

function ModernLayoutWidget() {
  return <div className="widget modern-layout">Modern Layout</div>;
}

function CompactLayoutWidget() {
  return <div className="widget compact-layout">Compact Layout</div>;
}

function DefaultLayoutWidget() {
  return <div className="widget default-layout">Default Layout</div>;
}

function PremiumFeaturesWidget() {
  return <div className="widget premium-features">Premium Features</div>;
}

export default App;