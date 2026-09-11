package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldBypassEmbeddedFrontend(t *testing.T) {
	t.Parallel()

	bypass := []string{
		"/api/v1/auth/login",
		"/v1/chat/completions",
		"/v1beta/models",
		"/backend-api/conversation",
		"/antigravity/foo",
		"/setup/status",
		"/health",
		"/models",
		"/responses",
		"/responses/abc",
		"/alpha/search",
		"/images/generations",
		"/videos/generations",
	}
	for _, path := range bypass {
		require.True(t, shouldBypassEmbeddedFrontend(path), "path=%s", path)
	}

	serve := []string{"/", "/login", "/register", "/dashboard", "/setup", "/favicon.ico"}
	for _, path := range serve {
		require.False(t, shouldBypassEmbeddedFrontend(path), "path=%s", path)
	}
}

func TestResolveFrontendRedirectTarget(t *testing.T) {
	t.Parallel()

	mustURL := func(raw string) *url.URL {
		t.Helper()
		parsed, err := url.Parse(raw)
		require.NoError(t, err)
		return parsed
	}

	t.Run("joins_path_and_query", func(t *testing.T) {
		t.Parallel()
		target, ok := resolveFrontendRedirectTarget("http://127.0.0.1:3000", mustURL("/login?next=/dashboard"))
		require.True(t, ok)
		assert.Equal(t, "http://127.0.0.1:3000/login?next=/dashboard", target)
	})

	t.Run("keeps_configured_base_path", func(t *testing.T) {
		t.Parallel()
		target, ok := resolveFrontendRedirectTarget("https://example.com/app", mustURL("/login"))
		require.True(t, ok)
		assert.Equal(t, "https://example.com/app/login", target)
	})

	t.Run("rejects_empty_or_invalid_base", func(t *testing.T) {
		t.Parallel()
		_, ok := resolveFrontendRedirectTarget("", mustURL("/login"))
		assert.False(t, ok)
		_, ok = resolveFrontendRedirectTarget("/relative", mustURL("/login"))
		assert.False(t, ok)
		_, ok = resolveFrontendRedirectTarget("javascript:alert(1)", mustURL("/login"))
		assert.False(t, ok)
	})

	t.Run("rejects_protocol_relative_request", func(t *testing.T) {
		t.Parallel()
		_, ok := resolveFrontendRedirectTarget("http://127.0.0.1:3000", mustURL("//evil.example/phish"))
		assert.False(t, ok)
	})
}

func TestNonEmbeddedFrontendFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newRouter := func(frontendURL string) *gin.Engine {
		r := gin.New()
		r.GET("/health", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})
		r.NoRoute(NonEmbeddedFrontendFallback(func(*gin.Context) string {
			return frontendURL
		}))
		return r
	}

	t.Run("login_without_frontend_url_returns_html_help", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/login", nil)
		newRouter("").ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
		assert.Contains(t, w.Body.String(), "前端未挂在这个端口")
		assert.Contains(t, w.Body.String(), "http://localhost:3000/login")
		assert.NotEqual(t, ginDefaultNotFound, w.Body.String())
	})

	t.Run("login_with_frontend_url_redirects", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/login?next=/dashboard", nil)
		newRouter("http://127.0.0.1:3000").ServeHTTP(w, req)

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "http://127.0.0.1:3000/login?next=/dashboard", w.Header().Get("Location"))
	})

	t.Run("unknown_api_keeps_gin_404_body", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/does-not-exist", nil)
		newRouter("http://127.0.0.1:3000").ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, ginDefaultNotFound, w.Body.String())
		assert.Empty(t, w.Header().Get("Location"))
	})

	t.Run("post_login_keeps_gin_404_body", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		newRouter("http://127.0.0.1:3000").ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, ginDefaultNotFound, w.Body.String())
	})
}
