package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/gzip"
)

// GzipMiddleware 压缩中间件（排除图片等已压缩文件）
func GzipMiddleware() gin.HandlerFunc {
	return gzip.Gzip(
		gzip.DefaultCompression,
		gzip.WithExcludedPaths([]string{"/uploads/", "/images/"}),
	)
}
