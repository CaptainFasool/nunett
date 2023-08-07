// Package docker is marked for deletion. See `service` module.
//
// `docker` package is used to run docker containers for ML on GPU/CPU usecase.
package docker

import (
	"github.com/docker/docker/client"
	"gitlab.com/nunet/device-management-service/internal/logger"
)

var (
	dc   *client.Client
	zlog *logger.Logger
)

func init() {
	zlog = logger.New("docker")
}
