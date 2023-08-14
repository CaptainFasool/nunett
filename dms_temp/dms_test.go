package dms_temp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/libp2p"
)

func TestCheckOnboarding(t *testing.T) {
	dms := DMS{
		// Initialize or mock the dependencies here (DB, P2P, AFS, zlog, etc.)
	}

	// Test the CheckOnboarding function
	err := dms.CheckOnboarding()

	// Assert that there is no error returned
	assert.NoError(t, err, "Expected no error from CheckOnboarding")
}

func TestRunNode(t *testing.T) {
	// Create a mock instance of DMS with necessary dependencies
	dms := DMS{
		// Initialize or mock the dependencies here (DB, P2P, AFS, zlog, etc.)
	}

	privKey, _, _ := libp2p.GenerateKey(0)

	// Test the RunNode function
	dms.RunNode(privKey, true)

	assert.NotNil(t, dms.P2P)
	assert.NotNil(t, dms.DB)
}
