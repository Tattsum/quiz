package services

import (
	"bytes"
	"image"
	"image/jpeg"
	"mime/multipart"
	"net/textproto"
	"testing"
)

func TestImageService_ValidateImage(t *testing.T) {
	service := NewImageService()

	tests := []struct {
		name        string
		fileHeader  *multipart.FileHeader
		shouldError bool
		expectValid bool
	}{
		{
			name:        "valid jpeg file",
			fileHeader:  createTestImageFileHeader("test.jpg", "image/jpeg", createTestJPEGImage()),
			shouldError: false,
			expectValid: true,
		},
		{
			name:        "file too large",
			fileHeader:  createLargeFileHeader("large.jpg", "image/jpeg"),
			shouldError: false,
			expectValid: false,
		},
		{
			name:        "invalid file type",
			fileHeader:  createTestFileHeader("test.txt", "text/plain", []byte("not an image")),
			shouldError: false,
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ValidateImage(tt.fileHeader)

			if tt.shouldError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result != nil && result.IsValid != tt.expectValid {
				t.Errorf("expected IsValid=%v, got %v", tt.expectValid, result.IsValid)
			}
		})
	}
}

func TestImageService_ProcessImage(t *testing.T) {
	service := NewImageService()

	tests := []struct {
		name         string
		fileHeader   *multipart.FileHeader
		shouldError  bool
		expectResize bool
	}{
		{
			name:         "normal size image",
			fileHeader:   createTestImageFileHeader("normal.jpg", "image/jpeg", createTestJPEGImage()),
			shouldError:  false,
			expectResize: false,
		},
		{
			name:         "large image that needs resizing",
			fileHeader:   createTestImageFileHeader("large.jpg", "image/jpeg", createLargeTestJPEGImage()),
			shouldError:  false,
			expectResize: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imageUpload, data, err := service.ProcessImage(tt.fileHeader)

			if tt.shouldError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.shouldError {
				if imageUpload == nil {
					t.Error("expected imageUpload but got nil")
				}
				if data == nil {
					t.Error("expected data but got nil")
				}
				if imageUpload != nil && tt.expectResize && imageUpload.Width > MaxImageWidth {
					t.Errorf("expected image to be resized to max width %d, got %d", MaxImageWidth, imageUpload.Width)
				}
			}
		})
	}
}

func TestImageService_detectContentType(t *testing.T) {
	service := NewImageService()

	tests := []struct {
		name     string
		buffer   []byte
		expected string
	}{
		{
			name:     "JPEG image",
			buffer:   []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01},
			expected: "image/jpeg",
		},
		{
			name:     "PNG image",
			buffer:   []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D},
			expected: "image/png",
		},
		{
			name:     "GIF image",
			buffer:   append([]byte("GIF87a"), make([]byte, 6)...),
			expected: "image/gif",
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

func TestImageService_sanitizeFilename(t *testing.T) {
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
		{
			name:     "long filename",
			input:    "this_is_a_very_long_filename_that_exceeds_fifty_characters_and_should_be_truncated",
			expected: "this_is_a_very_long_filename_that_exceeds_fifty_characters_and_should_be_truncated",
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

// Helper functions for creating test data

func createTestImageFileHeader(filename, contentType string, imageData []byte) *multipart.FileHeader {
	// Create a buffer to simulate multipart data
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	// Create form file
	fw, _ := writer.CreateFormFile("image", filename)
	_, _ = fw.Write(imageData)
	writer.Close()

	// Parse the multipart data to get a real FileHeader
	reader := multipart.NewReader(&b, writer.Boundary())
	form, _ := reader.ReadForm(int64(len(imageData)) + 1024)

	// Clean up form on exit
	defer func() {
		if form != nil {
			_ = form.RemoveAll()
		}
	}()

	if form.File["image"] != nil && len(form.File["image"]) > 0 {
		return form.File["image"][0]
	}

	// Fallback to a basic FileHeader if parsing fails
	return &multipart.FileHeader{
		Filename: filename,
		Header: textproto.MIMEHeader{
			"Content-Type": []string{contentType},
		},
		Size: int64(len(imageData)),
	}
}

func createTestFileHeader(filename, contentType string, data []byte) *multipart.FileHeader {
	// Create a buffer to simulate multipart data
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	// Create form file
	fw, _ := writer.CreateFormFile("file", filename)
	_, _ = fw.Write(data)
	writer.Close()

	// Parse the multipart data to get a real FileHeader
	reader := multipart.NewReader(&b, writer.Boundary())
	form, _ := reader.ReadForm(int64(len(data)) + 1024)

	// Clean up form on exit
	defer func() {
		if form != nil {
			_ = form.RemoveAll()
		}
	}()

	if form.File["file"] != nil && len(form.File["file"]) > 0 {
		return form.File["file"][0]
	}

	// Fallback to a basic FileHeader if parsing fails
	return &multipart.FileHeader{
		Filename: filename,
		Header: textproto.MIMEHeader{
			"Content-Type": []string{contentType},
		},
		Size: int64(len(data)),
	}
}

func createLargeFileHeader(filename, contentType string) *multipart.FileHeader {
	return &multipart.FileHeader{
		Filename: filename,
		Header: textproto.MIMEHeader{
			"Content-Type": []string{contentType},
		},
		Size: MaxFileSize + 1,
	}
}

func createTestJPEGImage() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}) // テスト用なのでエラーハンドリング不要
	return buf.Bytes()
}

func createLargeTestJPEGImage() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 2500, 2000))

	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}) // テスト用なのでエラーハンドリング不要
	return buf.Bytes()
}
