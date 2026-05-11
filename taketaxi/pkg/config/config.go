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
	Upload   UploadConfig   `yaml:"upload"`
	Log      LogConfig      `yaml:"log"`
}

type ServerConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	GRPCHost string `yaml:"grpc_host"`
	GRPCPort int    `yaml:"grpc_port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
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

// UploadConfig 上传配置
type UploadConfig struct {
	StorageType string    `yaml:"storage_type" json:"storage_type"`
	MinIO       MinIOConf `yaml:"minio" json:"minio"`
	OSS         OSSConf   `yaml:"oss" json:"oss"`
	COS         COSConf   `yaml:"cos" json:"cos"`
}

// MinIOConf MinIO配置
type MinIOConf struct {
	Endpoint   string `yaml:"endpoint" json:"endpoint"`
	AccessKey  string `yaml:"access_key" json:"access_key"`
	SecretKey  string `yaml:"secret_key" json:"secret_key"`
	BucketName string `yaml:"bucket_name" json:"bucket_name"`
	UseSSL     bool   `yaml:"use_ssl" json:"use_ssl"`
	Region     string `yaml:"region" json:"region"`
}

// OSSConf 阿里云OSS配置
type OSSConf struct {
	Endpoint        string `yaml:"endpoint" json:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id" json:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret" json:"access_key_secret"`
	BucketName      string `yaml:"bucket_name" json:"bucket_name"`
}

// COSConf 腾讯云COS配置
type COSConf struct {
	BucketURL string `yaml:"bucket_url" json:"bucket_url"`
	SecretID  string `yaml:"secret_id" json:"secret_id"`
	SecretKey string `yaml:"secret_key" json:"secret_key"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level"`       // 日志级别: debug/info/warn/error
	Filename   string `yaml:"filename"`    // 日志文件路径
	MaxSize    int    `yaml:"max_size"`    // 单个日志文件最大大小(MB)
	MaxBackups int    `yaml:"max_backups"` // 保留的旧日志文件数量
	MaxAge     int    `yaml:"max_age"`     // 保留旧日志文件的最大天数
	Compress   bool   `yaml:"compress"`    // 是否压缩旧日志文件
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
