package libp2p

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	mrand "math/rand"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/models"
)

func GenerateKey(seed int64) (crypto.PrivKey, crypto.PubKey, error) {
	var r io.Reader
	if seed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(seed))
	}
	priv, pub, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, nil, err
	}
	return priv, pub, nil

}

func SaveKey(priv crypto.PrivKey, pub crypto.PubKey) error {
	var libp2pInfo models.Libp2pInfo
	libp2pInfo.ID = 1
	libp2pInfo.PrivateKey, _ = crypto.MarshalPrivateKey(priv)
	libp2pInfo.PublicKey, _ = crypto.MarshalPublicKey(pub)

	if res := db.DB.Find(&libp2pInfo); res.RowsAffected == 0 {
		result := db.DB.Create(&libp2pInfo)
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}

func NewHost(ctx context.Context, port int, priv crypto.PrivKey) (host.Host, *dht.IpfsDHT, error) {

	var idht *dht.IpfsDHT

	connmgr, err := connmgr.NewConnManager(
		100, // Lowwater
		400, // HighWater,
		connmgr.WithGracePeriod(time.Minute),
	)

	if err != nil {
		return nil, nil, err
	}

	host, err := libp2p.New(
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port),      // regular tcp connections
			fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", port), // a UDP endpoint for the QUIC transport
		),
		libp2p.Identity(priv),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err = dht.New(ctx, h)
			return idht, err
		}),
		libp2p.NATPortMap(),
		libp2p.EnableNATService(),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(noise.ID, noise.New),
		libp2p.DefaultTransports,
		libp2p.EnableNATService(),
		libp2p.ConnectionManager(connmgr),
		libp2p.EnableRelay(),
		libp2p.EnableRelayService(),
		// libp2p.EnableAutoRelay(autorelay.WithStaticRelays([]peer.AddrInfo{*relayPeer})),
		libp2p.EnableAutoRelay(
			autorelay.WithBootDelay(0),
			autorelay.WithPeerSource(
				func(ctx context.Context, numPeers int) <-chan peer.AddrInfo {
					r := make(chan peer.AddrInfo)
					go func() {
						defer close(r)
						for i := 0; i < numPeers; i++ {
							select {
							case p := <-make(chan peer.AddrInfo):
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
				0,
			),
			autorelay.WithMaxCandidates(3),
			autorelay.WithNumRelays(1),
			autorelay.WithBootDelay(0)),
	)

	if err != nil {
		return nil, nil, err
	}

	zlog.Sugar().Infof("Self Peer Info %s ------ %s\n", host.ID(), host.Addrs())

	return host, idht, nil
}
