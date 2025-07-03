package handlers

import (
	"fmt"
	"net/http"

	"github.com/Tattsum/quiz/internal/models"
	"github.com/Tattsum/quiz/internal/services"
	"github.com/gin-gonic/gin"
)

// ImageUploadHandler handles image upload requests
type ImageUploadHandler struct {
	imageService   *services.ImageService
	storageService services.StorageService
}

// NewImageUploadHandler creates a new ImageUploadHandler instance
func NewImageUploadHandler(storageService services.StorageService) *ImageUploadHandler {
	return &ImageUploadHandler{
		imageService:   services.NewImageService(),
		storageService: storageService,
	}
}

// UploadImage handles image file upload with validation, processing, and storage
func (h *ImageUploadHandler) UploadImage(c *gin.Context) {
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

	imageUpload, processedData, err := h.imageService.ProcessImage(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "PROCESSING_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	if err := h.storageService.Store(imageUpload, processedData); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "STORAGE_ERROR",
				Message: "Failed to store uploaded image",
			},
		})
		return
	}

	response := models.UploadResponse{
		URL:      imageUpload.URL,
		Filename: imageUpload.Filename,
		Size:     imageUpload.Size,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "画像がアップロードされました",
		Data:    response,
	})
}

// UploadImage is the legacy function that maintains compatibility
func UploadImage(c *gin.Context) {
	uploadDir := "uploads/images"
	baseURL := getBaseURL(c)

	storageService := services.NewLocalStorageService(uploadDir, baseURL)
	handler := NewImageUploadHandler(storageService)
	handler.UploadImage(c)
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
