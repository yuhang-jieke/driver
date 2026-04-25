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
