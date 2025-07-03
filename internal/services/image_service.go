package services

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/Tattsum/quiz/internal/models"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

const (
	// MaxFileSize defines the maximum allowed file size for image uploads (5MB)
	MaxFileSize    = 5 * 1024 * 1024
	MaxImageWidth  = 1920
	MaxImageHeight = 1080
	JpegQuality    = 80

	// FormatJPEG represents the JPEG image format
	FormatJPEG = "jpeg"
	// FormatPNG represents the PNG image format
	FormatPNG = "png"
	// FormatGIF represents the GIF image format
	FormatGIF = "gif"

	// ContentTypeJPEG represents the MIME type for JPEG images
	ContentTypeJPEG = "image/jpeg"
	// ContentTypePNG represents the MIME type for PNG images
	ContentTypePNG = "image/png"
	// ContentTypeGIF represents the MIME type for GIF images
	ContentTypeGIF = "image/gif"
)

// AllowedContentTypes defines the supported image content types
var AllowedContentTypes = map[string]bool{
	ContentTypeJPEG: true,
	ContentTypePNG:  true,
	ContentTypeGIF:  true,
}

// ImageService provides image processing and validation functionality
type ImageService struct{}

// NewImageService creates a new ImageService instance
func NewImageService() *ImageService {
	return &ImageService{}
}

// ImageValidationResult contains the result of image validation
type ImageValidationResult struct {
	IsValid     bool
	ContentType string
	Size        int64
	Width       int
	Height      int
	Error       error
}

// ValidateImage validates an uploaded image file
func (s *ImageService) ValidateImage(fileHeader *multipart.FileHeader) (*ImageValidationResult, error) {
	result := &ImageValidationResult{}

	if fileHeader.Size > MaxFileSize {
		result.Error = fmt.Errorf("file size %d exceeds maximum allowed size %d", fileHeader.Size, MaxFileSize)
		return result, nil
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer func() { _ = file.Close() }()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read file header: %w", err)
	}

	contentType := s.detectContentType(buffer)
	if !AllowedContentTypes[contentType] {
		result.Error = fmt.Errorf("unsupported file type: %s. Allowed types: jpeg, png, gif", contentType)
		return result, nil
	}

	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %w", err)
	}

	img, format, err := image.DecodeConfig(file)
	if err != nil {
		result.Error = fmt.Errorf("invalid image file: %w", err)
		return result, nil
	}

	expectedFormat := s.getExpectedFormat(contentType)
	if format != expectedFormat {
		result.Error = fmt.Errorf("content type mismatch: expected %s but got %s", expectedFormat, format)
		return result, nil
	}

	result.IsValid = true
	result.ContentType = contentType
	result.Size = fileHeader.Size
	result.Width = img.Width
	result.Height = img.Height

	return result, nil
}

// ProcessImage processes and resizes an uploaded image
func (s *ImageService) ProcessImage(fileHeader *multipart.FileHeader) (*models.ImageUpload, []byte, error) {
	validation, err := s.ValidateImage(fileHeader)
	if err != nil {
		return nil, nil, fmt.Errorf("validation failed: %w", err)
	}

	if !validation.IsValid {
		return nil, nil, validation.Error
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	originalImg, format, err := image.Decode(file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode image: %w", err)
	}

	processedImg := originalImg
	newWidth := validation.Width
	newHeight := validation.Height

	if validation.Width > MaxImageWidth {
		ratio := float64(MaxImageWidth) / float64(validation.Width)
		newWidth = MaxImageWidth
		newHeight = int(float64(validation.Height) * ratio)
		processedImg = imaging.Resize(originalImg, newWidth, newHeight, imaging.Lanczos)
	}

	var buf bytes.Buffer
	switch format {
	case FormatJPEG:
		err = jpeg.Encode(&buf, processedImg, &jpeg.Options{Quality: JpegQuality})
	case FormatPNG:
		err = png.Encode(&buf, processedImg)
	case FormatGIF:
		err = gif.Encode(&buf, processedImg, nil)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode processed image: %w", err)
	}

	filename := s.generateFilename(fileHeader.Filename, format)

	imageUpload := &models.ImageUpload{
		OriginalName: fileHeader.Filename,
		Filename:     filename,
		ContentType:  validation.ContentType,
		Size:         int64(buf.Len()),
		Width:        newWidth,
		Height:       newHeight,
		CreatedAt:    time.Now(),
	}

	return imageUpload, buf.Bytes(), nil
}

func (s *ImageService) detectContentType(buffer []byte) string {
	if len(buffer) < 12 {
		return ""
	}

	if bytes.HasPrefix(buffer, []byte{0xFF, 0xD8, 0xFF}) {
		return ContentTypeJPEG
	}

	if bytes.HasPrefix(buffer, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return ContentTypePNG
	}

	if bytes.HasPrefix(buffer, []byte("GIF87a")) || bytes.HasPrefix(buffer, []byte("GIF89a")) {
		return ContentTypeGIF
	}

	return ""
}

func (s *ImageService) getExpectedFormat(contentType string) string {
	switch contentType {
	case ContentTypeJPEG:
		return FormatJPEG
	case ContentTypePNG:
		return FormatPNG
	case ContentTypeGIF:
		return FormatGIF
	default:
		return ""
	}
}

func (s *ImageService) generateFilename(originalName, format string) string {
	ext := s.getFileExtension(format)
	name := strings.TrimSuffix(originalName, filepath.Ext(originalName))

	cleanName := s.sanitizeFilename(name)
	if len(cleanName) > 50 {
		cleanName = cleanName[:50]
	}

	timestamp := time.Now().Format("20060102_150405")
	uuid := uuid.New().String()[:8]

	return fmt.Sprintf("%s_%s_%s%s", cleanName, timestamp, uuid, ext)
}

func (s *ImageService) getFileExtension(format string) string {
	switch format {
	case "jpeg":
		return ".jpg"
	case "png":
		return ".png"
	case "gif":
		return ".gif"
	default:
		return ".jpg"
	}
}

func (s *ImageService) sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, " ", "_")

	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			result.WriteRune(r)
		}
	}

	sanitized := result.String()
	if sanitized == "" {
		sanitized = "image"
	}

	return sanitized
}

// GetImageInfo extracts image information from binary data
func (s *ImageService) GetImageInfo(data []byte) (*ImageValidationResult, error) {
	reader := bytes.NewReader(data)

	config, format, err := image.DecodeConfig(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image config: %w", err)
	}

	contentType := ""
	switch format {
	case "jpeg":
		contentType = "image/jpeg"
	case "png":
		contentType = "image/png"
	case "gif":
		contentType = "image/gif"
	}

	return &ImageValidationResult{
		IsValid:     true,
		ContentType: contentType,
		Size:        int64(len(data)),
		Width:       config.Width,
		Height:      config.Height,
	}, nil
}
