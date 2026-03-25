package mysql

import (
	"fmt"
	"log"

	"github.com/haolipeng/BeeGuard/server/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库连接
var DB *gorm.DB

// SetDB 设置共享的数据库连接（与 db.GetDB() 共享同一个连接池）
func SetDB(database *gorm.DB) {
	DB = database
	log.Println("mysql.DB 已设置为共享连接")
}

// parseGormLogLevel 解析 GORM 日志级别字符串
func parseGormLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Error // 默认只记录错误
	}
}

// InitDB 初始化数据库连接（独立连接模式，一般不使用）
func InitDB() error {
	cfg := config.AppConfig.Database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port)

	// 根据配置设置 GORM 日志级别
	logLevel := parseGormLogLevel(cfg.GormLogLevel)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	// 获取底层sql.DB对象以测试连接
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取底层DB对象失败: %v", err)
	}

	// 测试连接
	err = sqlDB.Ping()
	if err != nil {
		return fmt.Errorf("数据库ping失败: %v", err)
	}

	log.Println("数据库连接成功")
	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			sqlDB.Close()
			log.Println("数据库连接已关闭")
		}
	}
}
