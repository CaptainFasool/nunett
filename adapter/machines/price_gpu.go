package machines

import "gitlab.com/nunet/device-management-service/models"

const (
	lowNtxPerMinute      = 0.00008
	moderateNtxPerMinute = 0.00016
	highNtxPerMinute     = 0.00032

	wattPerMinNtx = 0.00016
)

func CalculateStaticNtxGpu(depReq models.DeploymentRequest) (estimatedNtx float64) {
	resourceUsage := depReq.Constraints

	// Calculate estimated NTX for each job complexity
	if resourceUsage.Complexity == "Low" {
		estimatedNtx = (lowNtxPerMinute * float64(resourceUsage.Time)) + (wattPerMinNtx * float64(resourceUsage.Power) * float64(resourceUsage.Time))
	}
	if resourceUsage.Complexity == "Moderate" {
		estimatedNtx = (moderateNtxPerMinute * float64(resourceUsage.Time)) + (wattPerMinNtx * float64(resourceUsage.Power) * float64(resourceUsage.Time))
	}
	if resourceUsage.Complexity == "High" {
		estimatedNtx = (highNtxPerMinute * float64(resourceUsage.Time)) + (wattPerMinNtx * float64(resourceUsage.Power) * float64(resourceUsage.Time))
	}

	return estimatedNtx
}

// CalculateDynamicNtxGpu works on the statistics/resource-usage reported by the job which could be different than the actual requested.
func CalculateDynamicNtxGpu(requestedResource, usedResource interface{}) (actualNtx float64) {
	// XXX: This function is a stub
	return 0
}
