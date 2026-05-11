package handler

import (
	"driver/taketaxi/pkg/logger"
	"driver/taketaxi/pkg/upload"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
	start := time.Now()

	// 1. 获取文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		logger.Warn("获取上传文件失败", zap.Error(err))
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
		logger.Warn("缺少业务类型参数")
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
		logger.Warn("无效的业务类型", zap.String("biz_type", bizType))
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
		logger.Error("上传失败",
			zap.String("biz_type", bizType),
			zap.String("filename", header.Filename),
			zap.Int64("size", header.Size),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)),
		)
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Code: 500,
			Msg:  err.Error(),
		})
		return
	}

	// 6. 返回结果
	logger.Info("上传成功",
		zap.String("biz_type", bizType),
		zap.String("filename", header.Filename),
		zap.Int64("size", header.Size),
		zap.String("path", result.Path),
		zap.Duration("duration", time.Since(start)),
	)
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
		logger.Warn("删除文件缺少路径参数")
		c.JSON(http.StatusBadRequest, UploadResponse{
			Code: 400,
			Msg:  "缺少文件路径参数",
		})
		return
	}

	if err := h.storage.Delete(c.Request.Context(), path); err != nil {
		logger.Error("删除文件失败", zap.String("path", path), zap.Error(err))
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Code: 500,
			Msg:  err.Error(),
		})
		return
	}

	logger.Info("删除文件成功", zap.String("path", path))
	c.JSON(http.StatusOK, UploadResponse{
		Code: 0,
		Msg:  "删除成功",
	})
}
