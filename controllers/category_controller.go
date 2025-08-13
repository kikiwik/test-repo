package controllers

import (
	"net/http"
	"personaltask/models"
	"personaltask/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CategoryController struct {
	DB *gorm.DB
}

func NewCategoryController(db *gorm.DB) *CategoryController {
	return &CategoryController{DB: db}
}

// 获取分类列表
func (cc *CategoryController) GetCategories(c *gin.Context) {
	userID := utils.GetUserID(c)

	var categories []models.Category
	query := cc.DB.Where("user_id = ?", userID)

	// 排序
	orderBy := c.DefaultQuery("order_by", "created_at")
	orderDir := c.DefaultQuery("order_dir", "asc")
	query = query.Order(orderBy + " " + orderDir)

	// 是否包含任务数量统计
	if c.Query("with_count") == "true" {
		type CategoryWithCount struct {
			models.Category
			TaskCount int64 `json:"task_count"`
		}

		var categoriesWithCount []CategoryWithCount
		if err := query.Find(&categories).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "查询分类失败", err)
			return
		}

		for _, category := range categories {
			var taskCount int64
			cc.DB.Model(&models.Task{}).Where("category_id = ? AND user_id = ?", category.ID, userID).Count(&taskCount)
			
			categoriesWithCount = append(categoriesWithCount, CategoryWithCount{
				Category:  category,
				TaskCount: taskCount,
			})
		}

		utils.SuccessResponse(c, categoriesWithCount)
		return
	}

	if err := query.Find(&categories).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询分类失败", err)
		return
	}

	utils.SuccessResponse(c, categories)
}

// 创建分类
func (cc *CategoryController) CreateCategory(c *gin.Context) {
	userID := utils.GetUserID(c)

	var req models.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err)
		return
	}

	// 检查分类名称是否已存在
	var existingCategory models.Category
	if err := cc.DB.Where("name = ? AND user_id = ?", req.Name, userID).First(&existingCategory).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "分类名称已存在", nil)
		return
	}

	category := models.Category{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		UserID:      userID,
	}

	// 如果没有设置颜色，使用默认颜色
	if category.Color == "" {
		category.Color = "#007bff"
	}

	if err := cc.DB.Create(&category).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "分类创建失败", err)
		return
	}

	utils.SuccessResponse(c, category)
}

// 获取分类详情
func (cc *CategoryController) GetCategory(c *gin.Context) {
	userID := utils.GetUserID(c)
	categoryID := c.Param("id")

	var category models.Category
	if err := cc.DB.Where("id = ? AND user_id = ?", categoryID, userID).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "分类不存在", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "查询分类失败", err)
		}
		return
	}

	// 如果需要包含任务信息
	if c.Query("with_tasks") == "true" {
		cc.DB.Preload("Tasks", "user_id = ?", userID).First(&category, category.ID)
	}

	utils.SuccessResponse(c, category)
}

// 更新分类
func (cc *CategoryController) UpdateCategory(c *gin.Context) {
	userID := utils.GetUserID(c)
	categoryID := c.Param("id")

	var req models.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err)
		return
	}

	// 查找分类
	var category models.Category
	if err := cc.DB.Where("id = ? AND user_id = ?", categoryID, userID).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "分类不存在", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "查询分类失败", err)
		}
		return
	}

	// 检查分类名称是否已存在（排除当前分类）
	var existingCategory models.Category
	if err := cc.DB.Where("name = ? AND user_id = ? AND id != ?", req.Name, userID, categoryID).First(&existingCategory).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "分类名称已存在", nil)
		return
	}

	// 更新分类
	category.Name = req.Name
	category.Description = req.Description
	if req.Color != "" {
		category.Color = req.Color
	}

	if err := cc.DB.Save(&category).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "分类更新失败", err)
		return
	}

	utils.SuccessResponse(c, category)
}

// 删除分类
func (cc *CategoryController) DeleteCategory(c *gin.Context) {
	userID := utils.GetUserID(c)
	categoryID := c.Param("id")

	// 检查分类是否存在
	var category models.Category
	if err := cc.DB.Where("id = ? AND user_id = ?", categoryID, userID).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "分类不存在", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "查询分类失败", err)
		}
		return
	}

	// 检查分类下是否有任务
	var taskCount int64
	cc.DB.Model(&models.Task{}).Where("category_id = ? AND user_id = ?", categoryID, userID).Count(&taskCount)

	if taskCount > 0 {
		// 如果有任务，询问是否强制删除
		if c.Query("force") != "true" {
			utils.ErrorResponse(c, http.StatusConflict, "分类下存在任务，无法删除。如需强制删除，请添加 force=true 参数", nil)
			return
		}

		// 强制删除：将关联任务的分类ID设为null
		if err := cc.DB.Model(&models.Task{}).Where("category_id = ? AND user_id = ?", categoryID, userID).Update("category_id", nil).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "清理关联任务失败", err)
			return
		}
	}

	// 删除分类
	if err := cc.DB.Delete(&category).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "分类删除失败", err)
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "分类删除成功"})
}

// 获取分类统计信息
func (cc *CategoryController) GetCategoryStats(c *gin.Context) {
	userID := utils.GetUserID(c)
	categoryID := c.Param("id")

	// 验证分类存在
	var category models.Category
	if err := cc.DB.Where("id = ? AND user_id = ?", categoryID, userID).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "分类不存在", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "查询分类失败", err)
		}
		return
	}

	// 统计任务数量
	var totalTasks, pendingTasks, inProgressTasks, completedTasks int64

	cc.DB.Model(&models.Task{}).Where("category_id = ? AND user_id = ?", categoryID, userID).Count(&totalTasks)
	cc.DB.Model(&models.Task{}).Where("category_id = ? AND user_id = ? AND status = ?", categoryID, userID, "pending").Count(&pendingTasks)
	cc.DB.Model(&models.Task{}).Where("category_id = ? AND user_id = ? AND status = ?", categoryID, userID, "in_progress").Count(&inProgressTasks)
	cc.DB.Model(&models.Task{}).Where("category_id = ? AND user_id = ? AND status = ?", categoryID, userID, "completed").Count(&completedTasks)

	stats := gin.H{
		"category":          category,
		"total_tasks":       totalTasks,
		"pending_tasks":     pendingTasks,
		"in_progress_tasks": inProgressTasks,
		"completed_tasks":   completedTasks,
		"completion_rate":   0.0,
	}

	if totalTasks > 0 {
		stats["completion_rate"] = float64(completedTasks) / float64(totalTasks) * 100
	}

	utils.SuccessResponse(c, stats)
}