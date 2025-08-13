package routes

import (
	"personaltask/config"
	"personaltask/controllers"
	"personaltask/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	// 创建Gin引擎
	router := gin.New()

	// 添加中间件
	router.Use(middleware.Logger())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit())

	// 初始化控制器
	authController := controllers.NewAuthController(db, cfg)
	taskController := controllers.NewTaskController(db)
	categoryController := controllers.NewCategoryController(db)
	projectController := controllers.NewProjectController(db)
	statsController := controllers.NewStatsController(db)

	// API路由组
	api := router.Group("/api")
	{
		// 认证路由（无需JWT认证）
		auth := api.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
		}

		// 需要JWT认证的路由
		protected := api.Group("/")
		protected.Use(middleware.JWTAuth(cfg))
		protected.Use(middleware.RequireAuth(db))
		{
			// 用户信息路由
			userGroup := protected.Group("/auth")
			{
				userGroup.GET("/profile", authController.GetProfile)
				userGroup.PUT("/profile", authController.UpdateProfile)
			}

			// 任务管理路由
			taskGroup := protected.Group("/tasks")
			{
				taskGroup.GET("", taskController.GetTasks)
				taskGroup.POST("", taskController.CreateTask)
				taskGroup.GET("/:id", middleware.ResourceOwnership(db, "task"), taskController.GetTask)
				taskGroup.PUT("/:id", middleware.ResourceOwnership(db, "task"), taskController.UpdateTask)
				taskGroup.DELETE("/:id", middleware.ResourceOwnership(db, "task"), taskController.DeleteTask)
				taskGroup.PATCH("/:id/status", middleware.ResourceOwnership(db, "task"), taskController.UpdateTaskStatus)
				
				// 批量操作
				taskGroup.PATCH("/batch/status", taskController.BatchUpdateTaskStatus)
				taskGroup.DELETE("/batch", taskController.BatchDeleteTasks)
			}

			// 分类管理路由
			categoryGroup := protected.Group("/categories")
			{
				categoryGroup.GET("", categoryController.GetCategories)
				categoryGroup.POST("", categoryController.CreateCategory)
				categoryGroup.GET("/:id", middleware.ResourceOwnership(db, "category"), categoryController.GetCategory)
				categoryGroup.PUT("/:id", middleware.ResourceOwnership(db, "category"), categoryController.UpdateCategory)
				categoryGroup.DELETE("/:id", middleware.ResourceOwnership(db, "category"), categoryController.DeleteCategory)
				categoryGroup.GET("/:id/stats", middleware.ResourceOwnership(db, "category"), categoryController.GetCategoryStats)
			}

			// 项目管理路由
			projectGroup := protected.Group("/projects")
			{
				projectGroup.GET("", projectController.GetProjects)
				projectGroup.POST("", projectController.CreateProject)
				projectGroup.GET("/:id", middleware.ResourceOwnership(db, "project"), projectController.GetProject)
				projectGroup.PUT("/:id", middleware.ResourceOwnership(db, "project"), projectController.UpdateProject)
				projectGroup.DELETE("/:id", middleware.ResourceOwnership(db, "project"), projectController.DeleteProject)
				projectGroup.GET("/:id/tasks", middleware.ResourceOwnership(db, "project"), projectController.GetProjectTasks)
				projectGroup.GET("/:id/stats", middleware.ResourceOwnership(db, "project"), projectController.GetProjectStats)
			}

			// 统计分析路由
			statsGroup := protected.Group("/stats")
			{
				statsGroup.GET("/overview", statsController.GetOverview)
				statsGroup.GET("/daily", statsController.GetDailyStats)
				statsGroup.GET("/weekly", statsController.GetWeeklyStats)
				statsGroup.GET("/productivity", statsController.GetProductivityStats)
				statsGroup.GET("/monthly", statsController.GetMonthlyReport)
			}
		}
	}

	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Personal Task Management API is running",
		})
	})

	// API文档端点（开发环境）
	if cfg.Environment == "development" {
		router.GET("/docs", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "API Documentation",
				"endpoints": gin.H{
					"auth": gin.H{
						"POST /api/auth/register":    "用户注册",
						"POST /api/auth/login":       "用户登录",
						"GET  /api/auth/profile":     "获取用户信息",
						"PUT  /api/auth/profile":     "更新用户信息",
					},
					"tasks": gin.H{
						"GET    /api/tasks":              "获取任务列表",
						"POST   /api/tasks":              "创建任务",
						"GET    /api/tasks/:id":          "获取任务详情",
						"PUT    /api/tasks/:id":          "更新任务",
						"DELETE /api/tasks/:id":          "删除任务",
						"PATCH  /api/tasks/:id/status":   "更新任务状态",
						"PATCH  /api/tasks/batch/status": "批量更新任务状态",
						"DELETE /api/tasks/batch":        "批量删除任务",
					},
					"categories": gin.H{
						"GET    /api/categories":        "获取分类列表",
						"POST   /api/categories":        "创建分类",
						"GET    /api/categories/:id":    "获取分类详情",
						"PUT    /api/categories/:id":    "更新分类",
						"DELETE /api/categories/:id":    "删除分类",
						"GET    /api/categories/:id/stats": "获取分类统计",
					},
					"projects": gin.H{
						"GET    /api/projects":           "获取项目列表",
						"POST   /api/projects":           "创建项目",
						"GET    /api/projects/:id":       "获取项目详情",
						"PUT    /api/projects/:id":       "更新项目",
						"DELETE /api/projects/:id":       "删除项目",
						"GET    /api/projects/:id/tasks": "获取项目任务",
						"GET    /api/projects/:id/stats": "获取项目统计",
					},
					"stats": gin.H{
						"GET /api/stats/overview":     "任务概览统计",
						"GET /api/stats/daily":        "每日任务统计",
						"GET /api/stats/weekly":       "每周任务统计",
						"GET /api/stats/productivity": "工作效率分析",
						"GET /api/stats/monthly":      "月度报告",
					},
				},
			})
		})
	}

	return router
}