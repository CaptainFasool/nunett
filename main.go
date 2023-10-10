package main

import (
	"gitlab.com/nunet/device-management-service/cmd"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/docker/watcher"
)

//	@title			Device Management Service
//	@version		0.4.130
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
	// Start the watcher client to monitor the DMS heartbeats
	go watcher.StartServerAndClient()
	go watcher.WatchForHeartbeats()
	db.ConnectDatabase()
	cmd.Execute()
}
