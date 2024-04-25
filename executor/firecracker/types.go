package firecracker

import (
	"encoding/json"
	"fmt"

	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils/validate"
)

const (
	EngineKeyKernelImage    = "kernel_image"
	EngineKeyKernelArgs     = "kernel_args"
	EngineKeyRootFileSystem = "root_file_system"
	EngineKeyMMDSMessage    = "mmds_message"
)

// EngineSpec contains necessary parameters to execute a firecracker job.
type EngineSpec struct {
	// KernelImage is the path to the kernel image file.
	KernelImage string `json:"kernel_image,omitempty"`
	// InitrdPath is the path to the initial ramdisk file.
	Initrd string `json:"initrd_path,omitempty"`
	// KernelArgs is the kernel command line arguments.
	KernelArgs string `json:"kernel_args,omitempty"`
	// RootFileSystem is the path to the root file system.
	RootFileSystem string `json:"root_file_system,omitempty"`
	// MMDSMessage is the MMDS message to be sent to the Firecracker VM.
	MMDSMessage string `json:"mmds_message,omitempty"`
}

// Validate checks if the engine spec is valid
func (c EngineSpec) Validate() error {
	if validate.IsBlank(c.RootFileSystem) {
		return fmt.Errorf("invalid firecracker engine params: root_file_system cannot be empty")
	}
	if validate.IsBlank(c.KernelImage) {
		return fmt.Errorf("invalid firecracker engine params: kernel_image cannot be empty")
	}
	return nil
}

// DecodeSpec decodes a spec config into a firecracker engine spec
// It converts the params into a firecracker EngineSpec struct and validates it
func DecodeSpec(spec *models.SpecConfig) (EngineSpec, error) {
	if !spec.IsType(models.ExecutorTypeFirecracker) {
		return EngineSpec{}, fmt.Errorf(
			"invalid firecracker engine type. expected %s, but recieved: %s",
			models.ExecutorTypeFirecracker,
			spec.Type,
		)
	}

	inputParams := spec.Params
	if inputParams == nil {
		return EngineSpec{}, fmt.Errorf("invalid firecracker engine params: params cannot be nil")
	}

	paramBytes, err := json.Marshal(inputParams)
	if err != nil {
		return EngineSpec{}, fmt.Errorf("failed to encode firecracker engine params: %w", err)
	}

	var firecrackerSpec *EngineSpec
	err = json.Unmarshal(paramBytes, &firecrackerSpec)
	if err != nil {
		return EngineSpec{}, fmt.Errorf("failed to decode firecracker engine params: %w", err)
	}

	return *firecrackerSpec, firecrackerSpec.Validate()
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
	eb := models.NewSpecConfig(models.ExecutorTypeFirecracker)
	eb.WithParam(EngineKeyRootFileSystem, rootFileSystem)
	return &FirecrackerEngineBuilder{eb: eb}
}

// WithRootFileSystem is a builder method that sets the Firecracker engine root file system.
// It returns the FirecrackerEngineBuilder for further chaining of builder methods.
func (b *FirecrackerEngineBuilder) WithRootFileSystem(e string) *FirecrackerEngineBuilder {
	b.eb.WithParam(EngineKeyRootFileSystem, e)
	return b
}

// WithKernelImage is a builder method that sets the Firecracker engine kernel image.
// It returns the FirecrackerEngineBuilder for further chaining of builder methods.
func (b *FirecrackerEngineBuilder) WithKernelImage(e string) *FirecrackerEngineBuilder {
	b.eb.WithParam(EngineKeyKernelImage, e)
	return b
}

// WithKernelArgs is a builder method that sets the Firecracker engine kernel arguments.
// It returns the FirecrackerEngineBuilder for further chaining of builder methods.
func (b *FirecrackerEngineBuilder) WithKernelArgs(e string) *FirecrackerEngineBuilder {
	b.eb.WithParam(EngineKeyKernelArgs, e)
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
