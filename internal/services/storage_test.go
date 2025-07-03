package services

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Tattsum/quiz/internal/models"
)

func TestLocalStorageService_Store(t *testing.T) {
	tempDir := t.TempDir()
	baseURL := "http://localhost:8080"

	service := NewLocalStorageService(tempDir, baseURL)

	imageUpload := &models.ImageUpload{
		OriginalName: "test.jpg",
		Filename:     "test_123.jpg",
		ContentType:  "image/jpeg",
		Size:         1024,
		Width:        800,
		Height:       600,
		CreatedAt:    time.Now(),
	}

	testData := []byte("test image data")

	err := service.Store(imageUpload, testData)
	if err != nil {
		t.Fatalf("failed to store image: %v", err)
	}

	if imageUpload.Path == "" {
		t.Error("expected Path to be set")
	}

	if imageUpload.URL == "" {
		t.Error("expected URL to be set")
	}

	fullPath := filepath.Join(tempDir, imageUpload.Path)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Error("expected file to exist")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("failed to read stored file: %v", err)
	}

	if string(content) != string(testData) {
		t.Error("stored content does not match original data")
	}
}

func TestLocalStorageService_GetURL(t *testing.T) {
	baseURL := "http://localhost:8080"
	service := NewLocalStorageService("/uploads", baseURL)

	imageUpload := &models.ImageUpload{
		Path: "te/st/test_123.jpg",
	}

	url := service.GetURL(imageUpload)
	expected := "http://localhost:8080/uploads/images/te/st/test_123.jpg"

	if url != expected {
		t.Errorf("expected URL %s, got %s", expected, url)
	}
}

func TestLocalStorageService_Delete(t *testing.T) {
	tempDir := t.TempDir()
	service := NewLocalStorageService(tempDir, "")

	imageUpload := &models.ImageUpload{
		Filename: "test_123.jpg",
		Path:     "te/st/test_123.jpg",
	}

	testData := []byte("test image data")
	err := service.Store(imageUpload, testData)
	if err != nil {
		t.Fatalf("failed to store image: %v", err)
	}

	fullPath := filepath.Join(tempDir, imageUpload.Path)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Fatal("file should exist before deletion")
	}

	err = service.Delete(imageUpload)
	if err != nil {
		t.Fatalf("failed to delete image: %v", err)
	}

	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		t.Error("file should not exist after deletion")
	}
}

func TestLocalStorageService_generateSubdirs(t *testing.T) {
	service := NewLocalStorageService("", "")

	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "normal filename",
			filename: "test_123.jpg",
			expected: "te/st",
		},
		{
			name:     "short filename",
			filename: "ab",
			expected: "misc",
		},
		{
			name:     "long filename",
			filename: "verylongfilename.jpg",
			expected: "ve/ry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.generateSubdirs(tt.filename)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
