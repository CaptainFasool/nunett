package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/models"
)

func TestNewNetwork(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		config *models.NetworkConfig
		expErr string
	}{
		"no config given": {
			expErr: "network configuration is nil",
		},
		"invalid network": {
			config: &models.NetworkConfig{Type: "invalid-type"},
			expErr: "unsupported network type: invalid-type",
		},
		"nats network": {
			config: &models.NetworkConfig{Type: models.NATSNetwork},
			expErr: "not implemented",
		},
		"libp2p network": {
			config: &models.NetworkConfig{Type: models.Libp2pNetwork},
		},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			network, err := NewNetwork(tt.config)
			if tt.expErr != "" {
				assert.Nil(t, network)
				assert.EqualError(t, err, tt.expErr)
			} else {
				assert.NotNil(t, network)
			}
		})
	}
}
