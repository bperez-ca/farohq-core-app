package httpserver

import (
	"net/http"
	"strings"
)

// Role names (lowercase) per Strategic Roadmap: Owner, Admin, Manager, Staff, Client Viewer.
const (
	RoleOwner        = "owner"
	RoleAdmin        = "admin"
	RoleManager      = "manager"
	RoleStaff        = "staff"
	RoleViewer       = "viewer"
	RoleClientViewer = "client_viewer"
)

// GetRoleFromContext returns the normalized role from request context (set by RequireAuth).
// Clerk may send "org:admin" or "admin"; we normalize to lowercase and strip "org:" prefix.
// Returns empty string if no role in context.
func GetRoleFromContext(r *http.Request) string {
	v := r.Context().Value("org_role")
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	s = strings.TrimSpace(strings.ToLower(s))
	if strings.HasPrefix(s, "org:") {
		s = strings.TrimPrefix(s, "org:")
	}
	return s
}

// GetAgencyIDFromContext returns the agency (tenant) ID from request context (set by RequireAuth).
// This is the same as org_id in Clerk terminology.
func GetAgencyIDFromContext(r *http.Request) (string, bool) {
	v := r.Context().Value("agency_id")
	if v == nil {
		v = r.Context().Value("org_id")
	}
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	if !ok {
		return "", false
	}
	return strings.TrimSpace(s), s != ""
}

// RequireRole returns middleware that allows only the given roles to proceed.
// If the request has no role or role is not in allowedRoles, responds with 403.
// Should be used after RequireAuth and TenantResolutionWithAuth.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, r := range allowedRoles {
		allowed[strings.ToLower(strings.TrimSpace(r))] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := GetRoleFromContext(r)
			if role == "" {
				http.Error(w, "Forbidden: no role in context", http.StatusForbidden)
				return
			}
			if _, ok := allowed[role]; !ok {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireOwnerOrAdmin is a convenience middleware that allows only owner or admin.
func RequireOwnerOrAdmin(next http.Handler) http.Handler {
	return RequireRole(RoleOwner, RoleAdmin)(next)
}
