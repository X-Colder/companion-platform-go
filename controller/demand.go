package controller

import (
	"net/http"

	"github.com/X-Colder/companion-backend/model"
	"github.com/X-Colder/companion-backend/service"
	"github.com/gin-gonic/gin"
)

// DemandController 陪诊需求控制器
type DemandController struct{}

// PublishDemand 发布陪诊需求
func (d *DemandController) PublishDemand(c *gin.Context) {
	// 接收参数
	var demand model.CompanionDemand
	if err := c.ShouldBindJSON(&demand); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "参数格式错误：" + err.Error(),
			"data": nil,
		})
		return
	}

	// 调用业务层
	ok, msg := service.DemandService{}.PublishDemand(&demand)
	if ok {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  msg,
			"data": nil,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  msg,
			"data": nil,
		})
	}
}
