package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"farohq-core-app/internal/domains/files/domain"
	"farohq-core-app/internal/domains/files/domain/ports/inbound"
)

// Handlers provides HTTP handlers for the files domain
type Handlers struct {
	logger     zerolog.Logger
	signUpload inbound.SignUpload
	deleteFile inbound.DeleteFile
}

// NewHandlers creates new files HTTP handlers
func NewHandlers(
	logger zerolog.Logger,
	signUpload inbound.SignUpload,
	deleteFile inbound.DeleteFile,
) *Handlers {
	return &Handlers{
		logger:     logger,
		signUpload: signUpload,
		deleteFile: deleteFile,
	}
}

// ListFilesHandler handles GET /api/v1/files
// Returns an empty list as file listing is not yet implemented
func (h *Handlers) ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement file listing use case
	// For now, return empty list to satisfy API contract
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

// SignHandler handles POST /api/v1/files/sign
func (h *Handlers) SignHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AgencyID string `json:"agency_id"`
		Asset    string `json:"asset"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	signReq := &inbound.SignUploadRequest{
		AgencyID: req.AgencyID,
		Asset:    req.Asset,
	}

	resp, err := h.signUpload.Execute(r.Context(), signReq)
	if err != nil {
		if err == domain.ErrInvalidAgencyID || err == domain.ErrInvalidAsset {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to generate presigned URL")
		http.Error(w, "Failed to generate presigned URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeleteFileHandler handles DELETE /api/v1/files/{key}
func (h *Handlers) DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		http.Error(w, "file key is required", http.StatusBadRequest)
		return
	}

	deleteReq := &inbound.DeleteFileRequest{
		Key: key,
	}

	resp, err := h.deleteFile.Execute(r.Context(), deleteReq)
	if err != nil {
		if err == domain.ErrFileNotFound {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to delete file")
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": resp.Success,
	})
}
