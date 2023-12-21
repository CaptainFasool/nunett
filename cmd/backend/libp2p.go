package backend

import (
	"github.com/libp2p/go-libp2p/core/peer"

	"gitlab.com/nunet/device-management-service/libp2p"
)

type P2P struct{}

func (p *P2P) ClearIncomingChatRequests() error {
	return libp2p.ClearIncomingChatRequests()
}

func (p *P2P) Decode(id string) (peer.ID, error) {
	return peer.Decode(id)
}
