package routes

import (
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/firecracker"
	"gitlab.com/nunet/device-management-service/firecracker/telemetry"
	"gitlab.com/nunet/device-management-service/integrations/tokenomics"
	"gitlab.com/nunet/device-management-service/internal/tracing"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/libp2p/machines"
	"gitlab.com/nunet/device-management-service/onboarding"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.Use(cors.New(getCustomCorsConfig()))

	router.Use(otelgin.Middleware(tracing.ServiceName))

	v1 := router.Group("/api/v1")

	onboardingRoute := v1.Group("/onboarding")
	{
		onboardingRoute.GET("/provisioned", onboarding.ProvisionedCapacity)
		onboardingRoute.GET("/address/new", onboarding.CreatePaymentAddress)
		onboardingRoute.POST("/onboard", onboarding.Onboard)
		onboardingRoute.GET("/status", onboarding.Status)
		onboardingRoute.DELETE("/offboard", onboarding.Offboard)
		onboardingRoute.POST("/resource-config", onboarding.ResourceConfig)
		onboardingRoute.GET("/metadata", onboarding.GetMetadata)
	}

	virtualmachine := v1.Group("/vm")
	{
		virtualmachine.POST("/start-default", firecracker.StartDefault)
		virtualmachine.POST("/start-custom", firecracker.StartCustom)
	}

	run := v1.Group("/run")
	{
		run.GET("/deploy", machines.HandleDeploymentRequest) // websocket
		run.POST("/request-service", machines.HandleRequestService)
	}

	tx := v1.Group("/transactions")
	{
		tx.GET("", tokenomics.GetJobTxHashes)
		tx.POST("/request-reward", tokenomics.HandleRequestReward)
		run.POST("/send-status", tokenomics.HandleSendStatus)
	}

	tele := v1.Group("/telemetry")
	{
		tele.GET("/free", telemetry.GetFreeResource)
	}

	if _, debugMode := os.LookupEnv("NUNET_DEBUG"); debugMode {
		dht := v1.Group("/dht")
		{
			dht.GET("", libp2p.DumpDHT)
			dht.GET("/update", libp2p.ManualDHTUpdateHandler)
		}
		kadDht := v1.Group("/kad-dht")
		{
			kadDht.GET("", libp2p.DumpKademliaDHT)
		}
		v1.GET("/ping", libp2p.PingPeerHandler)
		v1.GET("/oldping", libp2p.OldPingPeerHandler)
		v1.GET("/cleanup", libp2p.CleanupPeerHandler)
	}

	p2p := v1.Group("/peers")
	{
		// peer.GET("", machines.ListPeers)
		p2p.GET("", libp2p.ListPeers)
		p2p.GET("/dht", libp2p.ListDHTPeers)
		p2p.GET("/kad-dht", libp2p.ListKadDHTPeers)
		p2p.GET("/self", libp2p.SelfPeerInfo)
		p2p.GET("/chat", libp2p.ListChatHandler)
		p2p.GET("/depreq", libp2p.DefaultDepReqPeer)
		p2p.GET("/chat/start", libp2p.StartChatHandler)
		p2p.GET("/chat/join", libp2p.JoinChatHandler)
		p2p.GET("/chat/clear", libp2p.ClearChatHandler)
		p2p.GET("/file", libp2p.ListFileRequestsHandler)
		p2p.GET("/file/send", libp2p.InitiateFileTransferHandler)
		p2p.GET("/file/accept", libp2p.AcceptFileTransferHandler)
		p2p.GET("/file/clear", libp2p.ClearFileTransferRequestseHandler)
		p2p.GET("/dht/dump", libp2p.DumpDHT)
		// peer.GET("/shell", internal.HandleWebSocket)
		// peer.GET("/log", internal.HandleWebSocket)
	}

	return router
}

func getCustomCorsConfig() cors.Config {
	config := DefaultConfig()
	// FIXME: This is a security concern.
	config.AllowOrigins = []string{"http://localhost:9991", "http://localhost:9992"}
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
