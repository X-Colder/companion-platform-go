package model

import (
	"time"
)

// Demand 陪诊需求实体（对应数据库表：demands）
type Demand struct {
	ID             uint64    `gorm:"primary_key;auto_increment" json:"id"`
	PatientId      uint64    `gorm:"not null" json:"patient_id"`                 // 患者ID
	Hospital       string    `gorm:"type:varchar(100);not null" json:"hospital"` // 就诊医院
	HospitalAddr   string    `gorm:"type:varchar(255);not null" json:"hospital_addr"`
	ServiceTime    time.Time `gorm:"not null" json:"service_time"`                      // 服务时间
	ExpectedPrice  float64   `gorm:"type:decimal(10,2);not null" json:"expected_price"` // 期望价格
	ServiceContent string    `gorm:"type:text;not null" json:"service_content"`         // 服务内容
	ContactName    string    `gorm:"type:varchar(16);not null" json:"contact_name"`
	ContactPhone   string    `gorm:"type:varchar(11);not null" json:"contact_phone"`
	Status         int       `gorm:"type:tinyint;default:0;comment:'0-待接单，1-已接单，2-待服务，3-服务中，4-已完成，5-已取消'" json:"status"`
	OrderId        uint64    `gorm:"default:0" json:"order_id"` // 关联订单ID（接单后生成）
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      time.Time `gorm:"soft_delete;index" json:"-"` // GORM v1 软删除配置
}

func (d *Demand) TableName() string {
	return "demands"
}
