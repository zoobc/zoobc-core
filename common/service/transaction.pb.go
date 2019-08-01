// Code generated by protoc-gen-go. DO NOT EDIT.
// source: service/transaction.proto

package service // import "github.com/zoobc/zoobc-core/common/service"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import model "github.com/zoobc/zoobc-core/common/model"
import _ "google.golang.org/genproto/googleapis/api/annotations"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// TransactionServiceClient is the client API for TransactionService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TransactionServiceClient interface {
	GetTransactions(ctx context.Context, in *model.GetTransactionsRequest, opts ...grpc.CallOption) (*model.GetTransactionsResponse, error)
	GetTransaction(ctx context.Context, in *model.GetTransactionRequest, opts ...grpc.CallOption) (*model.Transaction, error)
	PostTransaction(ctx context.Context, in *model.PostTransactionRequest, opts ...grpc.CallOption) (*model.PostTransactionResponse, error)
}

type transactionServiceClient struct {
	cc *grpc.ClientConn
}

func NewTransactionServiceClient(cc *grpc.ClientConn) TransactionServiceClient {
	return &transactionServiceClient{cc}
}

func (c *transactionServiceClient) GetTransactions(ctx context.Context, in *model.GetTransactionsRequest, opts ...grpc.CallOption) (*model.GetTransactionsResponse, error) {
	out := new(model.GetTransactionsResponse)
	err := c.cc.Invoke(ctx, "/service.TransactionService/GetTransactions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *transactionServiceClient) GetTransaction(ctx context.Context, in *model.GetTransactionRequest, opts ...grpc.CallOption) (*model.Transaction, error) {
	out := new(model.Transaction)
	err := c.cc.Invoke(ctx, "/service.TransactionService/GetTransaction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *transactionServiceClient) PostTransaction(ctx context.Context, in *model.PostTransactionRequest, opts ...grpc.CallOption) (*model.PostTransactionResponse, error) {
	out := new(model.PostTransactionResponse)
	err := c.cc.Invoke(ctx, "/service.TransactionService/PostTransaction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TransactionServiceServer is the server API for TransactionService service.
type TransactionServiceServer interface {
	GetTransactions(context.Context, *model.GetTransactionsRequest) (*model.GetTransactionsResponse, error)
	GetTransaction(context.Context, *model.GetTransactionRequest) (*model.Transaction, error)
	PostTransaction(context.Context, *model.PostTransactionRequest) (*model.PostTransactionResponse, error)
}

func RegisterTransactionServiceServer(s *grpc.Server, srv TransactionServiceServer) {
	s.RegisterService(&_TransactionService_serviceDesc, srv)
}

func _TransactionService_GetTransactions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetTransactionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TransactionServiceServer).GetTransactions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.TransactionService/GetTransactions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TransactionServiceServer).GetTransactions(ctx, req.(*model.GetTransactionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TransactionService_GetTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetTransactionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TransactionServiceServer).GetTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.TransactionService/GetTransaction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TransactionServiceServer).GetTransaction(ctx, req.(*model.GetTransactionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TransactionService_PostTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.PostTransactionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TransactionServiceServer).PostTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.TransactionService/PostTransaction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TransactionServiceServer).PostTransaction(ctx, req.(*model.PostTransactionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _TransactionService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "service.TransactionService",
	HandlerType: (*TransactionServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetTransactions",
			Handler:    _TransactionService_GetTransactions_Handler,
		},
		{
			MethodName: "GetTransaction",
			Handler:    _TransactionService_GetTransaction_Handler,
		},
		{
			MethodName: "PostTransaction",
			Handler:    _TransactionService_PostTransaction_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service/transaction.proto",
}

func init() {
	proto.RegisterFile("service/transaction.proto", fileDescriptor_transaction_0e75a8e7f913159d)
}

var fileDescriptor_transaction_0e75a8e7f913159d = []byte{
	// 245 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x2c, 0x4e, 0x2d, 0x2a,
	0xcb, 0x4c, 0x4e, 0xd5, 0x2f, 0x29, 0x4a, 0xcc, 0x2b, 0x4e, 0x4c, 0x2e, 0xc9, 0xcc, 0xcf, 0xd3,
	0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x87, 0x4a, 0x49, 0x89, 0xe7, 0xe6, 0xa7, 0xa4, 0xe6,
	0x60, 0xaa, 0x90, 0x92, 0x49, 0xcf, 0xcf, 0x4f, 0xcf, 0x49, 0xd5, 0x4f, 0x2c, 0xc8, 0xd4, 0x4f,
	0xcc, 0xcb, 0xcb, 0x2f, 0x49, 0x04, 0x49, 0x16, 0x43, 0x64, 0x8d, 0xbe, 0x31, 0x71, 0x09, 0x85,
	0x20, 0xf4, 0x04, 0x43, 0x4c, 0x13, 0xaa, 0xe4, 0xe2, 0x77, 0x4f, 0x2d, 0x41, 0x92, 0x28, 0x16,
	0x92, 0xd5, 0x03, 0xdb, 0xa0, 0x87, 0x26, 0x1e, 0x94, 0x5a, 0x58, 0x9a, 0x5a, 0x5c, 0x22, 0x25,
	0x87, 0x4b, 0xba, 0xb8, 0x20, 0x3f, 0xaf, 0x38, 0x55, 0x49, 0xbd, 0xe9, 0xf2, 0x93, 0xc9, 0x4c,
	0x8a, 0x42, 0xf2, 0xfa, 0x65, 0x86, 0xc8, 0xae, 0xd4, 0x47, 0xb7, 0x27, 0x8b, 0x8b, 0x0f, 0x55,
	0x48, 0x48, 0x06, 0xab, 0xd1, 0x30, 0x8b, 0x85, 0xa0, 0xb2, 0x48, 0x52, 0x4a, 0x6a, 0x60, 0xcb,
	0x14, 0x84, 0xe4, 0xf0, 0x5b, 0x06, 0xf2, 0x66, 0x40, 0x7e, 0x31, 0x8a, 0x10, 0xcc, 0x9b, 0x68,
	0xe2, 0xe8, 0xde, 0xc4, 0x90, 0x46, 0xf5, 0xa6, 0x12, 0x86, 0x37, 0xd1, 0x34, 0x38, 0xe9, 0x44,
	0x69, 0xa5, 0x67, 0x96, 0x64, 0x94, 0x26, 0xe9, 0x25, 0xe7, 0xe7, 0xea, 0x57, 0xe5, 0xe7, 0x27,
	0x25, 0x43, 0x48, 0xdd, 0xe4, 0xfc, 0xa2, 0x54, 0xfd, 0xe4, 0xfc, 0xdc, 0xdc, 0xfc, 0x3c, 0x7d,
	0x68, 0xec, 0x26, 0xb1, 0x81, 0x63, 0xcb, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0x9a, 0x04, 0xcf,
	0x12, 0x0a, 0x02, 0x00, 0x00,
}
