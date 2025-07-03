package handlers

import "os"

const testEnvValue = "true"

// setupTestEnv sets up test environment variables if not already set
func setupTestEnv() {
	if os.Getenv("TEST_ENV") == testEnvValue {
		return
	}

	// Set default test database configuration
	if os.Getenv("DB_HOST") == "" {
		_ = os.Setenv("DB_HOST", "localhost")
	}
	if os.Getenv("DB_PORT") == "" {
		_ = os.Setenv("DB_PORT", "5433")
	}
	if os.Getenv("DB_USER") == "" {
		_ = os.Setenv("DB_USER", "quiz_user")
	}
	if os.Getenv("DB_PASSWORD") == "" {
		_ = os.Setenv("DB_PASSWORD", "quiz_password")
	}
	if os.Getenv("DB_NAME") == "" {
		_ = os.Setenv("DB_NAME", "quiz_db_test")
	}
	if os.Getenv("DB_SSLMODE") == "" {
		_ = os.Setenv("DB_SSLMODE", "disable")
	}
}
