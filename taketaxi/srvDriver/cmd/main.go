// srvDriver 服务端入口程序
//
// 启动流程（按顺序执行）：
//  1. 解析命令行参数（配置文件路径 -config）
//  2. 加载 YAML 配置文件
//  3. 初始化结构化日志（zap + rotate）
//  4. 初始化 Redis 客户端连接
//  5. 初始化 MySQL 数据库连接（GORM）
//  6. 创建 Repository → Service → Handler 依赖链
//  7. 注册 gRPC Server 并启动监听
//
// 架构层次：
//
//	main (入口)
//	  ├── config     ← YAML 配置加载
//	  ├── logger     ← zap 日志初始化
//	  ├── redis      ← Redis 连接
//	  ├── database   ← MySQL/GORM 连接
//	  └── handler    ← gRPC Handler（依赖 service → repository）
//	        └── service    ← 业务逻辑层（依赖 repo + rdb）
//	              └── repository ← 数据库操作层（依赖 *gorm.DB）
//
// 使用方式：
//
//	go run taketaxi/srvDriver/cmd/main.go -config=taketaxi/srvDriver/configs/config.yaml
package main

import (
	"driver/taketaxi/pkg/config"
	"driver/taketaxi/pkg/database"
	"driver/taketaxi/pkg/logger"
	"driver/taketaxi/pkg/redis"
	"flag"
	"fmt"
	"net"

	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/srvDriver/internal/handler"
	"driver/taketaxi/srvDriver/internal/repository"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var confPath string

func init() {
	// 注册命令行参数：-config 指定配置文件路径
	// 默认值 "../configs/config.yaml" 相对于 cmd 目录的上级
	flag.StringVar(&confPath, "config", "../configs/config.yaml", "config file")
}

func main() {
	flag.Parse()

	// ========== 第1步：加载配置 ==========
	cfg, err := config.Load(confPath)
	if err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}

	// ========== 第2步：初始化日志 ==========
	cleanup, err := logger.Init(&cfg.Log)
	if err != nil {
		panic(fmt.Sprintf("初始化日志失败: %v", err))
	}
	defer cleanup() // 程序退出时刷新日志缓冲区

	logger.Info("srvDriver starting",
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port),
	)

	// ========== 第3步：初始化 Redis ==========
	rdb := redis.NewRedisClient(&cfg.Redis)
	_ = rdb
	logger.Info("redis client initialized")

	// ========== 第4步：初始化数据库 ==========
	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		logger.Fatal("数据库初始化失败", zap.Error(err))
	}
	logger.Info("database initialized")

	// ========== 第5步：创建 gRPC 监听器 ==========
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("监听失败", zap.Error(err), zap.String("address", addr))
	}

	// ========== 第6步：组装依赖链 & 注册服务 ==========
	// 依赖注入链：db → Repo → Service(rdb) → Handler → gRPC Server
	s := grpc.NewServer()
	driver.RegisterDriverServiceServer(
		s,
		handler.NewDriverHandler( // 创建 gRPC Handler
			repository.NewDriverRepo(db), // 注入 Repository（含 DB）
			rdb,                          // 注入 Redis 客户端
		),
	)
	// 启用 gRPC 反射（用于 grpcurl 等调试工具调用）
	reflection.Register(s)

	// ========== 第7步：启动 gRPC 服务（阻塞） ==========
	logger.Info("gRPC server starting", zap.String("address", addr))
	if err := s.Serve(lis); err != nil {
		logger.Fatal("gRPC server failed", zap.Error(err))
	}
}
