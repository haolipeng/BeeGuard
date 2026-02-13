// SPDX-License-Identifier: GPL-2.0
package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger zap日志封装
type Logger struct {
	*zap.SugaredLogger
}

// New 创建新的日志记录器
func New() *Logger {
	// 配置日志编码器
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
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	// 文件输出（带轮转）
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "driver.log",
		MaxSize:    1,    // 每个日志文件最大 1 MB
		MaxBackups: 10,   // 保留 10 个旧文件
		Compress:   true, // 压缩旧文件
	})

	// 双输出：stderr + 文件
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), zapcore.InfoLevel),
		zapcore.NewCore(encoder, fileWriter, zapcore.InfoLevel),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &Logger{
		SugaredLogger: logger.Sugar(),
	}
}

// Info 记录信息日志（结构化 key-value 格式）
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Infow(msg, keysAndValues...)
}

// Warn 记录警告日志（结构化 key-value 格式）
func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Warnw(msg, keysAndValues...)
}

// Error 记录错误日志（结构化 key-value 格式）
func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Errorw(msg, keysAndValues...)
}

// Fatal 记录致命错误并退出
func (l *Logger) Fatal(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Fatalw(msg, keysAndValues...)
}
