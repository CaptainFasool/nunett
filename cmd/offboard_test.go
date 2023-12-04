package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_OffboardCmd(t *testing.T) {
	var (
		mockUtils *MockUtils
		response  string
	)

	mockUtils = newMockUtils()
	mockUtils.onboarded = true

	response = `{ "message":"offboarded" }`
	mockUtils.SetResponse("/api/v1/onboarding/offboard", []byte(response))

	in := bytes.NewBufferString("y\n")
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)

	cmd := NewOffboardCmd(mockUtils)
	cmd.SetIn(in)
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	assert.NoError(t, err)

	assert.Contains(t, out.String(), "offboarded")
}
