// middleware/logger.go
package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 自定义请求日志中间件
// 记录：请求时间、客户端IP、请求方法、请求路径、响应状态码、处理耗时
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 记录请求开始时间
		startTime := time.Now()

		// 2. 执行后续中间件/控制器逻辑
		c.Next()

		// 3. 请求结束后，计算耗时
		latencyTime := time.Since(startTime)

		// 4. 获取关键请求/响应信息
		clientIP := c.ClientIP()        // 客户端IP
		reqMethod := c.Request.Method   // 请求方法（GET/POST/PUT/DELETE等）
		reqUri := c.Request.RequestURI  // 请求路径
		statusCode := c.Writer.Status() // 响应状态码（200/404/500等）

		// 5. 格式化输出日志（按固定格式打印，方便后续日志分析）
		log.Printf(
			"[GIN] %s | %3d | %13v | %15s | %-7s %s",
			time.Now().Format("2006-01-02 15:04:05"), // 请求时间（格式化）
			statusCode,                               // 状态码
			latencyTime,                              // 处理耗时
			clientIP,                                 // 客户端IP
			reqMethod,                                // 请求方法
			reqUri,                                   // 请求路径
		)
	}
}
