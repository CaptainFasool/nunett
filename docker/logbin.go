package docker

// This files keeps all the functions related to Logbin communication

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/docker/docker/pkg/stdcopy"
	"gitlab.com/nunet/device-management-service/utils"
)

func newLogBin(title string) (LogbinResponse, error) {
	logbinToken, err := utils.GetLogbinToken()
	if err != nil {
		zlog.Sugar().Errorf("unable to fetch logbin token from db: %v", err)
		return LogbinResponse{}, err
	}

	log := NewLog{
		Title:  title,
		Stdout: "No updates from docker container to stdout stream",
		Stderr: "No updates from docker container to stderr stream",
	}

	logJson, err := json.Marshal(log)
	if err != nil {
		zlog.Sugar().Errorf("unable to marshal logbin create request: %v", err)
		return LogbinResponse{}, err
	}
	req, err := http.NewRequest(http.MethodPost, "https://log.nunet.io/api/v1/logbin", bytes.NewBuffer(logJson))

	if err != nil {
		zlog.Sugar().Errorf("unable to create http request: %v", err)
		return LogbinResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", logbinToken)
	var client *http.Client
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 201 {
		zlog.Sugar().Errorf("unable to create log at logbin: %v", err)
		return LogbinResponse{}, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		zlog.Sugar().Errorf("unable to read response body from logbin create: %v", err)
		return LogbinResponse{}, err
	}

	logbinResp := LogbinResponse{}
	err = json.Unmarshal(respBody, &logbinResp)
	if err != nil {
		zlog.Sugar().Errorf("unable to unmarshal logbin response: %v", err)
		return LogbinResponse{}, err
	}

	zlog.Info(fmt.Sprintf("[logbin]: RawUrl %s", *&logbinResp.RawUrl))

	return logbinResp, nil
}

func updateLogbin(ctx context.Context, logbinID string, containerID string) {

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	containerLog, err := GetLogs(ctx, containerID)
	if err != nil {
		zlog.Sugar().Errorf("failed to get logs from container - %v", err)
		return
	}
	stdcopy.StdCopy(&stdout, &stderr, containerLog)

	logAppend := LogAppend{}

	if stderr.String() != "" {
		logAppend.Stderr = cleanFlushInfo(&stderr)
	}
	if stdout.String() != "" {
		logAppend.Stdout = cleanFlushInfo(&stdout)
	}

	if logAppend.Stdout != "" || logAppend.Stderr != "" {
		logAppendJson, err := json.Marshal(logAppend)
		if err != nil {
			zlog.Sugar().Errorf("unable to marshal logbin append request: %v", err)
			return
		}
		logbinToken, err := utils.GetLogbinToken()
		if err != nil {
			zlog.Sugar().Errorf("unable to fetch logbin token from db: %v", err)
			return
		}

		req, err := http.NewRequest(http.MethodPut, "https://log.nunet.io/api/v1/logbin/"+logbinID, bytes.NewBuffer(logAppendJson))
		if err != nil {
			zlog.Sugar().Errorf("unable to create http request: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Header.Set("Authorization", logbinToken)
		var client *http.Client
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			zlog.Sugar().Errorf("unable to append log at logbin: %v", err)
			return
		}
		zlog.Info(fmt.Sprintf("[logbin]: Resp Code %d:", resp.StatusCode))
	}
}
