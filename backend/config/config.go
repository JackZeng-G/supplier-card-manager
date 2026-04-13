package config

import (
	"log"
	"os"
)

type Config struct {
	ServerPort       string
	DatabasePath     string
	TempUploadPath   string // 临时上传目录
	ImagePath        string // 永久图片存储目录
	TencentSecretID  string
	TencentSecretKey string
}

var AppConfig *Config

func InitConfig() {
	secretID := getEnv("TENCENT_SECRET_ID", "")
	secretKey := getEnv("TENCENT_SECRET_KEY", "")

	if secretID == "" || secretKey == "" {
		log.Println("警告: 未配置腾讯云OCR密钥(TENCENT_SECRET_ID/TENCENT_SECRET_KEY)，OCR功能将不可用")
	}

	AppConfig = &Config{
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		DatabasePath:     getEnv("DATABASE_PATH", "./data/suppliers.db"),
		TempUploadPath:   getEnv("TEMP_UPLOAD_PATH", "./data/uploads"),
		ImagePath:        getEnv("IMAGE_PATH", "./data/images"),
		TencentSecretID:  secretID,
		TencentSecretKey: secretKey,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
