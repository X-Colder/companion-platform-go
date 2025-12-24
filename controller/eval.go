// controller/eval.go
package controller

import (
	"strconv"
	"strings"

	"github.com/X-Colder/companion-backend/service"
	"github.com/X-Colder/companion-backend/utils"

	"github.com/gin-gonic/gin"
)

// EvalController 评价控制器
type EvalController struct{}

// -------------------------- 患者专属接口 --------------------------

// PatientEvalCompanion 患者评价陪诊师（仅已完成订单可评价）
func (e *EvalController) PatientEvalCompanion(c *gin.Context) {
	// 1. 获取当前登录患者ID与用户类型
	patientId, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "用户身份验证失败")
		return
	}
	userType, exists := c.Get("user_type")
	if !exists || userType.(int) != 1 {
		utils.Fail(c, "非患者账户，无评价权限")
		return
	}

	// 2. 接收评价参数
	var req struct {
		OrderId uint64   `json:"order_id" binding:"required,gt=0"`     // 关联订单ID
		Score   int      `json:"score" binding:"required,min=1,max=5"` // 评分（1-5星）
		Content string   `json:"content" binding:"required,max=500"`   // 评价内容
		ImgUrls []string `json:"img_urls"`                             // 评价图片（可选，前端传图片地址数组）
	}

	// 3. 参数绑定与校验
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数格式错误："+err.Error())
		return
	}

	// 4. 处理图片地址（数组转字符串，逗号分隔存储）
	imgUrlsStr := ""
	if len(req.ImgUrls) > 0 {
		imgUrlsStr = strings.Join(req.ImgUrls, ",")
	}

	// 5. 调用服务层评价方法
	err := (&service.EvalService{}).PatientEvalCompanion(
		req.OrderId,
		patientId.(uint64),
		req.Score,
		req.Content,
		imgUrlsStr,
	)
	if err != nil {
		utils.Fail(c, "评价失败："+err.Error())
		return
	}

	utils.Success(c, "评价成功")
}

// -------------------------- 陪诊师专属接口 --------------------------

// CompanionEvalPatient 陪诊师评价患者（仅已完成订单可评价）
func (e *EvalController) CompanionEvalPatient(c *gin.Context) {
	// 1. 获取当前登录陪诊师ID与用户类型
	companionId, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "用户身份验证失败")
		return
	}
	userType, exists := c.Get("user_type")
	if !exists || userType.(int) != 2 {
		utils.Fail(c, "非陪诊师账户，无评价权限")
		return
	}

	// 2. 接收评价参数
	var req struct {
		OrderId uint64 `json:"order_id" binding:"required,gt=0"`     // 关联订单ID
		Score   int    `json:"score" binding:"required,min=1,max=5"` // 评分（1-5星）
		Content string `json:"content" binding:"required,max=500"`   // 评价内容
	}

	// 3. 参数绑定与校验
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数格式错误："+err.Error())
		return
	}

	// 4. 调用服务层评价方法
	err := (&service.EvalService{}).CompanionEvalPatient(
		req.OrderId,
		companionId.(uint64),
		req.Score,
		req.Content,
	)
	if err != nil {
		utils.Fail(c, "评价失败："+err.Error())
		return
	}

	utils.Success(c, "评价成功")
}

// -------------------------- 通用接口 --------------------------

// GetUserEvalList 查询用户收到的评价列表（带分页）
func (e *EvalController) GetUserEvalList(c *gin.Context) {
	// 1. 获取当前登录用户ID（所有角色均可查询自己收到的评价）
	userId, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "用户身份验证失败")
		return
	}

	// 2. 接收分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	// 3. 调用服务层查询方法
	evalList, total, err := (&service.EvalService{}).GetUserReceivedEvalList(
		userId.(uint64),
		page,
		size,
	)
	if err != nil {
		utils.Fail(c, "查询评价列表失败："+err.Error())
		return
	}

	// 4. 返回分页结果
	utils.Success(c, gin.H{
		"list":  evalList,
		"total": total,
		"page":  page,
		"size":  size,
	})
}
