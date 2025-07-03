package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Tattsum/quiz/internal/models"
)

// UploadImage handles image file upload
func UploadImage(c *gin.Context) {
	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NO_FILE",
				Message: "No file uploaded",
			},
		})
		return
	}

	// Check file size
	maxSize := int64(5 * 1024 * 1024) // 5MB default
	if maxSizeStr := os.Getenv("UPLOAD_MAX_SIZE"); maxSizeStr != "" {
		if size, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
			maxSize = size
		}
	}

	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FILE_TOO_LARGE",
				Message: fmt.Sprintf("File size exceeds maximum allowed size of %d bytes", maxSize),
			},
		})
		return
	}

	// Check file type
	allowedTypes := []string{"image/jpeg", "image/png", "image/gif"}
	if allowedTypesStr := os.Getenv("UPLOAD_ALLOWED_TYPES"); allowedTypesStr != "" {
		allowedTypes = strings.Split(allowedTypesStr, ",")
	}

	// Open file to check content type
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FILE_READ_ERROR",
				Message: "Failed to read uploaded file",
			},
		})
		return
	}
	defer src.Close()

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FILE_READ_ERROR",
				Message: "Failed to read file content",
			},
		})
		return
	}

	contentType := http.DetectContentType(buffer)
	
	// Check if content type is allowed
	allowed := false
	for _, allowedType := range allowedTypes {
		if strings.TrimSpace(allowedType) == contentType {
			allowed = true
			break
		}
	}

	if !allowed {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_FILE_TYPE",
				Message: fmt.Sprintf("File type %s is not allowed. Allowed types: %s", 
					contentType, strings.Join(allowedTypes, ", ")),
			},
		})
		return
	}

	// Generate unique filename
	timestamp := time.Now().Format("20060102_150405")
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		// Determine extension from content type
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		default:
			ext = ".jpg"
		}
	}

	filename := fmt.Sprintf("quiz_image_%s_%d%s", timestamp, time.Now().UnixNano(), ext)
	
	// Ensure upload directory exists
	uploadDir := "uploads/images"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DIRECTORY_ERROR",
				Message: "Failed to create upload directory",
			},
		})
		return
	}

	// Save file
	filePath := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SAVE_ERROR",
				Message: "Failed to save uploaded file",
			},
		})
		return
	}

	// Generate URL (you might want to use a CDN or proper base URL in production)
	baseURL := getBaseURL(c)
	fileURL := fmt.Sprintf("%s/%s", baseURL, filePath)

	response := models.UploadResponse{
		URL:      fileURL,
		Filename: filename,
		Size:     file.Size,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "画像がアップロードされました",
		Data:    response,
	})
}

// getBaseURL returns the base URL for the application
func getBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	
	// Check for forwarded protocol header (for reverse proxies)
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	
	host := c.Request.Host
	if host == "" {
		host = "localhost:8080"
	}
	
	return fmt.Sprintf("%s://%s", scheme, host)
}