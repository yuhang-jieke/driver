package upload

import (
	"bytes"
	"context"
	"driver/taketaxi/pkg/config"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage MinIO存储实现
type MinIOStorage struct {
	client    *minio.Client
	bucket    string
	validator *Validator
}

// NewMinIOStorage 创建MinIO存储实例
func NewMinIOStorage(cfg *config.MinIOConf) (*MinIOStorage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("minio配置不能为空")
	}

	// 初始化MinIO客户端
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("初始化MinIO客户端失败: %w", err)
	}

	// 确保bucket存在
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("检查bucket失败: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{
			Region: cfg.Region,
		}); err != nil {
			return nil, fmt.Errorf("创建bucket失败: %w", err)
		}
	}

	return &MinIOStorage{
		client:    client,
		bucket:    cfg.BucketName,
		validator: NewValidator(),
	}, nil
}

// Upload 上传文件
func (s *MinIOStorage) Upload(ctx context.Context, req *UploadRequest) (*UploadResult, error) {
	if req == nil {
		return nil, fmt.Errorf("上传请求不能为空")
	}
	if req.File == nil {
		return nil, fmt.Errorf("文件不能为空")
	}

	// 读取前512字节用于校验
	header := make([]byte, 512)
	n, err := req.File.Read(header)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("读取文件头失败: %w", err)
	}
	header = header[:n]

	// 校验文件
	if err := s.validator.ValidateReader(header, req.FileName, req.Size); err != nil {
		return nil, err
	}

	// 生成对象路径: {bizType}/{YYYYMMDD}/{uuid}.{ext}
	ext := filepath.Ext(req.FileName)
	datePath := time.Now().Format("20060102")
	fileUUID := uuid.New().String()
	objectName := fmt.Sprintf("%s/%s/%s%s", req.BizType, datePath, fileUUID, ext)

	// 组合原始数据和剩余数据
	reader := io.MultiReader(bytes.NewReader(header), req.File)

	// 上传到MinIO
	_, err = s.client.PutObject(ctx, s.bucket, objectName, reader, req.Size, minio.PutObjectOptions{
		ContentType: s.getContentType(ext),
	})
	if err != nil {
		return nil, fmt.Errorf("上传文件到MinIO失败: %w", err)
	}

	// 返回结果
	return &UploadResult{
		URL:  fmt.Sprintf("/%s/%s", s.bucket, objectName),
		Path: objectName,
		Size: req.Size,
	}, nil
}

// Delete 删除文件
func (s *MinIOStorage) Delete(ctx context.Context, path string) error {
	if path == "" {
		return fmt.Errorf("文件路径不能为空")
	}
	return s.client.RemoveObject(ctx, s.bucket, path, minio.RemoveObjectOptions{})
}

// GetStorageType 获取存储类型
func (s *MinIOStorage) GetStorageType() StorageType {
	return StorageMinIO
}

// GetPresignedURL 获取预签名URL（用于临时访问）
func (s *MinIOStorage) GetPresignedURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucket, path, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("获取预签名URL失败: %w", err)
	}
	return url.String(), nil
}

// StatObject 获取对象信息
func (s *MinIOStorage) StatObject(ctx context.Context, path string) (minio.ObjectInfo, error) {
	return s.client.StatObject(ctx, s.bucket, path, minio.StatObjectOptions{})
}

func (s *MinIOStorage) getContentType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	default:
		return "application/octet-stream"
	}
}
