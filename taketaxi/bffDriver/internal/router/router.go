package router

import (
	"driver/taketaxi/bffDriver/internal/handler"
	"driver/taketaxi/bffDriver/internal/rpcClient"
	"driver/taketaxi/pkg/upload"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(client *rpcclient.DriverClient, storage upload.Storage) *gin.Engine {
	r := gin.Default()

	// CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Driver handlers
	driverHandler := handler.NewDriverHandler(client)
	r.GET("/api/v1/drivers", driverHandler.List)
	r.GET("/api/v1/drivers/:id", driverHandler.Get)
	r.POST("/api/v1/drivers", driverHandler.Create)
	r.PUT("/api/v1/drivers/:id", driverHandler.Update)
	r.DELETE("/api/v1/drivers/:id", driverHandler.Delete)
	r.GET("/api/v1/driver/profile", driverHandler.Profile)
	r.GET("/api/v1/driver/income", driverHandler.Income)

	// Upload handlers
	uploadHandler := handler.NewUploadHandler(storage)
	r.POST("/api/v1/upload", uploadHandler.Upload)
	r.DELETE("/api/v1/upload", uploadHandler.Delete)

	return r
}
