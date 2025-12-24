package middleware

import (
	"strings"

	"github.com/X-Colder/companion-backend/conf" // 配置读取（需先实现配置加载，下文main.go会提及）
	"github.com/X-Colder/companion-backend/utils"

	"github.com/gin-gonic/gin"
)

// JwtAuth JWT认证中间件
func JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求头中的Authorization
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		// 校验格式：Bearer xxx
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Unauthorized(c, "token格式错误")
			c.Abort()
			return
		}

		// 解析token
		tokenStr := parts[1]
		claims, err := utils.ParseToken(tokenStr, conf.AppConfig.Jwt.Secret)
		if err != nil {
			utils.Unauthorized(c, "token已过期或无效")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("user_type", claims.UserType)

		c.Next()
	}
}

// JwtAuthByRole 按角色认证（如：仅患者/仅陪诊师）
func JwtAuthByRole(role int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行通用JWT认证
		JwtAuth()(c)
		if c.IsAborted() {
			return
		}

		// 获取用户角色
		userType, _ := c.Get("user_type")
		if userType.(int) != role {
			utils.Forbidden(c, "你没有权限访问该接口")
			c.Abort()
			return
		}

		c.Next()
	}
}
