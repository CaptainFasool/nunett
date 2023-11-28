package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
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
	return &ExecCommand{
		cmd: exec.Command(name, arg...),
	}
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

type NVMLManager interface {
	Init() error
	Shutdown() error
}

type NVML struct{}

func (n *NVML) Init() error {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return fmt.Errorf("exited with code %d: %s", ret, nvml.ErrorString(ret))
	}
	return nil
}

func (n *NVML) Shutdown() error {
	ret := nvml.Shutdown()
	if ret != nvml.SUCCESS {
		return fmt.Errorf("exited with code %d: %s", ret, nvml.ErrorString(ret))
	}
	return nil
}

///////////////////////

type GPUController interface {
	CountDevices() (int, error)
	GetDeviceByIndex(i int) (GPU, error)
}

type NvidiaController struct{}

func (nm *NvidiaController) CountDevices() (int, error) {
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return 0, fmt.Errorf("exit code %d: %s", ret, nvml.ErrorString(ret))
	}

	return count, nil
}

func (nm *NvidiaController) GetDeviceByIndex(i int) (GPU, error) {
	var (
		device nvml.Device
		ret    nvml.Return
	)

	device, ret = nvml.DeviceGetHandleByIndex(i)
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("exit code %d: %s", ret, nvml.ErrorString(ret))
	}

	return &nvidiaGPU{device: device}, nil
}

type AMDController struct {
	executer Executer
}

func (ac *AMDController) CountDevices() (int, error) {
	var (
		cmd Commander
		out []byte

		pattern string
		re      *regexp.Regexp
		matches [][]string
		ids     []string

		err error
	)

	cmd = ac.executer.Execute("rocm-smi --showid")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}

	pattern = `GPU\[(\d+)\]`
	re = regexp.MustCompile(pattern)

	matches = re.FindAllStringSubmatch(string(out), -1)
	for _, match := range matches {
		ids = append(ids, match[1])
	}

	return len(ids), nil
}

func (ac *AMDController) GetDeviceByIndex(i int) (GPU, error) {
	return &amdGPU{index: i + 1}, nil
}

type GPU interface {
	Name() string
	UtilizationRate() uint32
	Memory() memoryInfo
	Temperature() float64
	PowerUsage() uint32
}

type memoryInfo struct {
	used  uint64
	free  uint64
	total uint64
}
