package libp2p

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	mrand "math/rand"
	"os"
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
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/afero"
	mafilt "github.com/whyrusleeping/multiaddr-filter"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/telemetry"
	"gitlab.com/nunet/device-management-service/utils"
)

type DMSp2p struct {
	Host  host.Host
	DHT   *dht.IpfsDHT
	PS    peerstore.Peerstore
	peers []peer.AddrInfo
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

	host, dht, err := NewHost(ctx, priv, server)
	if err != nil {
		panic(err)
	}

	p2p = *DMSp2pInit(host, dht)

	err = p2p.BootstrapNode(ctx)
	if err != nil {
		zlog.Sugar().Errorf("Bootstraping failed: %v", err)
	}

	host.SetStreamHandler(protocol.ID(PingProtocolID), PingHandler) // to be deprecated
	host.SetStreamHandler(protocol.ID("/ipfs/ping/1.0.0"), PingHandler)
	host.SetStreamHandler(protocol.ID(DepReqProtocolID), depReqStreamHandler)
	host.SetStreamHandler(protocol.ID(ChatProtocolID), chatStreamHandler)

	p2p.peers = discoverPeers(ctx, p2p.Host, p2p.DHT, utils.GetChannelName())
	go p2p.StartDiscovery(ctx, utils.GetChannelName())
	zlog.Sugar().Debugf("number of p2p.peers: %d", len(p2p.peers))

	content, err := AFS.ReadFile(fmt.Sprintf("%s/metadataV2.json", config.GetConfig().General.MetadataPath))
	if err != nil {
		zlog.Sugar().Errorf("metadata.json does not exists or not readable: %v", err)
	}
	var metadata2 models.MetadataV2
	err = json.Unmarshal(content, &metadata2)
	if err != nil {
		zlog.Sugar().Errorf("unable to parse metadata.json: %v", err)
	}

	if _, err := host.Peerstore().Get(host.ID(), "peer_info"); err != nil {
		peerInfo := models.PeerData{}
		peerInfo.PeerID = host.ID().String()
		peerInfo.AllowCardano = metadata2.AllowCardano
		peerInfo.TokenomicsAddress = metadata2.PublicKey
		peerInfo.EnabledPlugins = metadata2.Plugins
		if len(metadata2.GpuInfo) == 0 {
			peerInfo.HasGpu = false
			peerInfo.GpuInfo = metadata2.GpuInfo
		} else {
			peerInfo.GpuInfo = metadata2.GpuInfo
			peerInfo.HasGpu = true
		}

		host.Peerstore().Put(host.ID(), "peer_info", peerInfo)
	}

	err = telemetry.CalcFreeResources()
	if err != nil {
		zlog.Sugar().Errorf("Couldn't calculate the current free resources: %v", err)
	}

	// Start the DHT Update
	go UpdateKadDHT()
	go GetDHTUpdates(ctx)

	// Clean up the DHT every 5 minutes
	dhtHousekeepingTicker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-dhtHousekeepingTicker.C:
				zlog.Debug("Cleaning up DHT")
				CleanupOldPeers()
				UpdateConnections(host.Network().Conns())
			case <-stopDHTCleanup:
				zlog.Debug("Stopping DHT Cleanup")
				dhtHousekeepingTicker.Stop()
				return
			}
		}
	}()

	dhtUpdateTicker := time.NewTicker(10 * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				zlog.Debug("Context is done in DHT Update")
				return
			case <-dhtUpdateTicker.C:
				zlog.Debug("Calling UpdateKadDHT and GetDHTUpdates")

				if _, debug := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debug {
					zlog.Sugar().Info("Updating Kad DHT")
				}
				UpdateKadDHT()
				GetDHTUpdates(ctx)
			case <-stopDHTUpdate:
				zlog.Debug("Stopping DHT Update")
				dhtUpdateTicker.Stop()
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
				zlog.InfoContext(ctx,
					fmt.Sprintf("DMS Local addresses updated, reconnecting to peers. [PeerID: %s], [Current: %v], [Removed: %v]",
						host.ID().String(),
						evt.Current,
						evt.Removed))
				// Connect to saved peers
				savedConnections := GetConnections()
				for _, conn := range savedConnections {
					addr, err := multiaddr.NewMultiaddr(conn.Multiaddrs)
					if err != nil {
						zlog.Sugar().Errorf("Unable to convert multiaddr: %v", err)
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

func ShutdownNode() error {
	stopDHTUpdate <- true
	stopDHTCleanup <- true

	for _, node := range p2p.Host.Peerstore().Peers() {
		p2p.Host.Network().ClosePeer(node)
		p2p.Host.Peerstore().Put(node, "peer_info", nil)
		p2p.Host.Peerstore().ClearAddrs(node)
		p2p.Host.Peerstore().RemovePeer(node)
	}

	err := p2p.Host.Close()
	if err != nil {
		return err
	}
	err = p2p.DHT.Close()
	if err != nil {
		return err
	}

	var libp2pInfo models.Libp2pInfo
	result := db.DB.Where("id = ?", 1).Delete(&libp2pInfo)
	if result.Error != nil {
		return result.Error
	}

	p2p.Host = nil
	p2p.DHT = nil

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
		zlog.Error("filed to find public key")
		return nil, fmt.Errorf("failed to find public key")
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
		zlog.Error("failed to find private key")
		return nil, fmt.Errorf("failed to find private key")
	}
	privKey, err := crypto.UnmarshalPrivateKey(libp2pInfo.PrivateKey)
	if err != nil {
		zlog.Sugar().Errorf("Error: Unable to unmarshal Private Key: %v", err)
		return nil, err
	}
	return privKey, nil
}

func NewHost(ctx context.Context, priv crypto.PrivKey, server bool) (host.Host, *dht.IpfsDHT, error) {

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
			zlog.Sugar().Errorf("incorrectly formatted address filter in config: %s - %v", s, err)
		}
		filter.AddFilter(*f, multiaddr.ActionDeny)
	}
	var libp2pOpts []libp2p.Option
	baseOpts := []dht.Option{
		kadPrefix,
		dht.NamespacedValidator("nunet-dht", blankValidator{}),
		dht.Mode(dht.ModeServer),
	}
	libp2pOpts = append(libp2pOpts, libp2p.ListenAddrStrings(
		config.GetConfig().P2P.ListenAddress...,
	),
		libp2p.Identity(priv),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err = dht.New(ctx, h, baseOpts...)
			return idht, err
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

	host, err := libp2p.New(libp2pOpts...)

	if err != nil {
		zlog.Sugar().Errorf("Couldn't Create Host: %v", err)
		return nil, nil, err
	}

	zlog.Sugar().Infof("Self Peer Info %s -> %s", host.ID().String(), host.Addrs())

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
