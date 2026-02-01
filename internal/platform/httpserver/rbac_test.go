package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRoleFromContext(t *testing.T) {
	tests := []struct {
		name     string
		orgRole  interface{}
		wantRole string
	}{
		{"nil", nil, ""},
		{"empty", "", ""},
		{"admin", "admin", "admin"},
		{"owner", "owner", "owner"},
		{"org:admin", "org:admin", "admin"},
		{"org:owner", "org:owner", "owner"},
		{"Staff", "Staff", "staff"},
		{"client_viewer", "client_viewer", "client_viewer"},
		{"not_string", 123, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.orgRole != nil {
				ctx = context.WithValue(ctx, "org_role", tt.orgRole)
			}
			req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
			got := GetRoleFromContext(req)
			assert.Equal(t, tt.wantRole, got)
		})
	}
}

func TestGetAgencyIDFromContext(t *testing.T) {
	tests := []struct {
		name        string
		agencyID    interface{}
		orgID       interface{}
		wantID      string
		wantPresent bool
	}{
		{"none", nil, nil, "", false},
		{"agency_id", "agency-uuid-1", nil, "agency-uuid-1", true},
		{"org_id_only", nil, "org-uuid-2", "org-uuid-2", true},
		{"agency_preferred", "agency-uuid-1", "org-uuid-2", "agency-uuid-1", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.agencyID != nil {
				ctx = context.WithValue(ctx, "agency_id", tt.agencyID)
			}
			if tt.orgID != nil {
				ctx = context.WithValue(ctx, "org_id", tt.orgID)
			}
			req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
			got, ok := GetAgencyIDFromContext(req)
			assert.Equal(t, tt.wantPresent, ok)
			assert.Equal(t, tt.wantID, got)
		})
	}
}

func TestRequireRole(t *testing.T) {
	nextOK := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name         string
		roleInCtx    interface{}
		allowedRoles []string
		wantStatus   int
	}{
		{"owner_allowed", "owner", []string{"owner", "admin"}, http.StatusOK},
		{"admin_allowed", "admin", []string{"owner", "admin"}, http.StatusOK},
		{"viewer_forbidden", "viewer", []string{"owner", "admin"}, http.StatusForbidden},
		{"no_role_forbidden", nil, []string{"owner", "admin"}, http.StatusForbidden},
		{"staff_allowed", "staff", []string{"owner", "admin", "staff"}, http.StatusOK},
		{"org_admin_normalized", "org:admin", []string{"admin"}, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.roleInCtx != nil {
				ctx = context.WithValue(ctx, "org_role", tt.roleInCtx)
			}
			req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
			rec := httptest.NewRecorder()
			mw := RequireRole(tt.allowedRoles...)
			mw(nextOK).ServeHTTP(rec, req)
			require.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestRequireOwnerOrAdmin(t *testing.T) {
	nextOK := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name       string
		roleInCtx  interface{}
		wantStatus int
	}{
		{"owner_ok", "owner", http.StatusOK},
		{"admin_ok", "admin", http.StatusOK},
		{"staff_forbidden", "staff", http.StatusForbidden},
		{"viewer_forbidden", "viewer", http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.roleInCtx != nil {
				ctx = context.WithValue(ctx, "org_role", tt.roleInCtx)
			}
			req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
			rec := httptest.NewRecorder()
			RequireOwnerOrAdmin(nextOK).ServeHTTP(rec, req)
			require.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
