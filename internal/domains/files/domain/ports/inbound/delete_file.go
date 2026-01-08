package inbound

import (
	"context"
)

// DeleteFile is the inbound port for deleting files
type DeleteFile interface {
	Execute(ctx context.Context, req *DeleteFileRequest) (*DeleteFileResponse, error)
}

// DeleteFileRequest represents the request
type DeleteFileRequest struct {
	Key string
}

// DeleteFileResponse represents the response
type DeleteFileResponse struct {
	Success bool
}

