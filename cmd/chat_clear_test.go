package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
)

type MockP2PService struct {
	mockInboundStreams []int
}

func (mp *MockP2PService) ClearIncomingChatRequests() error {
	if len(mp.mockInboundStreams) == 0 {
		return fmt.Errorf("no inbound message streams")
	}
	mp.mockInboundStreams = nil
	return nil
}

func (mp *MockP2PService) Decode(s string) (peer.ID, error) {
	preffix := "Qm"

	if !strings.HasPrefix(s, preffix) {
		return "", fmt.Errorf("string does not match valid peer ID")
	}

	return "", nil
}

func Test_ChatClearCmd(t *testing.T) {
	streams := []int{0, 1, 2, 3, 4}

	mockP2P := &MockP2PService{mockInboundStreams: streams}
	buf := new(bytes.Buffer)

	cmd := NewChatClearCmd(mockP2P)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	assert.NoError(t, err)
}

func Test_ChatClearCmdEmpty(t *testing.T) {
	mockP2P := &MockP2PService{}
	buf := new(bytes.Buffer)

	cmd := NewChatClearCmd(mockP2P)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()

	// see if returned expected error
	assert.Errorf(t, err, "no inbound message streams")

	buf.Reset()
}
