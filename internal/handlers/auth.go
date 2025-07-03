// Package handlers provides HTTP handlers for the quiz application API endpoints.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Tattsum/quiz/internal/middleware"
	"github.com/Tattsum/quiz/internal/models"
	"github.com/Tattsum/quiz/internal/services"
)

// AdminLogin handles admin login
func AdminLogin(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid request data",
				Details: parseValidationErrors(err),
			},
		})
		return
	}

	authService := services.NewAuthService()
	response, err := authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_CREDENTIALS",
				Message: "Invalid username or password",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "ログインに成功しました",
		Data:    response,
	})
}

// AdminLogout handles admin logout
func AdminLogout(c *gin.Context) {
	// Get token from context (set by JWT middleware)
	token, exists := c.Get("token")
	if !exists {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_TOKEN",
				Message: "Token not found in request",
			},
		})
		return
	}

	// Add token to blacklist
	middleware.BlacklistToken(token.(string))

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "ログアウトしました",
	})
}

// VerifyToken verifies the current JWT token
func VerifyToken(c *gin.Context) {
	// Get admin info from context (set by JWT middleware)
	adminID, _ := c.Get("admin_id")
	username, _ := c.Get("username")

	authService := services.NewAuthService()
	admin, err := authService.GetAdminByID(adminID.(int64))
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "ADMIN_NOT_FOUND",
				Message: "Admin user not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"valid": true,
			"admin": admin,
			"context": map[string]interface{}{
				"admin_id": adminID,
				"username": username,
			},
		},
	})
}
