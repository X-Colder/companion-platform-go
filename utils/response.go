package utils

import "github.com/gin-gonic/gin"

// Response 统一返回结构体
type Response struct {
	Code int         `json:"code"` // 状态码：200-成功，500-失败，401-未认证，403-无权限
	Msg  string      `json:"msg"`  // 提示信息
	Data interface{} `json:"data"` // 返回数据
}

// Success 成功返回
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code: 200,
		Msg:  "操作成功",
		Data: data,
	})
}

// Fail 失败返回
func Fail(c *gin.Context, msg string) {
	c.JSON(200, Response{
		Code: 500,
		Msg:  msg,
		Data: nil,
	})
}

// Unauthorized 未认证返回（JWT失效/未登录）
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(200, Response{
		Code: 401,
		Msg:  msg,
		Data: nil,
	})
}

// Forbidden 无权限返回
func Forbidden(c *gin.Context, msg string) {
	c.JSON(200, Response{
		Code: 403,
		Msg:  msg,
		Data: nil,
	})
}
