package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/firecracker"
	"gitlab.com/nunet/device-management-service/onboarding"
	spoPackage "gitlab.com/nunet/device-management-service/spo"
	cardano "gitlab.com/nunet/device-management-service/cardano"
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
		spo.GET("/search_device", spoPackage.SearchDevice)
		spo.GET("/req_cardano_deploy/:peerID/auto", spoPackage.DeployAuto)
		spo.GET("/req_cardano_deploy/:peerID/manual", spoPackage.DeployManual)
	}

	cardano_route := v1.Group("/trigger_cardano")
	{
		cardano_route.GET("/:peerdID/manual", cardano.Deploy)
		cardano_route.GET("/:peerdID/auto", cardano.Deploy)
	}
	
	return router
}
