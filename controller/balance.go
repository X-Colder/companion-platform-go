// controller/balance.go
package controller

import (
	"strconv"

	"github.com/X-Colder/companion-backend/service"
	"github.com/X-Colder/companion-backend/utils"

	"github.com/gin-gonic/gin"
)

// BalanceController 余额控制器（仅陪诊师访问）
type BalanceController struct{}

// GetBalance 查询当前陪诊师账户余额
func (b *BalanceController) GetBalance(c *gin.Context) {
	// 1. 获取当前登录用户ID
	companionId, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "用户身份验证失败")
		return
	}

	// 2. 获取用户类型（校验是否为陪诊师）
	userType, exists := c.Get("user_type")
	if !exists || userType.(int) != 2 {
		utils.Fail(c, "非陪诊师账户，无余额查询权限")
		return
	}

	// 3. 调用服务层查询余额
	balance, err := (&service.BalanceService{}).GetCompanionBalance(companionId.(uint64))
	if err != nil {
		utils.Fail(c, "查询余额失败："+err.Error())
		return
	}

	// 4. 返回余额信息
	utils.Success(c, gin.H{
		"balance": balance, // 保留两位小数的余额
	})
}

// GetBalanceRecordList 查询余额明细列表（带分页、类型筛选）
func (b *BalanceController) GetBalanceRecordList(c *gin.Context) {
	// 1. 获取当前登录陪诊师ID
	companionId, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "用户身份验证失败")
		return
	}

	// 2. 校验用户类型
	userType, exists := c.Get("user_type")
	if !exists || userType.(int) != 2 {
		utils.Fail(c, "非陪诊师账户，无明细查询权限")
		return
	}

	// 3. 接收分页参数与类型筛选
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	recordTypeStr := c.DefaultQuery("type", "") // 筛选类型：1-收入，2-提现成功，3-提现失败
	var recordType int
	if recordTypeStr != "" {
		t, err := strconv.Atoi(recordTypeStr)
		if err == nil && (t == 1 || t == 2 || t == 3) {
			recordType = t
		}
	}

	// 4. 调用服务层查询明细
	recordList, total, err := (&service.BalanceService{}).GetCompanionBalanceRecordList(
		companionId.(uint64),
		recordType,
		page,
		size,
	)
	if err != nil {
		utils.Fail(c, "查询余额明细失败："+err.Error())
		return
	}

	// 5. 返回分页结果
	utils.Success(c, gin.H{
		"list":  recordList,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// ApplyWithdraw 申请提现（仅陪诊师访问）
func (b *BalanceController) ApplyWithdraw(c *gin.Context) {
	// 1. 获取当前登录陪诊师ID
	companionId, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "用户身份验证失败")
		return
	}

	// 2. 校验用户类型
	userType, exists := c.Get("user_type")
	if !exists || userType.(int) != 2 {
		utils.Fail(c, "非陪诊师账户，无提现权限")
		return
	}

	// 3. 接收提现参数
	var req struct {
		Amount   float64 `json:"amount" binding:"required,gt=0"` // 提现金额（大于0）
		Account  string  `json:"account" binding:"required"`     // 提现账户（如微信/支付宝账号）
		RealName string  `json:"real_name" binding:"required"`   // 真实姓名（与账户一致）
	}

	// 4. 参数绑定与校验
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数格式错误："+err.Error())
		return
	}

	// 5. 调用服务层申请提现
	serialNo, err := (&service.BalanceService{}).ApplyWithdraw(
		companionId.(uint64),
		req.Amount,
		req.Account,
		req.RealName,
	)
	if err != nil {
		utils.Fail(c, "提现申请失败："+err.Error())
		return
	}

	// 6. 返回提现明细编号
	utils.Success(c, gin.H{
		"serial_no": serialNo, // 提现明细编号（用于查询提现状态）
		"msg":       "提现申请已提交，等待审核处理",
	})
}
