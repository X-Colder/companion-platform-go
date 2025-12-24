// controller/order.go
package controller

import (
	"strconv"

	"github.com/X-Colder/companion-backend/service"
	"github.com/X-Colder/companion-backend/utils"

	"github.com/gin-gonic/gin"
)

// OrderController 订单控制器
type OrderController struct{}

// -------------------------- 陪诊师专属接口 --------------------------

// GetOrderHall 获取订单大厅（待接单需求列表，仅陪诊师访问）
func (o *OrderController) GetOrderHall(c *gin.Context) {
	// 接收分页参数（默认第1页，每页10条）
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	// 调用服务层查询待接单需求
	demandList, total, err := (&service.OrderService{}).GetUndertakeDemandList(page, size)
	if err != nil {
		utils.Fail(c, "查询订单大厅失败："+err.Error())
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

// TakeOrder 接单操作（仅陪诊师访问）
func (o *OrderController) TakeOrder(c *gin.Context) {
	// 1. 获取当前陪诊师ID
	companionId, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "用户身份验证失败")
		return
	}

	// 2. 接收前端参数
	var req struct {
		DemandId uint64 `json:"demand_id" binding:"required,gt=0"` // 需求ID
	}

	// 3. 参数绑定与校验
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数格式错误："+err.Error())
		return
	}

	// 4. 调用服务层接单方法
	err := (&service.OrderService{}).TakeOrder(req.DemandId, companionId.(uint64))
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, "接单成功，等待患者确认")
}

// GetCompanionServiceList 获取陪诊师服务列表（自己接的订单，仅陪诊师访问）
func (o *OrderController) GetCompanionServiceList(c *gin.Context) {
	// 1. 获取当前陪诊师ID
	companionId, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "用户身份验证失败")
		return
	}

	// 2. 接收分页参数与状态筛选
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	statusStr := c.DefaultQuery("status", "") // 可选筛选：1-待服务，2-服务中，3-待结算，4-已完成，5-已取消
	var status int
	if statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil {
			status = s
		}
	}

	// 3. 调用服务层查询
	orderList, total, err := (&service.OrderService{}).GetCompanionOrderList(companionId.(uint64), status, page, size)
	if err != nil {
		utils.Fail(c, "查询服务列表失败："+err.Error())
		return
	}

	// 返回分页结果
	utils.Success(c, gin.H{
		"list":  orderList,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// CompanionConfirmComplete 陪诊师确认服务完成（仅陪诊师访问）
func (o *OrderController) CompanionConfirmComplete(c *gin.Context) {
	// 1. 获取当前陪诊师ID
	companionId, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "用户身份验证失败")
		return
	}

	// 2. 接收订单ID
	var req struct {
		OrderId uint64 `json:"order_id" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数格式错误："+err.Error())
		return
	}

	// 3. 调用服务层方法
	err := (&service.OrderService{}).CompanionConfirmOrderComplete(req.OrderId, companionId.(uint64))
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, "已确认服务完成，等待患者确认")
}

// CompanionCancelOrder 陪诊师取消订单（仅待服务状态可取消，仅陪诊师访问）
func (o *OrderController) CompanionCancelOrder(c *gin.Context) {
	// 1. 获取当前陪诊师ID
	companionId, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "用户身份验证失败")
		return
	}

	// 2. 接收订单ID与取消原因
	var req struct {
		OrderId uint64 `json:"order_id" binding:"required,gt=0"`
		Reason  string `json:"reason" binding:"required,max=255"` // 取消原因
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数格式错误："+err.Error())
		return
	}

	// 3. 调用服务层方法
	err := (&service.OrderService{}).CompanionCancelOrder(req.OrderId, companionId.(uint64), req.Reason)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, "订单取消成功")
}

// -------------------------- 患者专属接口 --------------------------

// GetPatientOrderList 获取患者订单列表（自己发布的需求对应的订单，仅患者访问）
func (o *OrderController) GetPatientOrderList(c *gin.Context) {
	// 1. 获取当前患者ID
	patientId, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "用户身份验证失败")
		return
	}

	// 2. 接收分页参数与状态筛选
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	statusStr := c.DefaultQuery("status", "")
	var status int
	if statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil {
			status = s
		}
	}

	// 3. 调用服务层查询
	orderList, total, err := (&service.OrderService{}).GetPatientOrderList(patientId.(uint64), status, page, size)
	if err != nil {
		utils.Fail(c, "查询订单列表失败："+err.Error())
		return
	}

	// 返回分页结果
	utils.Success(c, gin.H{
		"list":  orderList,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// PatientConfirmComplete 患者确认服务完成（仅患者访问，确认后订单结算）
func (o *OrderController) PatientConfirmComplete(c *gin.Context) {
	// 1. 获取当前患者ID
	patientId, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "用户身份验证失败")
		return
	}

	// 2. 接收订单ID
	var req struct {
		OrderId uint64 `json:"order_id" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数格式错误："+err.Error())
		return
	}

	// 3. 调用服务层方法
	err := (&service.OrderService{}).PatientConfirmOrderComplete(req.OrderId, patientId.(uint64))
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, "已确认服务完成，订单已结算")
}

// PatientCancelOrder 患者取消订单（仅待服务状态可取消，仅患者访问）
func (o *OrderController) PatientCancelOrder(c *gin.Context) {
	// 1. 获取当前患者ID
	patientId, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "用户身份验证失败")
		return
	}

	// 2. 接收订单ID与取消原因
	var req struct {
		OrderId uint64 `json:"order_id" binding:"required,gt=0"`
		Reason  string `json:"reason" binding:"required,max=255"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数格式错误："+err.Error())
		return
	}

	// 3. 调用服务层方法
	err := (&service.OrderService{}).PatientCancelOrder(req.OrderId, patientId.(uint64), req.Reason)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, "订单取消成功")
}
