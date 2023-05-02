package libp2p

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"gitlab.com/nunet/device-management-service/internal"
	"gitlab.com/nunet/device-management-service/models"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var inboundChatStreams []network.Stream
var inboundDepReqStream network.Stream
var outboundDepReqStream network.Stream

type openStream struct {
	ID         int
	TimeOpened string
}

func depReqStreamHandler(stream network.Stream) {
	ctx := context.Background()
	defer ctx.Done()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("MsgType", "DepReq Recv"))
	span.SetAttributes(attribute.String("PeerID", p2p.Host.ID().String()))

	zlog.InfoContext(ctx, "Got a new depReq stream!")

	// limit to 1 request
	if inboundDepReqStream != nil {
		w := bufio.NewWriter(stream)
		_, err := w.WriteString("Open Stream Length Exceeded. Closing Stream.\n")
		if err != nil {
			zlog.Sugar().Errorf("fialed to write to stream after DepReq open stream length exceeded - %v", err)
		}

		err = w.Flush()
		if err != nil {
			zlog.Sugar().Errorf("failed to flush stream after DepReq open stream length exceeded - %v", err)
		}

		err = stream.Close()
		if err != nil {
			zlog.Sugar().Errorf("failed to close stream after DepReq open stream length exceeded - %v", err)
		}
		return
	}

	inboundDepReqStream = stream

	r := bufio.NewReader(stream)
	//XXX : see into disadvantages of using newline \n as a delimiter when reading and writing
	//      from/to the buffer. So far, all messages are sent with a \n at the end and the
	//      reader looks for it as a delimiter. See also DeploymentResponse - w.WriteString
	str, err := r.ReadString('\n')
	if err != nil {
		zlog.Sugar().Errorf("failed to read from buffer")
		panic(err)
	}

	depreqMessage := models.DeploymentRequest{}
	err = json.Unmarshal([]byte(str), &depreqMessage)
	if err != nil {
		zlog.ErrorContext(ctx, fmt.Sprintf("unable to decode deployment request: %v", err))
		// XXX : might be best to propagate context through depReq/depResp to encompass everything done starting with a single depReq
		DeploymentResponse("Unable to decode deployment request", true)
	} else {
		DepReqQueue <- depreqMessage
	}
}

func DeploymentResponseListener(stream network.Stream) {
	defer func() {
		if r := recover(); r != nil {
			zlog.Sugar().Warnf("Connection Error: %v\n", r)
			if stream != nil {
				stream.Close()
			}
		}
	}()

	r := bufio.NewReader(stream)
	for {
		resp, err := readData(r)

		zlog.Sugar().Debugf("received deployment response: %s", resp)

		if err != nil {
			panic(err)
		} else if resp == "" {
			// do nothing
		} else {
			depRespMessage := models.DeploymentResponse{}
			err = json.Unmarshal([]byte(resp), &depRespMessage)
			if err != nil {
				panic(err)
			} else {

				zlog.Sugar().Debugf("deployment response message model: %v", depRespMessage)

				DepResQueue <- depRespMessage
			}
		}
	}
}

func DeploymentResponse(msg string, close bool) error {
	ctx := context.Background()
	defer ctx.Done()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("MsgType", "DepResp Send"))
	span.SetAttributes(attribute.String("PeerID", p2p.Host.ID().String()))

	zlog.Sugar().DebugfContext(ctx, "send deployment response message: %s", msg)
	zlog.Sugar().DebugfContext(ctx, "send deployment response close stream: %v", close)

	if inboundDepReqStream == nil {
		zlog.Sugar().ErrorfContext(ctx, "no inbound deployment request to respond to")
		return fmt.Errorf("no inbound deployment request to respond to")
	}
	w := bufio.NewWriter(inboundDepReqStream)
	_, err := w.WriteString(fmt.Sprintf("%s\n", msg))
	if err != nil {
		zlog.Sugar().ErrorfContext(ctx, "failed to write deployment response to buffer")
		return err
	}

	err = w.Flush()
	if err != nil {
		zlog.Sugar().Errorf("failed to flush buffer")
		return err
	}

	if close {
		err = inboundDepReqStream.Close()
		if err != nil {
			zlog.Sugar().ErrorfContext(ctx, "failed to close inbound deployment request stream - %v", err)
		}
		inboundDepReqStream = nil
	}

	return nil
}

func SendDeploymentRequest(ctx context.Context, depReq models.DeploymentRequest) (network.Stream, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("MsgType", "DepReq Recv"))
	span.SetAttributes(attribute.String("PeerID", p2p.Host.ID().String()))
	zlog.InfoContext(ctx, "Creating a new depReq!")

	// limit to 1 request
	if outboundDepReqStream != nil {
		return nil, fmt.Errorf("couldn't create deployment request. a request already in progress")
	}

	peerID, err := peer.Decode(depReq.Params.NodeID)
	if err != nil {
		return nil, fmt.Errorf("couldn't decode input peer-id '%s', : %v", depReq.Params.NodeID, err)
	}

	outboundDepReqStream, err := GetP2P().Host.NewStream(ctx, peerID, protocol.ID(DepReqProtocolID))
	if err != nil {
		return nil, fmt.Errorf("couldn't create deployment request stream: %v", err)
	}

	msg, err := json.Marshal(depReq)
	if err != nil {
		return nil, fmt.Errorf("couldn't convert deployment request to json: %v", err)
	}

	w := bufio.NewWriter(outboundDepReqStream)

	zlog.Sugar().DebugfContext(ctx, "deployment request: %s", string(msg))

	_, err = w.WriteString(fmt.Sprintf("%s\n", msg))
	if err != nil {
		return nil, fmt.Errorf("couldn't write deployment request to stream: %v", err)
	}
	err = w.Flush()
	if err != nil {
		return nil, fmt.Errorf("couldn't flush deployment request to stream: %v", err)
	}

	return outboundDepReqStream, nil
}

func chatStreamHandler(stream network.Stream) {
	zlog.Info("Got a new chat stream!")

	// limit to 3 streams
	if len(inboundChatStreams) >= 3 {
		w := bufio.NewWriter(stream)
		_, err := w.WriteString("Unable to Accept Chat Request. Closing.\n")
		if err != nil {
			zlog.Sugar().Errorf("failed to write to stream after open chat stream length exceeded - %v", err)
		}

		err = w.Flush()
		if err != nil {
			zlog.Sugar().Errorf("failed to flush buffer after open chat stream length exceeded - %v", err)
		}

		err = stream.Close()
		if err != nil {
			zlog.Sugar().Errorf("failed to close stream after open chat stream length exceeded - %v", err)
		}
		return
	}

	if stream.Stat().Direction.String() == "Inbound" && !stream.Stat().Transient {
		zlog.Info("Adding Incoming Stream to Queue")
		inboundChatStreams = append(inboundChatStreams, stream)
	}
}

func incomingChatRequests() ([]openStream, error) {
	if len(inboundChatStreams) == 0 {
		return nil, fmt.Errorf("no incoming message stream")
	}

	var out []openStream
	for idx := 0; idx < len(inboundChatStreams); idx++ {
		out = append(out, openStream{ID: idx, TimeOpened: inboundChatStreams[idx].Stat().Opened.String()})
	}
	return out, nil
}

func clearIncomingChatRequests() error {
	if len(inboundChatStreams) == 0 {
		return fmt.Errorf("no inbound message streams")
	}
	inboundChatStreams = nil
	return nil
}

func readData(r *bufio.Reader) (string, error) {
	str, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	zlog.Sugar().Debugf("received raw data from stream: %s", str)
	if str == "\n" {
		return "", nil
	}
	return str, nil
}

func writeData(w *bufio.Writer, msg string) {

	zlog.Sugar().Debugf("writing raw data to stream: %s", msg)

	_, err := w.WriteString(fmt.Sprintf("%s\n", msg))
	if err != nil {
		zlog.Sugar().Errorf("failed to write to buffer: %v", err)
		panic(err)
	}
	err = w.Flush()
	if err != nil {
		zlog.Sugar().Errorf("failed to flush buffer: %v", err)
		panic(err)
	}
}

func SockReadStreamWrite(conn *internal.WebSocketConnection, stream network.Stream, w *bufio.Writer) {
	defer func() {
		if r := recover(); r != nil {
			zlog.Sugar().Errorf("Error: %v\n", r)
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			zlog.Sugar().Errorf("Error Reading From Websocket Connection - %v", err)
			conn.Close()
			stream.Close()
			break
		} else {
			writeData(w, string(msg))
		}
	}
}

func StreamReadSockWrite(conn *internal.WebSocketConnection, stream network.Stream, r *bufio.Reader) {
	defer func() {
		if r := recover(); r != nil {
			zlog.Sugar().Infof("Connection Error: %v\n", r)
			conn.Close()
			if stream != nil {
				stream.Close()
			}
		}
	}()

	for {
		reply, err := readData(r)
		if err != nil {
			panic(err)
		} else if reply == "" {
			// do nothing
		} else {
			conn.Conn.WriteMessage(websocket.TextMessage, []byte("Peer: "+reply))
		}
	}
}
