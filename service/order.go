package service

import (
	"time"

	"github.com/X-Colder/companion-backend/conf"
	"github.com/X-Colder/companion-backend/model"
)

// OrderService 订单业务层
type OrderService struct{}

// SettleOrder 订单结算（平台抽佣10%）
func (o *OrderService) SettleOrder(orderNo string) (bool, string) {
	// 开启事务
	tx := conf.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 查询订单
	var order model.Order
	if err := tx.Where("order_no = ? AND order_status = ?", orderNo, 3).First(&order).Error; err != nil {
		tx.Rollback()
		return false, "订单不存在或未达到结算条件：" + err.Error()
	}

	// 2. 计算金额（10%佣金）
	platformCommission := order.OrderAmount * 0.1
	companionIncome := order.OrderAmount - platformCommission

	// 3. 更新订单状态（已结算）
	order.OrderStatus = 4
	order.PayStatus = 2
	order.SettleTime = time.Now()
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return false, "更新订单状态失败：" + err.Error()
	}

	// 4. 增加陪诊师余额
	var companion model.User
	if err := tx.Where("id = ?", order.CompanionId).First(&companion).Error; err != nil {
		tx.Rollback()
		return false, "陪诊师不存在：" + err.Error()
	}
	companion.Balance += companionIncome
	if err := tx.Save(&companion).Error; err != nil {
		tx.Rollback()
		return false, "更新陪诊师余额失败：" + err.Error()
	}

	// 5. 记录资金流水
	// 患者支付流水（已在支付时记录，此处记录平台佣金和陪诊师收入）
	// 平台佣金流水
	platformFlow := model.FundFlow{
		UserId:     0, // 平台虚拟用户ID（可自定义）
		OrderNo:    orderNo,
		Amount:     platformCommission,
		FlowType:   2,
		Status:     1,
		Remark:     "订单佣金收入",
		CreateTime: time.Now(),
	}
	if err := tx.Create(&platformFlow).Error; err != nil {
		tx.Rollback()
		return false, "记录平台佣金流水失败：" + err.Error()
	}

	// 陪诊师收入流水
	companionFlow := model.FundFlow{
		UserId:     order.CompanionId,
		OrderNo:    orderNo,
		Amount:     companionIncome,
		FlowType:   3,
		Status:     1,
		Remark:     "订单结算收入",
		CreateTime: time.Now(),
	}
	if err := tx.Create(&companionFlow).Error; err != nil {
		tx.Rollback()
		return false, "记录陪诊师流水失败：" + err.Error()
	}

	// 提交事务
	tx.Commit()
	return true, "订单结算成功"
}
