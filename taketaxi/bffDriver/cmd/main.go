package main

import (
	"driver/taketaxi/bffDriver/internal/router"
	"driver/taketaxi/bffDriver/internal/rpcClient"
	"driver/taketaxi/pkg/config"
	"driver/taketaxi/pkg/upload"
	"flag"
	"fmt"
	"log"
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
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建gRPC客户端
	grpcAddr := fmt.Sprintf("%s:%d", cfg.Server.GRPCHost, cfg.Server.GRPCPort)
	client, err := rpcclient.NewDriverClient(grpcAddr)
	if err != nil {
		log.Fatalf("创建gRPC客户端失败: %v", err)
	}
	defer client.Close()

	// 创建存储实例
	storage, err := upload.NewStorage(&cfg.Upload)
	if err != nil {
		log.Fatalf("创建存储实例失败: %v", err)
	}

	// 启动HTTP服务
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("BFF starting on %s", addr)
	router.NewRouter(client, storage).Run(addr)
}
