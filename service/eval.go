// service/eval.go
package service

import (
	"errors"
	"strings"

	"github.com/X-Colder/companion-backend/model"
	"github.com/jinzhu/gorm"
)

// EvalService 评价服务
type EvalService struct{}

// -------------------------- 患者评价陪诊师 --------------------------

// PatientEvalCompanion 患者评价陪诊师（事务：创建评价+更新订单评价状态）
func (e *EvalService) PatientEvalCompanion(orderId uint64, patientId uint64, score int, content string, imgUrls string) error {
	// 开启事务（创建评价 + 更新订单的患者评价状态）
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	// 1. 查询订单信息，校验评价合法性
	var order model.Order
	if err := tx.Where("id = ? AND patient_id = ? AND status = ?", orderId, patientId, 4).First(&order).Error; err != nil {
		tx.Rollback()
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("订单不存在、非本人订单或非已完成状态，无法评价")
		}
		return errors.New("查询订单失败")
	}

	// 2. 校验是否已评价
	if order.HasPatientEval == 1 {
		tx.Rollback()
		return errors.New("该订单已评价，不可重复评价")
	}

	// 3. 构造评价实体（患者→陪诊师：from_user=患者ID，to_user=陪诊师ID）
	eval := model.Evaluation{
		RelateId:   orderId,
		FromUserId: patientId,
		ToUserId:   order.CompanionId,
		Score:      score,
		Content:    content,
		ImgUrls:    imgUrls,
	}

	// 4. 创建评价记录
	if err := tx.Create(&eval).Error; err != nil {
		tx.Rollback()
		return errors.New("创建评价失败")
	}

	// 5. 更新订单的患者评价状态（0-未评价 → 1-已评价）
	if err := tx.Model(&model.Order{}).Where("id = ?", orderId).Update("has_patient_eval", 1).Error; err != nil {
		tx.Rollback()
		return errors.New("更新订单评价状态失败")
	}

	// 6. 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return errors.New("评价事务提交失败")
	}

	return nil
}

// -------------------------- 陪诊师评价患者 --------------------------

// CompanionEvalPatient 陪诊师评价患者（事务：创建评价+更新订单评价状态）
func (e *EvalService) CompanionEvalPatient(orderId uint64, companionId uint64, score int, content string) error {
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

	// 1. 查询订单信息，校验评价合法性
	var order model.Order
	if err := tx.Where("id = ? AND companion_id = ? AND status = ?", orderId, companionId, 4).First(&order).Error; err != nil {
		tx.Rollback()
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("订单不存在、非本人订单或非已完成状态，无法评价")
		}
		return errors.New("查询订单失败")
	}

	// 2. 校验是否已评价
	if order.HasCompanionEval == 1 {
		tx.Rollback()
		return errors.New("该订单已评价，不可重复评价")
	}

	// 3. 构造评价实体（陪诊师→患者：from_user=陪诊师ID，to_user=患者ID）
	eval := model.Evaluation{
		RelateId:   orderId,
		FromUserId: companionId,
		ToUserId:   order.PatientId,
		Score:      score,
		Content:    content,
		ImgUrls:    "", // 陪诊师评价暂不支持图片，可后续扩展
	}

	// 4. 创建评价记录
	if err := tx.Create(&eval).Error; err != nil {
		tx.Rollback()
		return errors.New("创建评价失败")
	}

	// 5. 更新订单的陪诊师评价状态（0-未评价 → 1-已评价）
	if err := tx.Model(&model.Order{}).Where("id = ?", orderId).Update("has_companion_eval", 1).Error; err != nil {
		tx.Rollback()
		return errors.New("更新订单评价状态失败")
	}

	// 6. 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return errors.New("评价事务提交失败")
	}

	return nil
}

// -------------------------- 查询评价列表 --------------------------

// GetUserReceivedEvalList 查询用户收到的所有评价（带分页）
func (e *EvalService) GetUserReceivedEvalList(toUserId uint64, page int, size int) ([]model.Evaluation, int64, error) {
	var evalList []model.Evaluation
	var total int64

	// 1. 计算分页偏移量
	offset := (page - 1) * size

	// 2. 查询评价总数（当前用户是被评价人：to_user_id=当前用户ID）
	if err := model.DB.Where("to_user_id = ?", toUserId).Model(&model.Evaluation{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 3. 查询分页评价数据（按创建时间倒序，最新评价优先）
	if err := model.DB.Where("to_user_id = ?", toUserId).Order("created_at DESC").Offset(offset).Limit(size).Find(&evalList).Error; err != nil {
		return nil, 0, err
	}

	// 4. 处理图片地址（字符串转数组，便于前端展示）
	for i := range evalList {
		if evalList[i].ImgUrls != "" {
			evalList[i].ImgUrls = strings.Join(strings.Split(evalList[i].ImgUrls, ","), "|") // 前端可按|分割，或直接返回数组（需修改模型字段类型）
			// 若需直接返回数组，可将模型ImgUrls改为[]string，并用gorm标签：`gorm:"type:varchar(512);default:''" json:"img_urls"`
		}
	}

	return evalList, total, nil
}

// -------------------------- 辅助方法 --------------------------

// GetOrderEvalStatus 查询订单的评价状态（判断双方是否已评价）
func (e *EvalService) GetOrderEvalStatus(orderId uint64) (bool, bool, error) {
	var order model.Order
	if err := model.DB.Where("id = ?", orderId).First(&order).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, false, errors.New("订单不存在")
		}
		return false, false, errors.New("查询订单失败")
	}

	// 返回：患者是否已评价、陪诊师是否已评价
	return order.HasPatientEval == 1, order.HasCompanionEval == 1, nil
}
