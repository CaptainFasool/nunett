package oracle

import (
	context "context"
	"crypto/tls"
	"strings"
	"time"

	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func getAddress() string {
	channelName := utils.GetChannelName()
	var (
		addr string

		// Oracle Address
		nunetStagingAddr string = "oracle-staging.test.nunet.io:10052"
		nunetTestAddr    string = "oracle-test.test.nunet.io:20052"
		nunetEdgeAddr    string = "oracle-edge.dev.nunet.io:30052"
		nunetTeamAddr    string = "oracle-team.dev.nunet.io:40052"
	)

	if channelName == "nunet-staging" {
		addr = nunetStagingAddr
	} else if channelName == "nunet-test" {
		addr = nunetTestAddr
	} else if channelName == "nunet-edge" {
		addr = nunetEdgeAddr
	} else if channelName == "nunet-team" {
		addr = nunetTeamAddr
	} else {
		addr = nunetTeamAddr
	}

	return addr
}

func getOracleTlsCredentials(address string) credentials.TransportCredentials {
	serverName := strings.Split(address, ":")[0]
	creds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: false,
		ServerName:         serverName,
	})
	return creds
}

// WithdrawTokenRequest acts as a middleman between withdraw endpoint handler and Oracle to withdraw token
func WithdrawTokenRequest(service models.Services) (RewardResponse, error) {
	address := getAddress()
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(getOracleTlsCredentials(address)))
	if err != nil {
		return RewardResponse{}, err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oracleClient := NewOracleClient(conn)

	withdrawReq := RewardRequest{
		JobStatus:            service.JobStatus,
		JobDuration:          service.JobDuration,
		EstimatedJobDuration: service.EstimatedJobDuration,
		LogPath:              service.LogURL,
	}

	zlog.Sugar().Infof("sending withdraw request to oracle")
	res, err := oracleClient.ValidateRewardReq(ctx, &withdrawReq)
	if err != nil {
		zlog.Sugar().Infof("withdraw request failed %v", err)
		return RewardResponse{}, err
	}

	withdrawRes := RewardResponse{
		RewardType: res.GetRewardType(),
	}

	zlog.Sugar().Infof("withdraw response from oracle: %v", withdrawRes)
	return withdrawRes, nil
}

// FundContractRequest is called from the HandleRequestService to cummunicate Oracle for
// Signature and OracleMessage
func FundContractRequest() (FundingResponse, error) {
	address := getAddress()
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(getOracleTlsCredentials(address)))
	if err != nil {
		return FundingResponse{}, err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oracleClient := NewOracleClient(conn)

	zlog.Sugar().Infof("sending funding request to oracle")
	res, err := oracleClient.ValidateFundingReq(ctx, &FundingRequest{})
	if err != nil {
		zlog.Sugar().Infof("funding request failed %v", err)
		return FundingResponse{}, err
	}

	fundingRes := FundingResponse{
		MetadataHash:   res.MetadataHash,
		WithdrawHash:   res.WithdrawHash,
		RefundHash:     res.RefundHash,
		DistributeHash: res.DistributeHash,
	}

	zlog.Sugar().Infof("funding response from oracle: %v", fundingRes)
	return fundingRes, nil
}
