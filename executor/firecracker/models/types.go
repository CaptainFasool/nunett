package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bacalhau-project/bacalhau/pkg/lib/validate"
	"github.com/bacalhau-project/bacalhau/pkg/models"
)

// Constants defining engine type and specific configuration keys for Firecracker VMs.
const (
	EngineFirecracker                  = "firecracker"
	EngineKeyKernelImageFirecracker    = "KernelImage"
	EngineKeyKernelArgsFirecracker     = "KernelArgs"
	EngineKeyRootFileSystemFirecracker = "RootFileSystem"
	EngineKeyMMDSMessage               = "MMDSMessage"
)

// EngineSpec defines the structure for engine specifications with JSON serialization options.
type EngineSpec struct {
	// KernelImage is the path to the kernel image file.
	KernelImage    string `json:"KernelImage,omitempty"`
	// InitrdPath is the path to the initial ramdisk file.
	Initrd         string `json:"InitrdPath,omitempty"`
	// KernelArgs is the kernel command line arguments.
	KernelArgs     string `json:"KernelArgs,omitempty"`
	// RootFileSystem is the path to the root file system.
	RootFileSystem string `json:"RootFileSystem,omitempty"`
	// MMDSMessage is the MMDS message to be sent to the Firecracker VM.
	MMDSMessage    string `json:"MMDSMessage,omitempty"`
}

// Validate checks the integrity of EngineSpec fields, ensuring mandatory fields are not empty.
func (c EngineSpec) Validate() error {
	if validate.IsBlank(c.RootFileSystem) {
		return errors.New("invalid firecracker engine params: RootFileSystem cannot be empty")
	}
	if validate.IsBlank(c.KernelImage) {
		return errors.New("invalid firecracker engine params: KernelImage cannot be empty")
	}
	return nil
}

// DecodeSpec is a function that decodes a SpecConfig object into a Firecracker EngineSpec object.
// It returns an EngineSpec object and an error if the decoding process fails.
func DecodeSpec(spec *models.SpecConfig) (EngineSpec, error) {
	if !spec.IsType(EngineFirecracker) {
		//nolint:goconst
		return EngineSpec{}, errors.New(
			"invalid firecracker engine type. expected " + EngineFirecracker + ", but received: " + spec.Type,
		)
	}
	inputParams := spec.Params
	if inputParams == nil {
		return EngineSpec{}, errors.New("invalid firecracker engine params. cannot be nil")
	}

	paramsBytes, err := json.Marshal(inputParams)
	if err != nil {
		return EngineSpec{}, fmt.Errorf("failed to encode firecracker engine specs. %w", err)
	}

	var c *EngineSpec
	err = json.Unmarshal(paramsBytes, &c)
	if err != nil {
		return EngineSpec{}, fmt.Errorf("failed to decode firecracker engine specs. %w", err)
	}
	return *c, c.Validate()
}

// FirecrackerEngineBuilder is a struct that is used for constructing an EngineSpec object
// specifically for Firecracker engines using the Builder pattern.
// It embeds an EngineBuilder object for handling the common builder methods.
type FirecrackerEngineBuilder struct {
	eb *models.SpecConfig
}

// NewFirecrackerEngineBuilder function initializes a new FirecrackerEngineBuilder instance.
// It sets the engine type to EngineFirecracker.String() and kernel image path as per the input argument.
func NewFirecrackerEngineBuilder(rootFileSystem string) *FirecrackerEngineBuilder {
	eb := models.NewSpecConfig(EngineFirecracker)
	eb.WithParam(EngineKeyRootFileSystemFirecracker, rootFileSystem)
	return &FirecrackerEngineBuilder{eb: eb}
}

// WithRootFileSystem is a builder method that sets the Firecracker engine root file system.
// It returns the FirecrackerEngineBuilder for further chaining of builder methods.
func (b *FirecrackerEngineBuilder) WithRootFileSystem(e string) *FirecrackerEngineBuilder {
	b.eb.WithParam(EngineKeyRootFileSystemFirecracker, e)
	return b
}

// WithKernelImage is a builder method that sets the Firecracker engine kernel image.
// It returns the FirecrackerEngineBuilder for further chaining of builder methods.
func (b *FirecrackerEngineBuilder) WithKernelImage(e string) *FirecrackerEngineBuilder {
	b.eb.WithParam(EngineKeyKernelImageFirecracker, e)
	return b
}

// WithKernelArgs is a builder method that sets the Firecracker engine kernel arguments.
// It returns the FirecrackerEngineBuilder for further chaining of builder methods.
func (b *FirecrackerEngineBuilder) WithKernelArgs(e string) *FirecrackerEngineBuilder {
	b.eb.WithParam(EngineKeyKernelArgsFirecracker, e)
	return b
}

// WithMMDSMessage is a builder method that sets the Firecracker engine MMDS message.
// It returns the FirecrackerEngineBuilder for further chaining of builder methods.
func (b *FirecrackerEngineBuilder) WithMMDSMessage(e string) *FirecrackerEngineBuilder {
	b.eb.WithParam(EngineKeyMMDSMessage, e)
	return b
}

// Build method constructs the final SpecConfig object by calling the embedded EngineBuilder's Build method.
func (b *FirecrackerEngineBuilder) Build() *models.SpecConfig {
	return b.eb
}
