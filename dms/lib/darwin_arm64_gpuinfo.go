//go:build darwin && (arm64 || amd64)

package library

type GPUInfo struct {
	GPUName     string
	TotalMemory uint64
	UsedMemory  uint64
	FreeMemory  uint64
	Vendor      GPUVendor
}

func (v GPUVendor) String() string {
	switch v {
	case NVIDIA:
		return "NVIDIA"
	case AMD:
		return "AMD"
	default:
		return "Unknown"
	}
}

func GetGPUInfo() ([][]GPUInfo, error) {
	var gpu_infos [][]GPUInfo
	return gpu_infos, nil
}
