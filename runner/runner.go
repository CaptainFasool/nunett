package runner

import (
	"time"
)

// JobConfig defines the configuration necessary to start a job.
// This is a generic structure that can be extended or interpreted differently
// by various runner implementations.
type JobConfig struct {
	ID           string
	RunnerType   string            // Indicates the type of runner (Docker, VM, Firecracker, etc.)
	Source       string            // Image or VM identifier
	Command      []string          // Command and arguments to execute
	Environment  map[string]string // Environment variables
	Resources    ResourcesConfig
	Network      NetworkConfig
	Storage      StorageConfig
	OtherConfigs map[string]interface{} // Additional configs, flexible for various runner types
}

// ResourcesConfig specifies the resource allocation for the job.
type ResourcesConfig struct {
	CPU    float64
	Memory int64
	GPU    int
}

type NetworkConfig struct {
	PortBindings []PortBinding
	NetworkMode  string
	Labels       map[string]string
	SecurityOpts []string
}

type PortBinding struct {
	HostPort   string // Host port, e.g., "8080" or "0.0.0.0:8080"
	RunnerPort string // Container port, e.g., "80/tcp"
	Protocol   string // Protocol (tcp, udp, etc.)
}

type StorageConfig struct {
	Bindings []StorageBinding
	ReadOnly bool // Global read-only setting for all storages
	Labels   map[string]string
}

type StorageBinding struct {
	Source      string // Source path or identifier
	Destination string // Destination path in the job environment
	Mode        string // Mount mode (e.g., "rw" for read-write, "ro" for read-only)
	SizeLimit   int64  // Optional size limit for the storage
}

// JobStatus provides detailed information about a job's status.
type JobStatus struct {
	ID             string
	Status         JobStatusCode
	ExitCode       int
	StartTime      time.Time
	CompletionTime time.Time
	ErrorMessage   string
}

// JobStatusCode represents the status of a job.
type JobStatusCode string

const (
	JobStatusPending   JobStatusCode = "PENDING"
	JobStatusRunning   JobStatusCode = "RUNNING"
	JobStatusCompleted JobStatusCode = "COMPLETED"
	JobStatusFailed    JobStatusCode = "FAILED"
)

// JobStatusStream defines the interface for streaming job status updates.
type JobStatusStream interface {
	// Next returns the next JobStatus update. It should block until there is an update or an error.
	Next() (JobStatus, error)
}

// Runner defines the interface for any type of job runner.
type Runner interface {
	Capabilities() map[string]interface{}
	Start(job JobConfig) (JobStatus, error)
	Stop(jobID string) error
	Pause(jobID string) error
	Resume(jobID string) error
	HealthCheck() error
	Status(jobID string) (JobStatus, error)
	StatusStream(jobID string) (JobStatusStream, error)
}
