package machines

import (
	"testing"

	"gitlab.com/nunet/device-management-service/models"
)

func TestIsIPFSPluginNeeded(t *testing.T) {
	type testExample struct {
		requestedFuncs []string
		output         bool
	}

	tableTest := []testExample{
		{
			requestedFuncs: []string{"outputIPFS", "foo"},
			output:         true,
		},
		{
			requestedFuncs: []string{"foo", "bar"},
			output:         false,
		},
		{
			requestedFuncs: []string{},
			output:         false,
		},
	}

	for _, tt := range tableTest {
		depReq := getMockDepReq(tt.requestedFuncs)
		out := isIPFSPLuginNeeded(depReq)
		if out != tt.output {
			t.Errorf("for requested plugins %v | wanted %v, got %v", tt.requestedFuncs, tt.output, out)
		}
	}

}

func getMockDepReq(pluginIPFSFunctionalities []string) models.DeploymentRequest {
	var depReq models.DeploymentRequest
	depReq.Params.AdditionalFeatures = pluginIPFSFunctionalities
	return depReq
}
