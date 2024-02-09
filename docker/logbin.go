package docker

// This files keeps all the functions related to Logbin communication

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/docker/docker/pkg/stdcopy"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/utils"
)

func newLogBin(title string) (LogbinResponse, error) {
	logbinToken, err := utils.GetLogbinToken()
	if err != nil {
		zlog.Sugar().Errorf("unable to fetch logbin token from db: %v", err)
		return LogbinResponse{}, err
	}

	if logbinToken == "" {
		// backward compatibility: machines already onboarded without a logbin auth token
		logbinToken, err = utils.RegisterLogbin(utils.GetMachineUUID(), libp2p.GetP2P().Host.ID().String())
		if err != nil {
			zlog.Sugar().Errorf("unable to register logbin: %v", err)
			return LogbinResponse{}, err
		}
	}

	log := NewLog{
		Title:  title,
		Stdout: "nunet logs - no updates from docker container",
		Stderr: "nunet logs - no updates from docker container",
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
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		zlog.Sugar().Errorf("unable to create log at logbin: %v", err)
		return LogbinResponse{}, err
	}

	if resp.StatusCode != 201 {
		zlog.Sugar().Errorf("unable to create log at logbin - statusCode: %v", resp.Status)
		return LogbinResponse{}, errors.New("unable to create log at logbin - " + resp.Status)
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

func updateLogbin(ctx context.Context, logbinID string, containerID string) error {

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	containerLog, err := GetLogs(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to get logs from container - %v", err)
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
			return fmt.Errorf("unable to marshal logbin append request: %v", err)
		}
		logbinToken, err := utils.GetLogbinToken()
		if err != nil {
			return fmt.Errorf("unable to fetch logbin token from db: %v", err)
		}

		req, err := http.NewRequest(http.MethodPut, "https://log.nunet.io/api/v1/logbin/"+logbinID, bytes.NewBuffer(logAppendJson))
		if err != nil {
			return fmt.Errorf("unable to create http request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Header.Set("Authorization", logbinToken)
		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			return fmt.Errorf("unable to append log at logbin: %v", err)
		}
		zlog.Info(fmt.Sprintf("[logbin]: Resp Code %d:", resp.StatusCode))
	}
	return nil
}
