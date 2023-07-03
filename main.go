package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mitchellh/cli"
	"gitlab.com/nunet/device-management-service/db"
	_ "gitlab.com/nunet/device-management-service/docs"
	"gitlab.com/nunet/device-management-service/firecracker"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/internal/heartbeat"
	"gitlab.com/nunet/device-management-service/internal/messaging"
	"gitlab.com/nunet/device-management-service/internal/tracing"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/routes"
	"gitlab.com/nunet/device-management-service/utils"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Device Management Service
// @version         0.4.97
// @description     A dashboard application for computing providers.
// @termsOfService  https://nunet.io/tos

// @contact.name   Support
// @contact.url    https://devexchange.nunet.io/
// @contact.email  support@nunet.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:9999
// @BasePath  /api/v1
func main() {
	c := cli.NewCLI("nunet", utils.GetDMSVersion()) 
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"daemon": func() (cli.Command, error) {
			return &daemonCommand{}, nil
		},
	}
	exitStatus, err := c.Run()
	if err != nil {
		fmt.Println(err)
	}

	os.Exit(exitStatus)
}

type daemonCommand struct{}

func (c *daemonCommand) Help() string {
	return "Launch DMS"
}

func (c *daemonCommand) Run(args []string) int {
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
	wg.Wait()

	return 0
}

func (c *daemonCommand) Synopsis() string {
	return "Launch DMS daemon as background process"
}

func startServer(wg *sync.WaitGroup) {
	defer wg.Done()

	router := routes.SetupRouter()
	// router.Use(otelgin.Middleware(tracing.MachineName))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(fmt.Sprintf(":%d", config.GetConfig().Rest.Port))
}
