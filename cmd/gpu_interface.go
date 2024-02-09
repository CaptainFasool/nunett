package cmd

type GPU interface {
	name() string
	utilizationRate() uint32
	memory() memoryInfo
	temperature() float64
	powerUsage() uint32
}

type nvidiaGPU struct {
	index int
}

type amdGPU struct {
	index int
}

type memoryInfo struct {
	used  uint64
	free  uint64
	total uint64
}
