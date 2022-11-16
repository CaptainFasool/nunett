package main

import (
	"context"
	"sync"
	"time"

	"gitlab.com/nunet/device-management-service/adapter"
	"gitlab.com/nunet/device-management-service/db"
	_ "gitlab.com/nunet/device-management-service/docs"
	"gitlab.com/nunet/device-management-service/firecracker"
	"gitlab.com/nunet/device-management-service/routes"
	"go.opentelemetry.io/otel"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Device Management Service
// @version         0.4.5
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

	// Poll messages from the adapter
	// go adapter.PollAdapter()

	// wait for server to start properly before sending requests below
	time.Sleep(time.Second * 5)

	//export traces to jaeger
	tp, _ := adapter.TracerProvider("http://testserver.nunet.io:14268/api/traces")

	otel.SetTracerProvider(tp)
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			_ = err
		}
	}()

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
