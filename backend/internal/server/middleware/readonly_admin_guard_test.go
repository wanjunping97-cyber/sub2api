//go:build unit

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestReadOnlyAdminGuard_AllowsReadsAndBlocksWrites(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(ContextKeyUserRole), service.RoleReadonly)
		c.Next()
	})
	r.Use(ReadOnlyAdminGuard())
	r.GET("/api/v1/admin/accounts", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.POST("/api/v1/admin/accounts", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/accounts", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
	require.Contains(t, w.Body.String(), "READONLY_FORBIDDEN")
}

func TestReadOnlyAdminGuard_BlocksSecretGets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(ContextKeyUserRole), service.RoleReadonly)
		c.Next()
	})
	r.Use(ReadOnlyAdminGuard())
	r.GET("/api/v1/admin/accounts/data", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/api/v1/admin/proxies/data", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/api/v1/admin/users/8/api-keys", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/api/v1/admin/groups/3/api-keys", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/api/v1/admin/backups/9/download-url", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/api/v1/admin/accounts/1", func(c *gin.Context) { c.Status(http.StatusOK) })

	for _, path := range []string{
		"/api/v1/admin/accounts/data",
		"/api/v1/admin/proxies/data",
		"/api/v1/admin/users/8/api-keys",
		"/api/v1/admin/groups/3/api-keys",
		"/api/v1/admin/backups/9/download-url",
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusForbidden, w.Code, path)
		require.Contains(t, w.Body.String(), "READONLY_FORBIDDEN", path)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/accounts/1", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestReadOnlyAdminGuard_AllowsComplianceAccept(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(ContextKeyUserRole), service.RoleReadonly)
		c.Next()
	})
	r.Use(ReadOnlyAdminGuard())
	r.POST("/api/v1/admin/compliance/accept", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/compliance/accept", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestReadOnlyAdminGuard_SkipsFullAdmins(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(ContextKeyUserRole), service.RoleAdmin)
		c.Next()
	})
	r.Use(ReadOnlyAdminGuard())
	r.POST("/api/v1/admin/accounts", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/api/v1/admin/accounts/data", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/accounts/data", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}
