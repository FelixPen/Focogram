package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Userid      string `gorm:"primaryKey;type:varchar(36);not null;unique" json:"userid" validate:"required"`
	Username    string `gorm:"type:varchar(50);not null" json:"username" validate:"required,min=1,max=15"`
	Email       string `gorm:"type:varchar(100);not null;uniqueIndex" json:"email" validate:"required,email"`
	Password    string `gorm:"type:varchar(255);not null" json:"-" validate:"required,min=6"`
	Gender      string `gorm:"type:varchar(10);default:'unknown'" json:"gender" validate:"oneof=男 女 未知"`
	Age         int    `gorm:"type:int;default:0" json:"age" validate:"min=0,max=150"`
	BirthDate   string `gorm:"type:varchar(20);default:''" json:"birthDate"`
	Describe    string `gorm:"type:varchar(500)" json:"describe" validate:"max=100"`
	Address     string `gorm:"type:varchar(200);default:''" json:"address" validate:"max=50"`
	AvatarColor string `gorm:"type:varchar(200);default:''" json:"avatarColor"`
	AvatarUrl   string `gorm:"type:varchar(500);default:''" json:"avatarUrl"`
	BannerColor string `gorm:"type:varchar(200);default:''" json:"bannerColor"`
	PostNum     int    `gorm:"default:0" json:"postnum"`
	Posts       []Post `gorm:"-"`
}
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=1,max=15"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Gender   string `json:"gender" validate:"omitempty,oneof=男 女 未知"`
	Age      int    `json:"age" validate:"omitempty,min=0,max=150"`
}
type UpdatePasswordRequest struct {
	OldPassword string `json:"oldpassword" validate:"required"`       // 旧密码，用于身份二次验证
	NewPassword string `json:"newpassword" validate:"required,min=6"` // 新密码，至少6位
}

//增加贴文数量加1
