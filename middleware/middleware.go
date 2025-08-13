package middleware

import (
	"fmt"
	"net/http"
	"personaltask/config"
	"personaltask/models"
	"personaltask/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// JWT认证中间件
func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "缺少认证令牌", nil)
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "认证令牌格式错误", nil)
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析和验证token
		token, err := jwt.ParseWithClaims(tokenString, &utils.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWT.SecretKey), nil
		})

		if err != nil || !token.Valid {
			utils.ErrorResponse(c, http.StatusUnauthorized, "认证令牌无效", err)
			c.Abort()
			return
		}

		// 提取用户信息
		if claims, ok := token.Claims.(*utils.Claims); ok {
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
		} else {
			utils.ErrorResponse(c, http.StatusUnauthorized, "认证令牌解析失败", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORS跨域中间件
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 生产环境应该限制具体域名
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// 日志中间件
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 自定义日志格式
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 处理panic
		if err := recover(); err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "服务器内部错误", err)
		}
	}
}

// 限流中间件（简单实现）
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以实现更复杂的限流逻辑
		// 例如使用Redis存储访问计数
		c.Next()
	}
}

// 权限验证中间件
func RequireAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "用户未登录", nil)
			c.Abort()
			return
		}

		// 验证用户是否存在
		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "用户不存在", err)
			c.Abort()
			return
		}

		c.Set("current_user", user)
		c.Next()
	}
}

// 资源所有权验证中间件
func ResourceOwnership(db *gorm.DB, resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		resourceIDStr := c.Param("id")
		resourceID, err := strconv.ParseUint(resourceIDStr, 10, 32)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "无效的资源ID", err)
			c.Abort()
			return
		}

		var count int64
		switch resourceType {
		case "task":
			db.Model(&models.Task{}).Where("id = ? AND user_id = ?", resourceID, userID).Count(&count)
		case "category":
			db.Model(&models.Category{}).Where("id = ? AND user_id = ?", resourceID, userID).Count(&count)
		case "project":
			db.Model(&models.Project{}).Where("id = ? AND user_id = ?", resourceID, userID).Count(&count)
		default:
			utils.ErrorResponse(c, http.StatusBadRequest, "不支持的资源类型", nil)
			c.Abort()
			return
		}

		if count == 0 {
			utils.ErrorResponse(c, http.StatusForbidden, "无权访问该资源", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}