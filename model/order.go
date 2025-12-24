package model

import (
	"time"
)

// Order 订单实体（对应数据库表：orders）
type Order struct {
	ID               uint64    `gorm:"primary_key;auto_increment" json:"id"`
	OrderNo          string    `gorm:"type:varchar(32);unique_index;not null" json:"order_no"` // 订单编号（唯一）
	DemandId         uint64    `gorm:"not null" json:"demand_id"`                              // 关联需求ID
	PatientId        uint64    `gorm:"not null" json:"patient_id"`                             // 患者ID
	CompanionId      uint64    `gorm:"not null" json:"companion_id"`                           // 陪诊师ID
	OrderAmount      float64   `gorm:"type:decimal(10,2);not null" json:"order_amount"`        // 订单金额（与需求期望价格一致）
	CompanionIncome  float64   `gorm:"type:decimal(10,2);not null" json:"companion_income"`    // 陪诊师实际收入（扣除佣金后）
	Status           int       `gorm:"type:tinyint;default:1;comment:'1-待服务，2-服务中，3-待结算，4-已完成，5-已取消'" json:"status"`
	HasPatientEval   int       `gorm:"type:tinyint;default:0;comment:'0-未评价，1-已评价'" json:"has_patient_eval"`   // 患者是否评价
	HasCompanionEval int       `gorm:"type:tinyint;default:0;comment:'0-未评价，1-已评价'" json:"has_companion_eval"` // 陪诊师是否评价
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        time.Time `gorm:"soft_delete;index" json:"-"` // GORM v1 软删除配置
}

// TableName 指定订单表名
func (o *Order) TableName() string {
	return "orders"
}
