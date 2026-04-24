package main

import (
	"driver/taketaxi/bffDriver/internal/router"
	"driver/taketaxi/bffDriver/internal/rpcClient"
	"driver/taketaxi/pkg/config"
	"flag"
	"fmt"
	"log"
)

var confPath string

func init() {
	flag.StringVar(&confPath, "config", "C:\\Users\\35305\\Desktop\\driver\\driver\\taketaxi\\bffDriver\\configs\\config.yaml", "config file")
}

func main() {
	flag.Parse()
	cfg, err := config.Load(confPath)
	if err != nil {
		log.Fatal(err)
	}
	grpcAddr := fmt.Sprintf("%s:%d", cfg.Server.GRPCHost, cfg.Server.GRPCPort)
	client, err := rpcclient.NewDriverClient(grpcAddr)
	if err != nil {
		log.Fatalf("failed to create gRPC client: %v", err)
	}
	defer client.Close()
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("BFF starting on %s", addr)
	router.NewRouter(client).Run(addr)
}
