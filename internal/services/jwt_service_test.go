package services

import (
	"os"
	"testing"
	"time"

	"github.com/Tattsum/quiz/internal/models"
)

func TestJWTService_GenerateTokenPair(t *testing.T) {
	// Set test environment variables
	os.Setenv("JWT_ACCESS_SECRET", "test_access_secret")
	os.Setenv("JWT_REFRESH_SECRET", "test_refresh_secret")
	os.Setenv("JWT_ACCESS_EXPIRY", "15")   // 15 minutes
	os.Setenv("JWT_REFRESH_EXPIRY", "168") // 7 days

	jwtService := NewJWTService()

	admin := &models.Administrator{
		ID:       1,
		Username: "testadmin",
		Email:    "test@example.com",
	}

	response, err := jwtService.GenerateTokenPair(admin)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	if response.AccessToken == "" {
		t.Error("Access token should not be empty")
	}

	if response.RefreshToken == "" {
		t.Error("Refresh token should not be empty")
	}

	if response.ExpiresAt.Before(time.Now()) {
		t.Error("Token should not be expired immediately")
	}

	if response.Admin.ID != admin.ID {
		t.Error("Admin ID in response should match input")
	}

	if response.Admin.PasswordHash != "" {
		t.Error("Password hash should be cleared in response")
	}
}

func TestJWTService_ValidateAccessToken(t *testing.T) {
	os.Setenv("JWT_ACCESS_SECRET", "test_access_secret")
	os.Setenv("JWT_REFRESH_SECRET", "test_refresh_secret")

	jwtService := NewJWTService()

	admin := &models.Administrator{
		ID:       1,
		Username: "testadmin",
		Email:    "test@example.com",
	}

	response, err := jwtService.GenerateTokenPair(admin)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	// Test valid access token
	claims, err := jwtService.ValidateAccessToken(response.AccessToken)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}

	if claims.AdminID != admin.ID {
		t.Error("Admin ID in claims should match")
	}

	if claims.Username != admin.Username {
		t.Error("Username in claims should match")
	}

	if claims.Type != "access" {
		t.Error("Token type should be 'access'")
	}

	// Test invalid token
	_, err = jwtService.ValidateAccessToken("invalid_token")
	if err == nil {
		t.Error("Should fail with invalid token")
	}

	// Test refresh token with access validation (should fail)
	_, err = jwtService.ValidateAccessToken(response.RefreshToken)
	if err == nil {
		t.Error("Should fail when validating refresh token as access token")
	}
}

func TestJWTService_ValidateRefreshToken(t *testing.T) {
	os.Setenv("JWT_ACCESS_SECRET", "test_access_secret")
	os.Setenv("JWT_REFRESH_SECRET", "test_refresh_secret")

	jwtService := NewJWTService()

	admin := &models.Administrator{
		ID:       1,
		Username: "testadmin",
		Email:    "test@example.com",
	}

	response, err := jwtService.GenerateTokenPair(admin)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	// Test valid refresh token
	claims, err := jwtService.ValidateRefreshToken(response.RefreshToken)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}

	if claims.AdminID != admin.ID {
		t.Error("Admin ID in claims should match")
	}

	if claims.Type != "refresh" {
		t.Error("Token type should be 'refresh'")
	}

	// Test access token with refresh validation (should fail)
	_, err = jwtService.ValidateRefreshToken(response.AccessToken)
	if err == nil {
		t.Error("Should fail when validating access token as refresh token")
	}
}

func TestJWTService_RefreshTokens(t *testing.T) {
	os.Setenv("JWT_ACCESS_SECRET", "test_access_secret")
	os.Setenv("JWT_REFRESH_SECRET", "test_refresh_secret")

	jwtService := NewJWTService()

	admin := &models.Administrator{
		ID:       1,
		Username: "testadmin",
		Email:    "test@example.com",
	}

	// Generate initial token pair
	initialResponse, err := jwtService.GenerateTokenPair(admin)
	if err != nil {
		t.Fatalf("Failed to generate initial token pair: %v", err)
	}

	// Wait a moment to ensure new tokens have different issued time
	time.Sleep(1 * time.Second)

	// Refresh tokens
	refreshResponse, err := jwtService.RefreshTokens(initialResponse.RefreshToken, admin)
	if err != nil {
		t.Fatalf("Failed to refresh tokens: %v", err)
	}

	if refreshResponse.AccessToken == "" {
		t.Error("New access token should not be empty")
	}

	if refreshResponse.RefreshToken == "" {
		t.Error("New refresh token should not be empty")
	}

	if refreshResponse.AccessToken == initialResponse.AccessToken {
		t.Error("New access token should be different from original")
	}

	if refreshResponse.RefreshToken == initialResponse.RefreshToken {
		t.Error("New refresh token should be different from original")
	}

	// Validate new tokens
	_, err = jwtService.ValidateAccessToken(refreshResponse.AccessToken)
	if err != nil {
		t.Error("New access token should be valid")
	}

	_, err = jwtService.ValidateRefreshToken(refreshResponse.RefreshToken)
	if err != nil {
		t.Error("New refresh token should be valid")
	}
}

func TestJWTService_ExtractTokenFromHeader(t *testing.T) {
	jwtService := NewJWTService()

	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "Valid Bearer token",
			header:   "Bearer abc123def456",
			expected: "abc123def456",
		},
		{
			name:     "Invalid format - no Bearer",
			header:   "abc123def456",
			expected: "",
		},
		{
			name:     "Invalid format - wrong prefix",
			header:   "Token abc123def456",
			expected: "",
		},
		{
			name:     "Empty header",
			header:   "",
			expected: "",
		},
		{
			name:     "Only Bearer",
			header:   "Bearer",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jwtService.ExtractTokenFromHeader(tt.header)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestJWTService_GenerateSecureRandomString(t *testing.T) {
	jwtService := NewJWTService()

	// Test different lengths
	lengths := []int{16, 32, 64}

	for _, length := range lengths {
		result, err := jwtService.GenerateSecureRandomString(length)
		if err != nil {
			t.Fatalf("Failed to generate random string: %v", err)
		}

		if len(result) != length*2 { // hex encoding doubles the length
			t.Errorf("Expected length %d, got %d", length*2, len(result))
		}

		// Generate another string and ensure they're different
		result2, err := jwtService.GenerateSecureRandomString(length)
		if err != nil {
			t.Fatalf("Failed to generate second random string: %v", err)
		}

		if result == result2 {
			t.Error("Two random strings should be different")
		}
	}
}
