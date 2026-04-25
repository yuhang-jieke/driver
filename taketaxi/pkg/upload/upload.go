package upload

import (
	"context"
	"io"
)

// StorageType 存储类型
type StorageType string

const (
	StorageMinIO StorageType = "minio"
	StorageOSS   StorageType = "oss"
	StorageCOS   StorageType = "cos"
)

// BizType 业务类型
type BizType string

const (
	BizTypeAvatar  BizType = "avatar"
	BizTypeIDCard  BizType = "idcard"
	BizTypeLicense BizType = "license"
	BizTypeVehicle BizType = "vehicle"
	BizTypeFace    BizType = "face"
)

// Storage 存储接口 - 支持多存储实现
type Storage interface {
	// Upload 上传文件
	Upload(ctx context.Context, req *UploadRequest) (*UploadResult, error)
	// Delete 删除文件
	Delete(ctx context.Context, path string) error
	// GetStorageType 获取存储类型
	GetStorageType() StorageType
}

// UploadRequest 上传请求
type UploadRequest struct {
	File     io.Reader // 文件流
	FileName string    // 原始文件名
	Size     int64     // 文件大小(字节)
	BizType  BizType   // 业务类型
}

// UploadResult 上传结果
type UploadResult struct {
	URL   string // 访问URL
	Path  string // 存储路径
	Size  int64  // 文件大小
}
