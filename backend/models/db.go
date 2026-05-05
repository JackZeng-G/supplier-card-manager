package models

import (
	"log"
	"supplier-card-manager/config"
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

	// 迁移旧运输方式数据: 空→空运, 海→海运, 空/海→空运,海运
	migrateTransportType()

	log.Println("数据库初始化成功")
	return nil
}

func migrateTransportType() {
	migrations := map[string]string{
		"空/海": "空运,海运",
		"空":   "空运",
		"海":   "海运",
	}
	for old, new := range migrations {
		result := DB.Model(&Supplier{}).Where("transport_type = ?", old).Update("transport_type", new)
		if result.RowsAffected > 0 {
			log.Printf("迁移运输方式: %q → %q (%d条)", old, new, result.RowsAffected)
		}
	}
}
