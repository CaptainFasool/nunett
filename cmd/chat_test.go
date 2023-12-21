package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ChatCmd(t *testing.T) {
	conns := GetMockConn(true)
	mockConn := &MockConnection{conns: conns}

	cmd := NewChatCmd(mockConn)

	assert := assert.New(t)
	assert.True(cmd.HasAvailableSubCommands())

	subcmd := []string{"list", "clear", "join", "start"}

	cmds := cmd.Commands()
	for _, child := range cmds {
		assert.Contains(subcmd, child.Name())
	}

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	assert.NoError(err)
}
