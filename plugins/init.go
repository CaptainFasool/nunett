package plugins

import (
	"context"

	"github.com/docker/docker/client"
	"gitlab.com/nunet/device-management-service/internal/logger"
)

var (
	ctx  context.Context
	dc   *client.Client
	zlog *logger.Logger
)

// TODOs:

// 1. Check plugins enabled by CP
// 2. Init plugins enabled

// 1. Pull Container Image
// 2. Run Container Image
// 3. Update DHT with decreased free/available resources
// 4. (Optional) Do things while container is running
// 5. When job is finished, remove stored IPFS data for the specific job (send /delete call)
// 6. Free resources (delete container image)
// 7. Update DHT with increased free/available resources

func init() {
	var err error

	zlog = logger.New("plugins")
	ctx = context.Background()

	dc, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		zlog.Sugar().Errorf("Unable to initialize Docker client for plugins: %v", err)
		return
	}
}

func StartPlugins() {
	// TODO: Init all plugins
	go RunContainer()
}
