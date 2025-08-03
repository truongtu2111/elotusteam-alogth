package auth

import (
	"os"
	"strconv"
	"testing"
)

// TestLoadConfig tests the configuration loading
func TestLoadConfig(t *testing.T) {
	// Set test environment variables
	os.Setenv("SERVER_HOST", "testhost")
	os.Setenv("SERVER_PORT", "9999")
	os.Setenv("DB_HOST", "testdb")
	os.Setenv("JWT_SECRET", "testsecret")

	// Test environment variable loading
	host := getEnv("SERVER_HOST", "localhost")
	if host != "testhost" {
		t.Errorf("Expected host to be 'testhost', got '%s'", host)
	}

	port := getEnvAsInt("SERVER_PORT", 8080)
	if port != 9999 {
		t.Errorf("Expected port to be 9999, got %d", port)
	}

	// Test default values
	os.Unsetenv("UNKNOWN_VAR")
	defaultValue := getEnv("UNKNOWN_VAR", "default")
	if defaultValue != "default" {
		t.Errorf("Expected default value 'default', got '%s'", defaultValue)
	}

	defaultPort := getEnvAsInt("UNKNOWN_PORT", 3000)
	if defaultPort != 3000 {
		t.Errorf("Expected default port 3000, got %d", defaultPort)
	}

	// Clean up
	os.Unsetenv("SERVER_HOST")
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("JWT_SECRET")
}

// Helper functions (these would normally be imported from the main package)
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}