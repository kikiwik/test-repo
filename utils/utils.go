package utils

import (
	"net/http"
	"personaltask/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWT Claims结构
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// 生成JWT Token
func GenerateToken(userID uint, username, secretKey string, expiresIn int) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiresIn) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// 密码加密
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// 成功响应
func SuccessResponse(c *gin.Context, data interface{}) {
	response := models.Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      data,
		Timestamp: time.Now(),
	}
	c.JSON(http.StatusOK, response)
}

// 错误响应
func ErrorResponse(c *gin.Context, code int, message string, err interface{}) {
	response := models.Response{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
	}

	if err != nil {
		if e, ok := err.(error); ok {
			response.Error = e.Error()
		} else if s, ok := err.(string); ok {
			response.Error = s
		}
	}

	c.JSON(code, response)
}

// 分页响应
func PaginatedResponse(c *gin.Context, items interface{}, total int64, page, pageSize int) {
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	data := models.PaginatedResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	SuccessResponse(c, data)
}

// 获取分页参数
func GetPaginationParams(c *gin.Context) (int, int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	return page, pageSize, offset
}

// 获取用户ID
func GetUserID(c *gin.Context) uint {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return userID.(uint)
}

// 获取当前用户
func GetCurrentUser(c *gin.Context) (models.User, bool) {
	user, exists := c.Get("current_user")
	if !exists {
		return models.User{}, false
	}
	return user.(models.User), true
}

// 验证任务状态
func IsValidTaskStatus(status string) bool {
	validStatuses := []string{"pending", "in_progress", "completed"}
	for _, v := range validStatuses {
		if v == status {
			return true
		}
	}
	return false
}

// 验证任务优先级
func IsValidTaskPriority(priority string) bool {
	validPriorities := []string{"low", "medium", "high", "urgent"}
	for _, v := range validPriorities {
		if v == priority {
			return true
		}
	}
	return false
}

// 验证项目状态
func IsValidProjectStatus(status string) bool {
	validStatuses := []string{"active", "completed", "archived"}
	for _, v := range validStatuses {
		if v == status {
			return true
		}
	}
	return false
}

// 字符串数组包含检查
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// 安全的字符串转换
func SafeStringConvert(value interface{}) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

// 时间格式化
func FormatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// 日期格式化
func FormatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

// 安全的整数转换
func SafeIntConvert(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}