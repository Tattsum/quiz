// Package services provides business logic and service layer functionality.
package services

import (
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/Tattsum/quiz/internal/database"
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

// AuthenticateAdmin authenticates an admin user and returns admin details
func (s *AuthService) AuthenticateAdmin(username, password string) (*models.Administrator, error) {
	if username == "" || password == "" {
		return nil, errors.New("username and password are required")
	}

	if s.db == nil {
		return nil, errors.New("database connection not initialized")
	}

	admin, err := s.getAdminByUsername(username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Clear password hash for security
	admin.PasswordHash = ""
	return admin, nil
}

// GetAdminByID retrieves an administrator by their ID
func (s *AuthService) GetAdminByID(id int64) (*models.Administrator, error) {
	if id <= 0 {
		return nil, errors.New("invalid admin ID")
	}

	if s.db == nil {
		return nil, errors.New("database connection not initialized")
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
	if s.db == nil {
		return nil, errors.New("database connection not initialized")
	}

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

// HashPassword generates a bcrypt hash of the password
func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CreateAdmin creates a new administrator (for initial setup or testing)
func (s *AuthService) CreateAdmin(username, password, email string) (*models.Administrator, error) {
	if username == "" || password == "" || email == "" {
		return nil, errors.New("username, password, and email are required")
	}

	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	query := `INSERT INTO administrators (username, password_hash, email, created_at, updated_at) 
			  VALUES ($1, $2, $3, NOW(), NOW()) 
			  RETURNING id, username, email, created_at, updated_at`

	var admin models.Administrator
	err = s.db.QueryRow(query, username, hashedPassword, email).Scan(
		&admin.ID,
		&admin.Username,
		&admin.Email,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}

	return &admin, nil
}
