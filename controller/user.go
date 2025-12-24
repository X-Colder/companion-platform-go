package controller

import (
	"github.com/X-Colder/companion-backend/service"
	"github.com/X-Colder/companion-backend/utils"

	"github.com/gin-gonic/gin"
)

// UserController 用户控制器
type UserController struct{}

// Login 用户登录
func (u *UserController) Login(c *gin.Context) {
	// 接收前端参数
	var req struct {
		Phone    string `json:"phone" binding:"required,len=11"`
		Password string `json:"password" binding:"required,min=6,max=16"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数格式错误："+err.Error())
		return
	}

	// 调用服务层
	token, userInfo, err := (&service.UserService{}).Login(req.Phone, req.Password)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	// 返回结果
	utils.Success(c, gin.H{
		"token":     token,
		"user_info": userInfo,
	})
}

// Register 用户注册
func (u *UserController) Register(c *gin.Context) {
	// 接收前端参数
	var req struct {
		Phone      string `json:"phone" binding:"required,len=11"`
		UserType   int    `json:"userType" binding:"required,oneof=1 2"`
		Password   string `json:"password" binding:"required,min=6,max=16"`
		ConfirmPwd string `json:"confirmPassword" binding:"required,eqfield=Password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数格式错误："+err.Error())
		return
	}

	// 调用服务层
	err := (&service.UserService{}).Register(req.Phone, req.UserType, req.Password)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// UpdateProfile 修改个人资料
func (u *UserController) UpdateProfile(c *gin.Context) {
	// 获取当前用户ID
	userId, _ := c.Get("user_id")

	// 接收参数
	var req struct {
		Nickname string `json:"nickname" binding:"required,min=2,max=16"`
		Avatar   string `json:"avatar"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数格式错误："+err.Error())
		return
	}

	// 调用服务层
	err := (&service.UserService{}).UpdateProfile(userId.(uint64), req.Nickname, req.Avatar)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// GetUserInfo 获取用户信息
func (u *UserController) GetUserInfo(c *gin.Context) {
	userId, _ := c.Get("user_id")

	userInfo, err := (&service.UserService{}).GetUserInfo(userId.(uint64))
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, userInfo)
}
