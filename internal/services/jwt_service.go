package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Tattsum/quiz/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken is returned when a JWT token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when a JWT token has expired
	ErrExpiredToken = errors.New("token has expired")
	// ErrInvalidTokenType is returned when a JWT token has an invalid type
	ErrInvalidTokenType = errors.New("invalid token type")
)

// JWTService provides JWT token generation and validation functionality
type JWTService struct {
	accessSecretKey   string
	refreshSecretKey  string
	accessExpiryTime  time.Duration
	refreshExpiryTime time.Duration
}

// NewJWTService creates a new JWT service instance with configuration from environment variables
func NewJWTService() *JWTService {
	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	if accessSecret == "" {
		accessSecret = "default_access_secret_key_change_in_production"
	}

	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if refreshSecret == "" {
		refreshSecret = "default_refresh_secret_key_change_in_production"
	}

	accessExpiryStr := os.Getenv("JWT_ACCESS_EXPIRY")
	accessExpiry := 15 * time.Minute
	if accessExpiryStr != "" {
		if minutes, err := strconv.Atoi(accessExpiryStr); err == nil {
			accessExpiry = time.Duration(minutes) * time.Minute
		}
	}

	refreshExpiryStr := os.Getenv("JWT_REFRESH_EXPIRY")
	refreshExpiry := 7 * 24 * time.Hour
	if refreshExpiryStr != "" {
		if hours, err := strconv.Atoi(refreshExpiryStr); err == nil {
			refreshExpiry = time.Duration(hours) * time.Hour
		}
	}

	return &JWTService{
		accessSecretKey:   accessSecret,
		refreshSecretKey:  refreshSecret,
		accessExpiryTime:  accessExpiry,
		refreshExpiryTime: refreshExpiry,
	}
}

// GenerateTokenPair generates access and refresh token pair for an administrator
func (j *JWTService) GenerateTokenPair(admin *models.Administrator) (*models.LoginResponse, error) {
	now := time.Now()
	accessExpiresAt := now.Add(j.accessExpiryTime)
	refreshExpiresAt := now.Add(j.refreshExpiryTime)

	accessClaims := &models.JWTClaims{
		AdminID:  admin.ID,
		Username: admin.Username,
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "quiz-app",
			Subject:   fmt.Sprintf("admin:%d", admin.ID),
		},
	}

	refreshClaims := &models.JWTClaims{
		AdminID:  admin.ID,
		Username: admin.Username,
		Type:     "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "quiz-app",
			Subject:   fmt.Sprintf("admin:%d", admin.ID),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(j.accessSecretKey))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(j.refreshSecretKey))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	adminCopy := *admin
	adminCopy.PasswordHash = ""

	return &models.LoginResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessExpiresAt,
		Admin:        adminCopy,
	}, nil
}

// ValidateAccessToken validates an access token and returns its claims
func (j *JWTService) ValidateAccessToken(tokenString string) (*models.JWTClaims, error) {
	return j.validateToken(tokenString, j.accessSecretKey, "access")
}

// ValidateRefreshToken validates a refresh token and returns its claims
func (j *JWTService) ValidateRefreshToken(tokenString string) (*models.JWTClaims, error) {
	return j.validateToken(tokenString, j.refreshSecretKey, "refresh")
}

func (j *JWTService) validateToken(tokenString, secretKey, expectedType string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*models.JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Type != expectedType {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// RefreshTokens generates new token pair using a valid refresh token
func (j *JWTService) RefreshTokens(refreshTokenString string, admin *models.Administrator) (*models.RefreshTokenResponse, error) {
	claims, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	if claims.AdminID != admin.ID {
		return nil, ErrInvalidToken
	}

	response, err := j.GenerateTokenPair(admin)
	if err != nil {
		return nil, err
	}

	return &models.RefreshTokenResponse{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		ExpiresAt:    response.ExpiresAt,
	}, nil
}

// GenerateSecureRandomString generates a cryptographically secure random string
func (j *JWTService) GenerateSecureRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func (j *JWTService) ExtractTokenFromHeader(authHeader string) string {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}
