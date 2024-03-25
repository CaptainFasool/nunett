package dms

import (
	"context"
	"fmt"
	"os"
	"time"

	"gitlab.com/nunet/device-management-service/api"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/docker"
	"gitlab.com/nunet/device-management-service/firecracker"
	"gitlab.com/nunet/device-management-service/internal"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/internal/heartbeat"
	"gitlab.com/nunet/device-management-service/internal/messaging"
	"gitlab.com/nunet/device-management-service/internal/tracing"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/utils"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Run() {
	ctx := context.Background()
	config.LoadConfig()

	db.ConnectDatabase()

	docker.StartCleanup()

	cleanup := tracing.InitTracer()
	defer cleanup(context.Background())

	go startServer()

	go messaging.DeploymentWorker()

	go messaging.FileTransferWorker(ctx)

	heartbeat.Done = make(chan bool)
	go heartbeat.Heartbeat()
	// wait for server to start properly before sending requests below
	time.Sleep(time.Second * 5)

	// get managed VMs, assume previous run left some VM running
	firecracker.RunPreviouslyRunningVMs()

	// Recreate host with previous keys
	libp2p.CheckOnboarding()
	if libp2p.GetP2P().Host != nil {
		SanityCheck(db.DB)
		heartbeat.CheckToken(libp2p.GetP2P().Host.ID().String(), utils.GetChannelName())
	}

	// wait for SIGINT or SIGTERM
	select {
	case sig := <-internal.ShutdownChan:
		fmt.Printf("Shutting down after getting a %v...", sig)

		// add actual cleanup code here
		fmt.Println("Cleaning up before shutting down")

		// exit
		os.Exit(0)
	}
}

func startServer() {
	router := api.SetupRouter()
	// router.Use(otelgin.Middleware(tracing.MachineName))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(fmt.Sprintf(":%d", config.GetConfig().Rest.Port))

}
