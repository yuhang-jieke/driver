package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Registry RegistryConfig `yaml:"registry"`
	Mongo    MongoConfig    `yaml:"mongo"`
	Dispatch DispatchConfig `yaml:"dispatch"`
}

type ServerConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	GRPCHost string `yaml:"grpc_host"`
	GRPCPort int    `yaml:"grpc_port"`
}

type DatabaseConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	User string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type RedisConfig struct {
	Host     string `yaml:"Host"`
	Port     int    `yaml:"Port"`
	Password string `yaml:"Password"`
	Database int    `yaml:"Database"`
}

type RegistryConfig struct {
	Type    string `yaml:"type"`
	Address string `yaml:"address"`
}

type MongoConfig struct {
	Uri      string `yaml:"uri"`
	Database string `yaml:"database"`
}

type DispatchConfig struct {
	RadiusKm          float64 `yaml:"radius_km"`
	MinServiceScore   float64 `yaml:"min_service_score"`
	ArriveCheckRadius float64 `yaml:"arrive_check_radius"`
	EndTripCheckRadius float64 `yaml:"end_trip_check_radius"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
