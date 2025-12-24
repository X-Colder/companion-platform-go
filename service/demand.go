package service

import (
	"github.com/X-Colder/companion-backend/conf"
	"github.com/X-Colder/companion-backend/model"
)

// DemandService 陪诊需求业务层
type DemandService struct{}

// PublishDemand 发布需求
func (d *DemandService) PublishDemand(demand *model.CompanionDemand) (bool, string) {
	// 校验需求参数（省略详细校验，可自行补充）
	if demand.Hospital == "" || demand.ServiceTime.IsZero() {
		return false, "医院名称和服务时间不能为空"
	}

	// 保存到数据库
	err := conf.DB.Create(demand).Error
	if err != nil {
		return false, "需求发布失败：" + err.Error()
	}

	return true, "需求发布成功"
}
