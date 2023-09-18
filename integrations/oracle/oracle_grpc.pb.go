// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: specs/compute-api-spec/oracle.proto

package oracle

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// OracleClient is the client API for Oracle service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type OracleClient interface {
	ValidateFundingReq(ctx context.Context, in *FundingRequest, opts ...grpc.CallOption) (*FundingResponse, error)
	ValidateRewardReq(ctx context.Context, in *RewardRequest, opts ...grpc.CallOption) (*RewardResponse, error)
}

type oracleClient struct {
	cc grpc.ClientConnInterface
}

func NewOracleClient(cc grpc.ClientConnInterface) OracleClient {
	return &oracleClient{cc}
}

func (c *oracleClient) ValidateFundingReq(ctx context.Context, in *FundingRequest, opts ...grpc.CallOption) (*FundingResponse, error) {
	out := new(FundingResponse)
	err := c.cc.Invoke(ctx, "/oracle.Oracle/ValidateFundingReq", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *oracleClient) ValidateRewardReq(ctx context.Context, in *RewardRequest, opts ...grpc.CallOption) (*RewardResponse, error) {
	out := new(RewardResponse)
	err := c.cc.Invoke(ctx, "/oracle.Oracle/ValidateRewardReq", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OracleServer is the server API for Oracle service.
// All implementations must embed UnimplementedOracleServer
// for forward compatibility
type OracleServer interface {
	ValidateFundingReq(context.Context, *FundingRequest) (*FundingResponse, error)
	ValidateRewardReq(context.Context, *RewardRequest) (*RewardResponse, error)
	mustEmbedUnimplementedOracleServer()
}

// UnimplementedOracleServer must be embedded to have forward compatible implementations.
type UnimplementedOracleServer struct {
}

func (UnimplementedOracleServer) ValidateFundingReq(context.Context, *FundingRequest) (*FundingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateFundingReq not implemented")
}
func (UnimplementedOracleServer) ValidateRewardReq(context.Context, *RewardRequest) (*RewardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateRewardReq not implemented")
}
func (UnimplementedOracleServer) mustEmbedUnimplementedOracleServer() {}

// UnsafeOracleServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to OracleServer will
// result in compilation errors.
type UnsafeOracleServer interface {
	mustEmbedUnimplementedOracleServer()
}

func RegisterOracleServer(s grpc.ServiceRegistrar, srv OracleServer) {
	s.RegisterService(&Oracle_ServiceDesc, srv)
}

func _Oracle_ValidateFundingReq_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FundingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OracleServer).ValidateFundingReq(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/oracle.Oracle/ValidateFundingReq",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OracleServer).ValidateFundingReq(ctx, req.(*FundingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Oracle_ValidateRewardReq_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RewardRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OracleServer).ValidateRewardReq(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/oracle.Oracle/ValidateRewardReq",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OracleServer).ValidateRewardReq(ctx, req.(*RewardRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Oracle_ServiceDesc is the grpc.ServiceDesc for Oracle service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Oracle_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "oracle.Oracle",
	HandlerType: (*OracleServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ValidateFundingReq",
			Handler:    _Oracle_ValidateFundingReq_Handler,
		},
		{
			MethodName: "ValidateRewardReq",
			Handler:    _Oracle_ValidateRewardReq_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "specs/compute-api-spec/oracle.proto",
}
