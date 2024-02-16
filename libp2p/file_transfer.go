package libp2p

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/utils"
)

type IncomingFileTransfer struct {
	ID                int
	File              FileMetadata
	Time              time.Time
	Sender            peer.ID
	SenderPublicKey   crypto.PubKey
	InboundFileStream network.Stream
}

var CurrentFileTransfer IncomingFileTransfer

type FileTransferType uint8

const (
	FTDEPREQ FileTransferType = 0 // depreq related file transfer
	FTMISC   FileTransferType = 1 // misc file transfer
)

type FileMetadata struct {
	Name           string           `json:"name"`
	Size           int64            `json:"size"`
	Mod            os.FileMode      `json:"mod"`
	SHA256Checksum string           `json:"sha256_checksum"`
	TransferType   FileTransferType `json:"transfer_type"`
}

type FileTransferResult struct {
	FilePath     string
	TransferChan <-chan utils.IOProgress
	Error        error
}

// checkpoint is a struct to hold checkpoint_dir, filename, and timestamp
type checkpoint struct {
	CheckpointDir string `json:"checkpoint_dir"`
	FilenamePath  string `json:"filename_path"`
	LastModified  int64  `json:"last_modified"`
}

func fileStreamHandler(stream network.Stream) {
	zlog.Info("Got a new file stream!")

	// XXX bad limit to 1 request - temporary
	if CurrentFileTransfer.InboundFileStream != nil {
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

		zlog.Warn("Refusing to accept new file transfer request. Open Stream Length Exceeded.")
		return
	}

	r := bufio.NewReader(stream)
	fileMetadataRaw, err := readString(r)
	zlog.Sugar().Infof("received data from file transfer stream: %s", fileMetadataRaw)
	if err != nil {
		zlog.Sugar().Errorf("couldn't read file metadata from incoming file transfer stream: %v", err)
		stream.Reset()
		return
	}

	incomingFileTransfer := IncomingFileTransfer{}

	err = json.Unmarshal([]byte(fileMetadataRaw), &incomingFileTransfer.File)
	if err != nil {
		zlog.Sugar().Errorf("couldn't unmarshal file metadata from incoming file transfer stream: %v", err)
		stream.Reset()
		return
	}

	incomingFileTransfer.Time = time.Now()
	incomingFileTransfer.Sender = stream.Conn().RemotePeer()
	incomingFileTransfer.SenderPublicKey = stream.Conn().RemotePublicKey()
	incomingFileTransfer.InboundFileStream = stream

	// XXX bad limit to 1 request - temporary
	CurrentFileTransfer = incomingFileTransfer

	// only pass depreq related file transfer requests to queue
	if incomingFileTransfer.File.TransferType == FTDEPREQ {
		zlog.Debug("adding incoming file transfer request to queue")
		FileTransferQueue <- incomingFileTransfer
	}
}

func IncomingFileTransferRequests() (string, error) {
	if CurrentFileTransfer.InboundFileStream == nil {
		return "", fmt.Errorf("no incoming file transfer stream")
	}

	return fmt.Sprintf(
		"Time: %s\nFile Name: %s\nFile Size: %d bytes\n",
		CurrentFileTransfer.Time, CurrentFileTransfer.File.Name, CurrentFileTransfer.File.Size), nil
}

func ClearIncomingFileRequests() error {
	if CurrentFileTransfer.InboundFileStream == nil {
		return fmt.Errorf("no inbound file transfer stream")
	}
	CurrentFileTransfer.InboundFileStream = nil
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
	if err != nil {
		zlog.Sugar().Errorf("error in file transfer write - %v", err)
	}

	zlog.Info("transfer complete. Closing file.")
	file.Close()
}

func StreamReadFileWrite(ctxDone context.CancelFunc, incomingFileTransfer IncomingFileTransfer, file *os.File, r io.Reader) {
	zlog.Sugar().Debugf("in StreamReadFileWrite")

	defer func() {
		if r := recover(); r != nil {
			zlog.Sugar().Info("Connection Error: %v\n", r)
			file.Close()
			if incomingFileTransfer.InboundFileStream != nil {
				incomingFileTransfer.InboundFileStream.Reset()
			}
		}

		ctxDone()
	}()

	n, err := io.Copy(file, r)

	if err != nil {
		zlog.Sugar().Errorf("error in file transfer read - %v", err)
	}

	zlog.Sugar().Infof("file transfer complete - w=%d and err=%v", n, err)
	zlog.Sugar().Infof("file transfer complete - setting mod to %s", incomingFileTransfer.File.Mod.String())
	mod := incomingFileTransfer.File.Mod
	err = file.Chmod(os.FileMode(mod))
	if err != nil {
		zlog.Sugar().Errorf("error changing file mode: %v", err)
	}
	zlog.Sugar().Info("file transfer complete - closing file")
	file.Close()
	if incomingFileTransfer.InboundFileStream != nil {
		incomingFileTransfer.InboundFileStream.Reset()
	}
	CurrentFileTransfer = IncomingFileTransfer{}
}

// SendFileToPeer takes a libp2p peer id and a file path and sends the file to the peer.
func SendFileToPeer(ctx context.Context, peerID peer.ID, filePath string, transferType FileTransferType) (<-chan utils.IOProgress, error) {
	sha256Checksum, err := utils.CalculateSHA256Checksum(filePath)
	if err != nil {
		zlog.Sugar().Fatalf("Error calculating SHA-256 checksum:", err)
	}

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
		Name:           fileInfo.Name(),
		Size:           fileInfo.Size(),
		Mod:            fileInfo.Mode(),
		SHA256Checksum: sha256Checksum,
		TransferType:   transferType,
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

// AcceptFileTransfer accepts the file transfer and returns a file path of location where file is written
// as well as the progress channel with info on how much data is transferred.
func AcceptFileTransfer(ctx context.Context, incomingFileTransfer IncomingFileTransfer) (string, <-chan utils.IOProgress, error) {
	var storagePath = fmt.Sprintf("%s/received_checkpoints", config.GetConfig().General.DataDir)
	_, err := os.Stat(storagePath)
	if err != nil {
		err = os.Mkdir(storagePath, 0755)
		if err != nil {
			return "", nil, fmt.Errorf("unable to create folder %s", storagePath)
		}
	}

	if incomingFileTransfer.File.Name == "" && incomingFileTransfer.File.Size == 0 {
		return "", nil, fmt.Errorf("no file to receive")
	}

	file, err := os.Create(fmt.Sprintf("%s/%s", storagePath, incomingFileTransfer.File.Name))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create file - %v", err)
	}

	r := bufio.NewReader(incomingFileTransfer.InboundFileStream)
	w := bufio.NewWriter(incomingFileTransfer.InboundFileStream)

	writeData(w, []byte("ACK"))

	progressChan := make(chan utils.IOProgress)

	pR := utils.ReaderWithProgress(r, incomingFileTransfer.File.Size)
	ctxWithCancel, done := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ctxWithCancel.Done():
				close(progressChan)
				return
			case <-ticker.C:
				p := pR.Progress
				if p.Complete() {
					progressChan <- p
					close(progressChan)
					return
				}
				progressChan <- p
			}
		}
	}()
	go StreamReadFileWrite(done, incomingFileTransfer, file, pR)

	filePath := file.Name()
	return filePath, progressChan, nil
}

func ListCheckpoints() ([]checkpoint, error) {
	dataDir := config.GetConfig().General.DataDir
	checkpointDir := filepath.Join(dataDir, "received_checkpoints")

	ok, err := AFS.DirExists(checkpointDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoints directory: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("checkpoints directory does not exist")
	}

	checkpoints, err := filepath.Glob(filepath.Join(checkpointDir, "*.tar.gz"))
	if err != nil {
		return nil, fmt.Errorf("failed to find .tar.gz files: %w", err)
	}
	var list []checkpoint
	for _, c := range checkpoints {
		fileInfo, err := os.Stat(c)
		if err != nil {
			zlog.Sugar().Errorf("could not check info for file %s: %w", c, err)
			continue
		}
		lastMod := fileInfo.ModTime().Unix()
		entry := checkpoint{
			CheckpointDir: checkpointDir,
			FilenamePath:  filepath.Base(c),
			LastModified:  lastMod,
		}
		list = append(list, entry)
	}
	return list, nil
}
