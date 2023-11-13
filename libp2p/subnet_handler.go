package libp2p

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p/core/network"
)

// TODO: var vpns map[vpnID]*VPN
var vpn *VPN

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
		return
	}
	zlog.Sugar().Debugf("Received params: %+v\n", params)

	vpn, err := NewVPN(ctx, cancel, params.PeersIDs)
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
	if vpn != nil {
		// TODO: return a response in case of failure/success when entering invited vpn
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		zlog.Sugar().Errorf("No vpn found, checking if it's a vpn invite message")
		r := bufio.NewReader(stream)
		//XXX : see into disadvantages of using newline \n as a delimiter when reading and writing
		//      from/to the buffer. So far, all messages are sent with a \n at the end and the
		//      reader looks for it as a delimiter. See also DeploymentResponse - w.WriteString
		str, err := r.ReadString('\n')
		if err != nil {
			zlog.Sugar().Errorf("failed to read from new stream buffer - %v", err)
			w := bufio.NewWriter(stream)
			_, err := w.WriteString("unable to read VPN Message. Closing Stream.\n")
			if err != nil {
				zlog.Sugar().Errorf("failed to write to stream after unable to read VPN Message - %v", err)
			}

			err = w.Flush()
			if err != nil {
				zlog.Sugar().Errorf("failed to flush stream after unable to read VPN Message - %v", err)
			}

			err = stream.Close()
			if err != nil {
				zlog.Sugar().Errorf("failed to close stream after unable to read VPN Message - %v", err)
			}
			stream.Reset()
			return
		}

		zlog.Sugar().Debugf("[VPN stream] message: %s", str)

		vpnMsg := vpnMessage{}
		err = json.Unmarshal([]byte(str), &vpnMsg)
		if err != nil {
			zlog.Sugar().Errorf("Unable to decode vpn message. Closing stream. Error: %v", err)
			stream.Reset()
			return
		}

		if vpnMsg.MsgType == msgVPNCreationInvite {
			zlog.Sugar().Debugf("Received vpn creation invite")
			var routingTable *vpnRouter
			err := json.Unmarshal([]byte(vpnMsg.Msg), routingTable)
			if err != nil {
				zlog.Sugar().Errorf("Unable to decode vpn message. Closing stream. Error: %v", err)
				stream.Reset()
				return
			}
			vpn, err = JoinVPN(ctx, cancel, *routingTable)
			if err != nil {
				zlog.Sugar().Errorf("Unable to join vpn. Closing stream. Error: %v", err)
				stream.Reset()
				return
			}
			zlog.Sugar().Info("Successfully joined vpn")
			return
		}

		stream.Reset()
		return
	}

	// If tunneling device was not iniated yet, just close the stream
	if vpn.tunDev == nil {
		zlog.Sugar().Errorf("tunDev was not iniated")
		stream.Reset()
		return
	}
	// If the remote node ID isn't in the list of known nodes don't respond.
	if _, ok := vpn.routingTable[stream.Conn().RemotePeer()]; !ok {
		zlog.Sugar().Errorf("Peer %s not on the routing table",
			stream.Conn().RemotePeer().Pretty())
		stream.Reset()
		return
	}
	var packet = make([]byte, 1420)
	var packetSize = make([]byte, 2)
	for {
		// Read the incoming packet's size as a binary value.
		zlog.Sugar().Debug("[VPN] Receiving packet from libp2p stream")
		_, err := stream.Read(packetSize)
		if err != nil {
			zlog.Sugar().Errorf("[VPN] error reading size packet from stream: %v", err)
			stream.Close()
			return
		}

		// Decode the incoming packet's size from binary.
		size := binary.LittleEndian.Uint16(packetSize)

		// Read in the packet until completion.
		var plen uint16 = 0
		for plen < size {
			tmp, err := stream.Read(packet[plen:size])
			plen += uint16(tmp)
			if err != nil {
				zlog.Sugar().Errorf("[VPN] error reading packet's data from stream: %v", err)
				stream.Close()
				return
			}
		}
		zlog.Sugar().Debug("Writing packet to Tunneling interface")
		vpn.tunDev.Iface.Write(packet[:size])
	}
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
