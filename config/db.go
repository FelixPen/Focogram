package config

import (
	"Focogram/global"
	"Focogram/models"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 初始化数据库连接
func initDB() {
	dsn := AppConfig.Database.Dsn
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database, got error: %v", err)
	}
	sqlDB, err := db.DB()

	// SetMaxIdleConns 用于设置连接池中空闲连接的最大数量。
	sqlDB.SetMaxIdleConns(AppConfig.Database.Max_idle_conns)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(AppConfig.Database.Max_open_conns)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)
	if err != nil {
		log.Fatalf("failed to connect database, got error: %v", err)
	}
	global.Db = db

	// 禁用外键检查，删除所有错误的外键（MySQL 不支持 IF EXISTS，忽略报错）
	_ = global.Db.Exec("SET FOREIGN_KEY_CHECKS=0").Error
	_ = global.Db.Exec("ALTER TABLE users DROP FOREIGN KEY fk_posts_user").Error
	_ = global.Db.Exec("ALTER TABLE users DROP INDEX fk_posts_user").Error

	if err := global.Db.AutoMigrate(&models.User{}); err != nil {
		_ = global.Db.Exec("SET FOREIGN_KEY_CHECKS=1").Error
		log.Fatal("数据库迁移失败:", err)
	}
	if err := global.Db.AutoMigrate(&models.Post{}); err != nil {
		_ = global.Db.Exec("SET FOREIGN_KEY_CHECKS=1").Error
		log.Fatal("数据库迁移失败:", err)
	}
	if err := global.Db.AutoMigrate(&models.Like{}); err != nil {
		_ = global.Db.Exec("SET FOREIGN_KEY_CHECKS=1").Error
		log.Fatal("数据库迁移失败:", err)
	}
	if err := global.Db.AutoMigrate(&models.Comment{}); err != nil {
		_ = global.Db.Exec("SET FOREIGN_KEY_CHECKS=1").Error
		log.Fatal("数据库迁移失败:", err)
	}
	if err := global.Db.AutoMigrate(&models.Notification{}); err != nil {
		_ = global.Db.Exec("SET FOREIGN_KEY_CHECKS=1").Error
		log.Fatal("数据库迁移失败:", err)
	}
	// 迁移对话表
	if err := global.Db.AutoMigrate(&models.Conversation{}); err != nil {
		_ = global.Db.Exec("SET FOREIGN_KEY_CHECKS=1").Error
		log.Fatal("数据库迁移失败 (Conversation):", err)
	}

	// 迁移私信表
	if err := global.Db.AutoMigrate(&models.PrivateMessage{}); err != nil {
		_ = global.Db.Exec("SET FOREIGN_KEY_CHECKS=1").Error
		log.Fatal("数据库迁移失败 (PrivateMessage):", err)
	}
	// 迁移关注表（使用自定义迁移方法）
	follow := models.Follow{}
	if err := follow.Migration(global.Db); err != nil {
		log.Fatal("关注表迁移失败:", err)
	}

	// 所有迁移完成后，恢复外键检查
	_ = global.Db.Exec("SET FOREIGN_KEY_CHECKS=1").Error

}
