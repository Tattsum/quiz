package services

import (
	"testing"
)

func TestImageService_detectContentType_Simple(t *testing.T) {
	service := NewImageService()

	tests := []struct {
		name     string
		buffer   []byte
		expected string
	}{
		{
			name:     "JPEG image",
			buffer:   []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01},
			expected: ContentTypeJPEG,
		},
		{
			name:     "PNG image",
			buffer:   []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D},
			expected: ContentTypePNG,
		},
		{
			name:     "GIF image",
			buffer:   append([]byte("GIF87a"), make([]byte, 6)...),
			expected: ContentTypeGIF,
		},
		{
			name:     "unknown format",
			buffer:   []byte("unknown"),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.detectContentType(tt.buffer)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestImageService_sanitizeFilename_Simple(t *testing.T) {
	service := NewImageService()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal filename",
			input:    "test_image",
			expected: "test_image",
		},
		{
			name:     "filename with spaces",
			input:    "test image file",
			expected: "test_image_file",
		},
		{
			name:     "filename with special characters",
			input:    "test@#$%^&*()image",
			expected: "testimage",
		},
		{
			name:     "empty filename",
			input:    "",
			expected: "image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
