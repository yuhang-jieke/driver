package main

import (
	"driver/taketaxi/pkg/config"
	"driver/taketaxi/pkg/database"
	"driver/taketaxi/pkg/mongodb"
	"driver/taketaxi/pkg/redis"
	"driver/taketaxi/srvDriver/internal/handler"
	"driver/taketaxi/srvDriver/internal/repository"

	driver "driver/taketaxi/common/kitexGen"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var confPath string

func init() {
	flag.StringVar(&confPath, "config", "../configs/config.yaml", "config file")
}

func main() {
	flag.Parse()
	cfg, err := config.Load(confPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("init db: %v", err)
	}
	rdb := redis.NewRedisClient(&cfg.Redis)
	_ = rdb

	mongoDb, closeMongo := mongodb.NewMongoDB(cfg.Mongo.Uri, cfg.Mongo.Database)
	if closeMongo != nil {
		defer closeMongo()
	}

	repo := repository.NewDriverRepo(db)
	h := handler.NewDriverHandler(mongoDb, repo, &cfg.Dispatch)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	lis, _ := net.Listen("tcp", addr)
	s := grpc.NewServer()
	driver.RegisterDriverServiceServer(s, h)
	reflection.Register(s)
	log.Println("Starting on", addr)
	s.Serve(lis)
}
