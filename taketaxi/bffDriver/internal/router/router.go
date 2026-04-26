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
	r.GET("/api/v1/driver/profile", driverHandler.Profile)

	// Upload handlers
	uploadHandler := handler.NewUploadHandler(storage)
	r.POST("/api/v1/upload", uploadHandler.Upload)
	r.DELETE("/api/v1/upload", uploadHandler.Delete)

	return r
}
