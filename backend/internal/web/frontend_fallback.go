package web

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

const ginDefaultNotFound = "404 page not found"

const missingFrontendHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Frontend not embedded</title>
  <style>
    body { font-family: sans-serif; max-width: 40rem; margin: 3rem auto; padding: 0 1.25rem; line-height: 1.5; color: #111; }
    code { background: #f3f4f6; padding: 0.1rem 0.35rem; border-radius: 4px; }
    a { color: #4f46e5; }
  </style>
</head>
<body>
  <h1>前端未挂在这个端口</h1>
  <p>当前进程是 API 服务，没有嵌入前端（未使用 <code>-tags embed</code> 编译）。浏览器访问 <code>/</code> 或 <code>/login</code> 时不会出现登录页。</p>
  <p>开发环境请打开 Vite 前端（默认 <a href="http://localhost:3000/login">http://localhost:3000/login</a>），或设置 <code>server.frontend_url</code> / <code>SERVER_FRONTEND_URL</code> 把页面请求重定向到前端。</p>
  <p>生产环境请用 <code>go build -tags embed</code> 重新编译，把前端打进二进制。</p>
</body>
</html>
`

// NonEmbeddedFrontendFallback handles unmatched browser routes when the UI is
// not compiled into the binary. API-shaped paths keep Gin's default 404 body.
func NonEmbeddedFrontendFallback(resolveFrontendURL func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldBypassEmbeddedFrontend(c.Request.URL.Path) {
			c.String(http.StatusNotFound, ginDefaultNotFound)
			return
		}

		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.String(http.StatusNotFound, ginDefaultNotFound)
			return
		}

		frontendURL := ""
		if resolveFrontendURL != nil {
			frontendURL = resolveFrontendURL(c)
		}
		if target, ok := resolveFrontendRedirectTarget(frontendURL, c.Request.URL); ok {
			c.Redirect(http.StatusFound, target)
			return
		}

		c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(missingFrontendHTML))
	}
}

func resolveFrontendRedirectTarget(base string, reqURL *url.URL) (string, bool) {
	if reqURL == nil {
		return "", false
	}

	parsedBase, err := url.Parse(strings.TrimSpace(base))
	if err != nil || parsedBase.User != nil || parsedBase.RawQuery != "" || parsedBase.Fragment != "" {
		return "", false
	}
	if parsedBase.Scheme != "http" && parsedBase.Scheme != "https" {
		return "", false
	}
	if parsedBase.Host == "" {
		return "", false
	}

	if reqURL.IsAbs() || reqURL.Host != "" || reqURL.Scheme != "" || reqURL.User != nil {
		return "", false
	}

	path := reqURL.EscapedPath()
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		return "", false
	}

	parsedBase.Path = strings.TrimRight(parsedBase.Path, "/") + path
	parsedBase.RawQuery = reqURL.RawQuery
	parsedBase.Fragment = ""
	return parsedBase.String(), true
}
