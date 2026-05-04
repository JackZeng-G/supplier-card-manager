package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"supplier-card-system/config"
	"supplier-card-system/handlers"
	"supplier-card-system/middleware"
	"supplier-card-system/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

//go:embed frontend/dist frontend/dist/assets/*
var frontendFS embed.FS

// secureFileServer 创建安全的静态文件服务，防止目录遍历攻击
func secureFileServer(root string) gin.HandlerFunc {
	absRoot, _ := filepath.Abs(root)
	return func(c *gin.Context) {
		reqPath := c.Param("filepath")
		// 清理路径，防止目录遍历
		cleanPath := filepath.Clean(reqPath)
		if strings.Contains(cleanPath, "..") {
			c.JSON(http.StatusForbidden, gin.H{"error": "非法路径"})
			return
		}
		// 拼接并验证最终路径仍在根目录下
		fullPath := filepath.Join(absRoot, cleanPath)
		absPath, err := filepath.Abs(fullPath)
		if err != nil || !strings.HasPrefix(absPath, absRoot) {
			c.JSON(http.StatusForbidden, gin.H{"error": "非法路径"})
			return
		}
		c.File(absPath)
	}
}

func main() {
	// 加载.env文件
	if err := godotenv.Load(); err != nil {
		log.Println("未找到.env文件，使用系统环境变量")
	}

	// 初始化配置
	config.InitConfig()

	// 创建数据目录（包含子目录）
	if err := os.MkdirAll(config.AppConfig.TempUploadPath, 0755); err != nil {
		log.Fatalf("创建临时上传目录失败: %v", err)
	}
	if err := os.MkdirAll(config.AppConfig.ImagePath, 0755); err != nil {
		log.Fatalf("创建图片存储目录失败: %v", err)
	}

	// 初始化数据库
	if err := models.InitDB(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 创建Gin引擎
	r := gin.Default()

	// 使用CORS中间件
	r.Use(middleware.CORS())

	// 使用Gzip压缩中间件
	r.Use(middleware.GzipMiddleware())

	// 静态文件服务（带路径安全验证）
	log.Printf("临时上传目录: %s", config.AppConfig.TempUploadPath)
	r.GET("/uploads/*filepath", secureFileServer(config.AppConfig.TempUploadPath))
	log.Printf("永久图片目录: %s", config.AppConfig.ImagePath)
	r.GET("/images/*filepath", secureFileServer(config.AppConfig.ImagePath))

	// API路由
	api := r.Group("/api")
	{
		// 供应商相关接口
		api.POST("/suppliers/upload", handlers.UploadCard)
		api.POST("/suppliers/upload-back", handlers.UploadCardBack)
		api.GET("/suppliers", handlers.GetSuppliers)
		api.GET("/suppliers/:id", handlers.GetSupplier)
		api.POST("/suppliers", handlers.CreateSupplier)
		api.PUT("/suppliers/:id", handlers.UpdateSupplier)
		api.DELETE("/suppliers/:id", handlers.DeleteSupplier)
		api.GET("/suppliers/stats", handlers.GetSupplierStats)
		api.GET("/suppliers/export", handlers.ExportSuppliers)
	}

	// 健康检查（含数据库连接检测，异常日志3次、恢复1次）
	healthState := &struct {
		failCount int
		isDown    bool
	}{}
	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := models.DB.DB()
		if err != nil || sqlDB.Ping() != nil {
			healthState.failCount++
			if healthState.failCount <= 3 {
				log.Printf("[HEALTH] 数据库连接异常 (第%d次)", healthState.failCount)
			}
			healthState.isDown = true
			c.JSON(503, gin.H{"status": "error", "message": "数据库连接失败"})
			return
		}
		if healthState.isDown {
			log.Printf("[HEALTH] 数据库连接已恢复 (之前连续失败%d次)", healthState.failCount)
			healthState.failCount = 0
			healthState.isDown = false
		} else {
			healthState.failCount = 0
		}
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 前端静态文件服务（嵌入到Go二进制中）
	frontendDist, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Fatalf("获取前端文件失败: %v", err)
	}

	// 处理前端静态文件和SPA路由
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// API请求返回404
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(404, gin.H{"error": "Not Found"})
			return
		}

		// 尝试直接打开文件
		filePath := strings.TrimPrefix(path, "/")
		if filePath == "" {
			filePath = "index.html"
		}

		// 检查文件是否存在
		if _, err := fs.Stat(frontendDist, filePath); err == nil {
			// 文件存在，直接返回
			c.FileFromFS(path, http.FS(frontendDist))
			return
		}

		// 尝试作为 assets 文件
		if strings.HasPrefix(path, "/assets/") {
			assetPath := strings.TrimPrefix(path, "/assets/")
			if _, err := fs.Stat(frontendDist, "assets/"+assetPath); err == nil {
				c.FileFromFS("/assets/"+assetPath, http.FS(frontendDist))
				return
			}
			// assets 文件不存在返回404
			c.Status(404)
			return
		}

		// 其他情况返回 index.html（SPA 路由）
		c.Header("Cache-Control", "no-cache")
		c.FileFromFS("/index.html", http.FS(frontendDist))
	})

	// 启动服务器
	log.Printf("服务器启动在端口 %s", config.AppConfig.ServerPort)
	if err := r.Run(":" + config.AppConfig.ServerPort); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
