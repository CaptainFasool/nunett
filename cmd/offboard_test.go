package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
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
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if tc.args != nil {
			hasFlag, _ := cmd.Flags().GetBool("force")
			assert.True(hasFlag)
		}
		assert.Contains(out.String(), tc.want)
	}
}

func Test_OffboardCmdOnboardError(t *testing.T) {
	var (
		errOnboard string
		err        error

		out    *bytes.Buffer
		errOut *bytes.Buffer

		mockUtils *MockUtils
		cmd       *cobra.Command
	)
	errOnboard = "onboard failed"

	mockUtils = newMockUtils()
	mockUtils.onboardErr = fmt.Errorf(errOnboard)

	out = new(bytes.Buffer)
	errOut = new(bytes.Buffer)

	cmd = NewOffboardCmd(mockUtils)
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err = cmd.Execute()
	if err == nil {
		t.Fatalf("expected error while executing, got %v", err)
	}
	assert.ErrorContains(t, err, errOnboard)

	mockUtils.onboardErr = nil
	err = cmd.Execute()
	if err == nil {
		t.Fatalf("expected error while executing, got %v", err)
	}
	assert.ErrorContains(t, err, "machine is not onboarded")
}

func Test_OffboardCmdResponseError(t *testing.T) {
	type response struct {
		Err string `json:"error"`
	}

	var (
		errMsg string
		err    error

		responseJSON  response
		responseBytes []byte

		in     *bytes.Buffer
		out    *bytes.Buffer
		errOut *bytes.Buffer

		mockUtils *MockUtils
		cmd       *cobra.Command
	)
	errMsg = "offboard failed at some point"
	responseJSON = response{
		Err: errMsg,
	}
	responseBytes, err = json.Marshal(responseJSON)
	if err != nil {
		t.Fatalf("could not marshal response JSON: %v", err)
	}

	mockUtils = newMockUtils()
	mockUtils.SetResponse("/api/v1/onboarding/offboard", responseBytes)
	mockUtils.onboarded = true

	in = bytes.NewBufferString("y\n")
	out = new(bytes.Buffer)
	errOut = new(bytes.Buffer)

	cmd = NewOffboardCmd(mockUtils)
	cmd.SetIn(in)
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err = cmd.Execute()
	if err == nil {
		t.Fatalf("expected error while executing, got %v", err)
	}
	assert.ErrorContains(t, err, errMsg)
}
