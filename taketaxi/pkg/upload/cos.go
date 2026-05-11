package upload

import (
	"context"
	"driver/taketaxi/pkg/config"
	"fmt"
)

// COSStorage 腾讯云COS存储实现（预留）
type COSStorage struct {
	bucketURL string
	validator *Validator
}

// NewCOSStorage 创建COS存储实例
func NewCOSStorage(cfg *config.COSConf) (*COSStorage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("cos配置不能为空")
	}
	// TODO: 实现COS客户端初始化
	return &COSStorage{
		bucketURL: cfg.BucketURL,
		validator: NewValidator(),
	}, nil
}

// Upload 上传文件
func (s *COSStorage) Upload(ctx context.Context, req *UploadRequest) (*UploadResult, error) {
	return nil, fmt.Errorf("COS存储暂未实现")
}

// Delete 删除文件
func (s *COSStorage) Delete(ctx context.Context, path string) error {
	return fmt.Errorf("COS存储暂未实现")
}

// GetStorageType 获取存储类型
func (s *COSStorage) GetStorageType() StorageType {
	return StorageCOS
}
