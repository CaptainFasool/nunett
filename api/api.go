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
		onboarding.GET("/provisioned", HandleProvisionedCapacity)
		onboarding.GET("/address/new", HandleCreatePaymentAddress)
		onboarding.POST("/onboard", HandleOnboard)
		onboarding.GET("/status", HandleOnboardStatus)
		onboarding.DELETE("/offboard", HandleOffboard)
		onboarding.POST("/resource-config", HandleResourceConfig)
		onboarding.GET("/metadata", HandleGetMetadata)
	}

	device := v1.Group("/device")
	{
		device.GET("/status", DeviceStatusHandler)
		device.POST("/status", ChangeDeviceStatusHandler)
	}

	vm := v1.Group("/vm")
	{
		vm.POST("/start-default", HandleStartDefault)
		vm.POST("/start-custom", HandleStartCustom)
	}

	run := v1.Group("/run")
	{
		run.GET("/deploy", HandleDeploymentRequest) // websocket
		run.POST("/request-service", HandleRequestService)
		run.GET("/checkpoints", HandleListCheckpoint)
	}

	tx := v1.Group("/transactions")
	{
		tx.GET("", HandleGetJobTxHashes)
		tx.POST("/request-reward", HandleRequestReward)
		tx.POST("/send-status", HandleSendStatus)
		tx.POST("/update-status", HandleUpdateStatus)
	}

	tele := v1.Group("/telemetry")
	{
		tele.GET("/free", HandleGetFreeResources)
	}

	if _, debugMode := os.LookupEnv("NUNET_DEBUG"); debugMode {
		dht := v1.Group("/dht")
		{
			dht.GET("", HandleDumpDHT)
			dht.GET("/update", HandleManualDHTUpdate)
		}
		kadDHT := v1.Group("/kad-dht")
		{
			kadDHT.GET("", HandleDumpKademliaDHT)
		}
		v1.GET("/ping", HandlePingPeer)
		v1.GET("/oldping", HandleOldPingPeer)
		v1.GET("/cleanup", HandleCleanupPeer)
	}

	p2p := v1.Group("/peers")
	{
		// peer.GET("", machines.ListPeers)
		p2p.GET("", HandleListPeers)
		p2p.GET("/dht", HandleListDHTPeers)
		p2p.GET("/kad-dht", HandleListKadDHTPeers)
		p2p.GET("/self", HandleSelfPeerInfo)
		p2p.GET("/chat", HandleListChat)
		p2p.GET("/depreq", HandleDefaultDepReqPeer)
		// TODO: change to HandleStartChat
		p2p.GET("/chat/start", StartChatHandler)
		p2p.GET("/chat/join", HandleJoinChat)
		// TODO: change to HandleChatClear
		p2p.GET("/chat/clear", ClearChatHandler)
		p2p.GET("/file", HandleListFileRequests)
		// TODO: change name to HandleSendFileRequest
		p2p.GET("/file/send", HandleInitiateFileTransfer)
		p2p.GET("/file/accept", HandleAcceptFileTransfer)
		p2p.GET("/file/clear", HandleClearFileTransferRequests)
		// ?? duplicate func
		// p2p.GET("/dht/dump", libp2p.DumpDHT)
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
