package logger

import (
	"driver/taketaxi/pkg/config"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// L 全局 Logger
	L *zap.Logger
	// S 全局 SugaredLogger（支持 printf 风格）
	S *zap.SugaredLogger
)

// Init 初始化日志器
// 返回清理函数，用于优雅关闭时刷新缓冲区
func Init(cfg *config.LogConfig) (func(), error) {
	// 解析日志级别
	level := parseLevel(cfg.Level)

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 控制台输出
	consoleEncoder := zapcore.NewJSONEncoder(encoderConfig)
	consoleWriter := zapcore.AddSync(os.Stdout)

	// 文件输出
	var fileWriter zapcore.WriteSyncer
	if cfg.Filename != "" {
		// 确保日志目录存在
		if err := os.MkdirAll(filepath.Dir(cfg.Filename), 0755); err != nil {
			return nil, err
		}

		lj := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,    // MB
			MaxBackups: cfg.MaxBackups, // 保留旧文件数量
			MaxAge:     cfg.MaxAge,     // 保留天数
			Compress:   cfg.Compress,
		}
		fileWriter = zapcore.AddSync(lj)
	}

	// 多路输出: 控制台 + 文件
	var core zapcore.Core
	if cfg.Filename != "" {
		core = zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, consoleWriter, level),
			zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), fileWriter, level),
		)
	} else {
		core = zapcore.NewCore(consoleEncoder, consoleWriter, level)
	}

	// 创建 Logger
	L = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	S = L.Sugar()

	// 返回清理函数
	return func() {
		_ = L.Sync()
	}, nil
}

// parseLevel 解析日志级别字符串
func parseLevel(levelStr string) zapcore.Level {
	switch levelStr {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// Debug 记录调试日志
func Debug(msg string, fields ...zap.Field) {
	L.Debug(msg, fields...)
}

// Info 记录信息日志
func Info(msg string, fields ...zap.Field) {
	L.Info(msg, fields...)
}

// Warn 记录警告日志
func Warn(msg string, fields ...zap.Field) {
	L.Warn(msg, fields...)
}

// Error 记录错误日志
func Error(msg string, fields ...zap.Field) {
	L.Error(msg, fields...)
}

// Fatal 记录致命错误日志并退出
func Fatal(msg string, fields ...zap.Field) {
	L.Fatal(msg, fields...)
}

// With 创建带有预设字段的子 Logger
func With(fields ...zap.Field) *zap.Logger {
	return L.With(fields...)
}
