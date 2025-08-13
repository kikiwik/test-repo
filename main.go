package main

import (
	"log"
	"personaltask/config"
	"personaltask/models"
	"personaltask/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	db := config.InitDB(cfg)

	// 自动迁移数据库表
	err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Project{},
		&models.Task{},
	)
	if err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	// 设置Gin模式
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化路由
	router := routes.SetupRouter(db, cfg)

	// 启动服务器
	log.Printf("服务器启动在端口 %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}