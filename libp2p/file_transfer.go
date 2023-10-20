package libp2p

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/utils"
)

type IncomingFileTransfer struct {
	File              FileMetadata
	Time              time.Time
	Sender            peer.ID
	SenderPublicKey   crypto.PubKey
	InboundFileStream network.Stream
}

var incomingFileTransfer IncomingFileTransfer

type FileMetadata struct {
	Name string      `json:"name"`
	Size int64       `json:"size"`
	Mod  os.FileMode `json:"mod"`
}

func fileStreamHandler(stream network.Stream) {
	zlog.Info("Got a new file stream!")

	// limit to 1 request
	if incomingFileTransfer.InboundFileStream != nil {
		w := bufio.NewWriter(stream)
		_, err := w.WriteString("Open Stream Length Exceeded. Closing Stream.\n")
		if err != nil {
			zlog.Sugar().Errorln("Error Writing to Stream After File Transfer Open Stream Length Exceeded - ", err.Error())
		}

		err = w.Flush()
		if err != nil {
			zlog.Sugar().Errorln("Error Flushing Stream After File Transfer Open Stream Length Exceeded - ", err.Error())
		}

		err = stream.Reset()
		if err != nil {
			zlog.Sugar().Errorln("Error Closing Stream After File Transfer Open Stream Length Exceeded - ", err.Error())
		}
		return
	}

	incomingFileTransfer.InboundFileStream = stream

	r := bufio.NewReader(stream)
	fileMetadataRaw, err := readString(r)
	zlog.Sugar().Infof("received data from file transfer stream: %s", fileMetadataRaw)
	if err != nil {
		zlog.Sugar().Errorf("couldn't read file metadata from incoming file transfer stream: %v", err)
		incomingFileTransfer.InboundFileStream = nil
		incomingFileTransfer = IncomingFileTransfer{}

		stream.Reset()
	}

	err = json.Unmarshal([]byte(fileMetadataRaw), &incomingFileTransfer.File)
	if err != nil {
		zlog.Sugar().Errorf("couldn't unmarshal file metadata from incoming file transfer stream: %v", err)
		incomingFileTransfer.InboundFileStream = nil
		incomingFileTransfer = IncomingFileTransfer{}
		stream.Reset()
	}
	zlog.Sugar().Infof("unmarshalled file metadata: %+v", incomingFileTransfer)

	incomingFileTransfer.Time = time.Now()
	incomingFileTransfer.Sender = stream.Conn().RemotePeer()
	incomingFileTransfer.SenderPublicKey = stream.Conn().RemotePublicKey()

	FileTransferQueue <- incomingFileTransfer
}

func incomingFileTransferRequests() (string, error) {
	if incomingFileTransfer.InboundFileStream == nil {
		return "", fmt.Errorf("no incoming file transfer stream")
	}

	return fmt.Sprintf(
		"Time: %s\nFile Name: %s\nFile Size: %d bytes\n",
		incomingFileTransfer.Time, incomingFileTransfer.File.Name, incomingFileTransfer.File.Size), nil
}

func clearIncomingFileRequests() error {
	if incomingFileTransfer.InboundFileStream == nil {
		return fmt.Errorf("no inbound file transfer stream")
	}
	incomingFileTransfer.InboundFileStream = nil
	return nil
}

func FileReadStreamWrite(file *os.File, stream network.Stream, w io.Writer) {

	zlog.Sugar().Debugf("in FileReadStreamWrite")

	defer func() {
		if r := recover(); r != nil {
			zlog.Sugar().Errorf("Error: %v\n", r)
		}
	}()

	io.Copy(w, file)
	zlog.Info("transfer complete. Closing file.")
	file.Close()
}

func StreamReadFileWrite(file *os.File, stream network.Stream, r *bufio.Reader) {

	zlog.Sugar().Debugf("in StreamReadFileWrite")

	defer func() {
		if r := recover(); r != nil {
			zlog.Sugar().Info("Connection Error: %v\n", r)
			file.Close()
			if stream != nil {
				stream.Reset()
			}
		}
	}()

	// XXX - the following commented out code is left here to serve as an example
	//       when receiver side progress update is necessary. Replace the `io.Copy(file, r)`
	//       with commented out code.
	// newR := utils.ReaderWithProgress(r, incomingFileTransfer.File.Size)
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// go func() {
	// 	ticker := time.NewTicker(1 * time.Second)
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			return
	// 		case <-ticker.C:
	// 			p := newR.Progress
	// 			if p.Complete() {
	// 				return
	// 			}
	// 			fmt.Printf("receiving file - %.2f MB transferred so far\n", p.N()/1048576)
	// 		}
	// 	}
	// }()
	// n, err := io.Copy(file, newR)

	zlog.Info("starting copy-------")
	n, err := io.Copy(file, r)
	zlog.Sugar().Infof("file transfer complete - w=%d and err=%v", n, err)
	zlog.Sugar().Infof("file transfer complete - setting mod to %s", incomingFileTransfer.File.Mod.String())
	mod := incomingFileTransfer.File.Mod
	err = file.Chmod(os.FileMode(mod))
	if err != nil {
		zlog.Sugar().Errorf("error changing file mode: %v", err)
	}
	zlog.Sugar().Info("file transfer complete - closing file")
	file.Close()
	if stream != nil {
		stream.Reset()
	}
	incomingFileTransfer = IncomingFileTransfer{}
	incomingFileTransfer.InboundFileStream = nil
}

// SendFileToPeer takes a libp2p peer id and a file path and sends the file to the peer.
func SendFileToPeer(ctx context.Context, peerID peer.ID, filePath string) (<-chan utils.IOProgress, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("couldn't open file: %v", err)
	}

	zlog.Sugar().Debugf("sending '%s' to %s", filePath, peerID)

	stream, err := p2p.Host.NewStream(ctx, peerID, protocol.ID(FileTransferProtocolID))
	if err != nil {
		return nil, fmt.Errorf("could not create stream with peer for file transfer: %v", err)
	}

	zlog.Sugar().Debugf("stream : to %v", stream)

	w := bufio.NewWriter(stream)

	// send file metadata
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not get file info: %v", err)
	}

	fileMetadata := FileMetadata{
		Name: fileInfo.Name(),
		Size: fileInfo.Size(),
		Mod:  fileInfo.Mode(),
	}

	fileMetadataBytes, err := json.Marshal(fileMetadata)
	if err != nil {
		return nil, fmt.Errorf("could not marshal file metadata: %v", err)
	}

	n, err := writeString(w, string(fileMetadataBytes))
	if err != nil {
		return nil, fmt.Errorf("could not send file metadata: %v", err)
	}

	zlog.Sugar().Infof("file metadata: sent %d bytes", n)

	//wait for ack
	r := bufio.NewReader(stream)
	var resp = make([]byte, 3)
	n, err = readData(r, resp)
	if err != nil {
		return nil, fmt.Errorf("could not read ack after file metadata send: %v", err)
	}
	zlog.Sugar().Infof("received ack response (n = %d): %q", n, string(resp))
	if string(resp) == "ACK" {
		pW := utils.WriterWithProgress(w, fileInfo.Size())
		progressChan := make(chan utils.IOProgress)
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			for {
				select {
				case <-ctx.Done():
					close(progressChan)
					return
				case <-ticker.C:
					p := pW.Progress
					if p.Complete() {
						progressChan <- p
						close(progressChan)
						return
					}
					progressChan <- p
				}
			}
		}()
		go FileReadStreamWrite(file, stream, pW)

		return progressChan, nil
	} else {
		return nil, fmt.Errorf("peer denied file transfer: %v - response: %s", err, string(resp))
	}
}

// AcceptFile takes an IncomingFileTransfer and accepts the file transfer.
func AcceptFileTransfer() error {
	var storagePath = config.GetConfig().General.DataDir
	_, err := os.Stat(storagePath)
	if err != nil {
		err = os.Mkdir(storagePath, 0755)
		if err != nil {
			return fmt.Errorf("unable to create folder %s", storagePath)
		}
	}

	if incomingFileTransfer.File.Name == "" && incomingFileTransfer.File.Size == 0 {
		return fmt.Errorf("no file to receive")
	}

	file, err := os.Create(fmt.Sprintf("%s/%s", storagePath, incomingFileTransfer.File.Name))
	if err != nil {
		return fmt.Errorf("failed to create file - %v", err)
	}

	r := bufio.NewReader(incomingFileTransfer.InboundFileStream)
	w := bufio.NewWriter(incomingFileTransfer.InboundFileStream)
	writeData(w, []byte("ACK"))

	go StreamReadFileWrite(file, incomingFileTransfer.InboundFileStream, r)
	return nil
}
