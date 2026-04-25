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
	// 使用更大的图像确保PNG文件至少512字节
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to create test png: %v", err)
	}
	return buf.Bytes()
}
