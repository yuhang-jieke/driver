package upload

import (
	"context"
	"fmt"
)

// OSSStorage 阿里云OSS存储实现（预留）
type OSSStorage struct {
	bucket    string
	validator *Validator
}

// NewOSSStorage 创建OSS存储实例
func NewOSSStorage(cfg *OSSConf) (*OSSStorage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("oss配置不能为空")
	}
	// TODO: 实现OSS客户端初始化
	return &OSSStorage{
		bucket:    cfg.BucketName,
		validator: NewValidator(),
	}, nil
}

// Upload 上传文件
func (s *OSSStorage) Upload(ctx context.Context, req *UploadRequest) (*UploadResult, error) {
	return nil, fmt.Errorf("OSS存储暂未实现")
}

// Delete 删除文件
func (s *OSSStorage) Delete(ctx context.Context, path string) error {
	return fmt.Errorf("OSS存储暂未实现")
}

// GetStorageType 获取存储类型
func (s *OSSStorage) GetStorageType() StorageType {
	return StorageOSS
}
