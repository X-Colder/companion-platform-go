// service/order.go
package service

import (
	"errors"

	"github.com/X-Colder/companion-backend/model"
	"github.com/X-Colder/companion-backend/utils"

	"github.com/jinzhu/gorm"
)

// OrderService 订单服务
type OrderService struct{}

// -------------------------- 陪诊师相关业务 --------------------------

// GetUndertakeDemandList 获取订单大厅（待接单需求列表，带分页）
func (o *OrderService) GetUndertakeDemandList(page int, size int) ([]model.Demand, int64, error) {
	var demandList []model.Demand
	var total int64

	// 计算分页偏移量
	offset := (page - 1) * size

	// 查询待接单需求总数（status=0）
	if err := model.DB.Where("status = ?", 0).Model(&model.Demand{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询分页数据，按创建时间倒序
	if err := model.DB.Where("status = ?", 0).Order("created_at DESC").Offset(offset).Limit(size).Find(&demandList).Error; err != nil {
		return nil, 0, err
	}

	return demandList, total, nil
}

// TakeOrder 接单操作（生成订单，更新需求状态）
func (o *OrderService) TakeOrder(demandId uint64, companionId uint64) error {
	// 开启事务（多表操作，保证数据一致性）
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	// 1. 查询需求：必须是待接单状态（status=0），且未被其他陪诊师接单
	var demand model.Demand
	if err := tx.Where("id = ? AND status = ?", demandId, 0).First(&demand).Error; err != nil {
		tx.Rollback()
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("需求不存在或已被接单")
		}
		return errors.New("查询需求失败")
	}

	// 2. 校验陪诊师身份（避免自己接自己的需求，若需求发布者是陪诊师）
	if demand.PatientId == companionId {
		tx.Rollback()
		return errors.New("不能接自己发布的需求")
	}

	// 3. 生成唯一订单编号
	orderNo := utils.GenerateOrderNo()

	// 4. 计算订单金额与陪诊师收入（默认扣除10%平台佣金，可配置）
	orderAmount := utils.KeepTwoDecimal(demand.ExpectedPrice)
	commissionRate := 0.1 // 10%佣金
	companionIncome := utils.KeepTwoDecimal(orderAmount * (1 - commissionRate))

	// 5. 创建订单
	order := model.Order{
		OrderNo:         orderNo,
		DemandId:        demandId,
		PatientId:       demand.PatientId,
		CompanionId:     companionId,
		OrderAmount:     orderAmount,
		CompanionIncome: companionIncome,
		Status:          1, // 1-待服务
	}
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return errors.New("生成订单失败")
	}

	// 6. 更新需求状态（0-待接单 → 1-已接单），并关联订单ID
	if err := tx.Model(&model.Demand{}).Where("id = ?", demandId).Updates(map[string]interface{}{
		"status":   1,
		"order_id": order.ID,
	}).Error; err != nil {
		tx.Rollback()
		return errors.New("更新需求状态失败")
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return errors.New("接单事务提交失败")
	}

	return nil
}

// GetCompanionOrderList 获取陪诊师订单列表（带状态筛选、分页）
func (o *OrderService) GetCompanionOrderList(companionId uint64, status int, page int, size int) ([]model.Order, int64, error) {
	var orderList []model.Order
	var total int64

	// 计算分页偏移量
	offset := (page - 1) * size

	// 构造查询条件
	query := model.DB.Where("companion_id = ?", companionId)
	if status > 0 { // 状态为0时不筛选（查询所有状态）
		query = query.Where("status = ?", status)
	}

	// 查询总数
	if err := query.Model(&model.Order{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询分页数据，按创建时间倒序
	if err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&orderList).Error; err != nil {
		return nil, 0, err
	}

	return orderList, total, nil
}

// CompanionConfirmOrderComplete 陪诊师确认服务完成（仅服务中状态可操作）
func (o *OrderService) CompanionConfirmOrderComplete(orderId uint64, companionId uint64) error {
	// 1. 查询订单：必须是当前陪诊师的订单，且状态为2-服务中
	var order model.Order
	if err := model.DB.Where("id = ? AND companion_id = ? AND status = ?", orderId, companionId, 2).First(&order).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("订单不存在或非服务中状态，无法确认完成")
		}
		return errors.New("查询订单失败")
	}

	// 2. 更新订单状态（2-服务中 → 3-待结算）
	if err := model.DB.Model(&model.Order{}).Where("id = ?", orderId).Update("status", 3).Error; err != nil {
		return errors.New("更新订单状态失败")
	}

	return nil
}

// CompanionCancelOrder 陪诊师取消订单（仅待服务状态可操作）
func (o *OrderService) CompanionCancelOrder(orderId uint64, companionId uint64, reason string) error {
	// 开启事务
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	// 1. 查询订单：当前陪诊师的订单，状态为1-待服务
	var order model.Order
	if err := tx.Where("id = ? AND companion_id = ? AND status = ?", orderId, companionId, 1).First(&order).Error; err != nil {
		tx.Rollback()
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("订单不存在或非待服务状态，无法取消")
		}
		return errors.New("查询订单失败")
	}

	// 2. 更新订单状态（1-待服务 → 5-已取消）
	if err := tx.Model(&model.Order{}).Where("id = ?", orderId).Update("status", 5).Error; err != nil {
		tx.Rollback()
		return errors.New("更新订单状态失败")
	}

	// 3. 更新需求状态（1-已接单 → 0-待接单），清空订单ID
	if err := tx.Model(&model.Demand{}).Where("id = ?", order.DemandId).Updates(map[string]interface{}{
		"status":   0,
		"order_id": 0,
	}).Error; err != nil {
		tx.Rollback()
		return errors.New("更新需求状态失败")
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return errors.New("取消订单事务提交失败")
	}

	return nil
}

// -------------------------- 患者相关业务 --------------------------

// GetPatientOrderList 获取患者订单列表（带状态筛选、分页）
func (o *OrderService) GetPatientOrderList(patientId uint64, status int, page int, size int) ([]model.Order, int64, error) {
	var orderList []model.Order
	var total int64

	// 计算分页偏移量
	offset := (page - 1) * size

	// 构造查询条件
	query := model.DB.Where("patient_id = ?", patientId)
	if status > 0 { // 状态为0时不筛选
		query = query.Where("status = ?", status)
	}

	// 查询总数
	if err := query.Model(&model.Order{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询分页数据，按创建时间倒序
	if err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&orderList).Error; err != nil {
		return nil, 0, err
	}

	return orderList, total, nil
}

// PatientConfirmOrderComplete 患者确认服务完成（触发订单结算，陪诊师收款）
func (o *OrderService) PatientConfirmOrderComplete(orderId uint64, patientId uint64) error {
	// 开启事务（更新订单、更新陪诊师余额、生成收入明细）
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	// 1. 查询订单：当前患者的订单，状态为3-待结算
	var order model.Order
	if err := tx.Where("id = ? AND patient_id = ? AND status = ?", orderId, patientId, 3).First(&order).Error; err != nil {
		tx.Rollback()
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("订单不存在或非待结算状态，无法确认完成")
		}
		return errors.New("查询订单失败")
	}

	// 2. 查询陪诊师信息
	var companion model.User
	if err := tx.Where("id = ?", order.CompanionId).First(&companion).Error; err != nil {
		tx.Rollback()
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("陪诊师不存在")
		}
		return errors.New("查询陪诊师失败")
	}

	// 3. 更新订单状态（3-待结算 → 4-已完成）
	if err := tx.Model(&model.Order{}).Where("id = ?", orderId).Update("status", 4).Error; err != nil {
		tx.Rollback()
		return errors.New("更新订单状态失败")
	}

	// 4. 更新陪诊师余额（累加收入）
	newBalance := utils.KeepTwoDecimal(companion.Balance + order.CompanionIncome)
	if err := tx.Model(&model.User{}).Where("id = ?", order.CompanionId).Update("balance", newBalance).Error; err != nil {
		tx.Rollback()
		return errors.New("更新陪诊师余额失败")
	}

	// 5. 生成余额收入明细
	serialNo := utils.GenerateSerialNo("INC") // INC-收入前缀
	balanceRecord := model.BalanceRecord{
		SerialNo:    serialNo,
		CompanionId: order.CompanionId,
		Type:        1, // 1-服务收入
		Amount:      order.CompanionIncome,
		Remark:      "订单" + order.OrderNo + "服务收入",
	}
	if err := tx.Create(&balanceRecord).Error; err != nil {
		tx.Rollback()
		return errors.New("生成收入明细失败")
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return errors.New("确认完成事务提交失败")
	}

	return nil
}

// PatientCancelOrder 患者取消订单（仅待服务状态可操作）
func (o *OrderService) PatientCancelOrder(orderId uint64, patientId uint64, reason string) error {
	// 开启事务
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	// 1. 查询订单：当前患者的订单，状态为1-待服务
	var order model.Order
	if err := tx.Where("id = ? AND patient_id = ? AND status = ?", orderId, patientId, 1).First(&order).Error; err != nil {
		tx.Rollback()
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("订单不存在或非待服务状态，无法取消")
		}
		return errors.New("查询订单失败")
	}

	// 2. 更新订单状态（1-待服务 → 5-已取消）
	if err := tx.Model(&model.Order{}).Where("id = ?", orderId).Update("status", 5).Error; err != nil {
		tx.Rollback()
		return errors.New("更新订单状态失败")
	}

	// 3. 更新需求状态（1-已接单 → 0-待接单），清空订单ID
	if err := tx.Model(&model.Demand{}).Where("id = ?", order.DemandId).Updates(map[string]interface{}{
		"status":   0,
		"order_id": 0,
	}).Error; err != nil {
		tx.Rollback()
		return errors.New("更新需求状态失败")
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return errors.New("取消订单事务提交失败")
	}

	return nil
}

// -------------------------- 通用辅助方法 --------------------------

// UpdateOrderStatusToServing 订单状态更新为服务中（可在服务开始时调用）
func (o *OrderService) UpdateOrderStatusToServing(orderId uint64) error {
	// 校验订单状态（仅待服务状态可更新为服务中）
	var order model.Order
	if err := model.DB.Where("id = ? AND status = ?", orderId, 1).First(&order).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("订单不存在或非待服务状态，无法开始服务")
		}
		return errors.New("查询订单失败")
	}

	// 更新状态
	if err := model.DB.Model(&model.Order{}).Where("id = ?", orderId).Update("status", 2).Error; err != nil {
		return errors.New("更新订单状态为服务中失败")
	}

	return nil
}
