// service/balance.go
package service

import (
	"errors"
	"strconv"

	"github.com/X-Colder/companion-backend/model"
	"github.com/X-Colder/companion-backend/utils"
	"github.com/jinzhu/gorm"
)

// BalanceService 余额服务
type BalanceService struct{}

// GetCompanionBalance 查询陪诊师账户余额
func (b *BalanceService) GetCompanionBalance(companionId uint64) (float64, error) {
	// 1. 查询陪诊师用户信息
	var companion model.User
	if err := model.DB.Where("id = ? AND user_type = ?", companionId, 2).First(&companion).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return 0, errors.New("非陪诊师账户，无余额信息")
		}
		return 0, errors.New("查询用户信息失败")
	}

	// 2. 返回保留两位小数的余额
	return utils.KeepTwoDecimal(companion.Balance), nil
}

// GetCompanionBalanceRecordList 查询陪诊师余额明细（带分页、类型筛选）
func (b *BalanceService) GetCompanionBalanceRecordList(companionId uint64, recordType int, page int, size int) ([]model.BalanceRecord, int64, error) {
	var recordList []model.BalanceRecord
	var total int64

	// 1. 计算分页偏移量
	offset := (page - 1) * size

	// 2. 构造查询条件
	query := model.DB.Where("companion_id = ?", companionId)
	if recordType > 0 { // 筛选指定类型（1-收入，2-提现成功，3-提现失败）
		query = query.Where("type = ?", recordType)
	}

	// 3. 查询明细总数
	if err := query.Model(&model.BalanceRecord{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 4. 查询分页明细（按创建时间倒序，最新明细优先）
	if err := query.Order("create_time DESC").Offset(offset).Limit(size).Find(&recordList).Error; err != nil {
		return nil, 0, err
	}

	return recordList, total, nil
}

// ApplyWithdraw 陪诊师申请提现（事务处理：扣减余额+生成提现明细）
func (b *BalanceService) ApplyWithdraw(companionId uint64, amount float64, account string, realName string) (string, error) {
	// 1. 格式化提现金额（保留两位小数）
	withdrawAmount := utils.KeepTwoDecimal(amount)

	// 开启事务（多表操作：查询余额+扣减余额+生成明细）
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return "", err
	}

	// 2. 查询陪诊师信息与当前余额
	var companion model.User
	if err := tx.Where("id = ? AND user_type = ?", companionId, 2).First(&companion).Error; err != nil {
		tx.Rollback()
		if gorm.IsRecordNotFoundError(err) {
			return "", errors.New("非陪诊师账户，无法提现")
		}
		return "", errors.New("查询陪诊师信息失败")
	}

	// 3. 校验余额是否充足
	currentBalance := utils.KeepTwoDecimal(companion.Balance)
	if currentBalance < withdrawAmount {
		tx.Rollback()
		return "", errors.New("账户余额不足，当前余额：" + strconv.FormatFloat(currentBalance, 'f', 2, 64))
	}

	// 4. 生成提现明细编号
	serialNo := utils.GenerateSerialNo("WDR") // WDR-提现前缀

	// 5. 扣减陪诊师余额
	newBalance := utils.KeepTwoDecimal(currentBalance - withdrawAmount)
	if err := tx.Model(&model.User{}).Where("id = ?", companionId).Update("balance", newBalance).Error; err != nil {
		tx.Rollback()
		return "", errors.New("扣减账户余额失败")
	}

	// 6. 生成提现明细（状态：2-提现成功/3-提现失败，此处先记录为待审核，可后续扩展审核逻辑）
	// 备注：实际项目中可增加「提现审核」状态，此处简化为直接生成提现成功明细（或提现中）
	remark := "提现至" + account + "（姓名：" + realName + "）"
	balanceRecord := model.BalanceRecord{
		SerialNo:    serialNo,
		CompanionId: companionId,
		Type:        2,               // 2-提现成功（若需审核，可先设为4-提现中，审核后更新为2/3）
		Amount:      -withdrawAmount, // 提现金额为负数（收入为正，提现为负）
		Remark:      remark,
	}
	if err := tx.Create(&balanceRecord).Error; err != nil {
		tx.Rollback()
		return "", errors.New("生成提现明细失败")
	}

	// 7. 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return "", errors.New("提现事务提交失败")
	}

	// 返回提现明细编号
	return serialNo, nil
}

// UpdateWithdrawStatus 更新提现状态（后续扩展：审核/支付回调时调用）
func (b *BalanceService) UpdateWithdrawStatus(serialNo string, newType int) error {
	// 1. 校验提现类型（仅允许更新为2-成功/3-失败）
	if newType != 2 && newType != 3 {
		return errors.New("无效的提现状态，仅支持2-提现成功/3-提现失败")
	}

	// 2. 查询提现明细是否存在
	var record model.BalanceRecord
	if err := model.DB.Where("serial_no = ? AND type IN (4,2,3)", serialNo).First(&record).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("提现明细不存在")
		}
		return errors.New("查询提现明细失败")
	}

	// 3. 若提现失败，恢复余额
	if newType == 3 {
		tx := model.DB.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// 恢复陪诊师余额
		var companion model.User
		if err := tx.Where("id = ?", record.CompanionId).First(&companion).Error; err != nil {
			tx.Rollback()
			return errors.New("查询陪诊师信息失败")
		}
		restoreAmount := -record.Amount // 提现金额为负，恢复时取正
		newBalance := utils.KeepTwoDecimal(companion.Balance + restoreAmount)
		if err := tx.Model(&model.User{}).Where("id = ?", record.CompanionId).Update("balance", newBalance).Error; err != nil {
			tx.Rollback()
			return errors.New("恢复陪诊师余额失败")
		}

		// 更新提现明细状态
		if err := tx.Model(&model.BalanceRecord{}).Where("serial_no = ?", serialNo).Update("type", 3).Error; err != nil {
			tx.Rollback()
			return errors.New("更新提现状态失败")
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return errors.New("提现失败事务提交失败")
		}
		return nil
	}

	// 4. 提现成功，直接更新状态（若无需恢复余额）
	if err := model.DB.Model(&model.BalanceRecord{}).Where("serial_no = ?", serialNo).Update("type", 2).Error; err != nil {
		return errors.New("更新提现状态失败")
	}

	return nil
}
