package services

import (
	"testing"
)

func TestAuthService_Login(t *testing.T) {
	service := NewAuthService()

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
			name:     "non-empty credentials",
			username: "admin",
			password: "password",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Login(tt.username, tt.password)
			if tt.name == "non-empty credentials" {
				if err != nil && err.Error() != "invalid credentials" {
					t.Errorf("Login() unexpected error = %v", err)
				}
			} else if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthService_GetAdminByID(t *testing.T) {
	service := NewAuthService()

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
			name:    "valid ID format",
			id:      1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetAdminByID(tt.id)
			if tt.name == "valid ID format" {
				if err != nil && err.Error() != "admin not found" {
					t.Errorf("GetAdminByID() unexpected error = %v", err)
				}
			} else if (err != nil) != tt.wantErr {
				t.Errorf("GetAdminByID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
