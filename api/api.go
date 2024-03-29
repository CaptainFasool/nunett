package api

import (
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"gitlab.com/nunet/device-management-service/internal/tracing"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.Use(cors.New(getCustomCorsConfig()))

	router.Use(otelgin.Middleware(tracing.ServiceName))

	v1 := router.Group("/api/v1")

	onboarding := v1.Group("/onboarding")
	{
		onboarding.GET("/metadata", GetMetadataHandler)
		onboarding.GET("/provisioned", ProvisionedCapacityHandler)
		onboarding.GET("/address/new", CreatePaymentAddressHandler)
		onboarding.GET("/status", OnboardStatusHandler)
		onboarding.POST("/onboard", OnboardHandler)
		onboarding.POST("/resource-config", ResourceConfigHandler)
		onboarding.POST("/offboard", OffboardHandler)
	}

	device := v1.Group("/device")
	{
		device.GET("/status", DeviceStatusHandler)
		device.POST("/status", ChangeDeviceStatusHandler)
	}

	vm := v1.Group("/vm")
	{
		vm.POST("/start-default", StartDefaultHandler)
		vm.POST("/start-custom", StartCustomHandler)
	}

	run := v1.Group("/run")
	{
		run.GET("/deploy", DeploymentRequestHandler) // websocket
		run.GET("/checkpoints", ListCheckpointHandler)
		run.POST("/request-service", RequestServiceHandler)
	}

	tx := v1.Group("/transactions")
	{
		tx.GET("", GetJobTxHashesHandler)
		tx.POST("/request-reward", RequestRewardHandler)
		tx.POST("/send-status", SendTxStatusHandler)
		tx.POST("/update-status", UpdateTxStatusHandler)
	}

	tele := v1.Group("/telemetry")
	{
		tele.GET("/free", GetFreeResourcesHandler)
	}

	if _, debugMode := os.LookupEnv("NUNET_DEBUG"); debugMode {
		dht := v1.Group("/dht")
		{
			dht.GET("/update", ManualDHTUpdateHandler)
		}
		kadDHT := v1.Group("/kad-dht")
		{
			kadDHT.GET("", DumpKademliaDHTHandler)
		}
		v1.GET("/ping", PingPeerHandler)
		v1.GET("/oldping", OldPingPeerHandler)
		v1.GET("/cleanup", CleanupPeerHandler)
	}

	p2p := v1.Group("/peers")
	{
		p2p.GET("", ListPeersHandler)
		p2p.GET("/dht", ListDHTPeersHandler)
		p2p.GET("/dht/dump", DumpDHTHandler)
		p2p.GET("/kad-dht", ListKadDHTPeersHandler)
		p2p.GET("/self", SelfPeerInfoHandler)
		p2p.GET("/chat", ListChatHandler)
		p2p.GET("/depreq", DefaultDepReqPeerHandler)
		p2p.GET("/chat/start", StartChatHandler)
		p2p.GET("/chat/join", JoinChatHandler)
		p2p.GET("/chat/clear", ClearChatHandler)
		p2p.GET("/file", ListFileTransferRequestsHandler)
		p2p.GET("/file/send", SendFileTransferHandler)
		p2p.GET("/file/accept", AcceptFileTransferHandler)
		p2p.GET("/file/clear", ClearFileTransferRequestsHandler)
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
