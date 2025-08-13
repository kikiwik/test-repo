package controllers

import (
	"net/http"
	"personaltask/config"
	"personaltask/models"
	"personaltask/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct {
	DB     *gorm.DB
	Config *config.Config
}

func NewAuthController(db *gorm.DB, cfg *config.Config) *AuthController {
	return &AuthController{
		DB:     db,
		Config: cfg,
	}
}

// 用户注册
func (ac *AuthController) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err)
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if err := ac.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "用户名已存在", nil)
		return
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "密码加密失败", err)
		return
	}

	// 创建用户
	user := models.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
	}

	if err := ac.DB.Create(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "用户创建失败", err)
		return
	}

	// 生成JWT Token
	token, err := utils.GenerateToken(user.ID, user.Username, ac.Config.JWT.SecretKey, ac.Config.JWT.ExpiresIn)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "令牌生成失败", err)
		return
	}

	response := gin.H{
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"created_at": user.CreatedAt,
		},
		"token": token,
	}

	utils.SuccessResponse(c, response)
}

// 用户登录
func (ac *AuthController) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err)
		return
	}

	// 查找用户
	var user models.User
	if err := ac.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户名或密码错误", nil)
		return
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.Password) {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户名或密码错误", nil)
		return
	}

	// 生成JWT Token
	token, err := utils.GenerateToken(user.ID, user.Username, ac.Config.JWT.SecretKey, ac.Config.JWT.ExpiresIn)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "令牌生成失败", err)
		return
	}

	response := gin.H{
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"created_at": user.CreatedAt,
		},
		"token": token,
	}

	utils.SuccessResponse(c, response)
}

// 获取用户信息
func (ac *AuthController) GetProfile(c *gin.Context) {
	user, exists := utils.GetCurrentUser(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未登录", nil)
		return
	}

	response := gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	utils.SuccessResponse(c, response)
}

// 更新用户信息
func (ac *AuthController) UpdateProfile(c *gin.Context) {
	user, exists := utils.GetCurrentUser(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未登录", nil)
		return
	}

	var req struct {
		Email string `json:"email" binding:"omitempty,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err)
		return
	}

	// 更新用户信息
	if req.Email != "" {
		user.Email = req.Email
	}

	if err := ac.DB.Save(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "用户信息更新失败", err)
		return
	}

	response := gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"updated_at": user.UpdatedAt,
	}

	utils.SuccessResponse(c, response)
}