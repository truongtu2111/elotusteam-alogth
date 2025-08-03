// Feature Flags Service for Frontend Integration
// TypeScript implementation with React support

export interface UserContext {
  user_id?: string;
  email?: string;
  ip_address?: string;
  user_agent?: string;
  attributes?: Record<string, any>;
}

export interface FeatureFlagResult {
  enabled: boolean;
  variant?: string;
  metadata?: Record<string, any>;
}

export interface FeatureFlag {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  rollout: number;
  environment: string;
  service: string;
  created_at: string;
  updated_at: string;
}

class FeatureFlagService {
  private baseURL: string;
  private cache: Map<string, { result: FeatureFlagResult; timestamp: number }>;
  private cacheTimeout: number;
  private defaultTimeout: number;

  constructor(baseURL = '/api/v1/flags') {
    this.baseURL = baseURL;
    this.cache = new Map();
    this.cacheTimeout = 5 * 60 * 1000; // 5 minutes
    this.defaultTimeout = 3000; // 3 seconds
  }

  /**
   * Evaluate a single feature flag
   */
  async evaluateFlag(
    flagId: string,
    userContext: UserContext
  ): Promise<boolean> {
    const cacheKey = `${flagId}-${JSON.stringify(userContext)}`;
    const cached = this.cache.get(cacheKey);

    if (cached && Date.now() - cached.timestamp < this.cacheTimeout) {
      return cached.result.enabled;
    }

    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), this.defaultTimeout);

      const response = await fetch(`${this.baseURL}/${flagId}/evaluate`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(userContext),
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result: FeatureFlagResult = await response.json();

      this.cache.set(cacheKey, {
        result,
        timestamp: Date.now(),
      });

      return result.enabled;
    } catch (error) {
      console.error(`Error evaluating feature flag '${flagId}':`, error);
      return false; // Default to disabled on error
    }
  }

  /**
   * Evaluate multiple feature flags at once
   */
  async evaluateAllFlags(
    userContext: UserContext
  ): Promise<Record<string, FeatureFlagResult>> {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), this.defaultTimeout);

      const response = await fetch(`${this.baseURL}/evaluate`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(userContext),
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const results: Record<string, FeatureFlagResult> = await response.json();
      return results;
    } catch (error) {
      console.error('Error evaluating feature flags:', error);
      return {};
    }
  }

  /**
   * Get all available feature flags
   */
  async getAllFlags(): Promise<FeatureFlag[]> {
    try {
      const response = await fetch(`${this.baseURL}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const flags: FeatureFlag[] = await response.json();
      return flags;
    } catch (error) {
      console.error('Error fetching feature flags:', error);
      return [];
    }
  }

  /**
   * Get a specific feature flag
   */
  async getFlag(flagId: string): Promise<FeatureFlag | null> {
    try {
      const response = await fetch(`${this.baseURL}/${flagId}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        if (response.status === 404) {
          return null;
        }
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const flag: FeatureFlag = await response.json();
      return flag;
    } catch (error) {
      console.error(`Error fetching feature flag '${flagId}':`, error);
      return null;
    }
  }

  /**
   * Clear the cache
   */
  clearCache(): void {
    this.cache.clear();
  }

  /**
   * Set cache timeout
   */
  setCacheTimeout(timeout: number): void {
    this.cacheTimeout = timeout;
  }

  /**
   * Set request timeout
   */
  setRequestTimeout(timeout: number): void {
    this.defaultTimeout = timeout;
  }

  /**
   * Check if service is healthy
   */
  async healthCheck(): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseURL}/health`, {
        method: 'GET',
      });
      return response.ok;
    } catch (error) {
      console.error('Feature flags service health check failed:', error);
      return false;
    }
  }
}

// Export singleton instance
export default new FeatureFlagService();

// Export class for custom instances
export { FeatureFlagService };