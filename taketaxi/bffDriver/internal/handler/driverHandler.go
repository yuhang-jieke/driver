package handler

import (
	"driver/taketaxi/pkg/logger"
	"net/http"
	"strconv"

	"driver/taketaxi/bffDriver/internal/rpcClient"
	pb "driver/taketaxi/common/kitexGen"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DriverHandler struct {
	client *rpcclient.DriverClient
}

func NewDriverHandler(client *rpcclient.DriverClient) *DriverHandler {
	return &DriverHandler{client: client}
}

func (h *DriverHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("创建司机参数错误", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.Create(c.Request.Context(), &pb.CreateDriverReq{Name: req.Name})
	if err != nil {
		logger.Error("创建司机失败", zap.String("name", req.Name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("创建司机成功", zap.Int64("id", resp.Id), zap.String("name", req.Name))
	c.JSON(http.StatusOK, gin.H{"id": resp.Id})
}

func (h *DriverHandler) List(c *gin.Context) {
	resp, err := h.client.List(c.Request.Context(), &pb.ListDriverReq{})
	if err != nil {
		logger.Error("获取司机列表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("获取司机列表成功", zap.Int("count", len(resp.Items)))
	c.JSON(http.StatusOK, gin.H{"drivers": resp.Items})
}

func (h *DriverHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Warn("获取司机参数错误", zap.String("id", c.Param("id")), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	resp, err := h.client.Get(c.Request.Context(), &pb.GetDriverReq{Id: id})
	if err != nil {
		logger.Error("获取司机失败", zap.Int64("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("获取司机成功", zap.Int64("id", id))
	c.JSON(http.StatusOK, resp)
}

func (h *DriverHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Warn("更新司机参数错误", zap.String("id", c.Param("id")), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("更新司机参数错误", zap.Int64("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.Update(c.Request.Context(), &pb.UpdateDriverReq{Id: id, Name: req.Name})
	if err != nil {
		logger.Error("更新司机失败", zap.Int64("id", id), zap.String("name", req.Name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("更新司机成功", zap.Int64("id", id), zap.String("name", req.Name))
	c.JSON(http.StatusOK, resp)
}

func (h *DriverHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Warn("删除司机参数错误", zap.String("id", c.Param("id")), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	resp, err := h.client.Delete(c.Request.Context(), &pb.DeleteDriverReq{Id: id})
	if err != nil {
		logger.Error("删除司机失败", zap.Int64("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("删除司机成功", zap.Int64("id", id))
	c.JSON(http.StatusOK, resp)
}
