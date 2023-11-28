package cmd

import "github.com/NVIDIA/go-nvml/pkg/nvml"

const (
	sensorNVML nvml.TemperatureSensors = iota
)

type nvidiaGPU struct {
	device nvml.Device
}

func (n *nvidiaGPU) Name() string {
	name, ret := n.device.GetName()
	if ret != nvml.SUCCESS {
		return ""
	}
	return name
}

func (n *nvidiaGPU) UtilizationRate() uint32 {
	utilization, ret := n.device.GetUtilizationRates()
	if ret != nvml.SUCCESS {
		return 0
	}
	return utilization.Gpu
}

func (n *nvidiaGPU) Memory() memoryInfo {
	memory, ret := n.device.GetMemoryInfo()
	if ret != nvml.SUCCESS {
		return memoryInfo{}
	}
	return memoryInfo{
		used:  memory.Used,
		free:  memory.Free,
		total: memory.Total,
	}
}

func (n *nvidiaGPU) Temperature() float64 {
	temp, ret := n.device.GetTemperature(sensorNVML)
	if ret != nvml.SUCCESS {
		return 0
	}
	return float64(temp)
}

func (n *nvidiaGPU) PowerUsage() uint32 {
	power, ret := n.device.GetPowerUsage()
	if ret != nvml.SUCCESS {
		return 0
	}
	return power
}
