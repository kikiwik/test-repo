package controllers

import (
	"net/http"
	"personaltask/models"
	"personaltask/utils"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StatsController struct {
	DB *gorm.DB
}

func NewStatsController(db *gorm.DB) *StatsController {
	return &StatsController{DB: db}
}

// 任务概览统计
func (sc *StatsController) GetOverview(c *gin.Context) {
	userID := utils.GetUserID(c)

	var overview models.StatsOverview

	// 统计任务
	sc.DB.Model(&models.Task{}).Where("user_id = ?", userID).Count(&overview.TotalTasks)
	sc.DB.Model(&models.Task{}).Where("user_id = ? AND status = ?", userID, "pending").Count(&overview.PendingTasks)
	sc.DB.Model(&models.Task{}).Where("user_id = ? AND status = ?", userID, "in_progress").Count(&overview.InProgressTasks)
	sc.DB.Model(&models.Task{}).Where("user_id = ? AND status = ?", userID, "completed").Count(&overview.CompletedTasks)

	// 统计项目
	sc.DB.Model(&models.Project{}).Where("user_id = ?", userID).Count(&overview.TotalProjects)
	sc.DB.Model(&models.Project{}).Where("user_id = ? AND status = ?", userID, "active").Count(&overview.ActiveProjects)

	// 统计分类
	sc.DB.Model(&models.Category{}).Where("user_id = ?", userID).Count(&overview.TotalCategories)

	utils.SuccessResponse(c, overview)
}

// 每日任务统计
func (sc *StatsController) GetDailyStats(c *gin.Context) {
	userID := utils.GetUserID(c)

	// 获取日期范围参数
	daysStr := c.DefaultQuery("days", "7") // 默认最近7天
	days := 7
	if d := utils.SafeStringConvert(daysStr); d != "" {
		if parsed, err := utils.SafeIntConvert(d); err == nil && parsed > 0 && parsed <= 30 {
			days = parsed
		}
	}

	var dailyStats []models.DailyStats

	// 生成最近几天的统计数据
	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")

		var tasksCreated, tasksCompleted int64

		// 统计当天创建的任务
		sc.DB.Model(&models.Task{}).
			Where("user_id = ? AND DATE(created_at) = ?", userID, dateStr).
			Count(&tasksCreated)

		// 统计当天完成的任务
		sc.DB.Model(&models.Task{}).
			Where("user_id = ? AND DATE(completed_at) = ?", userID, dateStr).
			Count(&tasksCompleted)

		dailyStats = append(dailyStats, models.DailyStats{
			Date:           dateStr,
			TasksCreated:   tasksCreated,
			TasksCompleted: tasksCompleted,
		})
	}

	utils.SuccessResponse(c, dailyStats)
}

// 每周任务统计
func (sc *StatsController) GetWeeklyStats(c *gin.Context) {
	userID := utils.GetUserID(c)

	// 获取周数参数
	weeksStr := c.DefaultQuery("weeks", "4") // 默认最近4周
	weeks := 4
	if w := utils.SafeStringConvert(weeksStr); w != "" {
		if parsed, err := utils.SafeIntConvert(w); err == nil && parsed > 0 && parsed <= 12 {
			weeks = parsed
		}
	}

	type WeeklyStats struct {
		Week           string `json:"week"`
		TasksCreated   int64  `json:"tasks_created"`
		TasksCompleted int64  `json:"tasks_completed"`
	}

	var weeklyStats []WeeklyStats

	// 生成最近几周的统计数据
	for i := weeks - 1; i >= 0; i-- {
		// 计算周的开始和结束日期
		now := time.Now()
		weekStart := now.AddDate(0, 0, -int(now.Weekday())+1-i*7) // 本周周一
		weekEnd := weekStart.AddDate(0, 0, 6)                      // 本周周日

		weekStr := weekStart.Format("2006-01-02") + " 至 " + weekEnd.Format("2006-01-02")

		var tasksCreated, tasksCompleted int64

		// 统计本周创建的任务
		sc.DB.Model(&models.Task{}).
			Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, weekStart, weekEnd.Add(24*time.Hour)).
			Count(&tasksCreated)

		// 统计本周完成的任务
		sc.DB.Model(&models.Task{}).
			Where("user_id = ? AND completed_at >= ? AND completed_at <= ?", userID, weekStart, weekEnd.Add(24*time.Hour)).
			Count(&tasksCompleted)

		weeklyStats = append(weeklyStats, WeeklyStats{
			Week:           weekStr,
			TasksCreated:   tasksCreated,
			TasksCompleted: tasksCompleted,
		})
	}

	utils.SuccessResponse(c, weeklyStats)
}

// 工作效率分析
func (sc *StatsController) GetProductivityStats(c *gin.Context) {
	userID := utils.GetUserID(c)

	// 基础统计
	var totalTasks, completedTasks int64
	sc.DB.Model(&models.Task{}).Where("user_id = ?", userID).Count(&totalTasks)
	sc.DB.Model(&models.Task{}).Where("user_id = ? AND status = ?", userID, "completed").Count(&completedTasks)

	// 计算完成率
	completionRate := 0.0
	if totalTasks > 0 {
		completionRate = float64(completedTasks) / float64(totalTasks) * 100
	}

	// 优先级分布
	var lowPriority, mediumPriority, highPriority, urgentPriority int64
	sc.DB.Model(&models.Task{}).Where("user_id = ? AND priority = ?", userID, "low").Count(&lowPriority)
	sc.DB.Model(&models.Task{}).Where("user_id = ? AND priority = ?", userID, "medium").Count(&mediumPriority)
	sc.DB.Model(&models.Task{}).Where("user_id = ? AND priority = ?", userID, "high").Count(&highPriority)
	sc.DB.Model(&models.Task{}).Where("user_id = ? AND priority = ?", userID, "urgent").Count(&urgentPriority)

	// 每个优先级的完成率
	priorityCompletionRates := make(map[string]float64)
	priorities := []string{"low", "medium", "high", "urgent"}
	
	for _, priority := range priorities {
		var total, completed int64
		sc.DB.Model(&models.Task{}).Where("user_id = ? AND priority = ?", userID, priority).Count(&total)
		sc.DB.Model(&models.Task{}).Where("user_id = ? AND priority = ? AND status = ?", userID, priority, "completed").Count(&completed)
		
		rate := 0.0
		if total > 0 {
			rate = float64(completed) / float64(total) * 100
		}
		priorityCompletionRates[priority] = rate
	}

	// 平均完成时间（以小时为单位）
	var avgCompletionTime float64
	type CompletionTime struct {
		Hours float64
	}
	var result CompletionTime
	
	sc.DB.Raw(`
		SELECT AVG(TIMESTAMPDIFF(HOUR, created_at, completed_at)) as hours 
		FROM tasks 
		WHERE user_id = ? AND status = 'completed' AND completed_at IS NOT NULL
	`, userID).Scan(&result)
	
	avgCompletionTime = result.Hours

	// 最近7天的工作效率趋势
	var recentProductivity []gin.H
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")

		var created, completed int64
		sc.DB.Model(&models.Task{}).
			Where("user_id = ? AND DATE(created_at) = ?", userID, dateStr).
			Count(&created)
		sc.DB.Model(&models.Task{}).
			Where("user_id = ? AND DATE(completed_at) = ?", userID, dateStr).
			Count(&completed)

		efficiency := 0.0
		if created > 0 {
			efficiency = float64(completed) / float64(created) * 100
		} else if completed > 0 {
			efficiency = 100.0 // 没有创建但有完成，效率100%
		}

		recentProductivity = append(recentProductivity, gin.H{
			"date":       dateStr,
			"created":    created,
			"completed":  completed,
			"efficiency": efficiency,
		})
	}

	// 分类效率分析
	var categoryStats []gin.H
	var categories []models.Category
	sc.DB.Where("user_id = ?", userID).Find(&categories)

	for _, category := range categories {
		var total, completed int64
		sc.DB.Model(&models.Task{}).Where("user_id = ? AND category_id = ?", userID, category.ID).Count(&total)
		sc.DB.Model(&models.Task{}).Where("user_id = ? AND category_id = ? AND status = ?", userID, category.ID, "completed").Count(&completed)

		rate := 0.0
		if total > 0 {
			rate = float64(completed) / float64(total) * 100
		}

		categoryStats = append(categoryStats, gin.H{
			"category_name":   category.Name,
			"total_tasks":     total,
			"completed_tasks": completed,
			"completion_rate": rate,
		})
	}

	// 逾期任务统计
	var overdueTasks int64
	now := time.Now()
	sc.DB.Model(&models.Task{}).
		Where("user_id = ? AND status != ? AND due_date < ?", userID, "completed", now).
		Count(&overdueTasks)

	// 今日任务统计
	today := now.Format("2006-01-02")
	var todayTasks, todayCompleted int64
	sc.DB.Model(&models.Task{}).
		Where("user_id = ? AND DATE(due_date) = ?", userID, today).
		Count(&todayTasks)
	sc.DB.Model(&models.Task{}).
		Where("user_id = ? AND DATE(due_date) = ? AND status = ?", userID, today, "completed").
		Count(&todayCompleted)

	stats := gin.H{
		"overview": gin.H{
			"total_tasks":      totalTasks,
			"completed_tasks":  completedTasks,
			"completion_rate":  completionRate,
			"overdue_tasks":    overdueTasks,
		},
		"priority_distribution": gin.H{
			"low":    lowPriority,
			"medium": mediumPriority,
			"high":   highPriority,
			"urgent": urgentPriority,
		},
		"priority_completion_rates": priorityCompletionRates,
		"avg_completion_time_hours": avgCompletionTime,
		"recent_productivity":       recentProductivity,
		"category_efficiency":       categoryStats,
		"today": gin.H{
			"total_tasks":     todayTasks,
			"completed_tasks": todayCompleted,
			"completion_rate": func() float64 {
				if todayTasks > 0 {
					return float64(todayCompleted) / float64(todayTasks) * 100
				}
				return 0.0
			}(),
		},
	}

	utils.SuccessResponse(c, stats)
}

// 获取月度报告
func (sc *StatsController) GetMonthlyReport(c *gin.Context) {
	userID := utils.GetUserID(c)

	// 获取月份参数，默认当前月
	monthStr := c.DefaultQuery("month", time.Now().Format("2006-01"))
	month, err := time.Parse("2006-01", monthStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "月份格式错误，应为 YYYY-MM", err)
		return
	}

	// 计算月份的开始和结束日期
	monthStart := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

	// 月度基础统计
	var tasksCreated, tasksCompleted, tasksInProgress int64
	sc.DB.Model(&models.Task{}).
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, monthStart, monthEnd).
		Count(&tasksCreated)
	sc.DB.Model(&models.Task{}).
		Where("user_id = ? AND completed_at >= ? AND completed_at <= ?", userID, monthStart, monthEnd).
		Count(&tasksCompleted)
	sc.DB.Model(&models.Task{}).
		Where("user_id = ? AND status = ? AND created_at >= ? AND created_at <= ?", userID, "in_progress", monthStart, monthEnd).
		Count(&tasksInProgress)

	// 每日创建/完成趋势
	type DailyTrend struct {
		Day       int   `json:"day"`
		Created   int64 `json:"created"`
		Completed int64 `json:"completed"`
	}
	
	var dailyTrends []DailyTrend
	daysInMonth := monthEnd.Day()
	
	for day := 1; day <= daysInMonth; day++ {
		dayStart := time.Date(month.Year(), month.Month(), day, 0, 0, 0, 0, month.Location())
		dayEnd := dayStart.Add(24*time.Hour - time.Second)
		
		var created, completed int64
		sc.DB.Model(&models.Task{}).
			Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, dayStart, dayEnd).
			Count(&created)
		sc.DB.Model(&models.Task{}).
			Where("user_id = ? AND completed_at >= ? AND completed_at <= ?", userID, dayStart, dayEnd).
			Count(&completed)
			
		dailyTrends = append(dailyTrends, DailyTrend{
			Day:       day,
			Created:   created,
			Completed: completed,
		})
	}

	// 项目进展统计
	var projectProgress []gin.H
	var projects []models.Project
	sc.DB.Where("user_id = ?", userID).Find(&projects)
	
	for _, project := range projects {
		var total, completed int64
		sc.DB.Model(&models.Task{}).Where("project_id = ? AND user_id = ?", project.ID, userID).Count(&total)
		sc.DB.Model(&models.Task{}).Where("project_id = ? AND user_id = ? AND status = ?", project.ID, userID, "completed").Count(&completed)
		
		progress := 0.0
		if total > 0 {
			progress = float64(completed) / float64(total) * 100
		}
		
		projectProgress = append(projectProgress, gin.H{
			"project_name": project.Name,
			"total_tasks":  total,
			"completed":    completed,
			"progress":     progress,
		})
	}

	report := gin.H{
		"month": monthStr,
		"summary": gin.H{
			"tasks_created":    tasksCreated,
			"tasks_completed":  tasksCompleted,
			"tasks_in_progress": tasksInProgress,
			"completion_rate": func() float64 {
				if tasksCreated > 0 {
					return float64(tasksCompleted) / float64(tasksCreated) * 100
				}
				return 0.0
			}(),
		},
		"daily_trends":     dailyTrends,
		"project_progress": projectProgress,
	}

	utils.SuccessResponse(c, report)
}