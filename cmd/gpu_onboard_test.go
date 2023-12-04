package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	library "gitlab.com/nunet/device-management-service/lib"
)

// MockCmdExecutor implements Execute interface
type MockExecuter struct {
	CommandMocks map[string]Commander
}

// Execute returns a command object
func (me *MockExecuter) Execute(name string, arg ...string) Commander {
	var (
		key string
		cmd Commander
		ok  bool
	)

	key = me.makeCommandKey(name, arg)
	cmd, ok = me.CommandMocks[key]
	if !ok {
		return &MockCommand{}
	}
	return cmd
}

// Set output and error for given command based on key
func (me *MockExecuter) SetCommandOutput(name string, args []string, output []byte, err error) {
	key := me.makeCommandKey(name, args)
	me.CommandMocks[key] = &MockCommand{
		output: output,
		err:    err,
	}
}

// Concatenate name and args of command to use it as a key in CommandMocks map
func (me *MockExecuter) makeCommandKey(name string, args []string) string {
	return name + " " + strings.Join(args, " ")
}

// MockCommand implements Commander interface
type MockCommand struct {
	output []byte
	err    error
}

func (me *MockCommand) CombinedOutput() ([]byte, error) {
	if me.err != nil {
		return nil, me.err
	}
	return me.output, nil
}

func newMockExecuter() *MockExecuter {
	return &MockExecuter{
		CommandMocks: make(map[string]Commander),
	}
}

// utility
type MockUtils struct {
	wsl    bool
	wslErr error

	onboarded  bool
	onboardErr error

	response    map[string][]byte
	responseErr error
}

func (mu *MockUtils) CheckWSL() (bool, error) {
	if mu.wslErr != nil {
		return false, mu.wslErr
	} else if !mu.wsl {
		return false, nil
	} else {
		return true, nil
	}
}

func (mu *MockUtils) IsOnboarded() (bool, error) {
	if mu.onboardErr != nil {
		return false, mu.onboardErr
	} else if !mu.onboarded {
		return false, nil
	} else {
		return true, nil
	}
}

func (mu *MockUtils) ResponseBody(c *gin.Context, method, endpoint, query string, body []byte) ([]byte, error) {
	if mu.responseErr != nil {
		return nil, mu.responseErr
	}

	response, ok := mu.response[endpoint]
	if !ok {
		return nil, fmt.Errorf("response for endpoint does not exist")
	}

	return response, nil
}

func (mu *MockUtils) SetResponse(endpoint string, response []byte) {
	mu.response[endpoint] = response
}

func newMockUtils() *MockUtils {
	return &MockUtils{
		response: make(map[string][]byte),
	}
}

type MockFS struct {
	fs afero.Fs
}

func (mf *MockFS) ReadFile(name string) ([]byte, error) {
	return afero.ReadFile(mf.fs, name)
}

func (mf *MockFS) Create(name string) (afero.File, error) {
	return mf.fs.Create(name)
}

func newMockFS() *MockFS {
	return &MockFS{
		fs: afero.NewMemMapFs(),
	}
}

func Test_GPUOnboardCmdNvidia(t *testing.T) {
	var (
		mockUtils   *MockUtils
		mockLibrary *MockLibrary
		mockExec    *MockExecuter
		mockFS      *MockFS

		containerOut string
		driversOut   string

		cmd    *cobra.Command
		out    *bytes.Buffer
		errOut *bytes.Buffer
		input  *strings.Reader
		in     *bufio.Reader

		nvidiaSeries []library.GPUInfo
		err          error
	)
	assert := assert.New(t)

	mockUtils = newMockUtils()
	mockLibrary = newMockLibrary(withNvidia(1))
	mockExec = newMockExecuter()
	mockFS = newMockFS()

	mockFS.Create("/etc/os-release")

	// output for scripts
	containerOut = "container installed"
	driversOut = "nvidia drivers installed"
	mockExec.SetCommandOutput("/bin/bash", []string{"maint-scripts/install_container_runtime"}, []byte(containerOut), nil)
	mockExec.SetCommandOutput("/bin/bash", []string{"maint-scripts/install_nvidia_drivers"}, []byte(driversOut), nil)

	input = strings.NewReader("y\ny\n")
	in = bufio.NewReader(input)
	out = new(bytes.Buffer)
	errOut = new(bytes.Buffer)

	cmd = NewGPUOnboardCmd(mockUtils, mockLibrary, mockExec, mockFS)
	cmd.SetIn(in)
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err = cmd.Execute()
	assert.NoError(err)

	nvidiaSeries = newTestNvidia(1)
	for _, nvidia := range nvidiaSeries {
		assert.Contains(out.String(), nvidia.GPUName)
	}

	assert.Contains(out.String(), containerOut)
	assert.Contains(out.String(), driversOut)
}
