package libp2p

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p/core/network"
	"golang.org/x/sync/errgroup"
)

// TODO: var vpns map[vpnID]*VPN
var vpn *VPN

var packetPool = sync.Pool{
	New: func() interface{} {
		// Adjust the size according to your typical packet size
		return make([]byte, 1600) // Slightly larger than the MTU size to accommodate various packet sizes
	},
}

// CreateAndInviteHandler godoc
// @Summary      Creates a vpn inviting a list of peers
// @Description  Given a list of peers, vpn addresses will be assigned
// and the host will create a vpn and invite them to join
// @Tags         file
// @Accept       json
// @Produce      json
// @Success      200
// @Router       /network/vpn/create-and-invite [post]
func CreateAndInviteHandler(c *gin.Context) {
	ctx, cancel := context.WithCancel(context.Background())
	var params CreateAndInviteParams

	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		cancel()
		return
	}
	zlog.Sugar().Debugf("Received params: %+v\n", params)

	var err error
	vpn, err = NewVPN(ctx, cancel, params.PeersIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	vpn.invitePeersToVPN(params.PeersIDs)
	c.JSON(200, gin.H{"message": "VPN was created successfully"})
}

// DownHandler godoc
// @Summary      Removes a TUN interface
// @Description  Removes a TUN interface named dms-tun
// @Tags         file
// @Success      200
// @Router       /network/vpn/down [post]
func DownHandler(c *gin.Context) {
	// TODO: maybe instead of Down based on in-memory var, rely on
	// a given tun interface name
	if vpn != nil {
		vpn.cancel()
	}
	err := vpn.tunDev.SetDownAndDelete()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "dms-tun interface was deleted successfully"})
}

// vpnStreamHandler handles all incoming packets for streams
// following the protocol VPNProtocolID. It handles two types of messages:
// 1. VPN creation where there is no vpn yet; 2. VPN internal messaging
func vpnStreamHandler(stream network.Stream) {
	defer stream.Close()

	// Use a single goroutine per stream to reduce context switching overhead
	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if vpn == nil {
					zlog.Sugar().Debug("No vpn found, checking if it's a vpn invite message")

					ctx := context.Background()
					ctx, cancel := context.WithCancel(ctx)

					// First read the message size
					var length uint16
					err := binary.Read(stream, binary.LittleEndian, &length)
					if err != nil {
						stream.Close()
						zlog.Sugar().Errorf("Couldn't read message size from stream. Error: %w", err)
					}

					// Allocate enough space for the full message
					vpnMsgJson := make([]byte, length)

					// Read the message
					_, err = stream.Read(vpnMsgJson)
					if err != nil {
						stream.Close()
						zlog.Sugar().Errorf("Couldn't read message from stream. Error: %w", err)
					}

					// Unmarshal the JSON
					vpnMsg := vpnMessage{}
					err = json.Unmarshal(vpnMsgJson, &vpnMsg)
					if err != nil {
						stream.Close()
						zlog.Sugar().Errorf("Couldn't unmarshal vpn message. Error: %w", err)
					}

					if vpnMsg.MsgType == msgVPNCreationInvite {
						zlog.Sugar().Debugf("Received vpn creation invite")
						zlog.Sugar().Debugf("Routing table: %s", vpnMsg.Msg)
						routingTable := &vpnRouter{}
						err := json.Unmarshal([]byte(vpnMsg.Msg), &routingTable)
						if err != nil {
							zlog.Sugar().Errorf("Unable to decode vpn message. Closing stream. Error: %v", err)
							stream.Reset()
						}

						vpn, err = JoinVPN(ctx, cancel, *routingTable)
						if err != nil {
							zlog.Sugar().Errorf("Unable to join vpn. Closing stream. Error: %v", err)
							stream.Reset()
							cancel()
						}
						zlog.Sugar().Info("Successfully joined vpn")
					}
				}

				// If tunneling device was not iniated yet, just close the stream
				if vpn.tunDev == nil {
					zlog.Sugar().Errorf("tunDev was not iniated")
					stream.Reset()
				}
				// If the remote node ID isn't in the list of known nodes don't respond.
				if _, ok := vpn.routingTable[stream.Conn().RemotePeer()]; !ok {
					zlog.Sugar().Errorf("Peer %s not on the routing table",
						stream.Conn().RemotePeer().Pretty())
					stream.Reset()
				}
				packet := packetPool.Get().([]byte) // Retrieve a packet buffer from the pool
				defer packetPool.Put(&packet)       // Make sure to put the packet back into the pool

				// Read packet size
				packetSize := make([]byte, 2)
				if _, err := stream.Read(packetSize); err != nil {
					return err // Stream read error
				}

				// Decode the incoming packet's size from binary.
				size := binary.LittleEndian.Uint16(packetSize) // Decode packet size

				// Read the packet based on the decoded size
				if _, err := stream.Read(packet[:size]); err != nil {
					return err // Packet read error
				}

				// Process the packet, e.g., write to the TUN device or handle VPN logic
				if err := processPacket(size, packet[:size], stream); err != nil {
					return err // Packet processing error
				}

			}
		}
	})

	// Wait for the goroutine to finish and check for errors
	if err := g.Wait(); err != nil {
		// Log or handle the error as needed
		zlog.Sugar().Errorf("Error in vpnStreamHandler: %v", err)
	}
}

// processPacket placeholder function remains unchanged, implement as needed
func processPacket(size uint16, packet []byte, stream network.Stream) error {
	// Your packet processing logic here (e.g., writing to the TUN device)
	// Read in the packet until completion.
	var plen uint16 = 0
	for plen < size {
		tmp, err := stream.Read(packet[plen:size])
		plen += uint16(tmp)
		if err != nil {
			zlog.Sugar().Errorf("[VPN] error reading packet's data from stream: %v", err)
			stream.Close()
			return err
		}
	}
	zlog.Sugar().Debug("Writing packet to Tunneling interface")
	vpn.tunDev.Iface.Write(packet[:size])

	return nil
}

// JoinHandler godoc
// @Summary      Joins a vpn
// @Description  Joins a vpn given a list of peers
// @Tags         file
// @Accept       json
// @Produce      json
// @Success      200
// @Router       /network/vpn/join [post]
// func JoinHandler(c *gin.Context) {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	var params JoinVPNParams
//
// 	if err := c.BindJSON(&params); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
//
// 	fmt.Printf("Received params: %+v\n", params)
//
// 	// Create TUN interface
// 	tunDev, err := tun.New(
// 		"dms-tun",
// 		tun.Address(params.Addr),
// 		tun.MTU(1420),
// 	)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
//
// 	// Necessary data structures for the following
// 	// peersID []peerID
// 	// revLookup map[peerID]peerAddress
// 	// peerTable map[peerAddress]peerID
// 	// activeStreams map[peerAddress]stream
// 	var peersID []peer.ID
// 	revLookup := make(map[string]string, len(params.VPN.Peers))
// 	peerTable := make(map[string]peer.ID)
// 	for _, p := range params.VPN.Peers {
// 		pID, err := peer.Decode(p.ID)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		// Necessary data structures for some functions
// 		revLookup[p.ID] = p.Addr
// 		peersID = append(peersID, pID)
// 		peerTable[p.Addr] = pID
// 	}
//
// 	h := GetP2P().Host
// 	idht := GetP2P().DHT
//
// 	// Find and create connection with peers within VPN
// 	go dialPeersContinuously(ctx, h, idht, peersID)
//
// 	// TODO: check if peer support protocol creating and closing stream
//
// 	// Activate TUN interface to be ready to receive/send packets
// 	// tun.New() just created, it didn't make active.
// 	err = tunDev.Up()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
//
// 	// The following is responsible for SENDING packet to other peers
// 	// tunDev.Iface.Read() is reading all the packets coming to the TUN
// 	// interface, so if I do `ping 10.0.0.1`, the Iface.Read() will read
// 	// the packets that I'm trying to send to someone. As you can see,
// 	// `dst` will get the destination address before writing to the libp2p stream
// 	activeStreams := make(map[string]network.Stream)
// 	var packet = make([]byte, 1420)
// 	vpn = &VPN{
// 		ctx,
// 		cancel,
// 		params.VPN,
// 		tunDev,
// 		revLookup,
// 		activeStreams,
// 	}
// 	go func() {
// 		for {
// 			select {
// 			case <-vpn.ctx.Done():
// 				zlog.Sugar().Error("Closing all vpn streams if any")
// 				for dst, stream := range activeStreams {
// 					stream.Close()
// 					delete(activeStreams, dst)
// 				}
// 				return
// 			default:
// 				// ping 10.0.0.1
// 				// Read in a packet from the tun interface.
// 				plen, err := tunDev.Iface.Read(packet)
// 				if err != nil {
// 					zlog.Sugar().Errorf(
// 						"Error reading packet from TUN interface: %v", err)
// 					continue
// 				}
//
// 				// TODO: check if there is anything at all within the packet
//
// 				// Decode the packet's destination address
// 				dst := net.IPv4(packet[16], packet[17], packet[18], packet[19]).String()
// 				zlog.Sugar().Debugf("Send packet to destination peer: %v", dst)
//
// 				// Check if we already have an open connection to the destination peer.
// 				stream, ok := activeStreams[dst]
// 				if ok {
// 					// Write out the packet's length to the libp2p stream to ensure
// 					// we know the full size of the packet at the other end.
// 					err = binary.Write(stream, binary.LittleEndian, uint16(plen))
// 					if err == nil {
// 						// Write the packet out to the libp2p stream.
// 						// If everyting succeeds continue on to the next packet.
// 						_, err = stream.Write(packet[:plen])
// 						if err == nil {
// 							zlog.Sugar().Debugf(
// 								"Successfully sent packet to: %v", dst)
// 							continue
// 						}
// 					}
// 					// If we encounter an error when writing to a stream we should
// 					// close that stream and delete it from the active stream map.
// 					zlog.Sugar().Errorf(
// 						"Error writing to libp2p stream from tunneling: %v", err)
// 					stream.Close()
// 					delete(activeStreams, dst)
// 				}
//
// 				// Check if the destination of the packet is a known peer to
// 				// the interface.
// 				if peer, ok := peerTable[dst]; ok {
// 					zlog.Sugar().Debugf(
// 						"Didn't have an active stream with peer %v, creating one", dst)
// 					stream, err = h.NewStream(ctx, peer, VPNProtocolID)
// 					if err != nil {
// 						zlog.Sugar().Errorf(
// 							"Error creating stream with peer: %v", dst)
// 						continue
// 					}
// 					// Write packet length
// 					err = binary.Write(stream, binary.LittleEndian, uint16(plen))
// 					if err != nil {
// 						zlog.Sugar().Error("Error writing packet size")
// 						stream.Close()
// 						continue
// 					}
// 					// Write the packet
// 					_, err = stream.Write(packet[:plen])
// 					if err != nil {
// 						zlog.Sugar().Errorf(
// 							"Error writing to libp2p stream from tunneling: %v", err)
// 						stream.Close()
// 						continue
// 					}
// 					zlog.Sugar().Debugf(
// 						"Successfully sent packet to: %v", dst)
//
// 					// If all succeeds when writing the packet to the stream
// 					// we should reuse this stream by adding it active streams map.
// 					activeStreams[dst] = stream
// 				}
// 			}
// 		}
// 	}()
// 	c.JSON(200, gin.H{"message": "Successfully started vpn"})
// }
