package oracle

import (
	context "context"
	"time"

	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func getAddress() string {
	channelName := utils.GetChannelName()
	var (
		addr string

		// Oracle Address
		nunetStagingAddr string = "test.nunet.io:10052"
		nunetTestAddr    string = "test.nunet.io:20052"
		nunetTeamAddr    string = "dev.nunet.io:40052"
		nunetEdgeAddr    string = "dev.nunet.io:50052"
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
// WithdrawTokenRequest acts as a middleman between withdraw endpoint handler and Oracle to withdraw token
func WithdrawTokenRequest(service models.Services) (WithdrawResponse, error) {
	conn, err := grpc.Dial(getAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return WithdrawResponse{}, err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	oracleClient := NewOracleClient(conn)

	withdrawReq := WithdrawRequest{
		JobStatus:            service.JobStatus,
		JobDuration:          service.JobDuration,
		EstimatedJobDuration: service.EstimatedJobDuration,
		LogPath:              service.LogURL,
	}

	zlog.Sugar().Info("sending withdraw request to oracle")
	res, err := oracleClient.ValidateWithdrawReq(ctx, &withdrawReq)
	if err != nil {
		zlog.Sugar().Info("withdraw request failed %v", err)
		return WithdrawResponse{}, err
	}

	withdrawRes := WithdrawResponse{
		Signature:     res.GetSignature(),
		OracleMessage: res.GetOracleMessage(),
		RewardType:    res.GetRewardType(),
	}

	zlog.Sugar().Info("withdraw response from oracle: %v", withdrawRes)
	return withdrawRes, nil
}

// FundContractRequest is called from the HandleRequestService to cummunicate Oracle for
// Signature and OracleMessage
func FundContractRequest() (FundingResponse, error) {
	conn, err := grpc.Dial(getAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return FundingResponse{}, err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	oracleClient := NewOracleClient(conn)

	zlog.Sugar().Info("sending funding request to oracle")
	res, err := oracleClient.ValidateFundingReq(ctx, &FundingRequest{})
	if err != nil {
		zlog.Sugar().Info("funding request failed %v", err)
		return FundingResponse{}, err
	}

	fundingRes := FundingResponse{
		Signature:     res.GetSignature(),
		OracleMessage: res.GetOracleMessage(),
	}

	zlog.Sugar().Info("funding response from oracle: %v", fundingRes)
	return fundingRes, nil
}
