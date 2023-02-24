package libp2p

import (
	"bufio"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p/core/network"
	"gitlab.com/nunet/device-management-service/internal"
	"gitlab.com/nunet/device-management-service/models"
)

var inboundChatStreams []network.Stream
var inboundDepReqStream network.Stream

type openStream struct {
	ID         int
	TimeOpened string
}

func depReqStreamHandler(stream network.Stream) {
	zlog.Info("Got a new depReq stream!")

	// limit to 1 request
	if inboundDepReqStream != nil {
		w := bufio.NewWriter(stream)
		_, err := w.WriteString("Open Stream Length Exceeded. Closing Stream.\n")
		if err != nil {
			zlog.Sugar().Errorln("Error Writing to Stream After DepReq Open Stream Length Exceeded - ", err.Error())
		}

		err = w.Flush()
		if err != nil {
			zlog.Sugar().Errorln("Error Flushing Stream After DepReq Open Stream Length Exceeded - ", err.Error())
		}

		err = stream.Close()
		if err != nil {
			zlog.Sugar().Errorln("Error Closing Stream After DepReq Open Stream Length Exceeded - ", err.Error())
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
		zlog.Sugar().Errorln("Error reading from buffer")
		panic(err)
	}

	depreqMessage := models.DeploymentRequest{}
	err = json.Unmarshal([]byte(str), &depreqMessage)
	if err != nil {
		zlog.Error(err.Error())
		DeploymentResponse("Unable to decode deployment request", true)
	} else {
		DepReqQueue <- depreqMessage
	}
}

func DeploymentResponse(msg string, close bool) error {
	if inboundDepReqStream == nil {
		return fmt.Errorf("No Inbound Deployment Request to Respond to.")
	}
	w := bufio.NewWriter(inboundDepReqStream)
	_, err := w.WriteString(fmt.Sprintf("%s\n", msg))
	if err != nil {
		zlog.Sugar().Errorln("Error Writing Deployment Response to Buffer")
		return err
	}

	err = w.Flush()
	if err != nil {
		zlog.Sugar().Errorln("Error flushing buffer")
		return err
	}

	if close {
		err = inboundDepReqStream.Close()
		if err != nil {
			zlog.Sugar().Errorln("Error Closing InboundDepReqStream - ", err.Error())
		}
		inboundDepReqStream = nil
	}

	return nil
}

func chatStreamHandler(stream network.Stream) {
	zlog.Info("Got a new chat stream!")

	// limit to 3 streams
	if len(inboundChatStreams) >= 3 {
		w := bufio.NewWriter(stream)
		_, err := w.WriteString("Unable to Accept Chat Request. Closing.\n")
		if err != nil {
			zlog.Sugar().Errorln("Error Writing to Stream After Open Chat Stream Length Exceeded - ", err.Error())
		}

		err = w.Flush()
		if err != nil {
			zlog.Sugar().Errorln("Error Flushing Buffer After Open Chat Stream Length Exceeded - ", err.Error())
		}

		err = stream.Close()
		if err != nil {
			zlog.Sugar().Errorln("Error Closing Stream After Open Chat Stream Length Exceeded - ", err.Error())
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
		return nil, fmt.Errorf("No Incoming Message Stream.")
	}

	var out []openStream
	for idx := 0; idx < len(inboundChatStreams); idx++ {
		out = append(out, openStream{ID: idx, TimeOpened: inboundChatStreams[idx].Stat().Opened.String()})
	}
	return out, nil
}

func clearIncomingChatRequests() error {
	if len(inboundChatStreams) == 0 {
		return fmt.Errorf("No Inbound Message Streams.")
	}
	inboundChatStreams = nil
	return nil
}

func readData(r *bufio.Reader) (string, error) {
	str, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}

	if str == "\n" {
		return "", nil
	}
	return str, nil
}

func writeData(w *bufio.Writer, msg string) {
	_, err := w.WriteString(fmt.Sprintf("%s\n", msg))
	if err != nil {
		fmt.Println("Error writing to buffer:", err)
		panic(err)
	}
	err = w.Flush()
	if err != nil {
		fmt.Println("Error flushing buffer:", err)
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
			zlog.Sugar().Errorln("Error Reading From Websocket Connection - ", err.Error())
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
			zlog.Sugar().Info("Connection Error: %v\n", r)
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
