package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/firecracker"
	"gitlab.com/nunet/device-management-service/onboarding"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	v1 := router.Group("/api/v1")
	v1.POST("/echo", onboarding.Echo)

	onboardingRoute := v1.Group("/onboarding")
	{
		onboardingRoute.GET("/metadata", onboarding.GetMetadata)
		onboardingRoute.POST("/onboard", onboarding.Onboard)
		onboardingRoute.GET("/provisioned", onboarding.ProvisionedCapacity)
		onboardingRoute.GET("/address/new", onboarding.CreatePaymentAddress)
	}

	virtualmachine := v1.Group("/vm")
	{
		virtualmachine.GET("/ping", firecracker.Ping)
		virtualmachine.POST("/launch", firecracker.LaunchVM)
		// virtualmachine.GET("/health", firecracker.GetHealth)
		// virtualmachine.POST("/delete", firecracker.DeleteVM)

	}

	return router
}
