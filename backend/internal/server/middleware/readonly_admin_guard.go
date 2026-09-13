package middleware

import (
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// ReadOnlyAdminGuard lets read-only admins inspect admin data but blocks writes
// and GET endpoints that return OAuth tokens, API keys, or other secrets.
func ReadOnlyAdminGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := GetUserRoleFromContext(c)
		if !ok || role != service.RoleReadonly {
			c.Next()
			return
		}

		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			if isReadOnlyAdminSecretPath(c.Request.URL.Path) {
				AbortWithError(c, http.StatusForbidden, "READONLY_FORBIDDEN", "Read-only administrators cannot access credentials or export secrets")
				return
			}
			c.Next()
		default:
			if isReadOnlyAdminAllowedWritePath(c.Request.URL.Path) {
				c.Next()
				return
			}
			AbortWithError(c, http.StatusForbidden, "READONLY_FORBIDDEN", "Read-only administrators cannot modify data")
		}
	}
}

func isReadOnlyAdminSecretPath(path string) bool {
	path = strings.TrimSuffix(path, "/")
	switch {
	case strings.HasSuffix(path, "/accounts/data"),
		strings.HasSuffix(path, "/proxies/data"):
		return true
	case strings.HasSuffix(path, "/download-url") && strings.Contains(path, "/backups/"):
		return true
	case strings.HasSuffix(path, "/api-keys") && (strings.Contains(path, "/users/") || strings.Contains(path, "/groups/")):
		return true
	default:
		return false
	}
}

func isReadOnlyAdminAllowedWritePath(path string) bool {
	path = strings.TrimSuffix(path, "/")
	return strings.HasSuffix(path, "/admin/compliance/accept")
}

func IsReadOnlyAdminRequest(c *gin.Context) bool {
	role, ok := GetUserRoleFromContext(c)
	return ok && role == service.RoleReadonly
}
