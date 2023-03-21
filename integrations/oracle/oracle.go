package oracle

import (
	context "context"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	oracleAddr = "dev.nunet.io:40052"
)

// WithdrawTokenRequest acts as a middleman between withdraw endpoint handler and Oracle to withdraw token
func WithdrawTokenRequest() (WithdrawResponse, error) {
	conn, err := grpc.Dial(oracleAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return WithdrawResponse{}, err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	oracleClient := NewOracleClient(conn)

	withdrawReq := WithdrawRequest{
		JobStatus:            "finished without errors",
		JobDuration:          5,
		EstimatedJobDuration: 10,
		LogPath:              "https://gist.github.com/santosh/42e86f264c89be54e3351e2373c92edf",
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
	conn, err := grpc.Dial(oracleAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
