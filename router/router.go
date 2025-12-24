package router

import (
	"github.com/X-Colder/companion-backend/controller"
	"github.com/X-Colder/companion-backend/middleware"

	"github.com/gin-gonic/gin"
)

// InitRouter 初始化路由
func InitRouter() *gin.Engine {
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.Cors()) // 跨域中间件
	r.Use(middleware.Logger())
	// 静态资源路由（访问上传的图片）
	r.Static("/static", "./static")

	// 公共路由（无需登录）
	publicGroup := r.Group("/api")
	{
		// 用户接口
		publicGroup.POST("/user/login", &controller.UserController{}.Login)
		publicGroup.POST("/user/register", &controller.UserController{}.Register)
		// 文件上传接口（头像，无需登录，可根据需求调整）
		publicGroup.POST("/upload/img", &controller.UploadController{}.UploadAvatar)
	}

	// 需要登录的路由
	authGroup := r.Group("/api")
	authGroup.Use(middleware.JwtAuth()) // JWT认证中间件
	{
		// 用户接口
		authGroup.POST("/user/updateProfile", &controller.UserController{}.UpdateProfile)
		authGroup.GET("/user/info", &controller.UserController{}.GetUserInfo)

		// 患者专属接口（仅用户类型1可访问）
		patientGroup := authGroup.Group("/patient")
		patientGroup.Use(middleware.JwtAuthByRole(1))
		{
			patientGroup.POST("/demand/publish", &controller.DemandController{}.Publish)
			patientGroup.POST("/demand/update", &controller.DemandController{}.Update)
			patientGroup.GET("/demand/list", &controller.DemandController{}.GetMyDemandList)
			patientGroup.GET("/order/list", &controller.OrderController{}.GetPatientOrderList)
			patientGroup.POST("/order/confirm", &controller.OrderController{}.PatientConfirmComplete)
			patientGroup.POST("/order/cancel", &controller.OrderController{}.PatientCancelOrder)
		}

		// 陪诊师专属接口（仅用户类型2可访问）
		companionGroup := authGroup.Group("/companion")
		companionGroup.Use(middleware.JwtAuthByRole(2))
		{
			companionGroup.GET("/order/hall", &controller.OrderController{}.GetOrderHall)
			companionGroup.POST("/order/take", &controller.OrderController{}.TakeOrder)
			companionGroup.GET("/order/list", &controller.OrderController{}.GetCompanionServiceList)
			companionGroup.POST("/order/confirm", &controller.OrderController{}.CompanionConfirmComplete)
			companionGroup.POST("/order/cancel", &controller.OrderController{}.CompanionCancelOrder)
		}

		// 评价接口
		authGroup.POST("/eval/submit", &controller.EvalController{}.SubmitEval)

		// 余额接口（仅陪诊师）
		balanceGroup := authGroup.Group("/balance")
		balanceGroup.Use(middleware.JwtAuthByRole(2))
		{
			balanceGroup.GET("/info", &controller.BalanceController{}.GetBalanceInfo)
			balanceGroup.GET("/record/list", &controller.BalanceController{}.GetBalanceRecordList)
			balanceGroup.POST("/withdraw", &controller.BalanceController{}.ApplyWithdraw)
		}

		// 评价图片上传
		authGroup.POST("/upload/eval", &controller.UploadController{}.UploadEvalImg)
	}

	return r
}
