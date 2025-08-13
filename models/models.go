package models

import (
	"time"

	"gorm.io/gorm"
)

// 用户模型
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Password  string         `json:"-" gorm:"size:255;not null"`
	Email     string         `json:"email" gorm:"size:100"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Categories []Category `json:"categories,omitempty" gorm:"foreignKey:UserID"`
	Projects   []Project  `json:"projects,omitempty" gorm:"foreignKey:UserID"`
	Tasks      []Task     `json:"tasks,omitempty" gorm:"foreignKey:UserID"`
}

// 任务分类模型
type Category struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:50;not null"`
	Description string         `json:"description" gorm:"type:text"`
	Color       string         `json:"color" gorm:"size:7;default:#007bff"`
	UserID      uint           `json:"user_id" gorm:"not null"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	User  User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Tasks []Task `json:"tasks,omitempty" gorm:"foreignKey:CategoryID"`
}

// 项目模型
type Project struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:100;not null"`
	Description string         `json:"description" gorm:"type:text"`
	Status      string         `json:"status" gorm:"type:enum('active','completed','archived');default:active"`
	StartDate   *time.Time     `json:"start_date" gorm:"type:date"`
	EndDate     *time.Time     `json:"end_date" gorm:"type:date"`
	UserID      uint           `json:"user_id" gorm:"not null"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	User  User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Tasks []Task `json:"tasks,omitempty" gorm:"foreignKey:ProjectID"`
}

// 任务模型
type Task struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title" gorm:"size:200;not null"`
	Description string         `json:"description" gorm:"type:text"`
	Status      string         `json:"status" gorm:"type:enum('pending','in_progress','completed');default:pending"`
	Priority    string         `json:"priority" gorm:"type:enum('low','medium','high','urgent');default:medium"`
	DueDate     *time.Time     `json:"due_date"`
	CompletedAt *time.Time     `json:"completed_at"`
	UserID      uint           `json:"user_id" gorm:"not null"`
	CategoryID  *uint          `json:"category_id"`
	ProjectID   *uint          `json:"project_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	User     User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Category *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Project  *Project  `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

// 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"omitempty,email"`
}

// 用户登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 任务创建/更新请求
type TaskRequest struct {
	Title       string     `json:"title" binding:"required,max=200"`
	Description string     `json:"description"`
	Priority    string     `json:"priority" binding:"omitempty,oneof=low medium high urgent"`
	DueDate     *time.Time `json:"due_date"`
	CategoryID  *uint      `json:"category_id"`
	ProjectID   *uint      `json:"project_id"`
}

// 任务状态更新请求
type TaskStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending in_progress completed"`
}

// 分类创建/更新请求
type CategoryRequest struct {
	Name        string `json:"name" binding:"required,max=50"`
	Description string `json:"description"`
	Color       string `json:"color" binding:"omitempty,len=7"`
}

// 项目创建/更新请求
type ProjectRequest struct {
	Name        string     `json:"name" binding:"required,max=100"`
	Description string     `json:"description"`
	Status      string     `json:"status" binding:"omitempty,oneof=active completed archived"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}

// API响应结构
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// 分页响应结构
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// 统计响应结构
type StatsOverview struct {
	TotalTasks      int64 `json:"total_tasks"`
	PendingTasks    int64 `json:"pending_tasks"`
	InProgressTasks int64 `json:"in_progress_tasks"`
	CompletedTasks  int64 `json:"completed_tasks"`
	TotalProjects   int64 `json:"total_projects"`
	ActiveProjects  int64 `json:"active_projects"`
	TotalCategories int64 `json:"total_categories"`
}

// 每日统计
type DailyStats struct {
	Date           string `json:"date"`
	TasksCreated   int64  `json:"tasks_created"`
	TasksCompleted int64  `json:"tasks_completed"`
}

// JWT Claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}