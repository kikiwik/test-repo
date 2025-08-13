package controllers

import (
	"net/http"
	"personaltask/models"
	"personaltask/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProjectController struct {
	DB *gorm.DB
}

func NewProjectController(db *gorm.DB) *ProjectController {
	return &ProjectController{DB: db}
}

// 获取项目列表
func (pc *ProjectController) GetProjects(c *gin.Context) {
	userID := utils.GetUserID(c)
	page, pageSize, offset := utils.GetPaginationParams(c)

	// 构建查询
	query := pc.DB.Model(&models.Project{}).Where("user_id = ?", userID)

	// 状态过滤
	if status := c.Query("status"); status != "" {
		if utils.IsValidProjectStatus(status) {
			query = query.Where("status = ?", status)
		}
	}

	// 关键词搜索
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 排序
	orderBy := c.DefaultQuery("order_by", "created_at")
	orderDir := c.DefaultQuery("order_dir", "desc")
	query = query.Order(orderBy + " " + orderDir)

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页查询
	var projects []models.Project
	if err := query.Offset(offset).Limit(pageSize).Find(&projects).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询项目失败", err)
		return
	}

	// 如果需要包含任务统计
	if c.Query("with_stats") == "true" {
		type ProjectWithStats struct {
			models.Project
			TotalTasks     int64 `json:"total_tasks"`
			CompletedTasks int64 `json:"completed_tasks"`
			Progress       float64 `json:"progress"`
		}

		var projectsWithStats []ProjectWithStats
		for _, project := range projects {
			var totalTasks, completedTasks int64
			pc.DB.Model(&models.Task{}).Where("project_id = ? AND user_id = ?", project.ID, userID).Count(&totalTasks)
			pc.DB.Model(&models.Task{}).Where("project_id = ? AND user_id = ? AND status = ?", project.ID, userID, "completed").Count(&completedTasks)

			progress := 0.0
			if totalTasks > 0 {
				progress = float64(completedTasks) / float64(totalTasks) * 100
			}

			projectsWithStats = append(projectsWithStats, ProjectWithStats{
				Project:        project,
				TotalTasks:     totalTasks,
				CompletedTasks: completedTasks,
				Progress:       progress,
			})
		}

		utils.PaginatedResponse(c, projectsWithStats, total, page, pageSize)
		return
	}

	utils.PaginatedResponse(c, projects, total, page, pageSize)
}

// 创建项目
func (pc *ProjectController) CreateProject(c *gin.Context) {
	userID := utils.GetUserID(c)

	var req models.ProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err)
		return
	}

	// 检查项目名称是否已存在
	var existingProject models.Project
	if err := pc.DB.Where("name = ? AND user_id = ?", req.Name, userID).First(&existingProject).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "项目名称已存在", nil)
		return
	}

	project := models.Project{
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		UserID:      userID,
	}

	// 如果没有设置状态，使用默认状态
	if project.Status == "" {
		project.Status = "active"
	}

	if err := pc.DB.Create(&project).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "项目创建失败", err)
		return
	}

	utils.SuccessResponse(c, project)
}

// 获取项目详情
func (pc *ProjectController) GetProject(c *gin.Context) {
	userID := utils.GetUserID(c)
	projectID := c.Param("id")

	var project models.Project
	if err := pc.DB.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "项目不存在", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "查询项目失败", err)
		}
		return
	}

	// 如果需要包含任务信息
	if c.Query("with_tasks") == "true" {
		pc.DB.Preload("Tasks", "user_id = ?", userID).First(&project, project.ID)
	}

	utils.SuccessResponse(c, project)
}

// 更新项目
func (pc *ProjectController) UpdateProject(c *gin.Context) {
	userID := utils.GetUserID(c)
	projectID := c.Param("id")

	var req models.ProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err)
		return
	}

	// 查找项目
	var project models.Project
	if err := pc.DB.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "项目不存在", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "查询项目失败", err)
		}
		return
	}

	// 检查项目名称是否已存在（排除当前项目）
	var existingProject models.Project
	if err := pc.DB.Where("name = ? AND user_id = ? AND id != ?", req.Name, userID, projectID).First(&existingProject).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "项目名称已存在", nil)
		return
	}

	// 更新项目
	project.Name = req.Name
	project.Description = req.Description
	if req.Status != "" {
		project.Status = req.Status
	}
	project.StartDate = req.StartDate
	project.EndDate = req.EndDate

	if err := pc.DB.Save(&project).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "项目更新失败", err)
		return
	}

	utils.SuccessResponse(c, project)
}

// 删除项目
func (pc *ProjectController) DeleteProject(c *gin.Context) {
	userID := utils.GetUserID(c)
	projectID := c.Param("id")

	// 检查项目是否存在
	var project models.Project
	if err := pc.DB.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "项目不存在", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "查询项目失败", err)
		}
		return
	}

	// 检查项目下是否有任务
	var taskCount int64
	pc.DB.Model(&models.Task{}).Where("project_id = ? AND user_id = ?", projectID, userID).Count(&taskCount)

	if taskCount > 0 {
		// 如果有任务，询问是否强制删除
		if c.Query("force") != "true" {
			utils.ErrorResponse(c, http.StatusConflict, "项目下存在任务，无法删除。如需强制删除，请添加 force=true 参数", nil)
			return
		}

		// 强制删除：将关联任务的项目ID设为null
		if err := pc.DB.Model(&models.Task{}).Where("project_id = ? AND user_id = ?", projectID, userID).Update("project_id", nil).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "清理关联任务失败", err)
			return
		}
	}

	// 删除项目
	if err := pc.DB.Delete(&project).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "项目删除失败", err)
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "项目删除成功"})
}

// 获取项目下的任务
func (pc *ProjectController) GetProjectTasks(c *gin.Context) {
	userID := utils.GetUserID(c)
	projectID := c.Param("id")
	page, pageSize, offset := utils.GetPaginationParams(c)

	// 验证项目存在
	var project models.Project
	if err := pc.DB.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "项目不存在", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "查询项目失败", err)
		}
		return
	}

	// 构建查询
	query := pc.DB.Model(&models.Task{}).Where("project_id = ? AND user_id = ?", projectID, userID)

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

	// 排序
	orderBy := c.DefaultQuery("order_by", "created_at")
	orderDir := c.DefaultQuery("order_dir", "desc")
	query = query.Order(orderBy + " " + orderDir)

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页查询
	var tasks []models.Task
	if err := query.Preload("Category").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询任务失败", err)
		return
	}

	utils.PaginatedResponse(c, tasks, total, page, pageSize)
}

// 获取项目统计信息
func (pc *ProjectController) GetProjectStats(c *gin.Context) {
	userID := utils.GetUserID(c)
	projectID := c.Param("id")

	// 验证项目存在
	var project models.Project
	if err := pc.DB.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "项目不存在", nil)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "查询项目失败", err)
		}
		return
	}

	// 统计任务数量
	var totalTasks, pendingTasks, inProgressTasks, completedTasks int64

	pc.DB.Model(&models.Task{}).Where("project_id = ? AND user_id = ?", projectID, userID).Count(&totalTasks)
	pc.DB.Model(&models.Task{}).Where("project_id = ? AND user_id = ? AND status = ?", projectID, userID, "pending").Count(&pendingTasks)
	pc.DB.Model(&models.Task{}).Where("project_id = ? AND user_id = ? AND status = ?", projectID, userID, "in_progress").Count(&inProgressTasks)
	pc.DB.Model(&models.Task{}).Where("project_id = ? AND user_id = ? AND status = ?", projectID, userID, "completed").Count(&completedTasks)

	// 统计优先级分布
	var lowPriorityTasks, mediumPriorityTasks, highPriorityTasks, urgentPriorityTasks int64
	pc.DB.Model(&models.Task{}).Where("project_id = ? AND user_id = ? AND priority = ?", projectID, userID, "low").Count(&lowPriorityTasks)
	pc.DB.Model(&models.Task{}).Where("project_id = ? AND user_id = ? AND priority = ?", projectID, userID, "medium").Count(&mediumPriorityTasks)
	pc.DB.Model(&models.Task{}).Where("project_id = ? AND user_id = ? AND priority = ?", projectID, userID, "high").Count(&highPriorityTasks)
	pc.DB.Model(&models.Task{}).Where("project_id = ? AND user_id = ? AND priority = ?", projectID, userID, "urgent").Count(&urgentPriorityTasks)

	stats := gin.H{
		"project":           project,
		"total_tasks":       totalTasks,
		"pending_tasks":     pendingTasks,
		"in_progress_tasks": inProgressTasks,
		"completed_tasks":   completedTasks,
		"completion_rate":   0.0,
		"priority_stats": gin.H{
			"low":    lowPriorityTasks,
			"medium": mediumPriorityTasks,
			"high":   highPriorityTasks,
			"urgent": urgentPriorityTasks,
		},
	}

	if totalTasks > 0 {
		stats["completion_rate"] = float64(completedTasks) / float64(totalTasks) * 100
	}

	utils.SuccessResponse(c, stats)
}