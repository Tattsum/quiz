// Package middleware provides HTTP middleware functions for authentication and request processing.
package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/Tattsum/quiz/internal/services"
)

var (
	tokenBlacklist = make(map[string]bool)
	blacklistMutex sync.RWMutex
	rateLimiters   = make(map[string]*rate.Limiter)
	limiterMutex   sync.Mutex
)

// CORS middleware using gin-contrib/cors
func CORS() gin.HandlerFunc {
	config := cors.DefaultConfig()

	if origins := os.Getenv("CORS_ALLOWED_ORIGINS"); origins != "" {
		if origins == "*" {
			config.AllowAllOrigins = true
		} else {
			config.AllowOrigins = strings.Split(origins, ",")
		}
	} else {
		config.AllowAllOrigins = true
	}

	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	return cors.New(config)
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

// RateLimit middleware using token bucket algorithm
func RateLimit() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Determine rate limit based on endpoint
		var limiter *rate.Limiter

		switch {
		case strings.HasPrefix(c.Request.URL.Path, "/api/admin"):
			// Admin endpoints: 500 requests per minute (increased for performance testing)
			limiter = getLimiter(clientIP+":admin", rate.Every(time.Minute/500), 500)
		case strings.HasPrefix(c.Request.URL.Path, "/api/auth"):
			// Auth endpoints: 20 requests per minute (increased for performance testing)
			limiter = getLimiter(clientIP+":auth", rate.Every(time.Minute/20), 20)
		case strings.HasPrefix(c.Request.URL.Path, "/api/answers"):
			// Answer endpoints: 300 requests per minute (increased for performance testing)
			limiter = getLimiter(clientIP+":answers", rate.Every(time.Second/5), 300)
		default:
			// Default: 500 requests per minute (increased for performance testing)
			limiter = getLimiter(clientIP+":default", rate.Every(time.Minute/500), 500)
		}

		if !limiter.Allow() {
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

		c.Next()
	})
}

// getLimiter returns a rate limiter for the given key
func getLimiter(key string, r rate.Limit, b int) *rate.Limiter {
	limiterMutex.Lock()
	defer limiterMutex.Unlock()

	if limiter, exists := rateLimiters[key]; exists {
		return limiter
	}

	limiter := rate.NewLimiter(r, b)
	rateLimiters[key] = limiter

	// Clean up old limiters periodically (simple cleanup)
	if len(rateLimiters) > 10000 {
		// Remove half of the limiters
		count := 0
		for k := range rateLimiters {
			if count > 5000 {
				break
			}
			delete(rateLimiters, k)
			count++
		}
	}

	return limiter
}

// JWTAuth middleware for admin authentication
func JWTAuth(jwtService *services.JWTService) gin.HandlerFunc {
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

		tokenString := jwtService.ExtractTokenFromHeader(authHeader)
		if tokenString == "" {
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

		// Validate token using JWT service
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			var errorCode, errorMessage string
			switch {
			case errors.Is(err, services.ErrExpiredToken):
				errorCode = "TOKEN_EXPIRED"
				errorMessage = "Token has expired"
			case errors.Is(err, services.ErrInvalidTokenType):
				errorCode = "INVALID_TOKEN_TYPE"
				errorMessage = "Invalid token type"
			default:
				errorCode = "INVALID_TOKEN"
				errorMessage = "Invalid token"
			}

			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    errorCode,
					"message": errorMessage,
				},
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("admin_id", claims.AdminID)
		c.Set("username", claims.Username)
		c.Set("token", tokenString)
		c.Next()
	})
}

// LogoutUser adds a token to the blacklist (logout functionality)
func LogoutUser(c *gin.Context) {
	if token, exists := c.Get("token"); exists {
		if tokenString, ok := token.(string); ok {
			BlacklistToken(tokenString)
		}
	}
}

// BlacklistToken adds a token to the blacklist
func BlacklistToken(tokenString string) {
	blacklistMutex.Lock()
	defer blacklistMutex.Unlock()
	tokenBlacklist[tokenString] = true
}
