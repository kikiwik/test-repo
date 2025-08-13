package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Environment string
	ServerPort  string
	Database    DatabaseConfig
	JWT         JWTConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type JWTConfig struct {
	SecretKey string
	ExpiresIn int // 小时
}

func Load() *Config {
	// 加载.env文件
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载.env文件: %v", err)
		log.Println("将使用系统环境变量或默认值")
	}

	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "personaltask"),
		},
		JWT: JWTConfig{
			SecretKey: getEnv("JWT_SECRET", "your-super-secret-key"),
			ExpiresIn: 24, // 24小时
		},
	}
}

func InitDB(cfg *Config) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
	)

	log.Printf("正在连接数据库: %s@%s:%s/%s", 
		cfg.Database.User, 
		cfg.Database.Host, 
		cfg.Database.Port, 
		cfg.Database.DBName)

	var logLevel logger.LogLevel
	if cfg.Environment == "production" {
		logLevel = logger.Silent
	} else {
		logLevel = logger.Info
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	log.Println("数据库连接成功！")
	return db
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}