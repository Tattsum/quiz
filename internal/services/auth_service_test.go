package services

import (
	"testing"
)

// TestAuthService_AuthenticateAdmin tests the AuthenticateAdmin method
func TestAuthService_AuthenticateAdmin(t *testing.T) {
	// Skip test if no database connection available
	service := &AuthService{db: nil}

	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "empty username",
			username: "",
			password: "password",
			wantErr:  true,
		},
		{
			name:     "empty password",
			username: "admin",
			password: "",
			wantErr:  true,
		},
		{
			name:     "both empty",
			username: "",
			password: "",
			wantErr:  true,
		},
		{
			name:     "non-empty credentials (no DB)",
			username: "admin",
			password: "password",
			wantErr:  true, // Expected to fail due to no DB connection
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.AuthenticateAdmin(tt.username, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthenticateAdmin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthService_HashPassword(t *testing.T) {
	service := NewAuthService()

	password := "testpassword123"
	hash, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	if hash == "" {
		t.Error("HashPassword() returned empty hash")
	}

	if hash == password {
		t.Error("HashPassword() returned plaintext password")
	}

	// Test that the same password generates different hashes (due to salt)
	hash2, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() failed on second call: %v", err)
	}

	if hash == hash2 {
		t.Error("HashPassword() should generate different hashes for same password")
	}
}

func TestAuthService_GetAdminByID(t *testing.T) {
	service := &AuthService{db: nil}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "invalid ID - zero",
			id:      0,
			wantErr: true,
		},
		{
			name:    "invalid ID - negative",
			id:      -1,
			wantErr: true,
		},
		{
			name:    "valid ID format (no DB)",
			id:      1,
			wantErr: true, // Expected to fail due to no DB connection
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetAdminByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAdminByID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
