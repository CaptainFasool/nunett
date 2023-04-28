package docker

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/google/go-github/github"
	"gitlab.com/nunet/device-management-service/internal/logger"
	"gitlab.com/nunet/device-management-service/utils"
	"golang.org/x/oauth2"
)

var (
	gh       *github.Client
	ctx      context.Context
	dc       *client.Client
	gHealthy bool
	zlog     *logger.Logger
)

func init() {
	zlog = logger.New("docker")

	// initialise GitHub client
	ctx = context.Background()

	gist_token, err := utils.ReadHttpString("https://d.nunet.io/gist_token")
	if err != nil {
		zlog.Sugar().Errorf("unable to read gist token: %v", err)
		gHealthy = false
	} else {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: gist_token},
		)
		tc := oauth2.NewClient(ctx, ts)
		gh = github.NewClient(tc)
		dc, _ = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		gHealthy = true
	}
}
