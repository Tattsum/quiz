package services

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	"github.com/Tattsum/quiz/internal/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3StorageService implements StorageService for AWS S3 storage
type S3StorageService struct {
	bucket   string
	region   string
	baseURL  string
	s3Client *s3.S3
}

// NewS3StorageService creates a new S3StorageService instance
func NewS3StorageService(bucket, region, baseURL string) (*S3StorageService, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &S3StorageService{
		bucket:   bucket,
		region:   region,
		baseURL:  baseURL,
		s3Client: s3.New(sess),
	}, nil
}

// Store saves image data to S3 storage
func (s *S3StorageService) Store(imageUpload *models.ImageUpload, data []byte) error {
	key := s.generateS3Key(imageUpload.Filename)

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(imageUpload.ContentType),
		ACL:         aws.String("public-read"),
		Metadata: map[string]*string{
			"original-name": aws.String(imageUpload.OriginalName),
			"width":         aws.String(fmt.Sprintf("%d", imageUpload.Width)),
			"height":        aws.String(fmt.Sprintf("%d", imageUpload.Height)),
		},
	}

	_, err := s.s3Client.PutObject(input)
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	imageUpload.Path = key
	imageUpload.URL = s.GetURL(imageUpload)

	return nil
}

// GetURL returns the public URL for an uploaded image in S3
func (s *S3StorageService) GetURL(imageUpload *models.ImageUpload) string {
	if s.baseURL != "" {
		return fmt.Sprintf("%s/%s", s.baseURL, imageUpload.Path)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, imageUpload.Path)
}

// Delete removes an image file from S3 storage
func (s *S3StorageService) Delete(imageUpload *models.ImageUpload) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(imageUpload.Path),
	}

	_, err := s.s3Client.DeleteObject(input)
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

func (s *S3StorageService) generateS3Key(filename string) string {
	subdirs := s.generateSubdirs(filename)
	return filepath.Join("images", subdirs, filename)
}

func (s *S3StorageService) generateSubdirs(filename string) string {
	if len(filename) < 4 {
		return "misc"
	}
	return fmt.Sprintf("%s/%s", filename[:2], filename[2:4])
}

// CopyToStorage uploads data from a reader to S3 storage
func (s *S3StorageService) CopyToStorage(source io.Reader, imageUpload *models.ImageUpload) error {
	key := s.generateS3Key(imageUpload.Filename)

	data, err := io.ReadAll(source)
	if err != nil {
		return fmt.Errorf("failed to read source data: %w", err)
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(imageUpload.ContentType),
		ACL:         aws.String("public-read"),
		Metadata: map[string]*string{
			"original-name": aws.String(imageUpload.OriginalName),
			"width":         aws.String(fmt.Sprintf("%d", imageUpload.Width)),
			"height":        aws.String(fmt.Sprintf("%d", imageUpload.Height)),
		},
	}

	_, err = s.s3Client.PutObject(input)
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	imageUpload.Path = key
	imageUpload.URL = s.GetURL(imageUpload)
	imageUpload.Size = int64(len(data))

	return nil
}

// HeadObject retrieves metadata for an object in S3
func (s *S3StorageService) HeadObject(imageUpload *models.ImageUpload) error {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(imageUpload.Path),
	}

	result, err := s.s3Client.HeadObject(input)
	if err != nil {
		return fmt.Errorf("failed to get object metadata: %w", err)
	}

	if result.ContentLength != nil {
		imageUpload.Size = *result.ContentLength
	}

	return nil
}
