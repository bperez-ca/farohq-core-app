package usecases

import (
	"context"
	"strings"
	"time"

	"farohq-core-app/internal/domains/files/domain"
	"farohq-core-app/internal/domains/files/domain/ports/inbound"
	"farohq-core-app/internal/domains/files/domain/ports/outbound"
	"farohq-core-app/internal/domains/files/domain/services"

	"github.com/google/uuid"
)

// SignUpload implements the SignUpload inbound port
type SignUpload struct {
	storage        outbound.Storage
	assetValidator *services.AssetValidator
	keyGenerator   *services.KeyGenerator
	bucket         string
	expiresIn      time.Duration
}

// NewSignUpload creates a new SignUpload use case
func NewSignUpload(
	storage outbound.Storage,
	assetValidator *services.AssetValidator,
	keyGenerator *services.KeyGenerator,
	bucket string,
	expiresIn time.Duration,
) inbound.SignUpload {
	return &SignUpload{
		storage:        storage,
		assetValidator: assetValidator,
		keyGenerator:   keyGenerator,
		bucket:         bucket,
		expiresIn:      expiresIn,
	}
}

// Execute executes the use case
func (uc *SignUpload) Execute(ctx context.Context, req *inbound.SignUploadRequest) (*inbound.SignUploadResponse, error) {
	// Validate request
	if req.AgencyID == "" {
		return nil, domain.ErrInvalidAgencyID
	}

	if req.Asset == "" {
		return nil, domain.ErrInvalidAsset
	}

	// Reject paths containing '/'
	if strings.Contains(req.Asset, "/") {
		return nil, domain.ErrInvalidAsset
	}

	// Validate asset is in allowed list
	if !uc.assetValidator.IsValidAsset(req.Asset) {
		return nil, domain.ErrInvalidAsset
	}

	// Parse agency ID
	agencyUUID, err := uuid.Parse(req.AgencyID)
	if err != nil {
		return nil, domain.ErrInvalidAgencyID
	}

	// Generate object key
	key := uc.keyGenerator.GenerateObjectKey(agencyUUID, req.Asset)

	// Generate pre-signed URL
	url, headers, err := uc.storage.GeneratePresignedURL(ctx, uc.bucket, key, uc.expiresIn)
	if err != nil {
		return nil, err
	}

	return &inbound.SignUploadResponse{
		URL:     url,
		Method:  "PUT",
		Headers: headers,
		Key:     key,
		Expires: time.Now().Add(uc.expiresIn),
	}, nil
}

