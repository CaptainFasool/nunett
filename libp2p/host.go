package libp2p

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	mrand "math/rand"
	"os"
	"strconv"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/afero"
	mafilt "github.com/whyrusleeping/multiaddr-filter"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/models"
)

type DMSp2p struct {
	Host host.Host
	DHT  *dht.IpfsDHT
	PS   peerstore.Peerstore
}

func DMSp2pInit(node host.Host, dht *dht.IpfsDHT) *DMSp2p {
	return &DMSp2p{Host: node, DHT: dht}
}

var p2p DMSp2p
var FS afero.Fs = afero.NewOsFs()
var AFS *afero.Afero = &afero.Afero{Fs: FS}

func GetP2P() DMSp2p {
	return p2p
}

func CheckOnboarding() {
	// Checks for saved metadata and create a new host
	var libp2pInfo models.Libp2pInfo
	result := db.DB.Where("id = ?", 1).Find(&libp2pInfo)
	if result.Error != nil {
		panic(result.Error)
	}
	if libp2pInfo.PrivateKey != nil {
		// Recreate private key
		priv, err := crypto.UnmarshalPrivateKey(libp2pInfo.PrivateKey)
		if err != nil {
			panic(err)
		}
		RunNode(priv, libp2pInfo.ServerMode)
	}
}

func RunNode(priv crypto.PrivKey, server bool) {
	ctx := context.Background()

	host, dht, err := NewHost(ctx, 9000, priv, server)
	if err != nil {
		panic(err)
	}

	p2p = *DMSp2pInit(host, dht)

	err = p2p.BootstrapNode(ctx)
	if err != nil {
		zlog.Sugar().Errorf("Bootstraping failed: %s\n", err)
	}

	host.SetStreamHandler(protocol.ID(PingProtocolID), PingHandler)
	host.SetStreamHandler(protocol.ID(DHTProtocolID), DhtUpdateHandler)
	host.SetStreamHandler(protocol.ID(DepReqProtocolID), depReqStreamHandler)
	host.SetStreamHandler(protocol.ID(ChatProtocolID), chatStreamHandler)

	go p2p.StartDiscovery(ctx, "nunet")

	content, err := AFS.ReadFile("/etc/nunet/metadataV2.json")
	if err != nil {
		zlog.Sugar().Errorf("metadata.json does not exists or not readable: %s\n", err)
	}
	var metadata2 models.MetadataV2
	err = json.Unmarshal(content, &metadata2)
	if err != nil {
		zlog.Sugar().Errorf("unable to parse metadata.json: %s\n", err)
	}

	if _, err := host.Peerstore().Get(host.ID(), "peer_info"); err != nil {
		peerInfo := models.PeerData{}
		peerInfo.PeerID = host.ID().String()
		peerInfo.AllowCardano = metadata2.AllowCardano
		peerInfo.TokenomicsAddress = metadata2.PublicKey
		if len(metadata2.GpuInfo) == 0 {
			peerInfo.HasGpu = false
			peerInfo.GpuInfo = metadata2.GpuInfo
		} else {
			peerInfo.GpuInfo = metadata2.GpuInfo
			peerInfo.HasGpu = true
		}

		host.Peerstore().Put(host.ID(), "peer_info", peerInfo)
	}

	// Broadcast DHT updates every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	if val, debugMode := os.LookupEnv("NUNET_DHT_UPDATE_INTERVAL"); debugMode {
		interval, err := strconv.Atoi(val)
		if err != nil {
			zlog.Sugar().DPanicf("invalid value for $NUNET_DHT_UPDATE_INTERVAL - %v", val)
		}
		ticker = time.NewTicker(time.Duration(interval) * time.Second)
		zlog.Sugar().Infof("setting DHT update interval to %v seconds", interval)
	}
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				UpdateDHT()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	// Clean up the DHT every 5 minutes
	ticker2 := time.NewTicker(5 * time.Minute)
	quit2 := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker2.C:
				CleanupOldPeers()
				UpdateConnections(host.Network().Conns())
			case <-quit2:
				ticker2.Stop()
				return
			}
		}
	}()

	// Reconnect to peers on address change
	events, _ := host.EventBus().Subscribe(new(event.EvtLocalAddressesUpdated))
	go func() {
		for evt := range events.Out() {
			var connections []network.Conn = host.Network().Conns()
			evt := evt.(event.EvtLocalAddressesUpdated)

			if evt.Diffs {
				// Connect to saved peers
				savedConnections := GetConnections()
				for _, conn := range savedConnections {
					addr, err := multiaddr.NewMultiaddr(conn.Multiaddrs)
					if err != nil {
						zlog.Sugar().Error("Unable to convert multiaddr: ", err)
					}
					if err := host.Connect(ctx, peer.AddrInfo{
						ID:    peer.ID(conn.PeerID),
						Addrs: []multiaddr.Multiaddr{addr},
					}); err != nil {
						continue
					}
				}
				for _, conn := range connections {

					// Attempt to reconnect
					if err := host.Connect(ctx, peer.AddrInfo{
						ID:    conn.RemotePeer(),
						Addrs: []multiaddr.Multiaddr{conn.RemoteMultiaddr()},
					}); err != nil {
						continue
					}
				}

			}
		}
	}()

}

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

func SaveNodeInfo(priv crypto.PrivKey, pub crypto.PubKey, serverMode bool) error {
	var libp2pInfo models.Libp2pInfo
	libp2pInfo.ID = 1
	libp2pInfo.PrivateKey, _ = crypto.MarshalPrivateKey(priv)
	libp2pInfo.PublicKey, _ = crypto.MarshalPublicKey(pub)
	libp2pInfo.ServerMode = serverMode

	if res := db.DB.Find(&libp2pInfo); res.RowsAffected == 0 {
		result := db.DB.Create(&libp2pInfo)
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}

func GetPublicKey() (crypto.PubKey, error) {
	var libp2pInfo models.Libp2pInfo
	result := db.DB.Where("id = ?", 1).Find(&libp2pInfo)
	if result.Error != nil {
		zlog.Sugar().Errorf("Error: Unable to Read from database: %v", result.Error)
		return nil, result.Error
	}
	if libp2pInfo.PublicKey == nil {
		zlog.Sugar().Errorf("Error: No Public Key Found")
		return nil, fmt.Errorf("No Public Key Found")
	}
	pubKey, err := crypto.UnmarshalPublicKey(libp2pInfo.PublicKey)
	if err != nil {
		zlog.Sugar().Errorf("Error: Unable to unmarshal Public Key: %v", err)
		return nil, err
	}
	return pubKey, nil
}

func GetPrivateKey() (crypto.PrivKey, error) {
	var libp2pInfo models.Libp2pInfo
	result := db.DB.Where("id = ?", 1).Find(&libp2pInfo)
	if result.Error != nil {
		zlog.Sugar().Errorf("Error: Unable to Read from database: %v", result.Error)
		return nil, result.Error
	}
	if libp2pInfo.PrivateKey == nil {
		zlog.Sugar().Errorf("Error: No Private Key Found")
		return nil, fmt.Errorf("No Private Key Found")
	}
	privKey, err := crypto.UnmarshalPrivateKey(libp2pInfo.PrivateKey)
	if err != nil {
		zlog.Sugar().Errorf("Error: Unable to unmarshal Private Key: %v", err)
		return nil, err
	}
	return privKey, nil
}

func NewHost(ctx context.Context, port int, priv crypto.PrivKey, server bool) (host.Host, *dht.IpfsDHT, error) {

	var idht *dht.IpfsDHT

	connmgr, err := connmgr.NewConnManager(
		100, // Lowwater
		400, // HighWater,
		connmgr.WithGracePeriod(time.Minute),
	)

	if err != nil {
		zlog.Sugar().Errorf("Error Creating Connection Manager: %v", err)
		return nil, nil, err
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
	libp2pOpts = append(libp2pOpts, libp2p.ListenAddrStrings(
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port),      // regular tcp connections
		fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", port), // a UDP endpoint for the QUIC transport
	),
		libp2p.Identity(priv),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err = dht.New(ctx, h)
			return idht, err
		}),
		libp2p.DefaultPeerstore,
		libp2p.EnableNATService(),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(noise.ID, noise.New),
		libp2p.DefaultTransports,
		libp2p.EnableNATService(),
		libp2p.ConnectionGater((*filtersConnectionGater)(filter)),
		libp2p.ConnectionManager(connmgr),
		libp2p.EnableRelay(),

		libp2p.EnableRelayService(),
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
			autorelay.WithBootDelay(0)))

	if server {
		libp2pOpts = append(libp2pOpts, libp2p.AddrsFactory(makeAddrsFactory([]string{}, []string{}, defaultServerFilters)))
	} else {
		libp2pOpts = append(libp2pOpts, libp2p.NATPortMap())
	}

	host, err := libp2p.New(libp2pOpts...)

	if err != nil {
		zlog.Sugar().Errorf("Couldn't Create Host: %v", err)
		return nil, nil, err
	}

	zlog.Sugar().Infof("Self Peer Info %s -> %s\n", host.ID(), host.Addrs())

	return host, idht, nil
}

func makeAddrsFactory(announce []string, appendAnnouce []string, noAnnounce []string) func([]multiaddr.Multiaddr) []multiaddr.Multiaddr {
	var err error                     // To assign to the slice in the for loop
	existing := make(map[string]bool) // To avoid duplicates

	annAddrs := make([]multiaddr.Multiaddr, len(announce))
	for i, addr := range announce {
		annAddrs[i], err = multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil
		}
		existing[addr] = true
	}

	var appendAnnAddrs []multiaddr.Multiaddr
	for _, addr := range appendAnnouce {
		if existing[addr] {
			// skip AppendAnnounce that is on the Announce list already
			continue
		}
		appendAddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil
		}
		appendAnnAddrs = append(appendAnnAddrs, appendAddr)
	}

	filters := multiaddr.NewFilters()
	noAnnAddrs := map[string]bool{}
	for _, addr := range noAnnounce {
		f, err := mafilt.NewMask(addr)
		if err == nil {
			filters.AddFilter(*f, multiaddr.ActionDeny)
			continue
		}
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil
		}
		noAnnAddrs[string(maddr.Bytes())] = true
	}

	return func(allAddrs []multiaddr.Multiaddr) []multiaddr.Multiaddr {
		var addrs []multiaddr.Multiaddr
		if len(annAddrs) > 0 {
			addrs = annAddrs
		} else {
			addrs = allAddrs
		}
		addrs = append(addrs, appendAnnAddrs...)

		var out []multiaddr.Multiaddr
		for _, maddr := range addrs {
			// check for exact matches
			ok := noAnnAddrs[string(maddr.Bytes())]
			// check for /ipcidr matches
			if !ok && !filters.AddrBlocked(maddr) {
				out = append(out, maddr)
			}
		}
		return out
	}
}
