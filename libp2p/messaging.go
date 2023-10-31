package libp2p

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/cardano-community/koios-go-client/v3"
	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/internal"
	kLogger "gitlab.com/nunet/device-management-service/internal/tracing"
	"gitlab.com/nunet/device-management-service/models"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Constants containing all message types happening between peers.
const (
	MsgDepResp   = "DepResp"
	MsgDepReq    = "DepReq"
	MsgJobStatus = "JobStatus"
	MsgLogStderr = "LogStderr"
	MsgLogStdout = "LogStdout"
)

// constants for job status messaging
const (
	ContainerJobPending               = "pending"
	ContainerJobRunning               = "running"
	ContainerJobFinishedWithErrors    = "finished with errors"
	ContainerJobFinishedWithoutErrors = "finished without errors"
	ContainerJobFailed                = "failed"
)

// constants for job status actions
const (
	JobSubmitted = "job-submitted"
	JobFailed    = "job-failed"
	JobCompleted = "job-completed"
)

var txHashPropogrationTime = 60 // seconds

var inboundChatStreams []network.Stream
var inboundDepReqStream network.Stream
var OutboundDepReqStream network.Stream

type OpenStream struct {
	ID         int    `json:"id"`
	StreamID   string `json:"stream_id"`
	FromPeer   string `json:"from_peer"`
	TimeOpened string `json:"time_opened"`
}

func writeToStream(stream network.Stream, msg string, failReason string) {
	w := bufio.NewWriter(stream)

	_, err := w.WriteString(fmt.Sprintf("%s\n", msg))
	if err != nil {
		zlog.Sugar().Errorf("failed to write to stream after %s - %v", failReason, err)
	}

	err = w.Flush()
	if err != nil {
		zlog.Sugar().Errorf("failed to flush stream after %s - %v", failReason, err)
	}

	err = stream.Close()
	if err != nil {
		zlog.Sugar().Errorf("failed to close stream after %s - %v", failReason, err)
	}
}

func depReqStreamHandler(stream network.Stream) {
	ctx := context.Background()
	defer ctx.Done()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("MsgType", MsgDepReq))
	span.SetAttributes(attribute.String("PeerID", p2p.Host.ID().String()))
	kLogger.Info("Deployment request stream handler", span)

	zlog.InfoContext(ctx, "Got a new depReq stream!")

	// limit to 1 request
	if inboundDepReqStream != nil {
		zlog.Sugar().Debugf("[depReq recv] depReq already in progress. Refusing to accept.")
		depRes := models.DeploymentResponse{Success: false, Content: "Open Stream Length Exceeded. Closing Stream."}
		depResJson, err := json.Marshal(depRes)
		if err != nil {
			zlog.Sugar().Errorf("failed to marshal depRes - %v", err)
			stream.Close()
			return
		}
		depUpdate := models.DeploymentUpdate{MsgType: MsgDepResp, Msg: string(depResJson)}
		depUpdateJson, err := json.Marshal(depUpdate)
		if err != nil {
			zlog.Sugar().Errorf("failed to marshal depUpdate - %v", err)
			stream.Close()
			return
		}

		writeToStream(stream, string(depUpdateJson), "DepReq open stream length exceeded")

		return
	}

	zlog.Sugar().DebugfContext(ctx, "[depReq recv] no existing depReq. Proceeding to read from stream.")
	r := bufio.NewReader(stream)
	//XXX : see into disadvantages of using newline \n as a delimiter when reading and writing
	//      from/to the buffer. So far, all messages are sent with a \n at the end and the
	//      reader looks for it as a delimiter. See also DeploymentResponse - w.WriteString
	str, err := r.ReadString('\n')
	if err != nil {
		zlog.Sugar().Errorf("failed to read from new stream buffer - %v", err)
		writeToStream(stream, "Unable to read DepReq. Closing Stream.", "unable to read depReq")
		return
	}

	zlog.Sugar().DebugfContext(ctx, "[depReq recv] message: %s", str)

	inboundDepReqStream = stream

	depreqMessage := models.DeploymentRequest{}
	err = json.Unmarshal([]byte(str), &depreqMessage)
	if err != nil {
		zlog.ErrorContext(ctx, fmt.Sprintf("unable to decode deployment request: %v", err))
		// XXX : might be best to propagate context through depReq/depResp to encompass everything done starting with a single depReq
		DeploymentUpdate(MsgDepResp, "Unable to decode deployment request", true)
	} else {
		// check if txhash is valid, but before that, we need to wait for txHashPropogrationTime
		zlog.Sugar().Infof("going to sleep for %d seconds while waiting for tx to propogate to other nodes", txHashPropogrationTime)
		time.Sleep(time.Duration(txHashPropogrationTime) * time.Second)

		doesHaveValidTxHash := checkTxValidity(depreqMessage.TxHash)
		if doesHaveValidTxHash {
			zlog.Sugar().Infof("tx_hash for %s is valid, proceeding with deployment", depreqMessage.TxHash)
			DepReqQueue <- depreqMessage
		} else {
			zlog.Sugar().Infof("tx_hash for %s is invalid or timed out. Stopping deployment process", depreqMessage.TxHash)
			DeploymentUpdate(MsgDepResp, "Invalid TxHash", true)
		}
	}
}

// checkTxValidity checks if transation really exists on the blockchain and not
// just some arbitrary value
func checkTxValidity(txHash string) (valid bool) {
	client, err := koios.New(koios.Host(koios.PreProdHost))
	if err != nil {
		log.Fatal(err)
	}

	txInfo, err := client.GetTxInfo(context.Background(), koios.TxHash(txHash), nil)
	if err != nil {
		zlog.Sugar().Infof("%v", err)
	}

	if txInfo.StatusCode == 200 {
		valid = true
	}

	return valid
}

// DeploymentUpdateListener listens for deployment response and service running status.
func DeploymentUpdateListener(stream network.Stream) {
	defer func() {
		if r := recover(); r != nil {
			zlog.Sugar().Errorf("connection error: closing stream and websocket %v", r)
			if stream != nil {
				stream.Close()
				OutboundDepReqStream = nil
			}
		}
	}()

	r := bufio.NewReader(stream)
	for {
		if stream.Conn().IsClosed() {
			zlog.Sugar().Info("stream closed")
			return
		}
		resp, err := readString(r)

		if err == io.EOF {
			zlog.Sugar().Debug("Stream closed with EOF, ending read loop")
			return
		} else if err != nil || resp == "" {
			zlog.Sugar().Errorf("failed to read deployment update: %v", err)
			panic(err)
		}

		zlog.Sugar().Debugf("received deployment update  -   msg: %s", resp)

		var depUpd models.DeploymentUpdate
		err = json.Unmarshal([]byte(resp), &depUpd)
		if err != nil {
			zlog.Sugar().Info("couldn't unmarshal deployment update")
			continue
		}

		var depReqFlat models.DeploymentRequestFlat

		switch depUpd.MsgType {
		case MsgJobStatus:
			service := models.Services{}
			err = json.Unmarshal([]byte(depUpd.Msg), &service)
			if err != nil {
				zlog.Sugar().Errorf("failed to unmarshal deployment JobStatus update: %v", err)
			}

			zlog.Sugar().Debugf("received deployment job status: %s", service.JobStatus)

			if strings.Contains(string(service.JobStatus), "finished") {
				if service.JobStatus == "finished without errors" {
					JobCompletedQueue <- ContainerJobFinishedWithoutErrors
				} else if service.JobStatus == "finished with errors" {
					JobFailedQueue <- ContainerJobFinishedWithErrors
				}

				// job finished, closing stream
				if stream != nil {
					stream.Close()
					OutboundDepReqStream = nil
				}
				zlog.Sugar().Infof("Deployed job finished. Deleting DepReqFlat record (id=%d) from DB", depReqFlat.ID)
				// XXX: Needs to be modified to take multiple deployment requests from same service provider
				// deletes all the record in table; deletes == mark as delete
				if err := db.DB.Where("deleted_at IS NULL").Delete(&depReqFlat).Error; err != nil {
					zlog.Sugar().Errorf("unable to delete record (id=%d) after job finish: %v", depReqFlat.ID, err)
				}

				// update service info into SP's DMS for claim Reward by SP user
				result := db.DB.Model(&models.Services{}).Where("tx_hash = ?", service.TxHash).Select("*").Updates(service)
				if result.Error != nil {
					zlog.Sugar().Errorf("Unable to update service info on SP side: %v", result.Error.Error())
				}

				return
			} else if strings.EqualFold(string(service.JobStatus), "running") {
				depRespMessage := models.DeploymentResponse{}
				depRespMessage.Content = service.LogURL
				depRespMessage.Success = true
				zlog.Sugar().Debugf("deployment update (jobstatus=running): %v", depRespMessage)
				DepResQueue <- depRespMessage
			}

			// update deplreqflat.jobstatus
			db.DB.Last(&depReqFlat)
			zlog.Sugar().Infof("Updating DepReqFlat Job Status record in DB (id=%d, jobstatus=%s)", depReqFlat.ID, service.JobStatus)
			depReqFlat.JobStatus = service.JobStatus
			if err := db.DB.Save(&depReqFlat).Error; err != nil {
				zlog.Sugar().Errorf("unable to update job status on finish. %v", err)
			}

			// update service info into SP's DMS for claim Reward by SP user
			result := db.DB.Model(&models.Services{}).Where("tx_hash = ?", service.TxHash).Select("*").Updates(service)
			if result.Error != nil {
				zlog.Sugar().Errorf("Unable to update service info on SP side: %v", result.Error.Error())
			}
		case MsgDepResp:
			zlog.Sugar().Debugf("received deployment response: %s", resp)

			if err != nil {
				zlog.Sugar().Errorf("failed to read deployment response: %v", err)
			} else if resp == "" {
				// do nothing
			} else {
				depRespMessage := models.DeploymentResponse{}
				err = json.Unmarshal([]byte(depUpd.Msg), &depRespMessage)
				if err != nil {
					zlog.Sugar().Errorf("failed to unmarshal deployment response: %v", err)
				}
				zlog.Sugar().Debugf("deployment update message model: %v", depRespMessage)

				if depRespMessage.Success {
					DepResQueue <- depRespMessage
				} else if !depRespMessage.Success {
					JobFailedQueue <- depRespMessage.Content
				}

				// XXX: Needs to be modified to take multiple deployment requests from same service provider
				// deletes all the record in table; deletes == mark as delete
				if err := db.DB.Where("deleted_at IS NULL").Delete(&depReqFlat).Error; err != nil {
					// TODO: Do not delete, update JobStatus
					zlog.Sugar().Errorf("%v", err)
				}
			}
		case MsgLogStdout, MsgLogStderr:
			zlog.Sugar().Debugf("received log response with length: %d", len(resp))

			if depUpd.MsgType == MsgLogStdout {
				JobLogStdoutQueue <- depUpd.Msg
			} else if depUpd.MsgType == MsgLogStderr {
				JobLogStderrQueue <- depUpd.Msg
			}
		default:
			zlog.Sugar().Errorf("received unknown type deployment update")
		}
	}
}

// DeploymentUpdate is an auxilary function to send updates from one machine to another
// Args:
//
//	msgType: one of MsgDepResp, MsgDepReq, MsgDepReqUpdate, MsgJobStatus, MsgLogStderr, MsgLogStdout
//	msg:     message to send
//	inbound: true if the depReq was inbound (DMS is CP) or false if depReq was outbound (DMS is SP)
//	close:   true if the depReq stream needs to be closed after sending the message
func DeploymentUpdate(msgType string, msg string, close bool) error {
	ctx := context.Background()
	defer ctx.Done()

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("MsgType", msgType))
	span.SetAttributes(attribute.String("PeerID", p2p.Host.ID().String()))
	kLogger.Info("Deployment update", span)

	zlog.Sugar().DebugfContext(ctx, "DeploymentUpdate -- msgType: %s -- closeStream: %t -- msg: %s", msgType, close, msg)

	// Construct the outer message before sending
	depUpdateMsg := &models.DeploymentUpdate{
		MsgType: msgType,
		Msg:     msg,
	}

	msgBytes, _ := json.Marshal(depUpdateMsg)

	if inboundDepReqStream == nil {
		zlog.Sugar().ErrorfContext(ctx, "no inbound deployment request stream to send an update to")
		return fmt.Errorf("no inbound deployment request to respond to")
	}

	w := bufio.NewWriter(inboundDepReqStream)
	_, err := w.WriteString(fmt.Sprintf("%s\n", string(msgBytes)))
	if err != nil {
		zlog.Sugar().ErrorfContext(ctx, "failed to write deployment update to buffer")
		return err
	}

	err = w.Flush()
	if err != nil {
		zlog.Sugar().Errorf("failed to flush buffer")
		return err
	}

	if close {
		zlog.Sugar().InfofContext(ctx, "closing deployment request stream from")
		err = inboundDepReqStream.Close()
		if err != nil {
			zlog.Sugar().ErrorfContext(ctx, "failed to close deployment request stream - %v", err)
		}
		inboundDepReqStream = nil
	}

	return nil
}

func SendDeploymentRequest(ctx context.Context, depReq models.DeploymentRequest) (network.Stream, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("MsgType", MsgDepReq))
	span.SetAttributes(attribute.String("PeerID", p2p.Host.ID().String()))
	zlog.InfoContext(ctx, "Creating a new depReq!")

	// limit to 1 request
	if OutboundDepReqStream != nil {
		return nil, fmt.Errorf("couldn't create deployment request. a request already in progress")
	}

	peerID, err := peer.Decode(depReq.Params.RemoteNodeID)
	if err != nil {
		return nil, fmt.Errorf("couldn't decode input peer-id '%s', : %v", depReq.Params.RemoteNodeID, err)
	}

	OutboundDepReqStream, err = GetP2P().Host.NewStream(ctx, peerID, protocol.ID(DepReqProtocolID))
	if err != nil {
		return nil, fmt.Errorf("couldn't create deployment request stream: %v", err)
	}

	msg, err := json.Marshal(depReq)
	if err != nil {
		return nil, fmt.Errorf("couldn't convert deployment request to json: %v", err)
	}

	w := bufio.NewWriter(OutboundDepReqStream)

	zlog.Sugar().DebugfContext(ctx, "deployment request: %s", string(msg))

	_, err = w.WriteString(fmt.Sprintf("%s\n", msg))
	if err != nil {
		return nil, fmt.Errorf("couldn't write deployment request to stream: %v", err)
	}
	err = w.Flush()
	if err != nil {
		return nil, fmt.Errorf("couldn't flush deployment request to stream: %v", err)
	}

	return OutboundDepReqStream, nil
}

func chatStreamHandler(stream network.Stream) {
	zlog.Info("Got a new chat stream!")

	// limit to 3 streams
	if len(inboundChatStreams) >= 3 {
		writeToStream(stream, "Unable to Accept Chat Request. Closing.", "open chat stream length exceeded")
		return
	}

	if stream.Stat().Direction.String() == "Inbound" && !stream.Stat().Transient {
		zlog.Info("Adding Incoming Stream to Queue")
		inboundChatStreams = append(inboundChatStreams, stream)
	}
}

func IncomingChatRequests() ([]OpenStream, error) {
	if len(inboundChatStreams) == 0 {
		return nil, fmt.Errorf("no incoming message stream")
	}

	var out []OpenStream
	for idx := 0; idx < len(inboundChatStreams); idx++ {
		out = append(out, OpenStream{
			ID:         idx,
			StreamID:   inboundChatStreams[idx].ID(),
			FromPeer:   inboundChatStreams[idx].Conn().RemotePeer().String(),
			TimeOpened: inboundChatStreams[idx].Stat().Opened.String()})
	}
	return out, nil
}

func ClearIncomingChatRequests() error {
	if len(inboundChatStreams) == 0 {
		return fmt.Errorf("no inbound message streams")
	}
	inboundChatStreams = nil
	return nil
}

func readData(r *bufio.Reader, data []byte) (int, error) {
	n, err := io.ReadFull(r, data)
	return n, err
}

func writeData(w *bufio.Writer, data []byte) (int, error) {
	n, err := w.Write(data)
	if err != nil {
		zlog.Sugar().Errorf("failed to write to buffer: %v", err)
		return 0, err
	}
	err = w.Flush()
	if err != nil {
		zlog.Sugar().Errorf("failed to flush buffer: %v", err)
		return 0, err
	}
	return n, nil
}

func readString(r *bufio.Reader) (string, error) {
	str, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}

	if str == "\n" {
		return "", nil
	}

	zlog.Sugar().Debugf("received raw data from stream: %s", str)

	return str, nil
}

func writeString(w *bufio.Writer, msg string) (int, error) {

	zlog.Sugar().Debugf("writing raw data to stream: %s", msg)

	n, err := w.WriteString(fmt.Sprintf("%s\n", msg))
	if err != nil {
		zlog.Sugar().Errorf("failed to write to buffer: %v", err)
		return 0, err
	}
	err = w.Flush()
	if err != nil {
		zlog.Sugar().Errorf("failed to flush buffer: %v", err)
		return 0, err
	}
	return n, nil
}

func SockReadStreamWrite(conn *internal.WebSocketConnection, stream network.Stream, w *bufio.Writer) {
	defer func() {
		if r := recover(); r != nil {
			zlog.Sugar().Errorf("Error: closing stream and conn after panic -  %v", r)
			if conn != nil {
				conn.Close()
			}
			if stream != nil {
				stream.Close()
			}
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			zlog.Sugar().Errorf("Error Reading From Websocket Connection.  - %v", err)
			panic(err)
		}

		if string(msg) != "\n" {
			writeString(w, string(msg))
		}
	}
}

func StreamReadSockWrite(conn *internal.WebSocketConnection, stream network.Stream, r *bufio.Reader) {
	defer func() {
		if r := recover(); r != nil {
			zlog.Sugar().Infof("Error: closing stream and conn after panic - %v", r)
			if conn != nil {
				conn.Close()
			}
			if stream != nil {
				stream.Close()
			}
		}
	}()

	for {
		reply, err := readString(r)
		if err != nil {
			panic(err)
		}

		reply = strings.TrimSuffix(reply, "\n")

		if reply != "" {
			conn.Conn.WriteMessage(websocket.TextMessage, []byte("Peer: "+reply))
		}
	}
}

func IsDepReqStreamOpen() bool {
	return OutboundDepReqStream != nil
}

func IsDepRespStreamOpen() bool {
	return inboundDepReqStream != nil
}
