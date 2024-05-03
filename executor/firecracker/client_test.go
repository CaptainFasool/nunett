package firecracker_test

import (
	"context"
	"testing"
	"time"

	firecrackerSdk "github.com/firecracker-microvm/firecracker-go-sdk"
	firecrackerModels "github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gitlab.com/nunet/device-management-service/executor/firecracker"
)

const (
	defaultSocketPath = "/tmp/firecracker.sock"
	rootDrivePath     = "testdata/rootfs.ext4"
	kernelImagePath   = "testdata/vmlinux.bin"
)

// ClientTestSuite is the test suite for the Firecracker client.
type ClientTestSuite struct {
	suite.Suite
	client *firecracker.Client
}

// SetupTest sets up the test suite by initializing a new Firecracker client.
func (s *ClientTestSuite) SetupTest() {
	c, err := firecracker.NewFirecrackerClient()
	require.NoError(s.T(), err)
	s.client = c
}

// TestClientTestSuite runs the test suite for the Firecracker client.
func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

// createTestVM is a helper method to create a VM for testing.
func (s *ClientTestSuite) createTestVM(socketPath string) *firecrackerSdk.Machine {
	
	cfg := firecrackerSdk.Config{
		SocketPath: socketPath,
		KernelImagePath: kernelImagePath,
		Drives: []firecrackerModels.Drive{
			{
				DriveID: firecrackerSdk.String("1"),
				PathOnHost:   firecrackerSdk.String(rootDrivePath),
				IsRootDevice: firecrackerSdk.Bool(true),
				IsReadOnly:   firecrackerSdk.Bool(true),
			},
		},
		MachineCfg: firecrackerModels.MachineConfiguration{
			VcpuCount:  firecrackerSdk.Int64(1),
			MemSizeMib: firecrackerSdk.Int64(1024),
		},
	}
	m, err := s.client.CreateVM(context.Background(), cfg)
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		_ = s.client.DestroyVM(context.Background(), m, 10*time.Second)
	})
	go func(m *firecrackerSdk.Machine) {
		time.Sleep(3 * time.Second)
		m.StopVMM()
	}(m)
	return m
}

// TestIsInstalled tests the IsInstalled method of the Firecracker client.
func (s *ClientTestSuite) TestIsInstalled() {
	assert.True(s.T(), s.client.IsInstalled())
}

// TestCreateVM tests the CreateVM method of the Firecracker client.
func (s *ClientTestSuite) TestCreateVM() {
	m := s.createTestVM(defaultSocketPath)
	require.NotNil(s.T(), m)
}

// TestStartVM tests the StartVM method of the Firecracker client.
func (s *ClientTestSuite) TestStartVM() {
	m := s.createTestVM(defaultSocketPath)
	require.NotNil(s.T(), m)

	err := s.client.StartVM(context.Background(), m)
	require.NoError(s.T(), err)
}

// TestFindVM tests the FindVM method of the Firecracker client.
func (s *ClientTestSuite) TestFindVM() {
	m := s.createTestVM(defaultSocketPath)
	require.NotNil(s.T(), m)

	err := s.client.StartVM(context.Background(), m)
	require.NoError(s.T(), err)

	vm, err := s.client.FindVM(context.Background(), m.Cfg.SocketPath)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), vm)
}
