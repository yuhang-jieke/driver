# 日志系统设计文档

## 概述

为 driver 项目添加结构化日志系统，基于 zap 日志库，支持控制台和文件双输出，采用 JSON 格式。

## 技术选型

| 项目 | 选择 | 理由 |
|------|------|------|
| 日志库 | zap | 高性能结构化日志，Uber开源，生产环境首选 |
| 输出目标 | 控制台 + 文件 | 便于开发调试和生产排查 |
| 日志格式 | JSON | 结构化日志，便于日志系统解析和查询 |

## 架构设计

### 目录结构

```
pkg/logger/
├── logger.go          # 核心日志初始化和配置
└── logger_test.go     # 单元测试
```

### 配置结构

在 `pkg/config/config.go` 中新增日志配置：

```go
type LogConfig struct {
    Level      string `yaml:"level"`       // 日志级别: debug/info/warn/error
    Filename   string `yaml:"filename"`    // 日志文件路径
    MaxSize    int    `yaml:"max_size"`    // 单个日志文件最大大小(MB)
    MaxBackups int    `yaml:"max_backups"` // 保留的旧日志文件数量
    MaxAge     int    `yaml:"max_age"`     // 保留旧日志文件的最大天数
    Compress   bool   `yaml:"compress"`    // 是否压缩旧日志文件
}
```

配置文件示例 (`config.yaml`)：

```yaml
log:
  level: info
  filename: logs/app.log
  max_size: 100
  max_backups: 3
  max_age: 7
  compress: true
```

### Logger 模块设计

**初始化函数**：

```go
// Init 初始化日志器
// 返回清理函数，用于优雅关闭时刷新缓冲区
func Init(cfg *LogConfig) (func(), error)
```

**日志级别**：
- `debug`: 调试信息
- `info`: 一般信息（默认）
- `warn`: 警告信息
- `error`: 错误信息

**输出格式**：

```json
{
  "level": "info",
  "time": "2026-04-25T10:00:00+08:00",
  "caller": "handler/uploadHandler.go:45",
  "msg": "upload success",
  "method": "POST",
  "path": "/api/v1/upload",
  "biz_type": "avatar",
  "file_size": 1024,
  "duration": "15ms"
}
```

## 日志记录点

### bffDriver 层

| 文件 | 记录点 | 级别 | 字段 |
|------|--------|------|------|
| `cmd/main.go` | 服务启动、配置加载、依赖初始化 | info | address, component |
| `handler/uploadHandler.go` | 上传成功/失败 | info/error | biz_type, file_size, duration, error |
| `handler/driverHandler.go` | CRUD 操作成功/失败 | info/error | operation, id, name, error |

### srvDriver 层

| 文件 | 记录点 | 级别 | 字段 |
|------|--------|------|------|
| `cmd/main.go` | 服务启动、gRPC监听 | info | address |
| `handler/driverHandler.go` | 关键业务操作 | info/error | operation, error |
| `service/driverService.go` | 业务逻辑执行 | info/error | operation, error |
| `repository/driverRepo.go` | 数据库操作错误 | error | operation, error |

## 错误处理

- 日志初始化失败：返回错误，主程序退出
- 日志文件创建失败：返回错误，主程序退出
- 日志写入失败：zap 内部处理，不影响主程序

## 依赖

```go
require (
    go.uber.org/zap v1.27.0
    gopkg.in/natefinch/lumberjack.v2 v2.2.1
)
```

- `zap`: 核心日志库
- `lumberjack`: 日志文件轮转

## 实施步骤

1. 添加 zap 和 lumberjack 依赖
2. 在 `pkg/config/config.go` 添加日志配置结构
3. 重构 `pkg/logger/logger.go`，实现 zap 初始化
4. 更新配置文件 `config.yaml`
5. 在 `bffDriver/cmd/main.go` 集成日志
6. 在 `srvDriver/cmd/main.go` 集成日志
7. 在 `bffDriver` handlers 中添加日志记录
8. 在 `srvDriver` handlers/services 中添加日志记录
9. 编写单元测试
