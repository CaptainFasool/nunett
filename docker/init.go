package docker

import (
	"context"
	"os"

	"github.com/docker/docker/client"
	"github.com/google/go-github/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"

	"gitlab.com/nunet/device-management-service/internal/logger"
)

var (
	gh  *github.Client
	ctx context.Context
	dc  *client.Client

	zlog *logger.Logger
)

func init() {
	godotenv.Load()

	// initialise GitHub client
	ctx = context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_PAT")},
	)

	tc := oauth2.NewClient(ctx, ts)
	gh = github.NewClient(tc)
	dc, _ = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	zlog = logger.New("docker")
}
