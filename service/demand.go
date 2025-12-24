// service/demand.go
package service

import (
	"errors"
	"time"

	"github.com/X-Colder/companion-backend/model"
	"github.com/X-Colder/companion-backend/utils"

	"github.com/jinzhu/gorm"
)

// DemandService 需求服务（结构体大写公开）
type DemandService struct{}

// PublishDemand 发布陪诊需求（方法大写公开，可跨包调用）
func (d *DemandService) PublishDemand(
	patientId uint64,
	hospital string,
	hospitalAddr string,
	serviceTimeStr string,
	expectedPrice float64,
	serviceContent string,
	contactName string,
	contactPhone string,
) error {
	// 1. 解析服务时间字符串为time.Time类型
	serviceTime, err := time.Parse("2006-01-02 15:04:05", serviceTimeStr)
	if err != nil {
		return errors.New("服务时间格式错误，请传入：2006-01-02 15:04:05")
	}

	// 2. 校验服务时间（不能早于当前时间）
	if serviceTime.Before(time.Now()) {
		return errors.New("服务时间不能早于当前时间")
	}

	// 3. 构造需求实体
	demand := model.Demand{
		PatientId:      patientId,
		Hospital:       hospital,
		HospitalAddr:   hospitalAddr,
		ServiceTime:    serviceTime,
		ExpectedPrice:  utils.KeepTwoDecimal(expectedPrice),
		ServiceContent: serviceContent,
		ContactName:    contactName,
		ContactPhone:   contactPhone,
		Status:         0, // 0-待接单
	}

	// 4. 存入数据库
	if err := model.DB.Create(&demand).Error; err != nil {
		return errors.New("发布需求失败")
	}

	return nil
}

// UpdateDemand 修改陪诊需求（仅待接单状态可修改）
func (d *DemandService) UpdateDemand(
	demandId uint64,
	patientId uint64,
	hospital string,
	hospitalAddr string,
	serviceTimeStr string,
	expectedPrice float64,
	serviceContent string,
	contactName string,
	contactPhone string,
) error {
	// 1. 查询需求是否存在，且属于当前患者，且状态为待接单
	var existDemand model.Demand
	if err := model.DB.Where("id = ? AND patient_id = ? AND status = 0", demandId, patientId).First(&existDemand).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("需求不存在或已接单，无法修改")
		}
		return errors.New("查询需求失败")
	}

	// 2. 解析服务时间
	serviceTime, err := time.Parse("2006-01-02 15:04:05", serviceTimeStr)
	if err != nil {
		return errors.New("服务时间格式错误，请传入：2006-01-02 15:04:05")
	}
	if serviceTime.Before(time.Now()) {
		return errors.New("服务时间不能早于当前时间")
	}

	// 3. 构造更新参数
	updateData := map[string]interface{}{
		"hospital":        hospital,
		"hospital_addr":   hospitalAddr,
		"service_time":    serviceTime,
		"expected_price":  utils.KeepTwoDecimal(expectedPrice),
		"service_content": serviceContent,
		"contact_name":    contactName,
		"contact_phone":   contactPhone,
	}

	// 4. 更新数据库
	if err := model.DB.Model(&model.Demand{}).Where("id = ?", demandId).Updates(updateData).Error; err != nil {
		return errors.New("修改需求失败")
	}

	return nil
}

// GetPatientDemandList 获取患者的需求列表（带分页）
func (d *DemandService) GetPatientDemandList(patientId uint64, page int, size int) ([]model.Demand, int64, error) {
	var demandList []model.Demand
	var total int64

	// 计算分页偏移量
	offset := (page - 1) * size

	// 查询总数
	if err := model.DB.Where("patient_id = ?", patientId).Model(&model.Demand{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询分页数据
	if err := model.DB.Where("patient_id = ?", patientId).Order("created_at DESC").Offset(offset).Limit(size).Find(&demandList).Error; err != nil {
		return nil, 0, err
	}

	return demandList, total, nil
}
