package libp2p

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"
)

// Bootstrap using a list.
func (l *Libp2p) Bootstrap(ctx context.Context, bootstrapPeers []multiaddr.Multiaddr) error {
	if err := l.DHT.Bootstrap(ctx); err != nil {
		return fmt.Errorf("failed to prepare this node for bootstraping: %w", err)
	}

	// bootstrap all nodes at the same time.
	if len(bootstrapPeers) > 0 {
		var wg sync.WaitGroup
		for _, addr := range bootstrapPeers {
			wg.Add(1)
			go func(peerAddr multiaddr.Multiaddr) {
				defer wg.Done()
				addrInfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
				if err != nil {
					zlog.Sugar().Errorf("failed to convert multi addr to addr info %v - %v", peerAddr, err)
					return
				}
				if err := l.Host.Connect(ctx, *addrInfo); err != nil {
					zlog.Sugar().Errorf("failed to connect to bootstrap node %s - %v", addrInfo.ID.String(), err)
				} else {
					zlog.Sugar().Infof("connected to Bootstrap Node %s", addrInfo.ID.String())
				}
			}(addr)
		}
		wg.Wait()
	}

	return nil
}

type dhtValidator struct {
	PS              peerstore.Peerstore
	customNamespace string
}

func (d dhtValidator) Validate(key string, value []byte) error {
	// Check if the key has the correct namespace
	if !strings.HasPrefix(key, d.customNamespace) {
		return errors.New("invalid key namespace")
	}

	// TODO: FIX this validation.
	return nil

	// components := strings.Split(key, "/")
	// key = components[len(components)-1]

	// var dhtUpdate models.KadDHTMachineUpdate

	// err := json.Unmarshal(value, &dhtUpdate)
	// if err != nil {
	// 	// zlog.Sugar().Errorf("Error unmarshalling value: %v", err)
	// 	return err
	// }

	// // Extract data and signature fields
	// data := dhtUpdate.Data
	// var peerInfo models.PeerData
	// err = json.Unmarshal(dhtUpdate.Data, &peerInfo)
	// if err != nil {
	// 	// zlog.Sugar().Errorf("Error unmarshalling value: %v", err)
	// 	return err
	// }

	// signature := dhtUpdate.Signature
	// remotePeerID, err := peer.Decode(key)
	// if err != nil {
	// 	// zlog.Sugar().Errorf("Error decoding peerID: %v", err)
	// 	return errors.New("error decoding peerID")
	// }

	// // Get the public key of the remote peer from the peerstore
	// remotePeerPublicKey := d.PS.PubKey(remotePeerID)
	// if remotePeerPublicKey == nil {
	// 	return errors.New("public key for remote peer not found in peerstore")
	// }
	// verify, err := remotePeerPublicKey.Verify(data, signature)
	// if err != nil {
	// 	// zlog.Sugar().Errorf("Error verifying signature: %v", err)
	// 	return err
	// }
	// if !verify {
	// 	// zlog.Sugar().Info("Invalid signature")
	// 	return errors.New("invalid signature")
	// }

	// if len(value) == 0 {
	// 	return errors.New("value cannot be empty")
	// }
}
func (dhtValidator) Select(_ string, _ [][]byte) (int, error) { return 0, nil }
