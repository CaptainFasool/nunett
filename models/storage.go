package models

const (
	StorageVolumeTypeBind = "bind"
)

// StorageVolume represents a prepared storage volume that can be mounted to an execution
type StorageVolume struct {
	// Type of the volume (e.g. bind)
	Type string `json:"type"`
	// Source path of the volume on the host
	Source string `json:"source"`
	// Target path of the volume in the execution
	Target string `json:"target"`
	// ReadOnly flag to mount the volume as read-only
	ReadOnly bool `json:"readonly"`
}
