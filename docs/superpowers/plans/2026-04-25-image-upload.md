1# 图片上传功能实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现企业级图片上传功能，支持MinIO存储，预留OSS/COS扩展接口

**Architecture:** 采用接口抽象模式，定义 `Storage` 接口，通过工厂模式创建存储实例。校验器独立封装，Handler层处理HTTP请求。配置统一管理。

**Tech Stack:** Go 1.26, Gin, MinIO (minio-go/v7), Google UUID

---

## 文件结构

| 文件 | 职责 |
|------|------|
| `taketaxi/pkg/upload/upload.go` | 核心接口定义：Storage接口、类型常量、请求/响应结构体 |
| `taketaxi/pkg/upload/config.go` | 上传配置结构体：MinIO/OSS/COS配置 |
| `taketaxi/pkg/upload/validator.go` | 文件校验器：大小、扩展名、魔数验证 |
| `taketaxi/pkg/upload/errors.go` | 错误定义：统一错误码和消息 |
| `taketaxi/pkg/upload/minio.go` | MinIO存储实现 |
| `taketaxi/pkg/upload/factory.go` | 存储工厂：根据配置创建存储实例 |
| `taketaxi/pkg/upload/oss.go` | OSS存储实现（预留骨架） |
| `taketaxi/pkg/upload/cos.go` | COS存储实现（预留骨架） |
| `taketaxi/pkg/upload/upload_test.go` | 单元测试 |
| `taketaxi/bffDriver/internal/handler/uploadHandler.go` | HTTP Handler |
| `taketaxi/pkg/config/config.go` | 修改：添加UploadConfig |
| `taketaxi/configs/config.yaml` | 修改：添加upload配置节 |
| `go.mod` | 修改：添加minio-go、uuid依赖 |

---

## Task 1: 添加依赖

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: 添加 minio-go 和 uuid 依赖**

Run:
```bash
cd D:/software/GoWork/src/driver
go get github.com/minio/minio-go/v7
go get github.com/google/uuid
go mod tidy
```

Expected: 依赖添加成功，go.mod 更新

- [ ] **Step 2: 验证依赖**

Run:
```bash
cd D:/software/GoWork/src/driver
go mod download
```

Expected: 无错误输出

---

## Task 2: 核心接口定义

**Files:**
- Create: `taketaxi/pkg/upload/upload.go`

- [ ] **Step 1: 创建 upload 包目录**

Run:
```bash
mkdir -p D:/software/GoWork/src/driver/taketaxi/pkg/upload
```

Expected: 目录创建成功

- [ ] **Step 2: 编写核心接口文件**

Create `taketaxi/pkg/upload/upload.go`:

```go
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
```

- [ ] **Step 3: 验证编译**

Run:
```bash
cd D:/software/GoWork/src/driver
go build ./taketaxi/pkg/upload/...
```

Expected: 无错误输出

- [ ] **Step 4: 提交**

```bash
cd D:/software/GoWork/src/driver
git add taketaxi/pkg/upload/upload.go
git commit -m "feat(upload): add core interface definitions

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Task 3: 错误定义

**Files:**
- Create: `taketaxi/pkg/upload/errors.go`

- [ ] **Step 1: 编写错误定义文件**

Create `taketaxi/pkg/upload/errors.go`:

```go
package upload

import "errors"

var (
	// ErrFileSizeExceed 文件大小超过限制
	ErrFileSizeExceed = errors.New("文件大小超过限制，最大允许2MB")
	// ErrInvalidExt 不支持的文件扩展名
	ErrInvalidExt = errors.New("不支持的文件格式，仅支持jpg/png")
	// ErrMimeTypeMismatch 文件内容类型不匹配
	ErrMimeTypeMismatch = errors.New("文件内容类型不匹配")
	// ErrEmptyFile 空文件
	ErrEmptyFile = errors.New("文件内容为空")
	// ErrUnsupportedStorage 不支持的存储类型
	ErrUnsupportedStorage = errors.New("不支持的存储类型")
)
```

- [ ] **Step 2: 验证编译**

Run:
```bash
cd D:/software/GoWork/src/driver
go build ./taketaxi/pkg/upload/...
```

Expected: 无错误输出

- [ ] **Step 3: 提交**

```bash
cd D:/software/GoWork/src/driver
git add taketaxi/pkg/upload/errors.go
git commit -m "feat(upload): add error definitions

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Task 4: 配置定义

**Files:**
- Create: `taketaxi/pkg/upload/config.go`

- [ ] **Step 1: 编写配置文件**

Create `taketaxi/pkg/upload/config.go`:

```go
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
```

- [ ] **Step 2: 验证编译**

Run:
```bash
cd D:/software/GoWork/src/driver
go build ./taketaxi/pkg/upload/...
```

Expected: 无错误输出

- [ ] **Step 3: 提交**

```bash
cd D:/software/GoWork/src/driver
git add taketaxi/pkg/upload/config.go
git commit -m "feat(upload): add config definitions

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Task 5: 校验器实现与测试

**Files:**
- Create: `taketaxi/pkg/upload/validator.go`
- Create: `taketaxi/pkg/upload/validator_test.go`

- [ ] **Step 1: 编写校验器实现**

Create `taketaxi/pkg/upload/validator.go`:

```go
package upload

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

const (
	// MaxFileSize 最大文件大小 2MB
	MaxFileSize int64 = 2 * 1024 * 1024
	// AllowedExts 允许的扩展名
	AllowedExts = ".jpg,.png"
	// AllowedMimeTypes 允许的MIME类型
	AllowedMimeTypes = "image/jpeg,image/png"
)

// Validator 文件校验器
type Validator struct {
	maxSize      int64
	allowedExts  []string
	allowedMimes []string
}

// NewValidator 创建校验器
func NewValidator() *Validator {
	return &Validator{
		maxSize:      MaxFileSize,
		allowedExts:  strings.Split(AllowedExts, ","),
		allowedMimes: strings.Split(AllowedMimeTypes, ","),
	}
}

// Validate 校验文件
func (v *Validator) Validate(file io.ReadSeeker, filename string, size int64) error {
	// 文件大小校验
	if size > v.maxSize {
		return ErrFileSizeExceed
	}

	// 扩展名校验
	ext := strings.ToLower(filepath.Ext(filename))
	if !v.isAllowedExt(ext) {
		return ErrInvalidExt
	}

	// 真实文件类型校验（魔数验证）
	if err := v.validateMimeType(file); err != nil {
		return err
	}

	// 重置文件读取位置
	file.Seek(0, io.SeekStart)
	return nil
}

func (v *Validator) isAllowedExt(ext string) bool {
	for _, allowed := range v.allowedExts {
		if ext == allowed {
			return true
		}
	}
	return false
}

func (v *Validator) validateMimeType(file io.ReadSeeker) error {
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return err
	}
	if n == 0 {
		return ErrEmptyFile
	}

	// 检测实际MIME类型
	contentType := http.DetectContentType(buffer[:n])
	for _, mime := range v.allowedMimes {
		if strings.HasPrefix(contentType, mime) {
			// 重置读取位置供后续使用
			file.Seek(0, io.SeekStart)
			return nil
		}
	}

	return ErrMimeTypeMismatch
}

// ValidateReader 校验 io.Reader（无法 Seek，需要调用者提供缓冲）
func (v *Validator) ValidateReader(buffer []byte, filename string, size int64) error {
	// 文件大小校验
	if size > v.maxSize {
		return ErrFileSizeExceed
	}

	// 扩展名校验
	ext := strings.ToLower(filepath.Ext(filename))
	if !v.isAllowedExt(ext) {
		return ErrInvalidExt
	}

	// 真实文件类型校验
	if len(buffer) == 0 {
		return ErrEmptyFile
	}

	contentType := http.DetectContentType(buffer)
	for _, mime := range v.allowedMimes {
		if strings.HasPrefix(contentType, mime) {
			return nil
		}
	}

	return ErrMimeTypeMismatch
}

// IsImageFile 判断buffer是否为有效图片
func IsImageFile(buffer []byte) bool {
	if len(buffer) < 512 {
		return false
	}
	contentType := http.DetectContentType(buffer[:512])
	return strings.HasPrefix(contentType, "image/jpeg") ||
		strings.HasPrefix(contentType, "image/png")
}

// DetectImageType 检测图片类型返回扩展名
func DetectImageType(buffer []byte) string {
	if len(buffer) < 512 {
		return ""
	}
	contentType := http.DetectContentType(buffer[:512])
	switch {
	case strings.HasPrefix(contentType, "image/jpeg"):
		return ".jpg"
	case strings.HasPrefix(contentType, "image/png"):
		return ".png"
	default:
		return ""
	}
}

// HasValidImageMagic 检查文件头魔数
func HasValidImageMagic(data []byte) bool {
	if len(data) < 8 {
		return false
	}
	// JPEG: FF D8 FF
	if bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}) {
		return true
	}
	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if bytes.Equal(data[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return true
	}
	return false
}
```

- [ ] **Step 2: 编写校验器测试**

Create `taketaxi/pkg/upload/validator_test.go`:

```go
package upload

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"
)

func TestValidator_Validate(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name      string
		filename  string
		size      int64
		content   []byte
		wantError error
	}{
		{
			name:      "有效JPG图片",
			filename:  "test.jpg",
			size:      1024,
			content:   createTestJPEG(t),
			wantError: nil,
		},
		{
			name:      "有效PNG图片",
			filename:  "test.png",
			size:      1024,
			content:   createTestPNG(t),
			wantError: nil,
		},
		{
			name:      "文件过大",
			filename:  "large.jpg",
			size:      3 * 1024 * 1024, // 3MB
			content:   createTestJPEG(t),
			wantError: ErrFileSizeExceed,
		},
		{
			name:      "不支持格式gif",
			filename:  "test.gif",
			size:      1024,
			content:   createTestJPEG(t),
			wantError: ErrInvalidExt,
		},
		{
			name:      "不支持格式bmp",
			filename:  "test.bmp",
			size:      1024,
			content:   createTestJPEG(t),
			wantError: ErrInvalidExt,
		},
		{
			name:      "扩展名大写JPG",
			filename:  "test.JPG",
			size:      1024,
			content:   createTestJPEG(t),
			wantError: nil,
		},
		{
			name:      "扩展名大写PNG",
			filename:  "test.PNG",
			size:      1024,
			content:   createTestPNG(t),
			wantError: nil,
		},
		{
			name:      "内容与扩展名不匹配",
			filename:  "fake.jpg",
			size:      1024,
			content:   []byte("this is not an image"),
			wantError: ErrMimeTypeMismatch,
		},
		{
			name:      "空文件",
			filename:  "empty.jpg",
			size:      0,
			content:   []byte{},
			wantError: ErrEmptyFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.content)
			err := validator.Validate(reader, tt.filename, tt.size)

			if tt.wantError != nil {
				if err != tt.wantError {
					t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidator_ValidateReader(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name      string
		filename  string
		size      int64
		buffer    []byte
		wantError error
	}{
		{
			name:      "有效JPG",
			filename:  "test.jpg",
			size:      1024,
			buffer:    createTestJPEG(t),
			wantError: nil,
		},
		{
			name:      "有效PNG",
			filename:  "test.png",
			size:      1024,
			buffer:    createTestPNG(t),
			wantError: nil,
		},
		{
			name:      "文件过大",
			filename:  "test.jpg",
			size:      3 * 1024 * 1024,
			buffer:    createTestJPEG(t),
			wantError: ErrFileSizeExceed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateReader(tt.buffer, tt.filename, tt.size)

			if tt.wantError != nil {
				if err != tt.wantError {
					t.Errorf("ValidateReader() error = %v, wantError %v", err, tt.wantError)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateReader() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestHasValidImageMagic(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    bool
	}{
		{
			name: "有效JPEG魔数",
			data: append([]byte{0xFF, 0xD8, 0xFF}, make([]byte, 509)...),
			want: true,
		},
		{
			name: "有效PNG魔数",
			data: append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, make([]byte, 504)...),
			want: true,
		},
		{
			name: "无效魔数",
			data: []byte(strings.Repeat("x", 512)),
			want: false,
		},
		{
			name: "数据太短",
			data: []byte{0xFF, 0xD8},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasValidImageMagic(tt.data); got != tt.want {
				t.Errorf("HasValidImageMagic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectImageType(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		want   string
	}{
		{
			name: "JPEG类型",
			data: createTestJPEG(t),
			want: ".jpg",
		},
		{
			name: "PNG类型",
			data: createTestPNG(t),
			want: ".png",
		},
		{
			name: "非图片",
			data: []byte(strings.Repeat("x", 512)),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectImageType(tt.data); got != tt.want {
				t.Errorf("DetectImageType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsImageFile(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "有效JPEG",
			data: createTestJPEG(t),
			want: true,
		},
		{
			name: "有效PNG",
			data: createTestPNG(t),
			want: true,
		},
		{
			name: "非图片",
			data: []byte(strings.Repeat("x", 512)),
			want: false,
		},
		{
			name: "数据太短",
			data: []byte{0xFF, 0xD8},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsImageFile(tt.data); got != tt.want {
				t.Errorf("IsImageFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

// createTestJPEG 创建测试用JPEG图片字节
func createTestJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("failed to create test jpeg: %v", err)
	}
	return buf.Bytes()
}

// createTestPNG 创建测试用PNG图片字节
func createTestPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to create test png: %v", err)
	}
	return buf.Bytes()
}
```

- [ ] **Step 3: 运行测试**

Run:
```bash
cd D:/software/GoWork/src/driver
go test -v ./taketaxi/pkg/upload/... -run TestValidator
```

Expected: 所有测试通过

- [ ] **Step 4: 提交**

```bash
cd D:/software/GoWork/src/driver
git add taketaxi/pkg/upload/validator.go taketaxi/pkg/upload/validator_test.go
git commit -m "feat(upload): implement file validator with tests

- File size limit: 2MB
- Allowed formats: jpg, png
- Magic number validation
- Comprehensive unit tests

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Task 6: MinIO 存储实现

**Files:**
- Create: `taketaxi/pkg/upload/minio.go`

- [ ] **Step 1: 编写 MinIO 存储实现**

Create `taketaxi/pkg/upload/minio.go`:

```go
package upload

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/credentials"
)

// MinIOStorage MinIO存储实现
type MinIOStorage struct {
	client    *minio.Client
	bucket    string
	validator *Validator
}

// NewMinIOStorage 创建MinIO存储实例
func NewMinIOStorage(cfg *MinIOConf) (*MinIOStorage, error) {
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
```

- [ ] **Step 2: 添加 bytes 导入并修复编译**

Edit `taketaxi/pkg/upload/minio.go`, add import:

```go
import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/credentials"
)
```

- [ ] **Step 3: 验证编译**

Run:
```bash
cd D:/software/GoWork/src/driver
go build ./taketaxi/pkg/upload/...
```

Expected: 无错误输出

- [ ] **Step 4: 提交**

```bash
cd D:/software/GoWork/src/driver
git add taketaxi/pkg/upload/minio.go
git commit -m "feat(upload): implement MinIO storage

- Upload with validation
- Delete object
- Presigned URL support
- Auto-create bucket

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Task 7: 存储工厂

**Files:**
- Create: `taketaxi/pkg/upload/factory.go`

- [ ] **Step 1: 编写存储工厂**

Create `taketaxi/pkg/upload/factory.go`:

```go
package upload

import "fmt"

// NewStorage 根据配置创建存储实例
func NewStorage(cfg *UploadConfig) (Storage, error) {
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
```

- [ ] **Step 2: 验证编译**

Run:
```bash
cd D:/software/GoWork/src/driver
go build ./taketaxi/pkg/upload/...
```

Expected: 报错 NewOSSStorage/NewCOSStorage 未定义（正常，下一步创建骨架）

---

## Task 8: OSS/COS 预留骨架

**Files:**
- Create: `taketaxi/pkg/upload/oss.go`
- Create: `taketaxi/pkg/upload/cos.go`

- [ ] **Step 1: 编写 OSS 骨架实现**

Create `taketaxi/pkg/upload/oss.go`:

```go
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
```

- [ ] **Step 2: 编写 COS 骨架实现**

Create `taketaxi/pkg/upload/cos.go`:

```go
package upload

import (
	"context"
	"fmt"
)

// COSStorage 腾讯云COS存储实现（预留）
type COSStorage struct {
	bucketURL string
	validator *Validator
}

// NewCOSStorage 创建COS存储实例
func NewCOSStorage(cfg *COSConf) (*COSStorage, error) {
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
```

- [ ] **Step 3: 验证编译**

Run:
```bash
cd D:/software/GoWork/src/driver
go build ./taketaxi/pkg/upload/...
```

Expected: 无错误输出

- [ ] **Step 4: 提交**

```bash
cd D:/software/GoWork/src/driver
git add taketaxi/pkg/upload/factory.go taketaxi/pkg/upload/oss.go taketaxi/pkg/upload/cos.go
git commit -m "feat(upload): add storage factory and OSS/COS skeleton

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Task 9: MinIO 集成测试

**Files:**
- Create: `taketaxi/pkg/upload/minio_test.go`

- [ ] **Step 1: 编写 MinIO 集成测试**

Create `taketaxi/pkg/upload/minio_test.go`:

```go
package upload

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"os"
	"testing"
	"time"
)

// 集成测试需要真实MinIO服务
// 运行前请确保MinIO服务已启动
// 可通过环境变量配置: MINIO_ENDPOINT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY, MINIO_BUCKET

func getMinIOConfig() *MinIOConf {
	return &MinIOConf{
		Endpoint:   getEnv("MINIO_ENDPOINT", "127.0.0.1:9000"),
		AccessKey:  getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		SecretKey:  getEnv("MINIO_SECRET_KEY", "minioadmin"),
		BucketName: getEnv("MINIO_BUCKET", "test-driver-images"),
		UseSSL:     false,
		Region:     "",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestMinIOStorage_NewMinIOStorage(t *testing.T) {
	cfg := getMinIOConfig()
	storage, err := NewMinIOStorage(cfg)
	if err != nil {
		t.Fatalf("NewMinIOStorage() error = %v", err)
	}
	if storage == nil {
		t.Fatal("NewMinIOStorage() returned nil")
	}
	if storage.GetStorageType() != StorageMinIO {
		t.Errorf("GetStorageType() = %v, want %v", storage.GetStorageType(), StorageMinIO)
	}
}

func TestMinIOStorage_Upload(t *testing.T) {
	cfg := getMinIOConfig()
	storage, err := NewMinIOStorage(cfg)
	if err != nil {
		t.Fatalf("NewMinIOStorage() error = %v", err)
	}

	// 创建测试PNG图片
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}

	tests := []struct {
		name     string
		filename string
		bizType  BizType
		wantErr  bool
	}{
		{
			name:     "上传头像",
			filename: "avatar.png",
			bizType:  BizTypeAvatar,
			wantErr:  false,
		},
		{
			name:     "上传身份证",
			filename: "idcard.png",
			bizType:  BizTypeIDCard,
			wantErr:  false,
		},
		{
			name:     "上传驾驶证",
			filename: "license.png",
			bizType:  BizTypeLicense,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(buf.Bytes())
			req := &UploadRequest{
				File:     reader,
				FileName: tt.filename,
				Size:     int64(buf.Len()),
				BizType:  tt.bizType,
			}

			result, err := storage.Upload(context.Background(), req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Upload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Fatal("Upload() returned nil result")
				}
				if result.URL == "" {
					t.Error("Upload() result.URL is empty")
				}
				if result.Path == "" {
					t.Error("Upload() result.Path is empty")
				}
				t.Logf("Upload success: URL=%s, Path=%s", result.URL, result.Path)

				// 测试删除
				if err := storage.Delete(context.Background(), result.Path); err != nil {
					t.Errorf("Delete() error = %v", err)
				}
			}
		})
	}
}

func TestMinIOStorage_Upload_InvalidFile(t *testing.T) {
	cfg := getMinIOConfig()
	storage, err := NewMinIOStorage(cfg)
	if err != nil {
		t.Fatalf("NewMinIOStorage() error = %v", err)
	}

	tests := []struct {
		name     string
		filename string
		content  []byte
		size     int64
		wantErr  error
	}{
		{
			name:     "文件过大",
			filename: "large.png",
			content:  buf.Bytes(),
			size:     3 * 1024 * 1024, // 3MB
			wantErr:  ErrFileSizeExceed,
		},
		{
			name:     "格式不支持",
			filename: "test.gif",
			content:  []byte("fake gif content"),
			size:     1024,
			wantErr:  ErrInvalidExt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.content)
			req := &UploadRequest{
				File:     reader,
				FileName: tt.filename,
				Size:     tt.size,
				BizType:  BizTypeAvatar,
			}

			_, err := storage.Upload(context.Background(), req)
			if err != tt.wantErr {
				t.Errorf("Upload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMinIOStorage_GetPresignedURL(t *testing.T) {
	cfg := getMinIOConfig()
	storage, err := NewMinIOStorage(cfg)
	if err != nil {
		t.Fatalf("NewMinIOStorage() error = %v", err)
	}

	// 先上传一个文件
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}

	reader := bytes.NewReader(buf.Bytes())
	req := &UploadRequest{
		File:     reader,
		FileName: "test.png",
		Size:     int64(buf.Len()),
		BizType:  BizTypeAvatar,
	}

	result, err := storage.Upload(context.Background(), req)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	defer storage.Delete(context.Background(), result.Path)

	// 获取预签名URL
	url, err := storage.GetPresignedURL(context.Background(), result.Path, time.Hour)
	if err != nil {
		t.Errorf("GetPresignedURL() error = %v", err)
	} else {
		t.Logf("PresignedURL: %s", url)
	}
}

func TestMinIOStorage_Delete(t *testing.T) {
	cfg := getMinIOConfig()
	storage, err := NewMinIOStorage(cfg)
	if err != nil {
		t.Fatalf("NewMinIOStorage() error = %v", err)
	}

	// 测试删除不存在的文件（MinIO不会报错）
	err = storage.Delete(context.Background(), "nonexistent/path/file.png")
	if err != nil {
		t.Logf("Delete nonexistent file: %v", err)
	}
}
```

- [ ] **Step 2: 验证编译**

Run:
```bash
cd D:/software/GoWork/src/driver
go build ./taketaxi/pkg/upload/...
```

Expected: 无错误输出

- [ ] **Step 3: 提交**

```bash
cd D:/software/GoWork/src/driver
git add taketaxi/pkg/upload/minio_test.go
git commit -m "test(upload): add MinIO integration tests

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Task 10: 更新配置模块

**Files:**
- Modify: `taketaxi/pkg/config/config.go`

- [ ] **Step 1: 读取当前配置文件**

Run:
```bash
cat D:/software/GoWork/src/driver/taketaxi/pkg/config/config.go
```

- [ ] **Step 2: 添加 UploadConfig 到配置结构体**

Edit `taketaxi/pkg/config/config.go`, update Config struct:

```go
package config

import (
	"os"

	"driver/taketaxi/pkg/upload"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig      `yaml:"server"`
	Database DatabaseConfig    `yaml:"database"`
	Redis    RedisConfig       `yaml:"redis"`
	Registry RegistryConfig    `yaml:"registry"`
	Upload   upload.UploadConfig `yaml:"upload"`
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
```

- [ ] **Step 3: 验证编译**

Run:
```bash
cd D:/software/GoWork/src/driver
go build ./taketaxi/pkg/...
```

Expected: 无错误输出

- [ ] **Step 4: 提交**

```bash
cd D:/software/GoWork/src/driver
git add taketaxi/pkg/config/config.go
git commit -m "feat(config): add upload config support

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Task 11: 创建配置文件

**Files:**
- Modify: `taketaxi/configs/config.yaml`

- [ ] **Step 1: 检查配置文件是否存在**

Run:
```bash
ls -la D:/software/GoWork/src/driver/taketaxi/configs/
```

- [ ] **Step 2: 创建或更新配置文件**

If exists, read and update. If not, create `taketaxi/configs/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  grpc_host: "127.0.0.1"
  grpc_port: 50051

database:
  host: "127.0.0.1"
  port: "3306"
  user: "root"
  password: "password"
  database: "driver"

redis:
  Host: "127.0.0.1"
  Port: 6379
  Password: ""
  Database: 0

registry:
  type: "etcd"
  address: "127.0.0.1:2379"

upload:
  storage_type: minio
  minio:
    endpoint: "127.0.0.1:9000"
    access_key: "minioadmin"
    secret_key: "minioadmin"
    bucket_name: "driver-images"
    use_ssl: false
    region: ""
  oss:
    endpoint: ""
    access_key_id: ""
    access_key_secret: ""
    bucket_name: ""
  cos:
    bucket_url: ""
    secret_id: ""
    secret_key: ""
```

- [ ] **Step 3: 提交**

```bash
cd D:/software/GoWork/src/driver
git add taketaxi/configs/config.yaml
git commit -m "chore: add upload config to config.yaml

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Task 12: Handler 实现

**Files:**
- Create: `taketaxi/bffDriver/internal/handler/uploadHandler.go`

- [ ] **Step 1: 编写 Handler**

Create `taketaxi/bffDriver/internal/handler/uploadHandler.go`:

```go
package handler

import (
	"net/http"

	"driver/taketaxi/pkg/upload"

	"github.com/gin-gonic/gin"
)

// UploadHandler 上传处理器
type UploadHandler struct {
	storage upload.Storage
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler(storage upload.Storage) *UploadHandler {
	return &UploadHandler{storage: storage}
}

// UploadResponse 上传响应
type UploadResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data *UploadData `json:"data,omitempty"`
}

// UploadData 上传数据
type UploadData struct {
	URL      string `json:"url"`
	Path     string `json:"path"`
	FileSize int64  `json:"file_size"`
}

// Upload 图片上传接口
// @Summary 图片上传
// @Description 上传图片到MinIO/OSS/COS
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "图片文件"
// @Param biz_type formData string true "业务类型: avatar/idcard/license/vehicle/face"
// @Success 200 {object} UploadResponse
// @Failure 400 {object} UploadResponse
// @Failure 500 {object} UploadResponse
// @Router /api/v1/upload [post]
func (h *UploadHandler) Upload(c *gin.Context) {
	// 1. 获取文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Code: 400,
			Msg:  "请选择要上传的文件",
		})
		return
	}
	defer file.Close()

	// 2. 获取业务类型
	bizType := c.PostForm("biz_type")
	if bizType == "" {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Code: 400,
			Msg:  "缺少业务类型参数",
		})
		return
	}

	// 3. 验证业务类型
	validBizTypes := map[string]bool{
		string(upload.BizTypeAvatar):  true,
		string(upload.BizTypeIDCard):  true,
		string(upload.BizTypeLicense): true,
		string(upload.BizTypeVehicle): true,
		string(upload.BizTypeFace):    true,
	}
	if !validBizTypes[bizType] {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Code: 400,
			Msg:  "无效的业务类型，支持: avatar/idcard/license/vehicle/face",
		})
		return
	}

	// 4. 构建上传请求
	req := &upload.UploadRequest{
		File:     file,
		FileName: header.Filename,
		Size:     header.Size,
		BizType:  upload.BizType(bizType),
	}

	// 5. 执行上传
	result, err := h.storage.Upload(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Code: 500,
			Msg:  err.Error(),
		})
		return
	}

	// 6. 返回结果
	c.JSON(http.StatusOK, UploadResponse{
		Code: 0,
		Msg:  "上传成功",
		Data: &UploadData{
			URL:      result.URL,
			Path:     result.Path,
			FileSize: result.Size,
		},
	})
}

// Delete 删除文件接口
// @Summary 删除文件
// @Description 删除已上传的文件
// @Produce json
// @Param path query string true "文件路径"
// @Success 200 {object} UploadResponse
// @Failure 400 {object} UploadResponse
// @Failure 500 {object} UploadResponse
// @Router /api/v1/upload [delete]
func (h *UploadHandler) Delete(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Code: 400,
			Msg:  "缺少文件路径参数",
		})
		return
	}

	if err := h.storage.Delete(c.Request.Context(), path); err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Code: 500,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UploadResponse{
		Code: 0,
		Msg:  "删除成功",
	})
}
```

- [ ] **Step 2: 验证编译**

Run:
```bash
cd D:/software/GoWork/src/driver
go build ./taketaxi/bffDriver/...
```

Expected: 无错误输出

- [ ] **Step 3: 提交**

```bash
cd D:/software/GoWork/src/driver
git add taketaxi/bffDriver/internal/handler/uploadHandler.go
git commit -m "feat(handler): add upload handler

- POST /api/v1/upload - upload image
- DELETE /api/v1/upload - delete file

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Task 13: 更新路由

**Files:**
- Modify: `taketaxi/bffDriver/internal/router/router.go`

- [ ] **Step 1: 更新路由文件**

Edit `taketaxi/bffDriver/internal/router/router.go`:

```go
package router

import (
	"driver/taketaxi/bffDriver/internal/handler"
	"driver/taketaxi/bffDriver/internal/rpcClient"
	"driver/taketaxi/pkg/upload"

	"github.com/gin-gonic/gin"
)

func NewRouter(client *rpcclient.DriverClient, storage upload.Storage) *gin.Engine {
	r := gin.Default()

	// Driver handlers
	driverHandler := handler.NewDriverHandler(client)
	r.GET("/api/v1/drivers", driverHandler.List)
	r.GET("/api/v1/drivers/:id", driverHandler.Get)
	r.POST("/api/v1/drivers", driverHandler.Create)
	r.PUT("/api/v1/drivers/:id", driverHandler.Update)
	r.DELETE("/api/v1/drivers/:id", driverHandler.Delete)

	// Upload handlers
	uploadHandler := handler.NewUploadHandler(storage)
	r.POST("/api/v1/upload", uploadHandler.Upload)
	r.DELETE("/api/v1/upload", uploadHandler.Delete)

	return r
}
```

- [ ] **Step 2: 验证编译**

Run:
```bash
cd D:/software/GoWork/src/driver
go build ./taketaxi/bffDriver/...
```

Expected: 报错 NewRouter 参数不匹配（正常，下一步更新 main.go）

---

## Task 14: 更新 main.go

**Files:**
- Modify: `taketaxi/bffDriver/cmd/main.go`

- [ ] **Step 1: 更新 main.go**

Edit `taketaxi/bffDriver/cmd/main.go`:

```go
package main

import (
	"driver/taketaxi/bffDriver/internal/router"
	"driver/taketaxi/bffDriver/internal/rpcClient"
	"driver/taketaxi/pkg/config"
	"driver/taketaxi/pkg/upload"
	"flag"
	"fmt"
	"log"
)

var confPath string

func init() {
	flag.StringVar(&confPath, "config", "../configs/config.yaml", "config file")
}

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(confPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建gRPC客户端
	grpcAddr := fmt.Sprintf("%s:%d", cfg.Server.GRPCHost, cfg.Server.GRPCPort)
	client, err := rpcclient.NewDriverClient(grpcAddr)
	if err != nil {
		log.Fatalf("创建gRPC客户端失败: %v", err)
	}
	defer client.Close()

	// 创建存储实例
	storage, err := upload.NewStorage(&cfg.Upload)
	if err != nil {
		log.Fatalf("创建存储实例失败: %v", err)
	}

	// 启动HTTP服务
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("BFF starting on %s", addr)
	router.NewRouter(client, storage).Run(addr)
}
```

- [ ] **Step 2: 验证编译**

Run:
```bash
cd D:/software/GoWork/src/driver
go build ./taketaxi/bffDriver/...
```

Expected: 无错误输出

- [ ] **Step 3: 提交**

```bash
cd D:/software/GoWork/src/driver
git add taketaxi/bffDriver/internal/router/router.go taketaxi/bffDriver/cmd/main.go
git commit -m "feat(router): integrate upload handler into router

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Task 15: 整体验证

- [ ] **Step 1: 运行所有测试**

Run:
```bash
cd D:/software/GoWork/src/driver
go test -v ./taketaxi/pkg/upload/...
```

Expected: 校验器测试通过，MinIO集成测试需要MinIO服务运行

- [ ] **Step 2: 编译整个项目**

Run:
```bash
cd D:/software/GoWork/src/driver
go build ./...
```

Expected: 无错误输出

- [ ] **Step 3: 最终提交**

```bash
cd D:/software/GoWork/src/driver
git add .
git commit -m "feat(upload): complete image upload feature

- MinIO storage with validation
- File size limit: 2MB, formats: jpg/png
- OSS/COS skeleton for future extension
- Integration tests included

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## 使用说明

### 启动 MinIO（Docker）

```bash
docker run -d \
  --name minio \
  -p 9000:9000 \
  -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data --console-address ":9001"
```

### 启动服务

```bash
cd D:/software/GoWork/src/driver
go run ./taketaxi/bffDriver/cmd/main.go -config ./taketaxi/configs/config.yaml
```

### 测试上传接口

```bash
# 上传图片
curl -X POST http://localhost:8080/api/v1/upload \
  -F "file=@test.png" \
  -F "biz_type=avatar"

# 删除文件
curl -X DELETE "http://localhost:8080/api/v1/upload?path=avatar/20260425/xxx.png"
```

### 运行集成测试

```bash
# 确保 MinIO 服务运行
cd D:/software/GoWork/src/driver
go test -v ./taketaxi/pkg/upload/... -run TestMinIO
```
