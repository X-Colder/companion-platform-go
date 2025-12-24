package router

import (
	"github.com/X-Colder/companion-backend/controller"
	"github.com/X-Colder/companion-backend/middleware"

	"github.com/gin-gonic/gin"
)

// InitRouter 初始化路由
func InitRouter() *gin.Engine {
	// 创建Gin引擎（发布环境可使用 gin.ReleaseMode）
	// gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// -------------------------- 全局中间件 --------------------------
	// 跨域中间件
	r.Use(middleware.Cors())
	// 日志中间件（自定义日志格式，记录请求信息）
	r.Use(middleware.Logger())
	// 异常恢复中间件（防止程序崩溃）
	r.Use(gin.Recovery())

	// -------------------------- 静态资源路由 --------------------------
	// 访问路径：/static/xxx → 对应本地目录 ./static/xxx（上传的图片等资源）
	r.Static("/static", "./static")

	// -------------------------- 无需JWT鉴权的公开接口 --------------------------
	publicGroup := r.Group("/api/public")
	{
		// 用户相关公开接口
		userPublic := publicGroup.Group("/user")
		{
			userPublic.POST("/register", (&controller.UserController{}).Register) // 用户注册
			userPublic.POST("/login", (&controller.UserController{}).Login)       // 用户登录
			userPublic.GET("/captcha", (&controller.UserController{}).GetCaptcha) // 获取验证码（可选）
		}

		// 健康检查接口（用于服务监控）
		publicGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"code": 200,
				"msg":  "service is running",
				"data": nil,
			})
		})
	}

	// -------------------------- 需要JWT鉴权的私有接口 --------------------------
	authGroup := r.Group("/api")
	authGroup.Use(middleware.JwtAuth()) // 全局JWT鉴权中间件，验证用户身份
	{
		// -------------------------- 通用用户接口（所有登录用户均可访问） --------------------------
		userGroup := authGroup.Group("/user")
		{
			userGroup.GET("/info", (&controller.UserController{}).GetUserInfo)              // 获取当前用户信息
			userGroup.POST("/info/update", (&controller.UserController{}).UpdateProfile)    // 修改用户信息
			userGroup.POST("/password/reset", (&controller.UserController{}).ResetPassword) // 重置密码
			userGroup.GET("/eval/list", (&controller.EvalController{}).GetUserEvalList)     // 查询用户收到的评价列表
		}

		// -------------------------- 文件上传接口（所有登录用户均可访问） --------------------------
		uploadGroup := authGroup.Group("/upload")
		{
			uploadGroup.POST("/avatar", (&controller.UploadController{}).UploadAvatar)      // 上传用户头像
			uploadGroup.POST("/eval/imgs", (&controller.UploadController{}).UploadEvalImgs) // 批量上传评价图片
			uploadGroup.POST("/file/delete", (&controller.UploadController{}).DeleteFile)   // 删除单个文件
		}

		// -------------------------- 患者专属接口 --------------------------
		patientGroup := authGroup.Group("/patient")
		{
			// 需求相关
			patientDemand := patientGroup.Group("/demand")
			{
				patientDemand.POST("/publish", (&controller.DemandController{}).Publish)        // 发布陪诊需求
				patientDemand.POST("/update", (&controller.DemandController{}).Update)          // 修改陪诊需求
				patientDemand.GET("/my/list", (&controller.DemandController{}).GetMyDemandList) // 查询我的需求列表
			}

			// 订单相关
			patientOrder := patientGroup.Group("/order")
			{
				patientOrder.GET("/list", (&controller.OrderController{}).GetPatientOrderList)        // 查询我的订单列表
				patientOrder.POST("/confirm", (&controller.OrderController{}).PatientConfirmComplete) // 确认服务完成
				patientOrder.POST("/cancel", (&controller.OrderController{}).PatientCancelOrder)      // 取消订单
			}

			// 评价相关
			patientEval := patientGroup.Group("/eval")
			{
				patientEval.POST("/companion", (&controller.EvalController{}).PatientEvalCompanion) // 评价陪诊师
			}
		}

		// -------------------------- 陪诊师专属接口 --------------------------
		companionGroup := authGroup.Group("/companion")
		{
			// 订单大厅/接单相关
			companionOrder := companionGroup.Group("/order")
			{
				companionOrder.GET("/hall", (&controller.OrderController{}).GetOrderHall)                 // 查询订单大厅（待接单需求）
				companionOrder.POST("/take", (&controller.OrderController{}).TakeOrder)                   // 接单操作
				companionOrder.GET("/list", (&controller.OrderController{}).GetCompanionServiceList)      // 查询我的服务列表
				companionOrder.POST("/confirm", (&controller.OrderController{}).CompanionConfirmComplete) // 确认服务完成
				companionOrder.POST("/cancel", (&controller.OrderController{}).CompanionCancelOrder)      // 取消订单
			}

			// 余额相关
			companionBalance := companionGroup.Group("/balance")
			{
				companionBalance.GET("", (&controller.BalanceController{}).GetBalance)                   // 查询账户余额
				companionBalance.GET("/records", (&controller.BalanceController{}).GetBalanceRecordList) // 查询余额明细
				companionBalance.POST("/withdraw", (&controller.BalanceController{}).ApplyWithdraw)      // 申请提现
			}

			// 评价相关
			companionEval := companionGroup.Group("/eval")
			{
				companionEval.POST("/patient", (&controller.EvalController{}).CompanionEvalPatient) // 评价患者
			}
		}
	}

	// 返回Gin引擎
	return r
}
