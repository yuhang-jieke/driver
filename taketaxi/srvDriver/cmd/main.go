package main

import (
	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/pkg/config"
	"driver/taketaxi/pkg/database"
	"driver/taketaxi/srvDriver/internal/handler"
	"driver/taketaxi/srvDriver/internal/repository"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var confPath string

func init() {
	flag.StringVar(&confPath, "config", "C:\\Users\\35305\\Desktop\\driver\\driver\\taketaxi\\srvDriver\\configs\\config.yaml", "config file")
}

func main() {
	flag.Parse()
	cfg, _ := config.Load(confPath)
	db, _ := database.NewDB(&cfg.Database)
	repo := repository.NewDriverRepo(db)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	lis, _ := net.Listen("tcp", addr)
	s := grpc.NewServer()
	driver.RegisterDriverServiceServer(s, handler.NewDriverHandler(repo))
	reflection.Register(s)
	log.Printf("Starting on %s", addr)
	s.Serve(lis)
}
