// Code generated by protoc-gen-go. DO NOT EDIT.
// source: service/accountBalance.proto

package service

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	model "github.com/zoobc/zoobc-core/common/model"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

func init() { proto.RegisterFile("service/accountBalance.proto", fileDescriptor_8b38d5f230566dd1) }

var fileDescriptor_8b38d5f230566dd1 = []byte{
	// 228 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x29, 0x4e, 0x2d, 0x2a,
	0xcb, 0x4c, 0x4e, 0xd5, 0x4f, 0x4c, 0x4e, 0xce, 0x2f, 0xcd, 0x2b, 0x71, 0x4a, 0xcc, 0x49, 0xcc,
	0x4b, 0x4e, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x87, 0xca, 0x4a, 0x49, 0xe5, 0xe6,
	0xa7, 0xa4, 0xe6, 0x60, 0x55, 0x24, 0x25, 0x93, 0x9e, 0x9f, 0x9f, 0x9e, 0x93, 0xaa, 0x9f, 0x58,
	0x90, 0xa9, 0x9f, 0x98, 0x97, 0x97, 0x5f, 0x92, 0x58, 0x92, 0x99, 0x9f, 0x57, 0x0c, 0x91, 0x35,
	0x9a, 0xc5, 0xc4, 0x25, 0xea, 0x88, 0xa2, 0x2d, 0x18, 0x62, 0xa6, 0x50, 0x23, 0x23, 0x97, 0x90,
	0x7b, 0x6a, 0x09, 0xaa, 0x64, 0xb1, 0x90, 0x82, 0x1e, 0xd8, 0x2e, 0x3d, 0x4c, 0xa9, 0xa0, 0xd4,
	0xc2, 0xd2, 0xd4, 0xe2, 0x12, 0x29, 0x45, 0x3c, 0x2a, 0x8a, 0x0b, 0xf2, 0xf3, 0x8a, 0x53, 0x95,
	0xd4, 0x9a, 0x2e, 0x3f, 0x99, 0xcc, 0xa4, 0x20, 0x24, 0xa7, 0x5f, 0x66, 0x08, 0x73, 0xb5, 0x3e,
	0x16, 0xcb, 0x6a, 0xb9, 0x04, 0x31, 0x44, 0x85, 0xe4, 0x71, 0x99, 0x0f, 0x73, 0x80, 0x02, 0x6e,
	0x05, 0x50, 0xfb, 0x55, 0xc1, 0xf6, 0xcb, 0x0b, 0xc9, 0xe2, 0xb5, 0xdf, 0x49, 0x27, 0x4a, 0x2b,
	0x3d, 0xb3, 0x24, 0xa3, 0x34, 0x49, 0x2f, 0x39, 0x3f, 0x57, 0xbf, 0x2a, 0x3f, 0x3f, 0x29, 0x19,
	0x42, 0xea, 0x26, 0xe7, 0x17, 0xa5, 0xea, 0x27, 0xe7, 0xe7, 0xe6, 0xe6, 0xe7, 0xe9, 0x43, 0x23,
	0x21, 0x89, 0x0d, 0x1c, 0xa2, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x5c, 0x32, 0x9a, 0x5b,
	0xb4, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// AccountBalanceServiceClient is the client API for AccountBalanceService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type AccountBalanceServiceClient interface {
	GetAccountBalances(ctx context.Context, in *model.GetAccountBalancesRequest, opts ...grpc.CallOption) (*model.GetAccountBalancesResponse, error)
	GetAccountBalance(ctx context.Context, in *model.GetAccountBalanceRequest, opts ...grpc.CallOption) (*model.GetAccountBalanceResponse, error)
}

type accountBalanceServiceClient struct {
	cc *grpc.ClientConn
}

func NewAccountBalanceServiceClient(cc *grpc.ClientConn) AccountBalanceServiceClient {
	return &accountBalanceServiceClient{cc}
}

func (c *accountBalanceServiceClient) GetAccountBalances(ctx context.Context, in *model.GetAccountBalancesRequest, opts ...grpc.CallOption) (*model.GetAccountBalancesResponse, error) {
	out := new(model.GetAccountBalancesResponse)
	err := c.cc.Invoke(ctx, "/service.AccountBalanceService/GetAccountBalances", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountBalanceServiceClient) GetAccountBalance(ctx context.Context, in *model.GetAccountBalanceRequest, opts ...grpc.CallOption) (*model.GetAccountBalanceResponse, error) {
	out := new(model.GetAccountBalanceResponse)
	err := c.cc.Invoke(ctx, "/service.AccountBalanceService/GetAccountBalance", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AccountBalanceServiceServer is the server API for AccountBalanceService service.
type AccountBalanceServiceServer interface {
	GetAccountBalances(context.Context, *model.GetAccountBalancesRequest) (*model.GetAccountBalancesResponse, error)
	GetAccountBalance(context.Context, *model.GetAccountBalanceRequest) (*model.GetAccountBalanceResponse, error)
}

func RegisterAccountBalanceServiceServer(s *grpc.Server, srv AccountBalanceServiceServer) {
	s.RegisterService(&_AccountBalanceService_serviceDesc, srv)
}

func _AccountBalanceService_GetAccountBalances_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetAccountBalancesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountBalanceServiceServer).GetAccountBalances(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.AccountBalanceService/GetAccountBalances",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountBalanceServiceServer).GetAccountBalances(ctx, req.(*model.GetAccountBalancesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountBalanceService_GetAccountBalance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetAccountBalanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountBalanceServiceServer).GetAccountBalance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.AccountBalanceService/GetAccountBalance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountBalanceServiceServer).GetAccountBalance(ctx, req.(*model.GetAccountBalanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _AccountBalanceService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "service.AccountBalanceService",
	HandlerType: (*AccountBalanceServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAccountBalances",
			Handler:    _AccountBalanceService_GetAccountBalances_Handler,
		},
		{
			MethodName: "GetAccountBalance",
			Handler:    _AccountBalanceService_GetAccountBalance_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service/accountBalance.proto",
}
