package inbound

import (
	"context"
	"time"
)

// SignUpload is the inbound port for signing file uploads
type SignUpload interface {
	Execute(ctx context.Context, req *SignUploadRequest) (*SignUploadResponse, error)
}

// SignUploadRequest represents the request
type SignUploadRequest struct {
	AgencyID string
	Asset    string
}

// SignUploadResponse represents the response
type SignUploadResponse struct {
	URL     string
	Method  string
	Headers map[string]string
	Key     string
	Expires time.Time
}

