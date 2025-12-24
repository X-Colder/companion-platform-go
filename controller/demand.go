// controller/demand.go（修正文件名：deman.go → demand.go）
package controller

import (
	"strconv"

	"github.com/X-Colder/companion-backend/service"
	"github.com/X-Colder/companion-backend/utils"

	"github.com/gin-gonic/gin"
)

// DemandController 需求控制器（确保结构体大写，公开可访问）
type DemandController struct{}

// Publish 发布陪诊需求（控制器方法，大写公开）
func (d *DemandController) Publish(c *gin.Context) {
	// 1. 获取当前登录用户ID（从JWT上下文获取）
	patientId, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "用户身份验证失败")
		return
	}

	// 2. 接收前端提交的参数（与前端请求体字段对应）
	var req struct {
		Hospital       string  `json:"hospital" binding:"required,max=100"`      // 就诊医院
		HospitalAddr   string  `json:"hospital_addr" binding:"required,max=255"` // 医院地址
		ServiceTime    string  `json:"service_time" binding:"required"`          // 服务时间（前端传格式化字符串，如：2025-12-25 09:30:00）
		ExpectedPrice  float64 `json:"expected_price" binding:"required,gt=0"`   // 期望价格（大于0）
		ServiceContent string  `json:"service_content" binding:"required"`       // 服务内容
		ContactName    string  `json:"contact_name" binding:"required,max=16"`   // 联系人姓名
		ContactPhone   string  `json:"contact_phone" binding:"required,len=11"`  // 联系人电话
	}

	// 3. 参数校验（绑定失败返回错误）
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数格式错误："+err.Error())
		return
	}

	// 注意：patientId需转为uint64类型，与服务层参数一致
	err := (&service.DemandService{}).PublishDemand(
		patientId.(uint64),
		req.Hospital,
		req.HospitalAddr,
		req.ServiceTime,
		req.ExpectedPrice,
		req.ServiceContent,
		req.ContactName,
		req.ContactPhone,
	)

	// 5. 处理业务结果
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// Update 修改陪诊需求（补充完整，便于后续使用）
func (d *DemandController) Update(c *gin.Context) {
	// 获取当前用户ID
	patientId, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "用户身份验证失败")
		return
	}

	// 接收参数
	var req struct {
		ID             uint64  `json:"id" binding:"required,gt=0"`               // 需求ID
		Hospital       string  `json:"hospital" binding:"required,max=100"`      // 就诊医院
		HospitalAddr   string  `json:"hospital_addr" binding:"required,max=255"` // 医院地址
		ServiceTime    string  `json:"service_time" binding:"required"`          // 服务时间
		ExpectedPrice  float64 `json:"expected_price" binding:"required,gt=0"`   // 期望价格
		ServiceContent string  `json:"service_content" binding:"required"`       // 服务内容
		ContactName    string  `json:"contact_name" binding:"required,max=16"`   // 联系人姓名
		ContactPhone   string  `json:"contact_phone" binding:"required,len=11"`  // 联系人电话
	}

	// 参数绑定
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数格式错误："+err.Error())
		return
	}

	// 调用服务层修改方法
	err := (&service.DemandService{}).UpdateDemand(
		req.ID,
		patientId.(uint64),
		req.Hospital,
		req.HospitalAddr,
		req.ServiceTime,
		req.ExpectedPrice,
		req.ServiceContent,
		req.ContactName,
		req.ContactPhone,
	)

	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// GetMyDemandList 获取当前患者的需求列表
func (d *DemandController) GetMyDemandList(c *gin.Context) {
	// 获取当前用户ID
	patientId, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "用户身份验证失败")
		return
	}

	// 接收分页参数（可选，默认第1页，每页10条）
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	// 调用服务层查询方法
	demandList, total, err := (&service.DemandService{}).GetPatientDemandList(patientId.(uint64), page, size)
	if err != nil {
		utils.Fail(c, "查询需求列表失败："+err.Error())
		return
	}

	// 返回分页结果
	utils.Success(c, gin.H{
		"list":  demandList,
		"total": total,
		"page":  page,
		"size":  size,
	})
}
