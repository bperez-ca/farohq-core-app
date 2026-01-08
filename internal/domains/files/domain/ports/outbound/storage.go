package outbound

import (
	"context"
	"time"
)

// Storage defines the interface for file storage operations
type Storage interface {
	// GeneratePresignedURL generates a pre-signed URL for uploading a file
	GeneratePresignedURL(ctx context.Context, bucket, key string, expiresIn time.Duration) (string, map[string]string, error)

	// DeleteFile deletes a file from storage
	DeleteFile(ctx context.Context, bucket, key string) error
}

