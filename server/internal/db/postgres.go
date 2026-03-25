package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/haolipeng/BeeGuard/server/internal/config"
	"github.com/haolipeng/BeeGuard/server/internal/log"
)

// DB 全局数据库连接
var db *gorm.DB

// Init 初始化数据库连接
func Init(cfg *config.DatabaseConfig) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port)

	// 根据配置设置 GORM 日志级别
	logLevel := parseGormLogLevel(cfg.GormLogLevel)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	// 获取底层sql.DB对象以配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取底层DB对象失败: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(cfg.PoolSize)
	sqlDB.SetMaxIdleConns(cfg.PoolSize / 2)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	log.Infof("[DB] PostgreSQL 连接成功: %s:%d/%s", cfg.Host, cfg.Port, cfg.Database)
	return nil
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return db
}

// Close 关闭数据库连接
func Close() error {
	if db != nil {
		log.Infof("[DB] 关闭数据库连接")
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// Health 健康检查
func Health() error {
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
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
