// React Hooks for Feature Flags
// TypeScript implementation with proper error handling and caching

import React, { useState, useEffect, useContext, useCallback, useRef } from 'react';
import featureFlagService, { UserContext, FeatureFlagResult } from '../services/featureFlags';

// Context for user information (you'll need to create this in your app)
export interface User {
  id: string;
  email: string;
  plan?: string;
  role?: string;
  attributes?: Record<string, any>;
}

// Hook for a single feature flag
export function useFeatureFlag(flagId: string, userContext?: UserContext) {
  const [enabled, setEnabled] = useState<boolean>(false);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<FeatureFlagResult | null>(null);
  const mountedRef = useRef(true);

  const evaluateFlag = useCallback(async () => {
    if (!userContext) {
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError(null);
      
      const isEnabled = await featureFlagService.evaluateFlag(flagId, userContext);
      
      if (mountedRef.current) {
        setEnabled(isEnabled);
        setResult({ enabled: isEnabled });
      }
    } catch (err) {
      if (mountedRef.current) {
        const errorMessage = err instanceof Error ? err.message : 'Unknown error';
        setError(errorMessage);
        setEnabled(false); // Default to disabled on error
      }
    } finally {
      if (mountedRef.current) {
        setLoading(false);
      }
    }
  }, [flagId, userContext]);

  useEffect(() => {
    mountedRef.current = true;
    evaluateFlag();
    
    return () => {
      mountedRef.current = false;
    };
  }, [evaluateFlag]);

  const refresh = useCallback(() => {
    featureFlagService.clearCache();
    evaluateFlag();
  }, [evaluateFlag]);

  return {
    enabled,
    loading,
    error,
    result,
    refresh,
  };
}

// Hook for multiple feature flags
export function useFeatureFlags(userContext?: UserContext) {
  const [flags, setFlags] = useState<Record<string, FeatureFlagResult>>({});
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const mountedRef = useRef(true);

  const evaluateFlags = useCallback(async () => {
    if (!userContext) {
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError(null);
      
      const results = await featureFlagService.evaluateAllFlags(userContext);
      
      if (mountedRef.current) {
        setFlags(results);
      }
    } catch (err) {
      if (mountedRef.current) {
        const errorMessage = err instanceof Error ? err.message : 'Unknown error';
        setError(errorMessage);
        setFlags({}); // Default to empty on error
      }
    } finally {
      if (mountedRef.current) {
        setLoading(false);
      }
    }
  }, [userContext]);

  useEffect(() => {
    mountedRef.current = true;
    evaluateFlags();
    
    return () => {
      mountedRef.current = false;
    };
  }, [evaluateFlags]);

  const refresh = useCallback(() => {
    featureFlagService.clearCache();
    evaluateFlags();
  }, [evaluateFlags]);

  const isEnabled = useCallback((flagId: string): boolean => {
    return flags[flagId]?.enabled || false;
  }, [flags]);

  const getVariant = useCallback((flagId: string): string | undefined => {
    return flags[flagId]?.variant;
  }, [flags]);

  const getMetadata = useCallback((flagId: string): Record<string, any> | undefined => {
    return flags[flagId]?.metadata;
  }, [flags]);

  return {
    flags,
    loading,
    error,
    refresh,
    isEnabled,
    getVariant,
    getMetadata,
  };
}

// Hook for feature flag with user context from React Context
export function useFeatureFlagWithUser(flagId: string, UserContext: React.Context<{ user: User | null }>) {
  const { user } = useContext(UserContext);
  
  const userContext: UserContext | undefined = user ? {
    user_id: user.id,
    email: user.email,
    attributes: {
      plan: user.plan,
      role: user.role,
      ...user.attributes,
    },
  } : undefined;

  return useFeatureFlag(flagId, userContext);
}

// Hook for multiple feature flags with user context from React Context
export function useFeatureFlagsWithUser(UserContext: React.Context<{ user: User | null }>) {
  const { user } = useContext(UserContext);
  
  const userContext: UserContext | undefined = user ? {
    user_id: user.id,
    email: user.email,
    attributes: {
      plan: user.plan,
      role: user.role,
      ...user.attributes,
    },
  } : undefined;

  return useFeatureFlags(userContext);
}

// Hook for feature flag health monitoring
export function useFeatureFlagHealth() {
  const [isHealthy, setIsHealthy] = useState<boolean>(true);
  const [lastCheck, setLastCheck] = useState<Date | null>(null);
  const [loading, setLoading] = useState<boolean>(false);

  const checkHealth = useCallback(async () => {
    setLoading(true);
    try {
      const healthy = await featureFlagService.healthCheck();
      setIsHealthy(healthy);
      setLastCheck(new Date());
    } catch (error) {
      setIsHealthy(false);
      console.error('Feature flag health check failed:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    checkHealth();
    
    // Check health every 5 minutes
    const interval = setInterval(checkHealth, 5 * 60 * 1000);
    
    return () => clearInterval(interval);
  }, [checkHealth]);

  return {
    isHealthy,
    lastCheck,
    loading,
    checkHealth,
  };
}

// Hook for feature flag management (admin features)
export function useFeatureFlagManagement() {
  const [flags, setFlags] = useState<any[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const loadFlags = useCallback(async () => {
    setLoading(true);
    setError(null);
    
    try {
      const allFlags = await featureFlagService.getAllFlags();
      setFlags(allFlags);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load flags';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  }, []);

  const getFlag = useCallback(async (flagId: string) => {
    try {
      return await featureFlagService.getFlag(flagId);
    } catch (err) {
      console.error(`Failed to get flag ${flagId}:`, err);
      return null;
    }
  }, []);

  useEffect(() => {
    loadFlags();
  }, [loadFlags]);

  return {
    flags,
    loading,
    error,
    loadFlags,
    getFlag,
  };
}

// Utility hook for conditional rendering based on feature flags
export function useConditionalRender(flagId: string, userContext?: UserContext) {
  const { enabled, loading } = useFeatureFlag(flagId, userContext);

  const ConditionalComponent = useCallback(
    ({ children, fallback = null }: { children: React.ReactNode; fallback?: React.ReactNode }) => {
      if (loading) {
        return fallback as React.ReactElement;
      }
      
      return enabled ? (children as React.ReactElement) : (fallback as React.ReactElement);
    },
    [enabled, loading]
  );

  return {
    enabled,
    loading,
    ConditionalComponent,
  };
}

// Hook for A/B testing with variants
export function useABTest(flagId: string, userContext?: UserContext) {
  const { enabled, result, loading, error } = useFeatureFlag(flagId, userContext);
  
  const variant = result?.variant || 'control';
  const isVariant = useCallback((variantName: string) => {
    return variant === variantName;
  }, [variant]);

  return {
    enabled,
    variant,
    isVariant,
    loading,
    error,
  };
}