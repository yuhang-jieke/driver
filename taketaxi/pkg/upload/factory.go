package upload

import (
	"driver/taketaxi/pkg/config"
	"fmt"
)

// NewStorage 根据配置创建存储实例
func NewStorage(cfg *config.UploadConfig) (Storage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("上传配置不能为空")
	}

	switch cfg.StorageType {
	case string(StorageMinIO):
		return NewMinIOStorage(&cfg.MinIO)
	case string(StorageOSS):
		return NewOSSStorage(&cfg.OSS)
	case string(StorageCOS):
		return NewCOSStorage(&cfg.COS)
	default:
		return nil, fmt.Errorf("不支持的存储类型: %s", cfg.StorageType)
	}
}
