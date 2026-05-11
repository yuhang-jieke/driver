package logger

import (
	"driver/taketaxi/pkg/config"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected zapcore.Level
	}{
		{"debug", zapcore.DebugLevel},
		{"info", zapcore.InfoLevel},
		{"warn", zapcore.WarnLevel},
		{"error", zapcore.ErrorLevel},
		{"unknown", zapcore.InfoLevel},
		{"", zapcore.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestInitWithFile(t *testing.T) {
	// 使用 os.TempDir() 创建临时日志文件路径
	// 注意：不使用 t.TempDir() 因为 Windows 上 lumberjack 会持有文件句柄导致清理失败
	logFile := filepath.Join(os.TempDir(), "logger_test.log")

	// 确保测试前删除旧文件
	os.Remove(logFile)

	cfg := &config.LogConfig{
		Level:      "info",
		Filename:   logFile,
		MaxSize:    10,
		MaxBackups: 1,
		MaxAge:     1,
		Compress:   false,
	}

	cleanup, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 测试日志输出
	Info("test message", zap.String("key", "value"))

	// 检查日志文件是否创建
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("log file not created: %s", logFile)
	}

	// 清理
	cleanup()

	// 尝试删除测试文件（忽略错误）
	os.Remove(logFile)
}

func TestInitWithoutFile(t *testing.T) {
	cfg := &config.LogConfig{
		Level:      "debug",
		Filename:   "", // 不写文件
		MaxSize:    10,
		MaxBackups: 1,
		MaxAge:     1,
		Compress:   false,
	}

	cleanup, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer cleanup()

	// 测试日志输出（仅控制台）
	Debug("debug message")
	Info("info message")
}

func TestLoggerMethods(t *testing.T) {
	// 使用 zaptest 创建测试 logger
	L = zaptest.NewLogger(t)
	S = L.Sugar()

	// 测试各日志方法
	Debug("debug message", zap.String("test", "debug"))
	Info("info message", zap.String("test", "info"))
	Warn("warn message", zap.String("test", "warn"))
	Error("error message", zap.String("test", "error"))

	// 测试 With
	childLogger := With(zap.String("component", "test"))
	if childLogger == nil {
		t.Error("With returned nil logger")
	}
}
