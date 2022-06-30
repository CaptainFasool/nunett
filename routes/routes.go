package routes

import (
	"gitlab.com/nunet/device-management-service/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	v1 := router.Group("/api/v1")
	{
		v1.GET("/metadata", handlers.GetMetadata)
		v1.POST("/onboard", handlers.Onboard)
		v1.GET("/provisioned", handlers.ProvisionedCapacity)
		v1.GET("/address/new", handlers.CreatePaymentAddress)

		v1.POST("/echo", handlers.Echo)
		// v1.POST("/testnomad", handlers.NomadTest)

	}

	return router
}
