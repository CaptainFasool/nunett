package dms_temp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/afero"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/internal/logger"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
)

type DMS struct {
	DB   db.DMSGormDB
	P2P  libp2p.P2P
	ctx  context.Context
	AFS  *afero.Afero
	zlog otelzap.Logger
}

func NewDMS() DMS {
	context := context.Background()
	priv, _, _ := libp2p.GenerateKey(0)
	var p2p libp2p.P2P
	p2p.NewHost(context, priv, true)
	zlog := logger.OtelZapLogger("libp2p")
	db := db.DMSGormDB{}
	db.ConnectDatabase(fmt.Sprintf("%s/nunet.db", config.GetConfig().General.MetadataPath))
	return DMS{
		DB:   db,
		P2P:  p2p,
		ctx:  context,
		zlog: zlog,
	}
}

func (dms *DMS) CheckOnboarding() error {
	// Check DB, Metadata and other prerequisites for safe running of node
	// Remove the private key check from this function
	// and move it to a separate function
	var libp2pInfo models.Libp2pInfo
	err := dms.DB.WhereFind(libp2pInfo, "id", "1")
	if err != nil {
		return err
	}
	if libp2pInfo.PrivateKey != nil {
		priv, err := crypto.UnmarshalPrivateKey(libp2pInfo.PrivateKey)
		if err != nil {
			panic(err)
		}
		dms.RunNode(priv, libp2pInfo.ServerMode)
	}

	return nil
}

func (dms *DMS) RunNode(priv crypto.PrivKey, server bool) {
	ctx := context.Background()

	err := dms.P2P.BootstrapNode(ctx)
	if err != nil {
		dms.zlog.Sugar().Errorf("Bootstraping failed: %s\n", err)
	}

	dms.P2P.Host.SetStreamHandler(protocol.ID(libp2p.PingProtocolID), libp2p.PingHandler) // to be deprecated
	dms.P2P.Host.SetStreamHandler(protocol.ID("/ipfs/ping/1.0.0"), libp2p.PingHandler)
	dms.P2P.Host.SetStreamHandler(protocol.ID(libp2p.DepReqProtocolID), libp2p.DepReqStreamHandler)
	dms.P2P.Host.SetStreamHandler(protocol.ID(libp2p.ChatProtocolID), libp2p.ChatStreamHandler)

	go dms.P2P.StartDiscovery(ctx, utils.GetChannelName())
	dms.zlog.Sugar().Debugf("number of p2p.peers: %d", len(dms.P2P.Peers))

	content, err := dms.AFS.ReadFile(fmt.Sprintf("%s/metadataV2.json", config.GetConfig().General.MetadataPath))
	if err != nil {
		dms.zlog.Sugar().Errorf("metadata.json does not exists or not readable: %s\n", err)
	}
	var metadata2 models.MetadataV2
	err = json.Unmarshal(content, &metadata2)
	if err != nil {
		dms.zlog.Sugar().Errorf("unable to parse metadata.json: %s\n", err)
	}

	if _, err := dms.P2P.Host.Peerstore().Get(dms.P2P.Host.ID(), "peer_info"); err != nil {
		peerInfo := models.PeerData{}
		peerInfo.PeerID = dms.P2P.Host.ID().String()
		peerInfo.AllowCardano = metadata2.AllowCardano
		peerInfo.TokenomicsAddress = metadata2.PublicKey
		if len(metadata2.GpuInfo) == 0 {
			peerInfo.HasGpu = false
			peerInfo.GpuInfo = metadata2.GpuInfo
		} else {
			peerInfo.GpuInfo = metadata2.GpuInfo
			peerInfo.HasGpu = true
		}

		dms.P2P.Host.Peerstore().Put(dms.P2P.Host.ID(), "peer_info", peerInfo)
	}

	// Start the DHT Update
	go libp2p.UpdateKadDHT()
	go libp2p.GetDHTUpdates(ctx)

	// Clean up the DHT every 5 minutes
	dhtHousekeepingTicker := time.NewTicker(5 * time.Minute)
	quit2 := make(chan struct{})
	go func() {
		for {
			select {
			case <-dhtHousekeepingTicker.C:
				libp2p.CleanupOldPeers()
				libp2p.UpdateConnections(dms.P2P.Host.Network().Conns())
			case <-quit2:
				dhtHousekeepingTicker.Stop()
				return
			}
		}
	}()

	dhtUpdateTicker := time.NewTicker(10 * time.Minute)
	quit3 := make(chan struct{})
	go func() {
		for {
			select {
			case <-dhtUpdateTicker.C:
				if _, debug := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debug {
					dms.zlog.Sugar().Info("Updating Kad DHT")
				}
				libp2p.UpdateKadDHT()
				libp2p.GetDHTUpdates(ctx)
			case <-quit3:
				dhtUpdateTicker.Stop()
				return
			}
		}
	}()

	// Reconnect to peers on address change
	events, _ := dms.P2P.Host.EventBus().Subscribe(new(event.EvtLocalAddressesUpdated))
	go func() {
		for evt := range events.Out() {
			var connections []network.Conn = dms.P2P.Host.Network().Conns()
			evt := evt.(event.EvtLocalAddressesUpdated)

			if evt.Diffs {
				dms.zlog.InfoContext(ctx,
					fmt.Sprintf("DMS Local addresses updated, reconnecting to peers. [PeerID: %s], [Current: %v], [Removed: %v]",
						dms.P2P.Host.ID().String(),
						evt.Current,
						evt.Removed))
				// Connect to saved peers
				savedConnections := libp2p.GetConnections()
				for _, conn := range savedConnections {
					addr, err := multiaddr.NewMultiaddr(conn.Multiaddrs)
					if err != nil {
						dms.zlog.Sugar().Error("Unable to convert multiaddr: ", err)
					}
					if err := dms.P2P.Host.Connect(ctx, peer.AddrInfo{
						ID:    peer.ID(conn.PeerID),
						Addrs: []multiaddr.Multiaddr{addr},
					}); err != nil {
						continue
					}
				}
				for _, conn := range connections {

					// Attempt to reconnect
					if err := dms.P2P.Host.Connect(ctx, peer.AddrInfo{
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
