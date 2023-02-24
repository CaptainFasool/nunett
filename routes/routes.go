package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/adapter/machines"
	"gitlab.com/nunet/device-management-service/firecracker"
	"gitlab.com/nunet/device-management-service/firecracker/telemetry"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/onboarding"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	// router.Use(cors.New(getCustomCorsConfig()))

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
		virtualmachine.POST("/start-default", firecracker.StartDefault)
		virtualmachine.POST("/start-custom", firecracker.StartCustom)
	}

	v1.GET("/run/deploy", machines.HandleDeploymentRequest)

	tele := v1.Group("/telemetry")
	{
		tele.GET("/free", telemetry.GetFreeResource)
	}

	p2p := v1.Group("/peers")
	{
		// peer.GET("", machines.ListPeers)
		p2p.GET("", libp2p.ListPeers)
		p2p.GET("/self", libp2p.SelfPeerInfo)
		p2p.GET("/chat", libp2p.ListChatHandler)
		p2p.GET("/chat/start", libp2p.StartChatHandler)
		p2p.GET("/chat/join", libp2p.JoinChatHandler)
		p2p.GET("/chat/clear", libp2p.ClearChatHandler)
		// peer.GET("/shell", internal.HandleWebSocket)
		// peer.GET("/log", internal.HandleWebSocket)
	}

	return router
}

// func getCustomCorsConfig() cors.Config {
// 	config := DefaultConfig()
// 	// FIXME: This is a security concern.
// 	config.AllowOrigins = []string{"http://localhost:9998"}
// 	return config
// }

// // DefaultConfig returns a generic default configuration mapped to localhost.
// func DefaultConfig() cors.Config {
// 	return cors.Config{
// 		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
// 		AllowHeaders:     []string{"Access-Control-Allow-Origin", "Origin", "Content-Length", "Content-Type"},
// 		AllowCredentials: false,
// 		MaxAge:           12 * time.Hour,
// 	}
// }
