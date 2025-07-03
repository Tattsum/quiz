package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents JWT token claims
type JWTClaims struct {
	AdminID  int64  `json:"admin_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var jwtSecret []byte
var tokenBlacklist = make(map[string]bool)
var blacklistMutex sync.RWMutex

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-key-change-in-production"
	}
	jwtSecret = []byte(secret)
}

// CORS middleware
func CORS() gin.HandlerFunc {
	allowedOrigins := getEnv("CORS_ALLOWED_ORIGINS", "*")
	allowedMethods := getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS")
	allowedHeaders := getEnv("CORS_ALLOWED_HEADERS", "Origin,Content-Type,Accept,Authorization")

	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", allowedOrigins)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", allowedHeaders)
		c.Header("Access-Control-Allow-Methods", allowedMethods)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

// Logger middleware
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// RateLimit middleware (simplified implementation)
func RateLimit() gin.HandlerFunc {
	// In production, use Redis or similar for distributed rate limiting
	rateLimitMap := make(map[string][]time.Time)
	rateLimitMutex := sync.RWMutex{}

	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// Determine rate limit based on endpoint
		limit := 100 // default: 100 requests per minute
		if strings.HasPrefix(c.Request.URL.Path, "/api/admin") {
			limit = 1000 // admin endpoints: 1000 requests per minute
		} else if strings.HasPrefix(c.Request.URL.Path, "/api/answers") {
			limit = 60 // answer endpoints: 60 requests per minute
		}

		rateLimitMutex.Lock()
		defer rateLimitMutex.Unlock()

		// Clean old requests (older than 1 minute)
		if requests, exists := rateLimitMap[clientIP]; exists {
			var validRequests []time.Time
			cutoff := now.Add(-time.Minute)
			for _, reqTime := range requests {
				if reqTime.After(cutoff) {
					validRequests = append(validRequests, reqTime)
				}
			}
			rateLimitMap[clientIP] = validRequests
		}

		// Check if rate limit exceeded
		if len(rateLimitMap[clientIP]) >= limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "Rate limit exceeded. Please try again later.",
				},
			})
			c.Abort()
			return
		}

		// Add current request
		rateLimitMap[clientIP] = append(rateLimitMap[clientIP], now)
		c.Next()
	})
}

// JWTAuth middleware for admin authentication
func JWTAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "MISSING_TOKEN",
					"message": "Authorization header is required",
				},
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN_FORMAT",
					"message": "Invalid authorization header format",
				},
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Check if token is blacklisted
		blacklistMutex.RLock()
		if tokenBlacklist[tokenString] {
			blacklistMutex.RUnlock()
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "TOKEN_REVOKED",
					"message": "Token has been revoked",
				},
			})
			c.Abort()
			return
		}
		blacklistMutex.RUnlock()

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "Invalid token",
				},
			})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
			// Set user information in context
			c.Set("admin_id", claims.AdminID)
			c.Set("username", claims.Username)
			c.Set("token", tokenString)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_CLAIMS",
					"message": "Invalid token claims",
				},
			})
			c.Abort()
			return
		}
	})
}

// GenerateJWT generates a new JWT token for admin
func GenerateJWT(adminID int64, username string) (string, time.Time, error) {
	expiresHours := 24
	if hoursStr := os.Getenv("JWT_EXPIRES_HOURS"); hoursStr != "" {
		if hours, err := strconv.Atoi(hoursStr); err == nil {
			expiresHours = hours
		}
	}

	expirationTime := time.Now().Add(time.Duration(expiresHours) * time.Hour)

	claims := &JWTClaims{
		AdminID:  adminID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "quiz-system",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}

// BlacklistToken adds a token to the blacklist
func BlacklistToken(tokenString string) {
	blacklistMutex.Lock()
	defer blacklistMutex.Unlock()
	tokenBlacklist[tokenString] = true
}

// getEnv gets environment variable with fallback to default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
