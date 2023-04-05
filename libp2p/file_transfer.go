package libp2p

import (
	"bufio"
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p/core/network"
)

var inboundFileStream network.Stream

func fileStreamHandler(stream network.Stream) {
	zlog.Info("Got a new file stream!")

	// limit to 1 request
	if inboundFileStream != nil {
		w := bufio.NewWriter(stream)
		_, err := w.WriteString("Open Stream Length Exceeded. Closing Stream.\n")
		if err != nil {
			zlog.Sugar().Errorln("Error Writing to Stream After File Transfer Open Stream Length Exceeded - ", err.Error())
		}

		err = w.Flush()
		if err != nil {
			zlog.Sugar().Errorln("Error Flushing Stream After File Transfer Open Stream Length Exceeded - ", err.Error())
		}

		err = stream.Close()
		if err != nil {
			zlog.Sugar().Errorln("Error Closing Stream After File Transfer Open Stream Length Exceeded - ", err.Error())
		}
		return
	}

	inboundFileStream = stream

	r := bufio.NewReader(stream)
	var data []byte
	_, err := r.Read(data)
	if err != nil {
		zlog.Sugar().Errorln("Error reading file data from buffer")
		panic(err)
	}

}

func incomingFileTransferRequests() (string, error) {
	if inboundFileStream == nil {
		return "", fmt.Errorf("no incoming file transfer stream")
	}

	return inboundFileStream.Stat().Opened.GoString(), nil
}

func clearIncomingFileRequests() error {
	if inboundFileStream == nil {
		return fmt.Errorf("no inbound file transfer stream")
	}
	inboundFileStream = nil
	return nil
}

func readFileStream(r *bufio.Reader) ([]byte, error) {
	var data []byte
	_, err := r.Read(data)
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}
	return data, nil
}

func writeFileStream(w *bufio.Writer, data []byte) {
	_, err := w.Write(data)
	if err != nil {
		zlog.Sugar().Errorf("failed to write file data to buffer: %v", err)
		panic(err)
	}
	err = w.Flush()
	if err != nil {
		zlog.Sugar().Errorf("failed to flush buffer: %v", err)
		panic(err)
	}
}

func FileReadStreamWrite(file *os.File, stream network.Stream, w *bufio.Writer) {
	defer func() {
		if r := recover(); r != nil {
			zlog.Sugar().Errorf("Error: %v\n", r)
		}
	}()

	for {
		var data []byte
		_, err := file.Read(data)
		if err != nil {
			zlog.Sugar().Errorln("Error Reading File Data - ", err.Error())
			file.Close()
			stream.Close()
			break
		} else {
			writeFileStream(w, data)
		}
	}
}

func StreamReadFileWrite(file *os.File, stream network.Stream, r *bufio.Reader) {
	defer func() {
		if r := recover(); r != nil {
			zlog.Sugar().Info("Connection Error: %v\n", r)
			file.Close()
			if stream != nil {
				stream.Close()
			}
		}
	}()

	for {
		data, err := readFileStream(r)
		if err != nil {
			panic(err)
		} else if data == nil {
			// do nothing
		} else {
			_, err := file.Write(data)
			if err != nil {
				zlog.Sugar().Errorln("Error Writing Data to File - ", err.Error())
			}
		}
	}
}
