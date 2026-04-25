package upload

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
