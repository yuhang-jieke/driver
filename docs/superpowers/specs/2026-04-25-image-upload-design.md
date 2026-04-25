# 图片上传功能设计文档

## 1. 概述

### 1.1 背景
网约车司机端需要图片上传功能，支持头像、身份证、驾驶证、车辆照片、人脸照片等多种业务场景。

### 1.2 目标
- 企业级代码结构，遵循标准项目目录分层
- 基于 interface 和 struct 实现封装与多态
- 预留云存储扩展接口（OSS、COS）
- 生产级别可用，代码健壮、无明显漏洞
- 提供可调用的单元测试接口

## 2. 需求说明

### 2.1 业务类型
| 类型 | 标识 | 说明 |
|------|------|------|
| 头像 | avatar | 司机头像 |
| 身份证 | idcard | 身份证正反面 |
| 驾驶证 | license | 驾驶证照片 |
| 车辆照片 | vehicle | 车辆外观照片 |
| 人脸照片 | face | 人脸核验照片 |

### 2.2 安全校验
- 文件格式：仅支持 jpg、png
- 文件大小：最大 2MB
- 真实类型验证：通过魔数校验，防止伪造扩展名

### 2.3 存储方案
- 主存储：MinIO（S3兼容对象存储）
- 扩展支持：阿里云OSS、腾讯云COS
- 存储路径：`{bizType}/{YYYYMMDD}/{uuid}.{ext}`

## 3. 架构设计

### 3.1 目录结构
```
taketaxi/
├── pkg/
│   └── upload/
│       ├── upload.go          # 核心接口定义
│       ├── minio.go           # MinIO存储实现
│       ├── oss.go             # 阿里云OSS实现（预留）
│       ├── cos.go             # 腾讯云COS实现（预留）
│       ├── factory.go         # 存储工厂
│       ├── validator.go       # 文件校验器
│       ├── config.go          # 配置定义
│       └── upload_test.go     # 单元测试
├── config/
│   └── config.yaml            # 配置文件
└── srvDriver/
    └── internal/
        └── handler/
            └── uploadHandler.go  # 上传接口Handler
```

### 3.2 核心接口
```go
// Storage 存储接口 - 支持多存储实现
type Storage interface {
    Upload(ctx context.Context, req *UploadRequest) (*UploadResult, error)
    Delete(ctx context.Context, fileURL string) error
    GetStorageType() StorageType
}
```

### 3.3 数据流
```
客户端请求 -> Handler -> 校验器 -> Storage接口 -> MinIO/OSS/COS
                                    ↓
                              生成存储路径
                              {bizType}/{date}/{uuid}.{ext}
```

## 4. 详细设计

### 4.1 核心接口定义 (upload.go)

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

// Storage 存储接口
type Storage interface {
    Upload(ctx context.Context, req *UploadRequest) (*UploadResult, error)
    Delete(ctx context.Context, fileURL string) error
    GetStorageType() StorageType
}

// UploadRequest 上传请求
type UploadRequest struct {
    File     io.Reader
    FileName string
    Size     int64
    BizType  BizType
}

// UploadResult 上传结果
type UploadResult struct {
    URL   string
    Path  string
    Size  int64
}
```

### 4.2 配置定义 (config.go)

```go
package upload

type UploadConfig struct {
    StorageType string    `yaml:"storage_type"`
    MinIO       MinIOConf `yaml:"minio"`
    OSS         OSSConf   `yaml:"oss"`
    COS         COSConf   `yaml:"cos"`
}

type MinIOConf struct {
    Endpoint   string `yaml:"endpoint"`
    AccessKey  string `yaml:"access_key"`
    SecretKey  string `yaml:"secret_key"`
    BucketName string `yaml:"bucket_name"`
    UseSSL     bool   `yaml:"use_ssl"`
    Region     string `yaml:"region"`
}

type OSSConf struct {
    Endpoint        string `yaml:"endpoint"`
    AccessKeyID     string `yaml:"access_key_id"`
    AccessKeySecret string `yaml:"access_key_secret"`
    BucketName      string `yaml:"bucket_name"`
}

type COSConf struct {
    BucketURL string `yaml:"bucket_url"`
    SecretID  string `yaml:"secret_id"`
    SecretKey string `yaml:"secret_key"`
}
```

### 4.3 校验器 (validator.go)

```go
package upload

import (
    "errors"
    "io"
    "net/http"
    "path/filepath"
    "strings"
)

const (
    MaxFileSize      int64 = 2 * 1024 * 1024
    AllowedExts            = ".jpg,.png"
    AllowedMimeTypes       = "image/jpeg,image/png"
)

type Validator struct {
    maxSize      int64
    allowedExts  []string
    allowedMimes []string
}

func NewValidator() *Validator {
    return &Validator{
        maxSize:      MaxFileSize,
        allowedExts:  strings.Split(AllowedExts, ","),
        allowedMimes: strings.Split(AllowedMimeTypes, ","),
    }
}

func (v *Validator) Validate(file io.ReadSeeker, filename string, size int64) error {
    // 文件大小校验
    if size > v.maxSize {
        return errors.New("文件大小超过限制，最大允许2MB")
    }

    // 扩展名校验
    ext := strings.ToLower(filepath.Ext(filename))
    if !v.isAllowedExt(ext) {
        return errors.New("不支持的文件格式，仅支持jpg/png")
    }

    // 真实文件类型校验
    if err := v.validateMimeType(file); err != nil {
        return err
    }

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
    _, err := file.Read(buffer)
    if err != nil {
        return err
    }

    contentType := http.DetectContentType(buffer)
    for _, mime := range v.allowedMimes {
        if strings.HasPrefix(contentType, mime) {
            return nil
        }
    }

    return errors.New("文件内容类型不匹配")
}
```

### 4.4 MinIO存储实现 (minio.go)

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

type MinIOStorage struct {
    client    *minio.Client
    bucket    string
    validator *Validator
}

func NewMinIOStorage(cfg *MinIOConf) (*MinIOStorage, error) {
    client, err := minio.New(cfg.Endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
        Secure: cfg.UseSSL,
        Region: cfg.Region,
    })
    if err != nil {
        return nil, fmt.Errorf("初始化MinIO客户端失败: %w", err)
    }

    ctx := context.Background()
    exists, err := client.BucketExists(ctx, cfg.BucketName)
    if err != nil {
        return nil, fmt.Errorf("检查bucket失败: %w", err)
    }
    if !exists {
        if err := client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{}); err != nil {
            return nil, fmt.Errorf("创建bucket失败: %w", err)
        }
    }

    return &MinIOStorage{
        client:    client,
        bucket:    cfg.BucketName,
        validator: NewValidator(),
    }, nil
}

func (s *MinIOStorage) Upload(ctx context.Context, req *UploadRequest) (*UploadResult, error) {
    // 校验文件
    if seeker, ok := req.File.(io.ReadSeeker); ok {
        if err := s.validator.Validate(seeker, req.FileName, req.Size); err != nil {
            return nil, err
        }
    }

    // 生成对象路径
    ext := filepath.Ext(req.FileName)
    datePath := time.Now().Format("20060102")
    fileUUID := uuid.New().String()
    objectName := fmt.Sprintf("%s/%s/%s%s", req.BizType, datePath, fileUUID, ext)

    // 上传到MinIO
    _, err := s.client.PutObject(ctx, s.bucket, objectName, req.File, req.Size, minio.PutObjectOptions{
        ContentType: s.getContentType(ext),
    })
    if err != nil {
        return nil, fmt.Errorf("上传文件失败: %w", err)
    }

    return &UploadResult{
        URL:  fmt.Sprintf("/%s/%s", s.bucket, objectName),
        Path: objectName,
        Size: req.Size,
    }, nil
}

func (s *MinIOStorage) Delete(ctx context.Context, objectName string) error {
    return s.client.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{})
}

func (s *MinIOStorage) GetStorageType() StorageType {
    return StorageMinIO
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

### 4.5 存储工厂 (factory.go)

```go
package upload

import "fmt"

func NewStorage(cfg *UploadConfig) (Storage, error) {
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

### 4.6 Handler层 (uploadHandler.go)

```go
package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "driver/taketaxi/pkg/upload"
)

type UploadHandler struct {
    storage upload.Storage
}

func NewUploadHandler(storage upload.Storage) *UploadHandler {
    return &UploadHandler{storage: storage}
}

type UploadResponse struct {
    Code int         `json:"code"`
    Msg  string      `json:"msg"`
    Data *UploadData `json:"data,omitempty"`
}

type UploadData struct {
    URL      string `json:"url"`
    Path     string `json:"path"`
    FileSize int64  `json:"file_size"`
}

// Upload 图片上传接口
func (h *UploadHandler) Upload(c *gin.Context) {
    file, header, err := c.Request.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, UploadResponse{Code: 400, Msg: "请选择要上传的文件"})
        return
    }
    defer file.Close()

    bizType := c.PostForm("biz_type")
    if bizType == "" {
        c.JSON(http.StatusBadRequest, UploadResponse{Code: 400, Msg: "缺少业务类型参数"})
        return
    }

    req := &upload.UploadRequest{
        File:     file,
        FileName: header.Filename,
        Size:     header.Size,
        BizType:  upload.BizType(bizType),
    }

    result, err := h.storage.Upload(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, UploadResponse{Code: 500, Msg: err.Error()})
        return
    }

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
```

## 5. 配置示例

```yaml
upload:
  storage_type: minio
  minio:
    endpoint: 127.0.0.1:9000
    access_key: minioadmin
    secret_key: minioadmin
    bucket_name: driver-images
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

## 6. API接口

### 上传接口
- **URL**: `POST /api/v1/upload`
- **Content-Type**: `multipart/form-data`

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | 图片文件 |
| biz_type | string | 是 | 业务类型: avatar/idcard/license/vehicle/face |

**响应示例**:
```json
{
    "code": 0,
    "msg": "上传成功",
    "data": {
        "url": "/driver-images/avatar/20260425/550e8400-e29b-41d4-a716-446655440000.jpg",
        "path": "avatar/20260425/550e8400-e29b-41d4-a716-446655440000.jpg",
        "file_size": 102400
    }
}
```

## 7. 单元测试

测试用例覆盖：
1. 校验器测试：文件大小、格式、真实类型
2. 存储上传测试：MinIO集成测试
3. 存储删除测试：文件删除功能

测试要求：
- 使用真实MinIO服务进行集成测试
- 不使用mock数据
- 可直接调用接口验证功能
