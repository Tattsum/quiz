package services

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Tattsum/quiz/internal/models"
)

// StorageService defines the interface for file storage operations
type StorageService interface {
	Store(imageUpload *models.ImageUpload, data []byte) error
	GetURL(imageUpload *models.ImageUpload) string
	Delete(imageUpload *models.ImageUpload) error
}

// LocalStorageService implements StorageService for local file system storage
type LocalStorageService struct {
	uploadDir string
	baseURL   string
}

// NewLocalStorageService creates a new LocalStorageService instance
func NewLocalStorageService(uploadDir, baseURL string) *LocalStorageService {
	return &LocalStorageService{
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}
}

// Store saves image data to local storage
func (s *LocalStorageService) Store(imageUpload *models.ImageUpload, data []byte) error {
	if err := s.ensureUploadDirExists(); err != nil {
		return fmt.Errorf("failed to create upload directory: %w", err)
	}

	subdirs := s.generateSubdirs(imageUpload.Filename)
	fullDir := filepath.Join(s.uploadDir, subdirs)

	if err := os.MkdirAll(fullDir, 0o750); err != nil {
		return fmt.Errorf("failed to create subdirectories: %w", err)
	}

	filePath := filepath.Join(fullDir, imageUpload.Filename)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			// Log error silently
		}
	}()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	imageUpload.Path = filepath.Join(subdirs, imageUpload.Filename)
	imageUpload.URL = s.GetURL(imageUpload)

	return nil
}

// GetURL returns the public URL for an uploaded image
func (s *LocalStorageService) GetURL(imageUpload *models.ImageUpload) string {
	return fmt.Sprintf("%s/uploads/images/%s", s.baseURL, imageUpload.Path)
}

// Delete removes an image file from local storage
func (s *LocalStorageService) Delete(imageUpload *models.ImageUpload) error {
	filePath := filepath.Join(s.uploadDir, imageUpload.Path)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *LocalStorageService) ensureUploadDirExists() error {
	if _, err := os.Stat(s.uploadDir); os.IsNotExist(err) {
		return os.MkdirAll(s.uploadDir, 0o750)
	}
	return nil
}

func (s *LocalStorageService) generateSubdirs(filename string) string {
	if len(filename) < 4 {
		return "misc"
	}

	return fmt.Sprintf("%s/%s", filename[:2], filename[2:4])
}

func (s *LocalStorageService) CopyToStorage(source io.Reader, imageUpload *models.ImageUpload) error {
	if err := s.ensureUploadDirExists(); err != nil {
		return fmt.Errorf("failed to create upload directory: %w", err)
	}

	subdirs := s.generateSubdirs(imageUpload.Filename)
	fullDir := filepath.Join(s.uploadDir, subdirs)

	if err := os.MkdirAll(fullDir, 0o750); err != nil {
		return fmt.Errorf("failed to create subdirectories: %w", err)
	}

	filePath := filepath.Join(fullDir, imageUpload.Filename)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			// Log error silently
		}
	}()

	size, err := io.Copy(file, source)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	imageUpload.Path = filepath.Join(subdirs, imageUpload.Filename)
	imageUpload.URL = s.GetURL(imageUpload)
	imageUpload.Size = size

	return nil
}
