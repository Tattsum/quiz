// Package services provides business logic and service layer functionality.
package services

import (
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/Tattsum/quiz/internal/database"
	"github.com/Tattsum/quiz/internal/middleware"
	"github.com/Tattsum/quiz/internal/models"
)

// AuthService provides authentication related business logic
type AuthService struct {
	db *sql.DB
}

// NewAuthService creates a new AuthService instance
func NewAuthService() *AuthService {
	return &AuthService{
		db: database.GetDB(),
	}
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(username, password string) (*models.LoginResponse, error) {
	if username == "" || password == "" {
		return nil, errors.New("username and password are required")
	}

	admin, err := s.getAdminByUsername(username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, expiresAt, err := middleware.GenerateJWT(admin.ID, admin.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	admin.PasswordHash = ""

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		Admin:     *admin,
	}, nil
}

// GetAdminByID retrieves an administrator by their ID
func (s *AuthService) GetAdminByID(id int64) (*models.Administrator, error) {
	if id <= 0 {
		return nil, errors.New("invalid admin ID")
	}

	var admin models.Administrator
	query := `SELECT id, username, email, created_at, updated_at 
			  FROM administrators WHERE id = $1`

	err := s.db.QueryRow(query, id).Scan(
		&admin.ID,
		&admin.Username,
		&admin.Email,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("admin not found")
		}
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}

	return &admin, nil
}

func (s *AuthService) getAdminByUsername(username string) (*models.Administrator, error) {
	var admin models.Administrator
	query := `SELECT id, username, password_hash, email, created_at, updated_at 
			  FROM administrators WHERE username = $1`

	err := s.db.QueryRow(query, username).Scan(
		&admin.ID,
		&admin.Username,
		&admin.PasswordHash,
		&admin.Email,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}

	return &admin, nil
}
