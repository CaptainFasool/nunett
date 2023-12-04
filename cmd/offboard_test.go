package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_OffboardCmd(t *testing.T) {
	type test struct {
		args  []string
		input string
		want  string
	}

	type response struct {
		Msg string `json:"message"`
	}

	var (
		mockUtils *MockUtils

		success       string
		abort         string
		responseJSON  response
		responseBytes []byte

		tests []test
		err   error
	)
	assert := assert.New(t)

	success = "Offboarded successfully!"
	abort = "Exiting..."

	tests = []test{
		{input: "y\n", want: success},
		{args: []string{"--force"}, input: "y\n", want: success},
		{input: "n\n", want: abort},
		{args: []string{"--force"}, input: "n\n", want: abort},
	}

	responseJSON = response{
		Msg: success,
	}
	responseBytes, err = json.Marshal(responseJSON)
	if err != nil {
		t.Fatalf("could not marshal response JSON: %v", err)
	}

	mockUtils = newMockUtils()
	mockUtils.SetResponse("/api/v1/onboarding/offboard", responseBytes)
	mockUtils.onboarded = true

	for _, tc := range tests {
		in := bytes.NewBufferString(tc.input)
		out := new(bytes.Buffer)
		errOut := new(bytes.Buffer)

		cmd := NewOffboardCmd(mockUtils)
		cmd.SetIn(in)
		cmd.SetOut(out)
		cmd.SetErr(errOut)
		cmd.SetArgs(tc.args)

		err = cmd.Execute()
		assert.NoError(err)

		if tc.args != nil {
			hasFlag, _ := cmd.Flags().GetBool("force")
			assert.True(hasFlag)
		}
		assert.Contains(out.String(), tc.want)
	}
}
