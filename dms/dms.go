package dms

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/firecracker"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/internal/heartbeat"
	"gitlab.com/nunet/device-management-service/internal/messaging"
	"gitlab.com/nunet/device-management-service/internal/tracing"
	"gitlab.com/nunet/device-management-service/libp2p"
	plugins "gitlab.com/nunet/device-management-service/plugins/plugins_startup"
	"gitlab.com/nunet/device-management-service/routes"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Run() {
	config.LoadConfig()

	wg := new(sync.WaitGroup)
	wg.Add(1)

	db.ConnectDatabase()

	cleanup := tracing.InitTracer()
	defer cleanup(context.Background())

	go startServer(wg)

	go messaging.DeploymentWorker()

	heartbeat.Done = make(chan bool)
	go heartbeat.Heartbeat()
	// wait for server to start properly before sending requests below
	time.Sleep(time.Second * 5)

	// get managed VMs, assume previous run left some VM running
	firecracker.RunPreviouslyRunningVMs()

	// Recreate host with previous keys
	libp2p.CheckOnboarding()

	// Iniate plugins if any enabled
	plugins.StartPlugins()

	wg.Wait()
}

func startServer(wg *sync.WaitGroup) {
	defer wg.Done()

	router := routes.SetupRouter()
	// router.Use(otelgin.Middleware(tracing.MachineName))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(fmt.Sprintf(":%d", config.GetConfig().Rest.Port))
}
