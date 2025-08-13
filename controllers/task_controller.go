package controllers

import (
	"net/http"
	"personaltask/models"
	"personaltask/utils"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TaskController struct {
	DB *gorm.DB
}

func NewTaskController(db *gorm.DB) *TaskController {
	return &TaskController{DB: db}
}

// 获取任务列表
func (tc *TaskController) GetTasks(c *gin.Context) {
	userID := utils.GetUserID(c)
	page, pageSize, offset := utils.GetPaginationParams(c)

	// 构建查询
	query := tc.DB.Model(&models.Task{}).Where("user_id = ?", userID)

	// 状态过滤
	if status := c.Query("status"); status != "" {
		if utils.IsValidTaskStatus(status) {
			query = query.Where("status = ?", status)
		}
	}

	// 优先级过滤
	if priority := c.Query("priority"); priority != "" {
		if utils.IsValidTaskPriority(priority) {
			query = query.Where("priority = ?", priority)
		}
	}

	// 分类过滤
	if categoryID := c.Query("category_id"); categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	// 项目过滤
	if projectID := c.Query("project_id"); projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	// 关键词搜索
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 日期范围过滤
	if startDate := c.Query("start_date"); startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}

	// 截止日期过滤
	if dueBefore := c.Query("due_before"); dueBefore != "" {
		query = query.Where("due_date <= ?", dueBefore)
	}

	// 排序
	orderBy := c.DefaultQuery("order_by", "created_at")
	orderDir := c.DefaultQuery("order_dir", "desc")
	query = query.Order(orderBy + " " + orderDir)

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页查询
	var tasks []models.Task
	if err := query.Preload("Category").Preload("Project").
		Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询任务失败", err)
		return
	}

	utils.PaginatedResponse(c, tasks, total, page, pageSize)
}

// 创建任务
func (tc *TaskController) CreateTask(c *gin.Context) {
	userID := utils.GetUserID(c)

	var req models.TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err)
		return
	}

	// 验证分类归属
	if req.CategoryID != nil {
		var category models.Category
		if err := tc.DB.Where("id = ? AND user_id = ?", *req.CategoryID, userID).First(&category).Error; err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "分类不存在或无权限", err)
			return
		}
	}

	// 验证项目归属
	if req.ProjectID != nil {
		var project models.Project
		if err := tc.DB.Where("id = ? AND user_id = ?", *req.ProjectID, userID).First(&project).Error; err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "项目不存在或无权限", err)
			return
		}
	}

	task := models.Task{
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		DueDate:     req.DueDate,
		UserID:      userID,
		CategoryID:  req.CategoryID,
		ProjectID:   req.ProjectID,
		Status:      "pending",
	}

	if err := tc.DB.Create(&task).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "任务创建失败", err)
		return
	}

	// 重新查询以获取关联数据
	tc.DB.Preload("Category").Preload("Project").First(&task, task.ID)

	utils.SuccessResponse(c, task)
}

// 获取任务详情
func (tc *TaskController) GetTask(c *gin.Context) {
	userID := utils.GetUserID(c)
	taskID := c.Param("id")

	var task models.Task
	if err := tc.DB.Preload("Category").Preload("Project").
		Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "任务不存在", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "查询任务失败", err)
		}
		return
	}

	utils.SuccessResponse(c, task)
}

// 更新任务
func (tc *TaskController) UpdateTask(c *gin.Context) {
	userID := utils.GetUserID(c)
	taskID := c.Param("id")

	var req models.TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err)
		return
	}

	// 查找任务
	var task models.Task
	if err := tc.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "任务不存在", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "查询任务失败", err)
		}
		return
	}

	// 验证分类归属
	if req.CategoryID != nil {
		var category models.Category
		if err := tc.DB.Where("id = ? AND user_id = ?", *req.CategoryID, userID).First(&category).Error; err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "分类不存在或无权限", err)
			return
		}
	}

	// 验证项目归属
	if req.ProjectID != nil {
		var project models.Project
		if err := tc.DB.Where("id = ? AND user_id = ?", *req.ProjectID, userID).First(&project).Error; err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "项目不存在或无权限", err)
			return
		}
	}

	// 更新任务
	task.Title = req.Title
	task.Description = req.Description
	task.Priority = req.Priority
	task.DueDate = req.DueDate
	task.CategoryID = req.CategoryID
	task.ProjectID = req.ProjectID

	if err := tc.DB.Save(&task).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "任务更新失败", err)
		return
	}

	// 重新查询以获取关联数据
	tc.DB.Preload("Category").Preload("Project").First(&task, task.ID)

	utils.SuccessResponse(c, task)
}

// 更新任务状态
func (tc *TaskController) UpdateTaskStatus(c *gin.Context) {
	userID := utils.GetUserID(c)
	taskID := c.Param("id")

	var req models.TaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err)
		return
	}

	// 查找任务
	var task models.Task
	if err := tc.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "任务不存在", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "查询任务失败", err)
		}
		return
	}

	// 更新状态
	task.Status = req.Status

	// 如果标记为完成，设置完成时间
	if req.Status == "completed" && task.CompletedAt == nil {
		now := time.Now()
		task.CompletedAt = &now
	} else if req.Status != "completed" {
		task.CompletedAt = nil
	}

	if err := tc.DB.Save(&task).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "状态更新失败", err)
		return
	}

	utils.SuccessResponse(c, task)
}

// 删除任务
func (tc *TaskController) DeleteTask(c *gin.Context) {
	userID := utils.GetUserID(c)
	taskID := c.Param("id")

	// 软删除任务
	if err := tc.DB.Where("id = ? AND user_id = ?", taskID, userID).Delete(&models.Task{}).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "任务删除失败", err)
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "任务删除成功"})
}

// 批量更新任务状态
func (tc *TaskController) BatchUpdateTaskStatus(c *gin.Context) {
	userID := utils.GetUserID(c)

	var req struct {
		TaskIDs []uint `json:"task_ids" binding:"required"`
		Status  string `json:"status" binding:"required,oneof=pending in_progress completed"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err)
		return
	}

	// 验证任务归属并更新
	updates := map[string]interface{}{
		"status": req.Status,
	}

	if req.Status == "completed" {
		updates["completed_at"] = time.Now()
	} else {
		updates["completed_at"] = nil
	}

	result := tc.DB.Model(&models.Task{}).
		Where("id IN ? AND user_id = ?", req.TaskIDs, userID).
		Updates(updates)

	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "批量更新失败", result.Error)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message":        "批量更新成功",
		"affected_count": result.RowsAffected,
	})
}

// 批量删除任务
func (tc *TaskController) BatchDeleteTasks(c *gin.Context) {
	userID := utils.GetUserID(c)

	var req struct {
		TaskIDs []uint `json:"task_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err)
		return
	}

	// 批量软删除
	result := tc.DB.Where("id IN ? AND user_id = ?", req.TaskIDs, userID).Delete(&models.Task{})

	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "批量删除失败", result.Error)
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message":        "批量删除成功",
		"affected_count": result.RowsAffected,
	})
}