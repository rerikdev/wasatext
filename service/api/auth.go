package api

import (
	"net/http"
	"strings"
)

// extractIdentifier gets the ID from the "Authorization: Bearer <ID>" header
func (rt *Router) extractIdentifier(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(authHeader, "Bearer ")
}