package main

import (
	"driver/taketaxi/pkg/config"
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

	logger.Info("srvDriver starting",
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port),
	)

	rdb := redis.NewRedisClient(&cfg.Redis)
	_ = rdb
	logger.Info("redis client initialized")

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("监听失败", zap.Error(err), zap.String("address", addr))
	}

	s := grpc.NewServer()
	driver.RegisterDriverServiceServer(s, handler.NewDriverHandler(repository.NewDriverRepo(nil)))
	reflection.Register(s)

	logger.Info("gRPC server starting", zap.String("address", addr))
	if err := s.Serve(lis); err != nil {
		logger.Fatal("gRPC server failed", zap.Error(err))
	}
}
