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

func TestGPUOnboardCmdNvidia(t *testing.T) {
	type test struct {
		nvidiaCount int
		input       string
		want        []string
		notWant     []string
	}

	var (
		tests []test

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

		nvidias []library.GPUInfo

		err error
	)
	assert := assert.New(t)

	containerOut = "container installed"
	driversOut = "nvidia drivers installed"

	mockUtils = newMockUtils()

	mockExec = newMockExecuter()
	mockExec.SetCommandOutput("/bin/bash", []string{"maint-scripts/install_container_runtime"}, []byte(containerOut), nil)
	mockExec.SetCommandOutput("/bin/bash", []string{"maint-scripts/install_nvidia_drivers"}, []byte(driversOut), nil)

	mockFS = newMockFS()
	mockFS.Create("/etc/os-release")

	tests = []test{
		{nvidiaCount: 1, input: "y\ny\n", want: []string{containerOut, driversOut}},
		{nvidiaCount: 2, input: "n\ny\n", want: []string{driversOut}, notWant: []string{containerOut}},
		{nvidiaCount: 3, input: "y\nn\n", want: []string{containerOut}, notWant: []string{driversOut}},
		{nvidiaCount: 4, input: "n\nn\n", notWant: []string{containerOut, driversOut}},
	}

	for _, tc := range tests {
		mockLibrary = newMockLibrary(withNvidia(tc.nvidiaCount))

		input = strings.NewReader(tc.input)
		in = bufio.NewReader(input)
		out = new(bytes.Buffer)
		errOut = new(bytes.Buffer)

		cmd = NewGPUOnboardCmd(mockUtils, mockLibrary, mockExec, mockFS)
		cmd.SetIn(in)
		cmd.SetOut(out)
		cmd.SetErr(errOut)

		err = cmd.Execute()
		assert.NoError(err)

		// check if all gpus are printed
		nvidias = newTestNvidia(tc.nvidiaCount)
		for _, nvidia := range nvidias {
			assert.Contains(out.String(), nvidia.GPUName)
		}

		// check for scripts output
		if tc.want != nil {
			for _, output := range tc.want {
				assert.Contains(out.String(), output)
			}
		}
		if tc.notWant != nil {
			for _, output := range tc.notWant {
				assert.NotContains(out.String(), output)
			}
		}
	}
}

func TestGPUOnboardCmdAMD(t *testing.T) {
	type test struct {
		amdCount int
		input    string
		want     []string
		notWant  []string
	}

	var (
		tests []test

		mockUtils   *MockUtils
		mockLibrary *MockLibrary
		mockExec    *MockExecuter
		mockFS      *MockFS

		driversOut string

		cmd    *cobra.Command
		out    *bytes.Buffer
		errOut *bytes.Buffer
		input  *strings.Reader
		in     *bufio.Reader

		amds []library.GPUInfo
		err  error
	)
	assert := assert.New(t)

	driversOut = "AMD drivers installed"

	mockUtils = newMockUtils()

	mockExec = newMockExecuter()
	mockExec.SetCommandOutput("/bin/bash", []string{"maint-scripts/install_amd_drivers"}, []byte(driversOut), nil)

	mockFS = newMockFS()
	mockFS.Create("/etc/os-release")

	tests = []test{
		{amdCount: 1, input: "y\n", want: []string{driversOut}},
		{amdCount: 2, input: "n\n", notWant: []string{driversOut}},
	}

	for _, tc := range tests {
		mockLibrary = newMockLibrary(withAMD(tc.amdCount))

		input = strings.NewReader(tc.input)
		in = bufio.NewReader(input)
		out = new(bytes.Buffer)
		errOut = new(bytes.Buffer)

		cmd = NewGPUOnboardCmd(mockUtils, mockLibrary, mockExec, mockFS)
		cmd.SetIn(in)
		cmd.SetOut(out)
		cmd.SetErr(errOut)

		err = cmd.Execute()
		assert.NoError(err)

		// check if all gpus are printed
		amds = newTestAMD(tc.amdCount)
		for _, amd := range amds {
			assert.Contains(out.String(), amd.GPUName)
		}

		// check for scripts output
		if tc.want != nil {
			for _, output := range tc.want {
				assert.Contains(out.String(), output)
			}
		}
		if tc.notWant != nil {
			for _, output := range tc.notWant {
				assert.NotContains(out.String(), output)
			}
		}
	}
}