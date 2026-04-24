package main

import (
	"driver/taketaxi/pkg/config"
	"driver/taketaxi/pkg/redis"
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
	cfg, _ := config.Load(confPath)
	//db, _ := database.NewDB(&cfg.Database)
	rdb := redis.NewRedisClient(&cfg.Redis)
	_ = rdb
	/*db.AutoMigrate(
		&model.Passenger{},
		&model.DriverVehicleInfo{},
		&model.DriverFace{},
		&model.DriverLevelConfig{},
		&model.DriverLevelRecord{},
		&model.DriverLocationCache{},
		&model.OrderEvaluation{},
		&model.TripService{},
		&model.TripTrajectory{},
		&model.DriverWallet{},
		&model.DriverIncomeLog{},
		&model.WalletTransactionLog{},
		&model.DriverWithdrawRecord{},
		&model.WithdrawRecord{},
		&model.ServiceScoreLog{},
		&model.DriverStatisticsSummary{},
		&model.PricingRuleConfig{},
		&model.DriverS{},
		&model.DriverRealname{},
		&model.DriverLicense{},
		&model.DriverVehicle{},
		&model.DriverStatusLog{},
		&model.DriverOnlineLog{},
		&model.Order{},
		&model.DispatchLog{},
		&model.DriverFaceAuthLog{},
	)*/
	//repo := repository.NewDriverRepo(db)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	lis, _ := net.Listen("tcp", addr)
	s := grpc.NewServer()
	//driver.RegisterDriverServiceServer(s, handler.NewDriverHandler(repo))
	reflection.Register(s)
	log.Printf("Starting on %s", addr)
	s.Serve(lis)
}
