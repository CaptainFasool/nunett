package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/firecracker"
	"gitlab.com/nunet/device-management-service/onboarding"
	spoPackage "gitlab.com/nunet/device-management-service/spo"
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
		virtualmachine.POST("/init/:vmID", firecracker.InitVM)
		virtualmachine.PUT("/boot-source/:vmID", firecracker.BootSource)
		virtualmachine.PUT("/drives/:vmID", firecracker.Drives)
		virtualmachine.PUT("/machine-config/:vmID", firecracker.MachineConfig)
		virtualmachine.PUT("/network-interfaces/:vmID", firecracker.NetworkInterfaces)
		virtualmachine.PUT("/start/:vmID", firecracker.StartVM)
		virtualmachine.PUT("/stop/:vmID", firecracker.StopVM)

		virtualmachine.POST("/start-default", firecracker.StartDefault)
		virtualmachine.POST("/start-custom", firecracker.StartCustom)
		virtualmachine.POST("/from-config", firecracker.RunFromConfig)
	}

	// SPO == Stake Pool Operator
	spo := v1.Group("/spo")
	{
		spo.GET("/devices", spoPackage.SearchDevice)
		// following route accept query param: deployment_type: auto/manual
		spo.POST("/deploy/:nodeID", spoPackage.SendDeploymentRequest)
	}

	return router
}
