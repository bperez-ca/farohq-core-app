package usecases

import (
	"context"

	"farohq-core-app/internal/domains/files/domain"
	"farohq-core-app/internal/domains/files/domain/ports/inbound"
	"farohq-core-app/internal/domains/files/domain/ports/outbound"
	"farohq-core-app/internal/domains/files/domain/services"
)

// DeleteFile implements the DeleteFile inbound port
type DeleteFile struct {
	storage      outbound.Storage
	keyGenerator *services.KeyGenerator
	bucket       string
}

// NewDeleteFile creates a new DeleteFile use case
func NewDeleteFile(
	storage outbound.Storage,
	keyGenerator *services.KeyGenerator,
	bucket string,
) inbound.DeleteFile {
	return &DeleteFile{
		storage:      storage,
		keyGenerator: keyGenerator,
		bucket:       bucket,
	}
}

// Execute executes the use case
func (uc *DeleteFile) Execute(ctx context.Context, req *inbound.DeleteFileRequest) (*inbound.DeleteFileResponse, error) {
	// Validate key
	if req.Key == "" {
		return nil, domain.ErrFileNotFound
	}

	if !uc.keyGenerator.ValidateKey(req.Key) {
		return nil, domain.ErrInvalidAsset
	}

	// Delete from storage
	if err := uc.storage.DeleteFile(ctx, uc.bucket, req.Key); err != nil {
		return nil, domain.ErrFileNotFound
	}

	return &inbound.DeleteFileResponse{
		Success: true,
	}, nil
}

