package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
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

func Test_GPUOnboardCmdBoth(t *testing.T) {
	type test struct {
		nvidiaCount int
		amdCount    int
		input       string
		want        []string
		notWant     []string
	}

	var (
		containerOut     string
		nvidiaDriversOut string
		amdDriversOut    string

		mockFS      *MockFS
		mockUtils   *MockUtils
		mockExec    *MockExecuter
		mockLibrary *MockLibrary

		input  *strings.Reader
		in     *bufio.Reader
		out    *bytes.Buffer
		errOut *bytes.Buffer

		cmd *cobra.Command
		err error

		amds    []library.GPUInfo
		nvidias []library.GPUInfo

		tests []test
	)
	assert := assert.New(t)

	containerOut = "container installed"
	nvidiaDriversOut = "Nvidia drivers installed"
	amdDriversOut = "AMD drivers installed"

	mockFS = newMockFS()
	mockFS.Create("/etc/os-release")

	mockUtils = newMockUtils()
	mockExec = newMockExecuter()

	mockExec.SetCommandOutput("/bin/bash", []string{"maint-scripts/install_container_runtime"}, []byte(containerOut), nil)
	mockExec.SetCommandOutput("/bin/bash", []string{"maint-scripts/install_nvidia_drivers"}, []byte(nvidiaDriversOut), nil)
	mockExec.SetCommandOutput("/bin/bash", []string{"maint-scripts/install_amd_drivers"}, []byte(amdDriversOut), nil)

	tests = []test{
		{nvidiaCount: 1, amdCount: 1, input: "y\ny\ny\n", want: []string{containerOut, nvidiaDriversOut, amdDriversOut}},
		{nvidiaCount: 2, amdCount: 3, input: "n\nn\nn\n", notWant: []string{containerOut, nvidiaDriversOut, amdDriversOut}},
		{nvidiaCount: 2, amdCount: 2, input: "y\ny\nn\n", want: []string{containerOut, nvidiaDriversOut}, notWant: []string{amdDriversOut}},
		{nvidiaCount: 3, amdCount: 2, input: "y\nn\nn\n", want: []string{containerOut}, notWant: []string{nvidiaDriversOut, amdDriversOut}},
		{nvidiaCount: 1, amdCount: 1, input: "y\nn\ny\n", want: []string{containerOut, amdDriversOut}, notWant: []string{nvidiaDriversOut}},
		{nvidiaCount: 3, amdCount: 3, input: "n\nn\ny\n", want: []string{amdDriversOut}, notWant: []string{containerOut, nvidiaDriversOut}},
		{nvidiaCount: 1, amdCount: 3, input: "n\ny\nn\n", want: []string{nvidiaDriversOut}, notWant: []string{containerOut, amdDriversOut}},
		{nvidiaCount: 2, amdCount: 2, input: "n\ny\ny\n", want: []string{nvidiaDriversOut, amdDriversOut}, notWant: []string{containerOut}},
	}

	for _, tc := range tests {
		mockLibrary = newMockLibrary(withAMD(tc.amdCount), withNvidia(tc.nvidiaCount))

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

func Test_GPUOnboardCmdNoGPUs(t *testing.T) {
	var (
		mockUtils   *MockUtils
		mockLibrary *MockLibrary
		mockExec    *MockExecuter
		mockFS      *MockFS

		cmd    *cobra.Command
		out    *bytes.Buffer
		errOut *bytes.Buffer

		err error
	)
	assert := assert.New(t)

	mockLibrary = newMockLibrary()
	mockUtils = newMockUtils()
	mockExec = newMockExecuter()
	mockFS = newMockFS()
	mockFS.Create(osFile)

	out = new(bytes.Buffer)
	errOut = new(bytes.Buffer)

	cmd = NewGPUOnboardCmd(mockUtils, mockLibrary, mockExec, mockFS)
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err = cmd.Execute()
	if err == nil {
		t.Fatalf("expected error while executing, got %v", err)
	}
	assert.ErrorContains(err, "no AMD or NVIDIA GPU(s) detected...")
}

func Test_GPUOnboardCmdWSL(t *testing.T) {
	type test struct {
		nvidiaCount int
		input       string
		want        string
		notWant     string
	}

	var (
		mockUtils   *MockUtils
		mockLibrary *MockLibrary
		mockExec    *MockExecuter
		mockFS      *MockFS

		containerOut string

		cmd    *cobra.Command
		in     *bytes.Buffer
		out    *bytes.Buffer
		errOut *bytes.Buffer

		err   error
		tests map[string]test
	)
	assert := assert.New(t)
	containerOut = "container installed"

	mockUtils = newMockUtils()
	mockUtils.wsl = true

	mockExec = newMockExecuter()
	mockExec.SetCommandOutput("/bin/bash", []string{containerPath}, []byte(containerOut), nil)

	mockFS = newMockFS()
	mockFS.Create("/etc/os-release")

	tests = map[string]test{
		"wsl with nvidia confirmed":     {nvidiaCount: 2, input: "y\n", want: containerOut},
		"wsl with nvidia not confirmed": {nvidiaCount: 1, input: "n\n", notWant: containerOut},
	}

	for _, tc := range tests {
		mockLibrary = newMockLibrary(withNvidia(tc.nvidiaCount))

		in = bytes.NewBufferString(tc.input)
		out = new(bytes.Buffer)
		errOut = new(bytes.Buffer)

		cmd = NewGPUOnboardCmd(mockUtils, mockLibrary, mockExec, mockFS)
		cmd.SetIn(in)
		cmd.SetOut(out)
		cmd.SetErr(errOut)

		err = cmd.Execute()
		assert.NoError(err)

		if tc.notWant != "" {
			assert.NotContains(out.String(), tc.notWant)
		}
		if tc.want != "" {
			assert.Contains(out.String(), tc.want)
		}
	}
}

func Test_GpuOnboardCmdWSLError(t *testing.T) {
	type test struct {
		amdCount int
		want     string
	}

	var (
		mockUtils   *MockUtils
		mockLibrary *MockLibrary
		mockExec    *MockExecuter
		mockFS      *MockFS

		cmd    *cobra.Command
		out    *bytes.Buffer
		errOut *bytes.Buffer

		err   error
		tests map[string]test
	)

	assert := assert.New(t)

	mockExec = newMockExecuter()
	mockUtils = newMockUtils()
	mockUtils.wsl = true

	mockFS = newMockFS()
	mockFS.Create("/etc/os-release")

	tests = map[string]test{
		"wsl without nvidia with amd": {amdCount: 2, want: "no NVIDIA GPU(s) detected..."},
	}

	for _, tc := range tests {
		mockLibrary = newMockLibrary(withAMD(tc.amdCount))

		out = new(bytes.Buffer)
		errOut = new(bytes.Buffer)

		cmd = NewGPUOnboardCmd(mockUtils, mockLibrary, mockExec, mockFS)
		cmd.SetOut(out)
		cmd.SetErr(errOut)

		err = cmd.Execute()
		assert.ErrorContains(err, tc.want)
	}
}

func Test_GPUOnboardCmdMining(t *testing.T) {
	type test struct {
		amdCount    int
		nvidiaCount int
		input       string
		want        string
	}

	var (
		mockUtils   *MockUtils
		mockLibrary *MockLibrary
		mockExec    *MockExecuter
		mockFS      *MockFS

		cmd    *cobra.Command
		in     *bytes.Buffer
		out    *bytes.Buffer
		errOut *bytes.Buffer

		err   error
		tests map[string]test
	)

	assert := assert.New(t)

	mockFS = newMockFS()
	mockExec = newMockExecuter()
	mockUtils = newMockUtils()

	skipMsg := "You are likely running a Mining OS. Skipping driver installation..."
	miningOSes := []string{"Hive", "Rave", "PiMP", "Minerstat", "SimpleMining", "NH", "Miner", "SM", "MMP"}

	tests = map[string]test{
		"mining OS confirmed": {amdCount: 1, nvidiaCount: 2, input: "y\n", want: skipMsg},
		"mining OS declined":  {amdCount: 2, nvidiaCount: 3, input: "n\n", want: skipMsg},
	}

	for _, tc := range tests {
		// write random string from mining OSes slice to file
		mockFS.Create(osFile)
		random := rand.Intn(len(miningOSes))
		pick := miningOSes[random]
		afero.WriteFile(mockFS.fs, osFile, []byte(pick), 0666)

		mockLibrary = newMockLibrary(withAMD(tc.amdCount), withNvidia(tc.nvidiaCount))

		in = bytes.NewBufferString(tc.input)
		out = new(bytes.Buffer)
		errOut = new(bytes.Buffer)

		cmd = NewGPUOnboardCmd(mockUtils, mockLibrary, mockExec, mockFS)
		cmd.SetIn(in)
		cmd.SetOut(out)
		cmd.SetErr(errOut)

		err = cmd.Execute()
		assert.NoError(err)
		assert.Contains(out.String(), tc.want)
	}
}
