package libp2p

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"gitlab.com/nunet/device-management-service/telemetry/logger"
)

const (
// Custom namespace for DHT protocol with version number
// customNamespace = "/nunet-dht-1/"
)

var (
	zlog otelzap.Logger
	// gettingDHTUpdate     = false
	// doneGettingDHTUpdate = make(chan bool) // XXX dirty hack to wait for DHT update to finish - should be removed

	newPeer = make(chan peer.AddrInfo)
)

func init() {
	zlog = logger.OtelZapLogger("network.libp2p")
}
