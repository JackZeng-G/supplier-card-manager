package middleware

import (
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	// 从环境变量获取允许的源，默认为本地开发端口
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:5173",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:3001",
		"http://127.0.0.1:5173",
	}

	// 支持通过环境变量添加额外的允许源
	if extraOrigin := os.Getenv("ALLOWED_ORIGIN"); extraOrigin != "" {
		allowedOrigins = append(allowedOrigins, extraOrigin)
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
