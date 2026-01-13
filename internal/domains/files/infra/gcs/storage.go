package gcs

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"farohq-core-app/internal/domains/files/domain/ports/outbound"
	"google.golang.org/api/option"
)

// Storage implements the outbound.Storage interface using GCP Cloud Storage
type Storage struct {
	client *storage.Client
	bucket string
}

// NewStorage creates a new GCS storage adapter
func NewStorage(bucket string, projectID string, credentialsPath string) (outbound.Storage, error) {
	ctx := context.Background()

	var client *storage.Client
	var err error

	// Use credentials file if provided, otherwise use default credentials
	if credentialsPath != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(credentialsPath))
	} else {
		// Use default credentials (GOOGLE_APPLICATION_CREDENTIALS env var or default credentials)
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &Storage{
		client: client,
		bucket: bucket,
	}, nil
}

// GeneratePresignedURL generates a signed URL for uploading a file to GCS
// Supports both logo and favicon asset types
func (s *Storage) GeneratePresignedURL(ctx context.Context, bucket, key string, expiresIn time.Duration) (string, map[string]string, error) {
	// Use provided bucket or default to instance bucket
	targetBucket := bucket
	if targetBucket == "" {
		targetBucket = s.bucket
	}

	// Generate signed URL for PUT operation (upload)
	opts := &storage.SignedURLOptions{
		Method:  "PUT",
		Expires: time.Now().Add(expiresIn),
	}

	// Set Content-Type based on file extension (for logo/favicon)
	// This ensures proper MIME type handling
	contentType := getContentTypeFromKey(key)
	if contentType != "" {
		opts.Headers = []string{
			fmt.Sprintf("Content-Type: %s", contentType),
		}
	}

	// Use Bucket.SignedURL method (correct GCS API)
	url, err := s.client.Bucket(targetBucket).SignedURL(key, opts)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate signed URL: %w", err)
	}

	// Convert headers to map format expected by interface
	headers := make(map[string]string)
	if contentType != "" {
		headers["Content-Type"] = contentType
	}

	return url, headers, nil
}

// DeleteFile deletes a file from GCS bucket
func (s *Storage) DeleteFile(ctx context.Context, bucket, key string) error {
	// Use provided bucket or default to instance bucket
	targetBucket := bucket
	if targetBucket == "" {
		targetBucket = s.bucket
	}

	obj := s.client.Bucket(targetBucket).Object(key)

	if err := obj.Delete(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			return fmt.Errorf("file not found: %w", err)
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// getContentTypeFromKey determines Content-Type from file extension
func getContentTypeFromKey(key string) string {
	if len(key) == 0 {
		return "application/octet-stream"
	}

	// Extract extension
	var ext string
	for i := len(key) - 1; i >= 0 && key[i] != '/'; i-- {
		if key[i] == '.' {
			ext = key[i:]
			break
		}
	}

	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".svg":
		return "image/svg+xml"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

// Close closes the GCS client
func (s *Storage) Close() error {
	return s.client.Close()
}
