package main

import (
	"sync"
	"time"

	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/firecracker"
	"gitlab.com/nunet/device-management-service/routes"

	_ "gitlab.com/nunet/device-management-service/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Device Management Service
// @version         0.3.2
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
	wg := new(sync.WaitGroup)
	wg.Add(1)

	go startServer(wg)

	// wait for server to start properly before sending requests below
	time.Sleep(time.Second * 5)

	// get managed VMs, assume previous run left some VM running
	firecracker.RunPreviouslyRunningVMs()

	wg.Wait()
}

func startServer(wg *sync.WaitGroup) {
	defer wg.Done()

	router := routes.SetupRouter()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	db.ConnectDatabase()

	router.Run(":9999")

}
