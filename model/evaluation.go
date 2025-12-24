package model

import (
	"time"
)

// Evaluation 评价实体（对应数据库表：evaluations）
type Evaluation struct {
	ID         uint64    `gorm:"primary_key;auto_increment" json:"id"`
	RelateId   uint64    `gorm:"not null" json:"relate_id"`                    // 关联订单ID（评价对象）
	FromUserId uint64    `gorm:"not null" json:"from_user_id"`                 // 评价人ID
	ToUserId   uint64    `gorm:"not null" json:"to_user_id"`                   // 被评价人ID
	Score      int       `gorm:"type:tinyint;not null" json:"score"`           // 评分（1-5星）
	Content    string    `gorm:"type:text;not null" json:"content"`            // 评价内容
	ImgUrls    string    `gorm:"type:varchar(512);default:''" json:"img_urls"` // 评价图片地址（逗号分隔，多个图片）
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  time.Time `gorm:"soft_delete;index" json:"-"` // GORM v1 软删除配置
}

// TableName 指定评价表名
func (e *Evaluation) TableName() string {
	return "evaluations"
}
