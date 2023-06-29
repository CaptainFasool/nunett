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

func getMockDepReq(additionalFeats []string) models.DeploymentRequest {
	var depReq models.DeploymentRequest
	depReq.Params.AdditionalFeatures = additionalFeats
	return depReq
}

func TestFilterByNeededPlugins(t *testing.T) {
	// TODO: this test needs improvement but it's because
	// the original function itself needs to be more dynamic
	peerOne := models.PeerData{
		EnabledPlugins: []string{"ipfs-plugin", "other-plugin"},
		PeerID:         "peerOne",
	}

	peerTwo := models.PeerData{
		EnabledPlugins: []string{"ipfs-plugin"},
		PeerID:         "peerTwo",
	}

	peerThree := models.PeerData{
		EnabledPlugins: []string{},
		PeerID:         "peerThree",
	}

	testCases := []struct {
		name           string
		peers          []models.PeerData
		depReqAddFeats []string
		expectedResult []models.PeerData
	}{
		{
			name:           "Test with one plugin needed",
			peers:          []models.PeerData{peerOne, peerTwo, peerThree},
			depReqAddFeats: []string{"outputIPFS"},
			expectedResult: []models.PeerData{peerOne, peerTwo},
		},
		{
			name:           "Test with no plugin needed",
			peers:          []models.PeerData{peerOne, peerTwo, peerThree},
			depReqAddFeats: []string{},
			expectedResult: []models.PeerData{peerOne, peerTwo, peerThree},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			depReq := getMockDepReq(tt.depReqAddFeats)
			result := filterByNeededPlugins(tt.peers, depReq)

			if len(result) != len(tt.expectedResult) {
				t.Errorf("wanted %d peers, got %d peers", len(result), len(tt.expectedResult))
			}

			matchedPeers := 0
		OuterLoop:
			for _, peer := range result {
				for _, expectedPeer := range tt.expectedResult {
					if peer.PeerID == expectedPeer.PeerID {
						matchedPeers++
						continue OuterLoop
					}
				}
			}

			if matchedPeers != len(tt.expectedResult) {
				t.Errorf("got %v, want %v", result, tt.expectedResult)
			}
		})
	}
}
