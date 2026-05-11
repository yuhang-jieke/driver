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
