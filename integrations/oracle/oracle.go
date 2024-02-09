package oracle

import (
	context "context"
	"crypto/tls"
	"strings"
	"time"

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

type OracleInterface interface {
	WithdrawTokenRequest(req *RewardRequest) (*RewardResponse, error)
}

var Oracle OracleInterface = &nunetOracle{}

type nunetOracle struct{}

// WithdrawTokenRequest acts as a middleman between withdraw endpoint handler and Oracle to withdraw token
func (a *nunetOracle) WithdrawTokenRequest(rewardReq *RewardRequest) (*RewardResponse, error) {
	address := getAddress()
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(getOracleTlsCredentials(address)))
	if err != nil {
		return &RewardResponse{}, err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oracleClient := NewOracleClient(conn)

	zlog.Sugar().Infof("sending withdraw request to oracle")
	res, err := oracleClient.ValidateRewardReq(ctx, rewardReq)
	if err != nil {
		zlog.Sugar().Infof("withdraw request failed %v", err)
		return &RewardResponse{}, err
	}

	rewardRes := &RewardResponse{
		RewardType:        res.GetRewardType(),
		SignatureDatum:    res.GetSignatureDatum(),
		MessageHashDatum:  res.GetMessageHashDatum(),
		Datum:             res.GetDatum(),
		SignatureAction:   res.GetSignatureAction(),
		MessageHashAction: res.GetMessageHashAction(),
		Action:            res.GetAction(),
	}

	zlog.Sugar().Infof("withdraw response from oracle: %v", rewardRes)
	return rewardRes, nil
}

// FundContractRequest is called from the HandleRequestService to cummunicate Oracle for
// MetadataHash, WithdrawHash, RefundHash, DistributeHash
func FundContractRequest(fundingReq *FundingRequest) (*FundingResponse, error) {
	address := getAddress()
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(getOracleTlsCredentials(address)))
	if err != nil {
		return &FundingResponse{}, err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oracleClient := NewOracleClient(conn)

	zlog.Sugar().Infof("sending funding request to oracle")
	res, err := oracleClient.ValidateFundingReq(ctx, fundingReq)
	if err != nil {
		zlog.Sugar().Infof("funding request failed %v", err)
		return &FundingResponse{}, err
	}

	fundingRes := &FundingResponse{
		MetadataHash:      res.GetMetadataHash(),
		WithdrawHash:      res.GetWithdrawHash(),
		RefundHash:        res.GetRefundHash(),
		Distribute_50Hash: res.GetDistribute_50Hash(),
		Distribute_75Hash: res.GetDistribute_75Hash(),
	}

	zlog.Sugar().Infof("funding response from oracle: %v", fundingRes)
	return fundingRes, nil
}
