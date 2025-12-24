package controller

import (
	"github.com/X-Colder/companion-backend/service"
	"github.com/X-Colder/companion-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
)

// UserController 用户控制器
type UserController struct{}

// 全局验证码存储（开发环境用内存存储，生产环境建议替换为Redis）
var store = base64Captcha.DefaultMemStore

// GetCaptcha 获取图形验证码
func (u *UserController) GetCaptcha(c *gin.Context) {
	// 配置验证码参数
	driver := base64Captcha.NewDriverDigit(
		80,  // 验证码图片高度
		240, // 验证码图片宽度
		6,   // 验证码长度
		0.7, // 干扰线比例
		80,  // 干扰点数量
	)
	// 创建验证码实例
	captcha := base64Captcha.NewCaptcha(driver, store)
	// 生成验证码（id：验证码唯一标识，b64s：base64格式图片，_：验证码文本）
	id, b64s, _, err := captcha.Generate()
	if err != nil {
		utils.Fail(c, "生成验证码失败："+err.Error())
		return
	}

	// 返回验证码信息（id用于后续验证，b64s用于前端展示）
	utils.Success(c, gin.H{
		"captcha_id":  id,
		"captcha_img": b64s,
	})
}

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

// ResetPassword 重置密码（验证旧密码，设置新密码）
func (u *UserController) ResetPassword(c *gin.Context) {
	// 1. 获取当前登录用户ID（从JWT上下文获取，确保是本人操作）
	userId, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "用户身份验证失败")
		return
	}

	// 2. 接收重置密码参数
	var req struct {
		OldPassword string `json:"old_password" binding:"required,min=6,max=20"` // 旧密码
		NewPassword string `json:"new_password" binding:"required,min=6,max=20"` // 新密码
		ConfirmPwd  string `json:"confirm_pwd" binding:"required,min=6,max=20"`  // 确认新密码
	}

	// 3. 参数绑定与校验
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数格式错误："+err.Error())
		return
	}

	// 4. 校验新密码与确认密码是否一致
	if req.NewPassword != req.ConfirmPwd {
		utils.Fail(c, "新密码与确认密码不一致")
		return
	}

	// 5. 校验新密码与旧密码是否相同（避免重复设置相同密码）
	if req.OldPassword == req.NewPassword {
		utils.Fail(c, "新密码不能与旧密码一致")
		return
	}

	// 6. 调用服务层重置密码方法
	err := (&service.UserService{}).ResetPassword(
		userId.(uint64),
		req.OldPassword,
		req.NewPassword,
	)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, "密码重置成功，请重新登录")
}
