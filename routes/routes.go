package routes

import (
	"device-management-service/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	v1 := router.Group("/api/v1")
	{
		v1.GET("/onboard", handlers.Onboarded)
		v1.POST("/onboard", handlers.Onboard)
	}

	return router
}
