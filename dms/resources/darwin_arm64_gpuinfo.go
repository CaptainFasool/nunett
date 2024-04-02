//go:build darwin && (arm64 || amd64)

package resources

type GPUInfo struct {
	GPUName     string
	TotalMemory uint64
	UsedMemory  uint64
	FreeMemory  uint64
	Vendor      GPUVendor
}

func GetGPUInfo() ([][]GPUInfo, error) {
	var gpu_infos [][]GPUInfo
	return gpu_infos, nil
}
