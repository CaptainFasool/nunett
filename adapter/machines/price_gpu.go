package machines

const (
	lowNtxPerMinute      = 0.00008
	moderateNtxPerMinute = 0.00016
	highNtxPerMinute     = 0.00032

	wattPerMinNtx = 0.00016
)

type ResourceUsage struct {
	Complexity string // Low/Moderate/High
	CPU        float64
	RAM        float64
	VRAM       float64
	Power      float64
	Time       float64
}

func CalculateStaticNtxGpu(resourceUsage *ResourceUsage) (estimatedNtx float64) {

	// Calculate estimated NTX for each job complexity
	if resourceUsage.Complexity == "Low" {
		estimatedNtx = (lowNtxPerMinute * resourceUsage.Time) + (wattPerMinNtx * resourceUsage.Power * resourceUsage.Time)
	}
	if resourceUsage.Complexity == "Moderate" {
		estimatedNtx = (moderateNtxPerMinute * resourceUsage.Time) + (wattPerMinNtx * resourceUsage.Power * resourceUsage.Time)
	}
	if resourceUsage.Complexity == "High" {
		estimatedNtx = (highNtxPerMinute * resourceUsage.Time) + (wattPerMinNtx * resourceUsage.Power * resourceUsage.Time)
	}

	return estimatedNtx
}

// CalculateDynamicNtxGpu works on the statistics/resource-usage reported by the job which could be different than the actual requested.
func CalculateDynamicNtxGpu(requestedResource, usedResource *ResourceUsage) (actualNtx float64) {
	// XXX: This function is a stub
	return 0
}

// Test CalculateStaticNtxGpu
// func main() {
// 	lowResource := &ResourceUsage{
// 		Complexity: "Low",
// 		CPU:        500,
// 		RAM:        2000,
// 		VRAM:       2000,
// 		Power:      170,
// 		Time:       10,
// 	}

// 	moderateResourceResource := &ResourceUsage{
// 		Complexity: "Moderate",
// 		CPU:        1500,
// 		RAM:        8000,
// 		VRAM:       8000,
// 		Power:      220,
// 		Time:       10,
// 	}

// 	highResource := &ResourceUsage{
// 		Complexity: "High",
// 		CPU:        2500,
// 		RAM:        16000,
// 		VRAM:       24000,
// 		Power:      350,
// 		Time:       10,
// 	}

// 	fmt.Println("10 mins of low resource price: ", CalculateStaticNtxGpu(lowResource))
// 	fmt.Println("10 mins of moderate resource price: ", CalculateStaticNtxGpu(moderateResourceResource))
// 	fmt.Println("10 mins of high resource price: ", CalculateStaticNtxGpu(highResource))
// }
