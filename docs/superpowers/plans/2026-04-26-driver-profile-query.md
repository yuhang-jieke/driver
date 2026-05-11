# Driver Profile Query Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a new API endpoint `GET /api/v1/driver/profile` that returns driver personal info, order statistics, and verification status.

**Architecture:** HTTP handler in bffDriver calls gRPC service in srvDriver, which queries `drivers` table for personal info and `driver_statistics_summary` table for aggregated order stats.

**Tech Stack:** Go, gRPC, Gin, GORM

---

## Task 1: Update Proto Definition

**Files:**
- Modify: `taketaxi/common/idl/driver.proto`
- Regenerate: `taketaxi/common/kitexGen/driver.pb.go`
- Regenerate: `taketaxi/common/kitexGen/driver_grpc.pb.go`

- [ ] **Step 1: Update proto file with new messages and RPC method**

Edit `taketaxi/common/idl/driver.proto`:

```protobuf
syntax = "proto3";
package driver;
option go_package = "taketaxi/common/kitexGen/driver";

service DriverService {
  rpc Create(CreateDriverReq) returns (CreateDriverResp);
  rpc Get(GetDriverReq) returns (GetDriverResp);
  rpc List(ListDriverReq) returns (ListDriverResp);
  rpc Update(UpdateDriverReq) returns (UpdateDriverResp);
  rpc Delete(DeleteDriverReq) returns (DeleteDriverResp);
  rpc GetProfile(GetDriverProfileReq) returns (GetDriverProfileResp);
}

message CreateDriverReq { string name = 1; }
message CreateDriverResp { int64 id = 1; }
message GetDriverReq { int64 id = 1; }
message GetDriverResp { int64 id = 1; string name = 2; int32 status = 3; }
message ListDriverReq {}
message DriverItem { int64 id = 1; string name = 2; int32 status = 3; }
message ListDriverResp { repeated DriverItem items = 1; }
message UpdateDriverReq { int64 id = 1; string name = 2; }
message UpdateDriverResp { bool success = 1; }
message DeleteDriverReq { int64 id = 1; }
message DeleteDriverResp { bool success = 1; }

// GetDriverProfile 新增：查询司机个人信息与接单统计
message GetDriverProfileReq {
  int64 driver_id = 1;
  string date = 2;
  int32 days = 3;
}

message GetDriverProfileResp {
  PersonalInfo personal_info = 1;
  OrderStats order_stats = 2;
  int32 verify_status = 3;
}

message PersonalInfo {
  string nickname = 1;
  string avatar = 2;
  double service_score = 3;
  int32 order_count = 4;
}

message OrderStats {
  int32 order_count = 1;
  double income = 2;
  int32 online_duration = 3;
}

//protoc driver.proto `
//>> --go_out=../kitexGen `
//>> --go_opt=paths=source_relative `
//>> --go-grpc_out=../kitexGen `
//>> --go-grpc_opt=paths=source_relative
```

- [ ] **Step 2: Regenerate Go code from proto**

Run from `taketaxi/common/idl` directory:

```bash
cd "D:/software/GoWork/src/driver/taketaxi/common/idl" && protoc driver.proto --go_out=../kitexGen --go_opt=paths=source_relative --go-grpc_out=../kitexGen --go-grpc_opt=paths=source_relative
```

Expected: No errors, `driver.pb.go` and `driver_grpc.pb.go` updated with new messages and `GetProfile` method.

- [ ] **Step 3: Commit proto changes**

```bash
git add taketaxi/common/idl/driver.proto taketaxi/common/kitexGen/driver.pb.go taketaxi/common/kitexGen/driver_grpc.pb.go
git commit -m "feat(proto): add GetProfile RPC and related messages"
```

---

## Task 2: Add Repository Method

**Files:**
- Modify: `taketaxi/srvDriver/internal/repository/driverRepo.go`

- [ ] **Step 1: Add DriverProfileResult struct and GetProfile method to repository**

Edit `taketaxi/srvDriver/internal/repository/driverRepo.go`, append after existing methods:

```go
package repository

import (
	"context"
	"driver/taketaxi/pkg/database"
	"driver/taketaxi/srvDriver/internal/model"
	"time"

	"gorm.io/gorm"
)

type DriverRepo struct{ db *gorm.DB }

func NewDriverRepo(db *gorm.DB) *DriverRepo {
	if db == nil {
		db, _ = database.NewDB(nil)
	}
	return &DriverRepo{db: db}
}

func (r *DriverRepo) Create(ctx context.Context, m *model.Driver) error {
	return r.db.WithContext(ctx).Create(m).Error
}
func (r *DriverRepo) GetByID(ctx context.Context, id uint) (*model.Driver, error) {
	var m model.Driver
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}
func (r *DriverRepo) List(ctx context.Context) ([]*model.Driver, error) {
	var list []*model.Driver
	return list, r.db.WithContext(ctx).Find(&list).Error
}
func (r *DriverRepo) Update(ctx context.Context, m *model.Driver) error {
	return r.db.WithContext(ctx).Save(m).Error
}
func (r *DriverRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Driver{}, id).Error
}

// DriverProfileResult 个人信息查询结果
type DriverProfileResult struct {
	Nickname     string
	Avatar       string
	ServiceScore float64
	OrderCount   int
	VerifyStatus int8
}

// OrderStatsResult 接单统计结果
type OrderStatsResult struct {
	OrderCount     int
	TotalIncome    float64
	OnlineDuration int
}

// GetDriverProfile 查询司机个人信息
func (r *DriverRepo) GetDriverProfile(ctx context.Context, driverID int64) (*DriverProfileResult, error) {
	var driver model.DriverS
	if err := r.db.WithContext(ctx).
		Select("nickname, avatar, service_score, order_count, verify_status").
		Where("driver_id = ?", driverID).
		First(&driver).Error; err != nil {
		return nil, err
	}
	return &DriverProfileResult{
		Nickname:     driver.Nickname,
		Avatar:       driver.Avatar,
		ServiceScore: driver.ServiceScore,
		OrderCount:   driver.OrderCount,
		VerifyStatus: driver.VerifyStatus,
	}, nil
}

// GetOrderStats 查询接单统计（指定日期范围内汇总）
func (r *DriverRepo) GetOrderStats(ctx context.Context, driverID int64, startDate, endDate time.Time) (*OrderStatsResult, error) {
	var result OrderStatsResult
	err := r.db.WithContext(ctx).
		Model(&model.DriverStatisticsSummary{}).
		Select("COALESCE(SUM(order_count), 0) as order_count, COALESCE(SUM(total_income), 0) as total_income, COALESCE(SUM(online_duration), 0) as online_duration").
		Where("driver_id = ? AND stat_date BETWEEN ? AND ?", driverID, startDate, endDate).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}
```

- [ ] **Step 2: Commit repository changes**

```bash
git add taketaxi/srvDriver/internal/repository/driverRepo.go
git commit -m "feat(repo): add GetDriverProfile and GetOrderStats methods"
```

---

## Task 3: Add Service Method

**Files:**
- Modify: `taketaxi/srvDriver/internal/service/driverService.go`

- [ ] **Step 1: Add GetProfile method to service**

Edit `taketaxi/srvDriver/internal/service/driverService.go`, append after existing methods:

```go
package service

import (
	"context"
	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/srvDriver/internal/model"
	"driver/taketaxi/srvDriver/internal/repository"
	"errors"
	"time"
)

type DriverService struct{ repo *repository.DriverRepo }

func NewDriverService(repo *repository.DriverRepo) *DriverService {
	return &DriverService{repo: repo}
}

func (s *DriverService) Create(ctx context.Context, req *driver.CreateDriverReq) (*driver.CreateDriverResp, error) {
	m := &model.Driver{Name: req.Name}
	return &driver.CreateDriverResp{Id: int64(m.ID)}, s.repo.Create(ctx, m)
}
func (s *DriverService) Get(ctx context.Context, req *driver.GetDriverReq) (*driver.GetDriverResp, error) {
	m, err := s.repo.GetByID(ctx, uint(req.Id))
	if err != nil {
		return nil, err
	}
	return &driver.GetDriverResp{Id: int64(m.ID), Name: m.Name, Status: int32(m.Status)}, nil
}
func (s *DriverService) List(ctx context.Context, req *driver.ListDriverReq) (*driver.ListDriverResp, error) {
	list, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	var items []*driver.DriverItem
	for _, m := range list {
		items = append(items, &driver.DriverItem{Id: int64(m.ID), Name: m.Name, Status: int32(m.Status)})
	}
	return &driver.ListDriverResp{Items: items}, nil
}
func (s *DriverService) Update(ctx context.Context, req *driver.UpdateDriverReq) (*driver.UpdateDriverResp, error) {
	m, err := s.repo.GetByID(ctx, uint(req.Id))
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		m.Name = req.Name
	}
	return &driver.UpdateDriverResp{Success: true}, s.repo.Update(ctx, m)
}
func (s *DriverService) Delete(ctx context.Context, req *driver.DeleteDriverReq) (*driver.DeleteDriverResp, error) {
	return &driver.DeleteDriverResp{Success: true}, s.repo.Delete(ctx, uint(req.Id))
}

// GetProfile 查询司机个人信息与接单统计
func (s *DriverService) GetProfile(ctx context.Context, req *driver.GetDriverProfileReq) (*driver.GetDriverProfileResp, error) {
	// 参数校验
	if req.DriverId <= 0 {
		return nil, errors.New("invalid driver_id")
	}

	// 解析日期参数，默认当天
	var queryDate time.Time
	if req.Date == "" {
		queryDate = time.Now()
	} else {
		parsed, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			return nil, errors.New("invalid date format")
		}
		queryDate = parsed
	}

	// 校验日期不能为未来
	if queryDate.After(time.Now()) {
		return nil, errors.New("date cannot be in the future")
	}

	// 校验天数参数
	days := req.Days
	if days <= 0 {
		days = 1
	}
	if days > 30 {
		return nil, errors.New("days cannot exceed 30")
	}

	// 查询个人信息
	profile, err := s.repo.GetDriverProfile(ctx, req.DriverId)
	if err != nil {
		return nil, err
	}

	// 计算日期范围
	endDate := time.Date(queryDate.Year(), queryDate.Month(), queryDate.Day(), 0, 0, 0, 0, time.Local)
	startDate := endDate.AddDate(0, 0, -int(days)+1)

	// 查询接单统计
	stats, err := s.repo.GetOrderStats(ctx, req.DriverId, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return &driver.GetDriverProfileResp{
		PersonalInfo: &driver.PersonalInfo{
			Nickname:     profile.Nickname,
			Avatar:       profile.Avatar,
			ServiceScore: profile.ServiceScore,
			OrderCount:   int32(profile.OrderCount),
		},
		OrderStats: &driver.OrderStats{
			OrderCount:     int32(stats.OrderCount),
			Income:         stats.TotalIncome,
			OnlineDuration: int32(stats.OnlineDuration),
		},
		VerifyStatus: int32(profile.VerifyStatus),
	}, nil
}
```

- [ ] **Step 2: Commit service changes**

```bash
git add taketaxi/srvDriver/internal/service/driverService.go
git commit -m "feat(service): add GetProfile method for driver profile query"
```

---

## Task 4: Add gRPC Handler

**Files:**
- Modify: `taketaxi/srvDriver/internal/handler/driverHandler.go`

- [ ] **Step 1: Add GetProfile handler method**

Edit `taketaxi/srvDriver/internal/handler/driverHandler.go`, append after existing methods:

```go
package handler

import (
	"context"
	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/pkg/logger"
	"driver/taketaxi/srvDriver/internal/repository"
	"driver/taketaxi/srvDriver/internal/service"
	"time"

	"go.uber.org/zap"
)

type DriverHandler struct {
	driver.UnimplementedDriverServiceServer
	svc *service.DriverService
}

func NewDriverHandler(repo *repository.DriverRepo) *DriverHandler {
	return &DriverHandler{svc: service.NewDriverService(repo)}
}

func (h *DriverHandler) Create(ctx context.Context, req *driver.CreateDriverReq) (*driver.CreateDriverResp, error) {
	start := time.Now()
	resp, err := h.svc.Create(ctx, req)
	if err != nil {
		logger.Error("gRPC Create failed", zap.String("method", "Create"), zap.String("name", req.Name), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC Create success", zap.String("method", "Create"), zap.Int64("id", resp.Id), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

func (h *DriverHandler) Get(ctx context.Context, req *driver.GetDriverReq) (*driver.GetDriverResp, error) {
	start := time.Now()
	resp, err := h.svc.Get(ctx, req)
	if err != nil {
		logger.Error("gRPC Get failed", zap.String("method", "Get"), zap.Int64("id", req.Id), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC Get success", zap.String("method", "Get"), zap.Int64("id", req.Id), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

func (h *DriverHandler) List(ctx context.Context, req *driver.ListDriverReq) (*driver.ListDriverResp, error) {
	start := time.Now()
	resp, err := h.svc.List(ctx, req)
	if err != nil {
		logger.Error("gRPC List failed", zap.String("method", "List"), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC List success", zap.String("method", "List"), zap.Int("count", len(resp.Items)), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

func (h *DriverHandler) Update(ctx context.Context, req *driver.UpdateDriverReq) (*driver.UpdateDriverResp, error) {
	start := time.Now()
	resp, err := h.svc.Update(ctx, req)
	if err != nil {
		logger.Error("gRPC Update failed", zap.String("method", "Update"), zap.Int64("id", req.Id), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC Update success", zap.String("method", "Update"), zap.Int64("id", req.Id), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

func (h *DriverHandler) Delete(ctx context.Context, req *driver.DeleteDriverReq) (*driver.DeleteDriverResp, error) {
	start := time.Now()
	resp, err := h.svc.Delete(ctx, req)
	if err != nil {
		logger.Error("gRPC Delete failed", zap.String("method", "Delete"), zap.Int64("id", req.Id), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC Delete success", zap.String("method", "Delete"), zap.Int64("id", req.Id), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

func (h *DriverHandler) GetProfile(ctx context.Context, req *driver.GetDriverProfileReq) (*driver.GetDriverProfileResp, error) {
	start := time.Now()
	resp, err := h.svc.GetProfile(ctx, req)
	if err != nil {
		logger.Error("gRPC GetProfile failed", zap.String("method", "GetProfile"), zap.Int64("driver_id", req.DriverId), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC GetProfile success", zap.String("method", "GetProfile"), zap.Int64("driver_id", req.DriverId), zap.Duration("duration", time.Since(start)))
	return resp, nil
}
```

- [ ] **Step 2: Commit handler changes**

```bash
git add taketaxi/srvDriver/internal/handler/driverHandler.go
git commit -m "feat(handler): add GetProfile gRPC handler in srvDriver"
```

---

## Task 5: Add RPC Client Method

**Files:**
- Modify: `taketaxi/bffDriver/internal/rpcClient/driverClient.go`

- [ ] **Step 1: Add GetProfile client method**

Edit `taketaxi/bffDriver/internal/rpcClient/driverClient.go`, append after existing methods:

```go
package rpcclient

import (
	"context"

	driver "driver/taketaxi/common/kitexGen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DriverClient struct {
	conn   *grpc.ClientConn
	client driver.DriverServiceClient
}

func NewDriverClient(addr string) (*DriverClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &DriverClient{conn: conn, client: driver.NewDriverServiceClient(conn)}, nil
}

func (c *DriverClient) Close() { c.conn.Close() }

func (c *DriverClient) Create(ctx context.Context, req *driver.CreateDriverReq) (*driver.CreateDriverResp, error) {
	return c.client.Create(ctx, req)
}

func (c *DriverClient) Get(ctx context.Context, req *driver.GetDriverReq) (*driver.GetDriverResp, error) {
	return c.client.Get(ctx, req)
}

func (c *DriverClient) List(ctx context.Context, req *driver.ListDriverReq) (*driver.ListDriverResp, error) {
	return c.client.List(ctx, req)
}

func (c *DriverClient) Update(ctx context.Context, req *driver.UpdateDriverReq) (*driver.UpdateDriverResp, error) {
	return c.client.Update(ctx, req)
}

func (c *DriverClient) Delete(ctx context.Context, req *driver.DeleteDriverReq) (*driver.DeleteDriverResp, error) {
	return c.client.Delete(ctx, req)
}

func (c *DriverClient) GetProfile(ctx context.Context, req *driver.GetDriverProfileReq) (*driver.GetDriverProfileResp, error) {
	return c.client.GetProfile(ctx, req)
}
```

- [ ] **Step 2: Commit client changes**

```bash
git add taketaxi/bffDriver/internal/rpcClient/driverClient.go
git commit -m "feat(client): add GetProfile RPC client method"
```

---

## Task 6: Add HTTP Handler

**Files:**
- Modify: `taketaxi/bffDriver/internal/handler/driverHandler.go`

- [ ] **Step 1: Add Profile HTTP handler method**

Edit `taketaxi/bffDriver/internal/handler/driverHandler.go`, append `Profile` method after existing methods:

```go
package handler

import (
	"driver/taketaxi/pkg/logger"
	"net/http"
	"strconv"

	"driver/taketaxi/bffDriver/internal/rpcclient"
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

// Profile 查询司机个人信息与接单统计
func (h *DriverHandler) Profile(c *gin.Context) {
	// 从登录态获取司机ID（此处暂时从查询参数获取，后续接入认证中间件）
	driverIDStr := c.Query("driver_id")
	if driverIDStr == "" {
		logger.Warn("Profile 缺少 driver_id 参数")
		c.JSON(http.StatusBadRequest, gin.H{"error": "driver_id is required"})
		return
	}
	driverID, err := strconv.ParseInt(driverIDStr, 10, 64)
	if err != nil {
		logger.Warn("Profile 参数错误", zap.String("driver_id", driverIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid driver_id"})
		return
	}

	date := c.Query("date")
	daysStr := c.Query("days")
	var days int32 = 1
	if daysStr != "" {
		d, err := strconv.ParseInt(daysStr, 10, 32)
		if err != nil {
			logger.Warn("Profile 参数错误", zap.String("days", daysStr), zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid days"})
			return
		}
		days = int32(d)
	}

	resp, err := h.client.GetProfile(c.Request.Context(), &pb.GetDriverProfileReq{
		DriverId: driverID,
		Date:     date,
		Days:     days,
	})
	if err != nil {
		logger.Error("Profile 查询失败", zap.Int64("driver_id", driverID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("Profile 查询成功", zap.Int64("driver_id", driverID))
	c.JSON(http.StatusOK, resp)
}
```

- [ ] **Step 2: Commit HTTP handler changes**

```bash
git add taketaxi/bffDriver/internal/handler/driverHandler.go
git commit -m "feat(handler): add Profile HTTP handler in bffDriver"
```

---

## Task 7: Update Router

**Files:**
- Modify: `taketaxi/bffDriver/internal/router/router.go`

- [ ] **Step 1: Add profile route**

Edit `taketaxi/bffDriver/internal/router/router.go`:

```go
package router

import (
	"driver/taketaxi/bffDriver/internal/handler"
	"driver/taketaxi/bffDriver/internal/rpcclient"
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
	r.GET("/api/v1/driver/profile", driverHandler.Profile)

	// Upload handlers
	uploadHandler := handler.NewUploadHandler(storage)
	r.POST("/api/v1/upload", uploadHandler.Upload)
	r.DELETE("/api/v1/upload", uploadHandler.Delete)

	return r
}
```

- [ ] **Step 2: Commit router changes**

```bash
git add taketaxi/bffDriver/internal/router/router.go
git commit -m "feat(router): add GET /api/v1/driver/profile route"
```

---

## Task 8: Build Verification

- [ ] **Step 1: Verify Go build for srvDriver**

```bash
cd "D:/software/GoWork/src/driver/taketaxi/srvDriver" && go build ./...
```

Expected: No build errors.

- [ ] **Step 2: Verify Go build for bffDriver**

```bash
cd "D:/software/GoWork/src/driver/taketaxi/bffDriver" && go build ./...
```

Expected: No build errors.

- [ ] **Step 3: Final commit if any fixes needed**

If there are compilation errors, fix them and commit:

```bash
git add -A
git commit -m "fix: resolve build issues"
```
