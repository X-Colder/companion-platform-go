package model

import (
	"time"
)

// User 用户实体（对应数据库表：users）
type User struct {
	ID        uint64    `gorm:"primary_key;auto_increment" json:"id"`
	Phone     string    `gorm:"type:varchar(11);unique_index;not null" json:"phone"` // 手机号（唯一）
	Password  string    `gorm:"type:varchar(64);not null" json:"-"`                  // 密码（不返回给前端）
	Nickname  string    `gorm:"type:varchar(16);default:'未设置昵称'" json:"nickname"`
	Avatar    string    `gorm:"type:varchar(255);default:''" json:"avatar"` // 头像地址
	UserType  int       `gorm:"type:tinyint;default:1;comment:'1-患者/家属，2-陪诊师'" json:"user_type"`
	IsAuth    int       `gorm:"type:tinyint;default:0;comment:'0-未实名认证，1-已实名认证'" json:"is_auth"`
	Balance   float64   `gorm:"type:decimal(10,2);default:0.00" json:"balance"` // 账户余额（仅陪诊师有效）
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt time.Time `gorm:"soft_delete;index" json:"-"` // GORM v1 软删除：time.Time + soft_delete 标签
}

// TableName 指定表名
func (u *User) TableName() string {
	return "users"
}
