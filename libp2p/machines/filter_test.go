package machines

import (
	"testing"

	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
)

func TestSolvePluginsNeeded(t *testing.T) {
	type testExample struct {
		requestedFuncs []string
		output         []string
	}

	tableTest := []testExample{
		{
			requestedFuncs: []string{"outputIPFS", "jobResumingIPFS", "foo"},
			output:         []string{"ipfs-plugin"},
		},
		{
			requestedFuncs: []string{"jobResumingIPFS", "foo"},
			output:         []string{"ipfs-plugin"},
		},
		{
			requestedFuncs: []string{"foo", "bar"},
			output:         []string{},
		},
		{
			requestedFuncs: []string{},
			output:         []string{},
		},
	}

	for _, tt := range tableTest {
		depReq := getMockDepReq(tt.requestedFuncs)
		out := solvePluginsNeeded(depReq)
		if !utils.AreSlicesEqual(tt.output, out) {
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
		EnabledPlugins: []string{"ipfs-plugin", "jobResumingIPFS"},
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
			name:           "Test with one plugin needed (but two functionalities for the same plugin)",
			peers:          []models.PeerData{peerOne, peerTwo, peerThree},
			depReqAddFeats: []string{"outputIPFS", "jobResumingIPFS"},
			expectedResult: []models.PeerData{peerOne, peerTwo},
		},
		{
			name:           "Test with functionality not related to any plugin",
			peers:          []models.PeerData{peerOne, peerTwo, peerThree},
			depReqAddFeats: []string{"non-valid-functionality"},
			expectedResult: []models.PeerData{peerOne, peerTwo, peerThree},
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
				t.Errorf("wanted %d peers, got %d peers", len(tt.expectedResult), len(result))
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
