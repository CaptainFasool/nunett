package main

import (
	"gitlab.com/nunet/device-management-service/cmd"
	"gitlab.com/nunet/device-management-service/db"
)

//	@title			Device Management Service
//	@version		0.4.119
//	@description	A dashboard application for computing providers.
//	@termsOfService	https://nunet.io/tos

//	@contact.name	Support
//	@contact.url	https://devexchange.nunet.io/
//	@contact.email	support@nunet.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @host		localhost:9999
// @BasePath	/api/v1
func main() {
	db.ConnectDatabase()
	cmd.Execute()
}

// To be used when migrating to the refactored version of the code

// func main() {
// 	config.LoadConfig()

// 	wg := new(sync.WaitGroup)
// 	wg.Add(1)
// 	dmsInstance := dms.NewDMS()

// 	cleanup := tracing.InitTracer()
// 	defer cleanup(context.Background())

// 	go startServer(wg)

// 	go messaging.DeploymentWorker()

// 	heartbeat.Done = make(chan bool)
// 	go heartbeat.Heartbeat()
// 	// wait for server to start properly before sending requests below
// 	time.Sleep(time.Second * 5)

// 	// get managed VMs, assume previous run left some VM running
// 	firecracker.RunPreviouslyRunningVMs()

// 	// Recreate host with previous keys
// 	dmsInstance.CheckOnboarding()
// 	wg.Wait()
// }
