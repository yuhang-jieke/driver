package router

import (
	"driver/taketaxi/bffDriver/internal/handler"
	"driver/taketaxi/bffDriver/internal/rpcClient"

	"github.com/gin-gonic/gin"
)

func NewRouter(client *rpcclient.DriverClient) *gin.Engine {
	r := gin.Default()
	h := handler.NewDriverHandler(client)
	r.GET("/api/v1/drivers", h.List)
	r.GET("/api/v1/drivers/:id", h.Get)
	r.POST("/api/v1/drivers", h.Create)
	r.PUT("/api/v1/drivers/:id", h.Update)
	r.DELETE("/api/v1/drivers/:id", h.Delete)
	r.GET("/api/v1/orders/:orderId/details", h.DriverDetails)
	r.GET("/api/driver/orders", h.DriverOrderList)

	return r
}
