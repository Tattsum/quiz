package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/Tattsum/quiz/internal/database"
	"github.com/Tattsum/quiz/internal/middleware"
	"github.com/Tattsum/quiz/internal/models"
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

	db := database.GetDB()

	// Get admin by username
	var admin models.Administrator
	query := `SELECT id, username, password_hash, email, created_at, updated_at 
			  FROM administrators WHERE username = $1`
	
	err := db.QueryRow(query, req.Username).Scan(
		&admin.ID,
		&admin.Username,
		&admin.PasswordHash,
		&admin.Email,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "INVALID_CREDENTIALS",
					Message: "Invalid username or password",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to query database",
			},
		})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_CREDENTIALS",
				Message: "Invalid username or password",
			},
		})
		return
	}

	// Generate JWT token
	token, expiresAt, err := middleware.GenerateJWT(admin.ID, admin.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TOKEN_GENERATION_ERROR",
				Message: "Failed to generate authentication token",
			},
		})
		return
	}

	// Prepare response (exclude password hash)
	admin.PasswordHash = ""
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "ログインに成功しました",
		Data: models.LoginResponse{
			Token:     token,
			ExpiresAt: expiresAt,
			Admin:     admin,
		},
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

	db := database.GetDB()

	// Get fresh admin data from database
	var admin models.Administrator
	query := `SELECT id, username, email, created_at, updated_at 
			  FROM administrators WHERE id = $1`
	
	err := db.QueryRow(query, adminID).Scan(
		&admin.ID,
		&admin.Username,
		&admin.Email,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "ADMIN_NOT_FOUND",
					Message: "Admin user not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to query database",
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