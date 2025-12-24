package model

import (
	"time"
)

// BalanceRecord 余额明细实体（对应数据库表：balance_records）
type BalanceRecord struct {
	ID          uint64    `gorm:"primary_key;auto_increment" json:"id"`
	SerialNo    string    `gorm:"type:varchar(32);unique_index;not null" json:"serial_no"` // 明细编号（唯一）
	CompanionId uint64    `gorm:"not null" json:"companion_id"`                            // 陪诊师ID（仅陪诊师有余额）
	Type        int       `gorm:"type:tinyint;not null;comment:'1-服务收入，2-提现成功，3-提现失败'" json:"type"`
	Amount      float64   `gorm:"type:decimal(10,2);not null" json:"amount"`            // 金额（收入为正，提现为负）
	Remark      string    `gorm:"type:varchar(255);default:''" json:"remark"`           // 明细备注（如“订单XXX收入”“提现至微信”）
	CreateTime  time.Time `gorm:"autoCreateTime;column:create_time" json:"create_time"` // 发生时间（字段名与SQL一致）
	DeletedAt   time.Time `gorm:"soft_delete;index" json:"-"`                           // GORM v1 软删除配置
}

// TableName 指定余额明细表名
func (br *BalanceRecord) TableName() string {
	return "balance_records"
}
