package main

import (
	"driver/taketaxi/bffDriver/internal/router"
	"driver/taketaxi/bffDriver/internal/rpcClient"
	"driver/taketaxi/pkg/config"
	"driver/taketaxi/pkg/logger"
	"driver/taketaxi/pkg/upload"
	"flag"
	"fmt"

	"go.uber.org/zap"
)

var confPath string

func init() {
	flag.StringVar(&confPath, "config", "../configs/config.yaml", "config file")
}

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(confPath)
	if err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}

	// 初始化日志
	cleanup, err := logger.Init(&cfg.Log)
	if err != nil {
		panic(fmt.Sprintf("初始化日志失败: %v", err))
	}
	defer cleanup()

	logger.Info("bffDriver starting",
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port),
	)

	// 创建gRPC客户端
	grpcAddr := fmt.Sprintf("%s:%d", cfg.Server.GRPCHost, cfg.Server.GRPCPort)
	client, err := rpcclient.NewDriverClient(grpcAddr)
	if err != nil {
		logger.Fatal("创建gRPC客户端失败", zap.Error(err))
	}
	defer client.Close()
	logger.Info("gRPC client created", zap.String("address", grpcAddr))

	// 创建存储实例
	storage, err := upload.NewStorage(&cfg.Upload)
	if err != nil {
		logger.Fatal("创建存储实例失败", zap.Error(err))
	}
	logger.Info("storage initialized", zap.String("type", cfg.Upload.StorageType))

	// 启动HTTP服务
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("HTTP server starting", zap.String("address", addr))
	router.NewRouter(client, storage).Run(addr)
}
