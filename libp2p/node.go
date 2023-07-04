package libp2p

import (
	"bufio"
	"context"
	"encoding/json"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/multiformats/go-multiaddr"
	mafilt "github.com/whyrusleeping/multiaddr-filter"
	"gitlab.com/nunet/device-management-service/internal/config"
)

type P2P struct {
	Host  host.Host
	DHT   *dht.IpfsDHT
	peers []peer.AddrInfo
}

func NewP2P(node host.Host, dht *dht.IpfsDHT) *P2P {
	return &P2P{Host: node, DHT: dht}
}

// Sends a message to a peer
func (p2p *P2P) SendMessage(payload []byte, peerID peer.ID, protocolID protocol.ID) error {

	ctx := context.Background()
	err := p2p.sendMessage(ctx, peerID, payload, protocolID)
	if err != nil {
		return err
	}
	return nil
}

func (p2p P2P) sendMessage(ctx context.Context, peerID peer.ID, payload []byte, protocolID protocol.ID) error {

	stream, err := p2p.Host.NewStream(ctx, peerID, protocolID)
	if err != nil {
		return err
	}
	defer stream.Close()

	err = writeToStream(stream, payload)
	if err != nil {
		return err
	}
	return nil
}

// Sends a signed message to a peer
func (p2p P2P) SendSignedMessage(peerID peer.ID, payload []byte, protocolID protocol.ID) error {
	ctx := context.Background()

	signature, err := signData(p2p.Host.Peerstore().PrivKey(p2p.Host.ID()), payload)
	if err != nil {
		return err
	}
	msg := struct {
		Data      []byte
		Signature []byte
	}{
		Data:      payload,
		Signature: signature,
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	err = p2p.sendMessage(ctx, peerID, msgBytes, protocolID)
	if err != nil {
		return err
	}
	return nil
}

func writeToStream(s network.Stream, data []byte) error {
	w := bufio.NewWriter(s)
	n, err := w.Write(data)
	if n != len(data) {
		return err
	}
	if err != nil {
		return err
	}
	err = w.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (p2p *P2P) NewHost(ctx context.Context, priv crypto.PrivKey, server bool) error {

	connmgr, err := connmgr.NewConnManager(
		100, // Lowwater
		400, // HighWater,
		connmgr.WithGracePeriod(time.Minute),
	)

	if err != nil {
		zlog.Sugar().Errorf("Error Creating Connection Manager: %v", err)
		return err
	}

	filter := multiaddr.NewFilters()
	for _, s := range defaultServerFilters {
		f, err := mafilt.NewMask(s)
		if err != nil {
			zlog.Sugar().Errorf("incorrectly formatted address filter in config: %s", s)
		}
		filter.AddFilter(*f, multiaddr.ActionDeny)
	}
	var libp2pOpts []libp2p.Option
	baseOpts := []dht.Option{
		kadPrefix,
		dht.NamespacedValidator("nunet-dht", blankValidator{
			P2p: p2p,
		}),
		dht.Mode(dht.ModeServer),
	}
	libp2pOpts = append(libp2pOpts, libp2p.ListenAddrStrings(
		config.GetConfig().P2P.ListenAddress...,
	),
		libp2p.Identity(priv),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			p2p.DHT, err = dht.New(ctx, h, baseOpts...)
			return p2p.DHT, err
		}),
		libp2p.DefaultPeerstore,
		libp2p.EnableNATService(),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(noise.ID, noise.New),
		libp2p.DefaultTransports,
		libp2p.EnableNATService(),
		libp2p.ConnectionManager(connmgr),
		libp2p.EnableRelay(),
		libp2p.EnableHolePunching(),
		libp2p.EnableRelayService(
			relay.WithResources(
				relay.Resources{
					MaxReservations:        256,
					MaxCircuits:            32,
					BufferSize:             4096,
					MaxReservationsPerPeer: 8,
					MaxReservationsPerIP:   16,
				},
			),
			relay.WithLimit(&relay.RelayLimit{
				Duration: 5 * time.Minute,
				Data:     1 << 21, // 2 MiB
			}),
		),
		libp2p.EnableAutoRelayWithPeerSource(
			func(ctx context.Context, num int) <-chan peer.AddrInfo {
				r := make(chan peer.AddrInfo)
				go func() {
					defer close(r)
					for i := 0; i < num; i++ {
						select {
						case p := <-newPeer:
							select {
							case r <- p:
							case <-ctx.Done():
								return
							}
						case <-ctx.Done():
							return
						}
					}
				}()
				return r
			},
			autorelay.WithBootDelay(time.Minute),
			autorelay.WithBackoff(30*time.Second),
			autorelay.WithMinCandidates(2),
			autorelay.WithMaxCandidates(3),
			autorelay.WithNumRelays(2),
		),
	)

	if server {
		libp2pOpts = append(libp2pOpts, libp2p.AddrsFactory(makeAddrsFactory([]string{}, []string{}, defaultServerFilters)))
		libp2pOpts = append(libp2pOpts, libp2p.ConnectionGater((*filtersConnectionGater)(filter)))
	} else {
		libp2pOpts = append(libp2pOpts, libp2p.NATPortMap())
	}

	p2p.Host, err = libp2p.New(libp2pOpts...)

	if err != nil {
		zlog.Sugar().Errorf("Couldn't Create Host: %v", err)
		return err
	}

	zlog.Sugar().Infof("Self Peer Info %s -> %s\n", p2p.Host.ID(), p2p.Host.Addrs())

	return nil
}

func (p2p *P2P) SetStreamHandler(protocol protocol.ID, handler network.StreamHandler) {
	p2p.Host.SetStreamHandler(protocol, handler)
}
