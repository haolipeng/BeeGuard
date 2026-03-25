package log

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/haolipeng/BeeGuard/server/internal/config"
)

// 全局 logger
var logger *zap.Logger
var sugar *zap.SugaredLogger

// Init 初始化日志
func Init(cfg *config.LogConfig) error {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return err
	}

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 控制台输出
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleWriter := zapcore.Lock(os.Stdout)
	consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, level)

	core := consoleCore

	// 文件输出（当 Dir 非空时启用）
	if cfg.Dir != "" {
		if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
			return fmt.Errorf("创建日志目录失败: %w", err)
		}

		filename := cfg.Filename
		if filename == "" {
			filename = "server.log"
		}

		fileWriter := &lumberjack.Logger{
			Filename:   filepath.Join(cfg.Dir, filename),
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
			LocalTime:  true,
			Compress:   false,
		}

		fileEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(fileWriter), level)

		core = zapcore.NewTee(consoleCore, fileCore)
	}

	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar = logger.Sugar()

	sugar.Infof("日志初始化成功, level=%s", cfg.Level)
	return nil
}

// parseLevel 解析日志级别
func parseLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("未知的日志级别: %s", level)
	}
}

// GetLogger 获取原始 logger
func GetLogger() *zap.Logger {
	return logger
}

// GetSugar 获取 SugaredLogger
func GetSugar() *zap.SugaredLogger {
	return sugar
}

// Sync 刷新日志缓冲
func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}

// Debug 输出 debug 级别日志
func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// Info 输出 info 级别日志
func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// Warn 输出 warn 级别日志
func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// Error 输出 error 级别日志
func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// Fatal 输出 fatal 级别日志并退出
func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

// Debugf 格式化输出 debug 级别日志
func Debugf(template string, args ...interface{}) {
	sugar.Debugf(template, args...)
}

// Infof 格式化输出 info 级别日志
func Infof(template string, args ...interface{}) {
	sugar.Infof(template, args...)
}

// Warnf 格式化输出 warn 级别日志
func Warnf(template string, args ...interface{}) {
	sugar.Warnf(template, args...)
}

// Errorf 格式化输出 error 级别日志
func Errorf(template string, args ...interface{}) {
	sugar.Errorf(template, args...)
}

// Fatalf 格式化输出 fatal 级别日志并退出
func Fatalf(template string, args ...interface{}) {
	sugar.Fatalf(template, args...)
}
