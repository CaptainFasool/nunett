package routes

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/adapter/machines"
	"gitlab.com/nunet/device-management-service/firecracker"
	"gitlab.com/nunet/device-management-service/firecracker/telemetry"
	"gitlab.com/nunet/device-management-service/onboarding"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.Use(cors.New(getCustomCorsConfig()))

	v1 := router.Group("/api/v1")

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

	// spo and gpu endpoint will merge.
	// devices endpoint will go away and exist as a function only

	run := v1.Group("/run")
	{
		run.POST("/deploy", machines.SendDeploymentRequest)
		run.GET("/deploy/receive", machines.ReceiveDeploymentRequest)
	}

	tele := v1.Group("/telemetry")
	{
		tele.GET("/free", telemetry.GetFreeResource)
	}

	peer := v1.Group("/peers")
	{
		peer.GET("", machines.ListPeers)
	}

	return router
}

func getCustomCorsConfig() cors.Config {
	config := DefaultConfig()
	// FIXME: This is a security concern.
	config.AllowOrigins = []string{"http://localhost:9998"}
	return config
}

// DefaultConfig returns a generic default configuration mapped to localhost.
func DefaultConfig() cors.Config {
	return cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Access-Control-Allow-Origin", "Origin", "Content-Length", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
}
