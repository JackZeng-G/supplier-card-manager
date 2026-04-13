package models

import (
	"log"
	"supplier-card-system/config"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	var err error
	DB, err = gorm.Open(sqlite.Open(config.AppConfig.DatabasePath), &gorm.Config{})
	if err != nil {
		return err
	}

	// 配置数据库连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(25)                  // 最大打开连接数
	sqlDB.SetMaxIdleConns(5)                   // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(10 * time.Minute) // 连接最大生命周期
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)  // 空闲连接最大生命周期

	// 自动迁移数据库表结构
	err = DB.AutoMigrate(&Supplier{})
	if err != nil {
		return err
	}

	log.Println("数据库初始化成功")
	return nil
}
