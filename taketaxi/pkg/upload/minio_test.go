package upload

import (
	"bytes"
	"context"
	"driver/taketaxi/pkg/config"
	"image"
	"image/png"
	"os"
	"testing"
	"time"
)

// 集成测试需要真实MinIO服务
// 运行前请确保MinIO服务已启动
// 可通过环境变量配置: MINIO_ENDPOINT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY, MINIO_BUCKET

func getMinIOConfig() *config.MinIOConf {
	return &config.MinIOConf{
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

	// 创建测试图片
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	png.Encode(&buf, img)

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
