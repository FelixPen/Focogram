package models

import (
	"gorm.io/gorm"
)

type Follow struct {
	gorm.Model
	Followerid string `gorm:"type:varchar(36);not null" json:"followerid"`
	Followedid string `gorm:"type:varchar(36);not null" json:"followedid"`
}

// TableName 定义数据库表名
func (Follow) TableName() string {
	return "follows"
}

// Migration 处理表迁移和索引创建（兼容MySQL低版本）
func (Follow) Migration(db *gorm.DB) error {
	// 1. 自动迁移表结构（创建表或更新字段）
	if err := db.AutoMigrate(&Follow{}); err != nil {
		return err
	}

	// 2. 检查联合唯一索引是否已存在（兼容MySQL 5.7及以下版本）
	var indexCount int64
	err := db.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.statistics 
		WHERE table_schema = DATABASE() 
		  AND table_name = 'follows' 
		  AND index_name = 'idx_unique_follow'
	`).Scan(&indexCount).Error
	if err != nil {
		return err
	}

	// 3. 索引不存在时才创建（避免重复创建报错）
	if indexCount == 0 {
		err = db.Exec(`
			CREATE UNIQUE INDEX idx_unique_follow 
			ON follows (followerid, followedid)
		`).Error
		if err != nil {
			return err
		}
	}

	return nil
}
