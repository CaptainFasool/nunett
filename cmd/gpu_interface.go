package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	library "gitlab.com/nunet/device-management-service/lib"
	"gitlab.com/nunet/device-management-service/utils"
)

type Utility interface {
	CheckWSL() (bool, error)
	IsOnboarded() (bool, error)
	ResponseBody(c *gin.Context, method, endpoint, query string, body []byte) ([]byte, error)
}

type Utils struct{}

func (u *Utils) CheckWSL() (bool, error) {
	return utils.CheckWSL()
}

func (u *Utils) IsOnboarded() (bool, error) {
	return utils.IsOnboarded()
}

func (u *Utils) ResponseBody(c *gin.Context, method, endpoint, query string, body []byte) ([]byte, error) {
	return utils.ResponseBody(c, method, endpoint, query, body)
}

////////////////////

type FileSystem interface {
	ReadFile(name string) ([]byte, error)
}

type FS struct{}

func (f *FS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

/////////////////////

// abstracts running commands
type Executer interface {
	Execute(name string, arg ...string) Commander
}

type CmdExecutor struct{}

func (c *CmdExecutor) Execute(name string, arg ...string) Commander {
	return &ExecCommand{cmd: exec.Command(name, arg...)}
}

type Commander interface {
	CombinedOutput() ([]byte, error)
}

type ExecCommand struct {
	cmd *exec.Cmd
}

func (e *ExecCommand) CombinedOutput() ([]byte, error) {
	return e.cmd.CombinedOutput()
}

////////////////////

type Docker interface {
	ContainerAttach(ctx context.Context, container string, opt types.ContainerAttachOptions) (types.HijackedResponse, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *specs.Platform, name string) (container.ContainerCreateCreatedBody, error)
	ContainerRemove(ctx context.Context, id string, opt types.ContainerRemoveOptions) error
	ContainerStart(ctx context.Context, id string, opt types.ContainerStartOptions) error
	ContainerWait(ctx context.Context, id string, condition container.WaitCondition) (<-chan container.ContainerWaitOKBody, <-chan error)
	ImageList(ctx context.Context, opt types.ImageListOptions) ([]types.ImageSummary, error)
	ImagePull(ctx context.Context, img string, opt types.ImagePullOptions) (io.ReadCloser, error)
}

type DockerClient struct {
	cli Docker
}

func (dc *DockerClient) NewDockerClient(cli Docker) *DockerClient {
	return &DockerClient{
		cli: cli,
	}
}

func (dc *DockerClient) ContainerAttach(ctx context.Context, container string, opt types.ContainerAttachOptions) (types.HijackedResponse, error) {
	return dc.cli.ContainerAttach(ctx, container, opt)
}

func (dc *DockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *specs.Platform, name string) (container.ContainerCreateCreatedBody, error) {
	return dc.cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, platform, name)
}

func (dc *DockerClient) ContainerRemove(ctx context.Context, id string, opt types.ContainerRemoveOptions) error {
	return dc.cli.ContainerRemove(ctx, id, opt)
}

func (dc *DockerClient) ContainerStart(ctx context.Context, id string, opt types.ContainerStartOptions) error {
	return dc.cli.ContainerStart(ctx, id, opt)
}

func (dc *DockerClient) ContainerWait(ctx context.Context, id string, condition container.WaitCondition) (<-chan container.ContainerWaitOKBody, <-chan error) {
	return dc.cli.ContainerWait(ctx, id, condition)
}

func (dc *DockerClient) ImageList(ctx context.Context, opt types.ImageListOptions) ([]types.ImageSummary, error) {
	return dc.cli.ImageList(ctx, opt)
}

func (dc *DockerClient) ImagePull(ctx context.Context, img string, opt types.ImagePullOptions) (io.ReadCloser, error) {
	return dc.cli.ImagePull(ctx, img, opt)
}

//////////////////

type Librarier interface {
	DetectGPUVendors() ([]library.GPUVendor, error)
	GetAMDGPUInfo() ([]library.GPUInfo, error)
	GetNVIDIAGPUInfo() ([]library.GPUInfo, error)
}

type Library struct{}

func (l *Library) DetectGPUVendors() ([]library.GPUVendor, error) {
	return library.DetectGPUVendors()
}

func (l *Library) GetAMDGPUInfo() ([]library.GPUInfo, error) {
	return library.GetAMDGPUInfo()
}

func (l *Library) GetNVIDIAGPUInfo() ([]library.GPUInfo, error) {
	return library.GetNVIDIAGPUInfo()
}

/////////////////////

type NvidiaManager interface {
	Init() nvml.Return
	Shutdown() nvml.Return
	DeviceGetCount() (int, nvml.Return)
}

type NVML struct{}

func (n *NVML) Init() nvml.Return {
	return nvml.Init()
}

func (n *NVML) Shutdown() nvml.Return {
	return nvml.Shutdown()
}

func (n *NVML) DeviceGetCount() (int, nvml.Return) {
	return nvml.DeviceGetCount()
}

///////////////////////

type GPU interface {
	name() string
	utilizationRate() uint32
	memory() memoryInfo
	temperature() float64
	powerUsage() uint32
}

type nvidiaGPU struct {
	index int
}

type amdGPU struct {
	index int
}

type memoryInfo struct {
	used  uint64
	free  uint64
	total uint64
}
