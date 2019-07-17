// Code generated by protoc-gen-go. DO NOT EDIT.
// source: service/transaction.proto

package service

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	model "github.com/zoobc/zoobc-core/common/model"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

func init() { proto.RegisterFile("service/transaction.proto", fileDescriptor_e672968ede58c6fc) }

var fileDescriptor_e672968ede58c6fc = []byte{
	// 222 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x2c, 0x4e, 0x2d, 0x2a,
	0xcb, 0x4c, 0x4e, 0xd5, 0x2f, 0x29, 0x4a, 0xcc, 0x2b, 0x4e, 0x4c, 0x2e, 0xc9, 0xcc, 0xcf, 0xd3,
	0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x87, 0x4a, 0x49, 0x89, 0xe7, 0xe6, 0xa7, 0xa4, 0xe6,
	0x60, 0xaa, 0x90, 0x92, 0x49, 0xcf, 0xcf, 0x4f, 0xcf, 0x49, 0xd5, 0x4f, 0x2c, 0xc8, 0xd4, 0x4f,
	0xcc, 0xcb, 0xcb, 0x2f, 0x49, 0x04, 0x49, 0x16, 0x43, 0x64, 0x8d, 0x7e, 0x33, 0x72, 0x09, 0x85,
	0x20, 0xf4, 0x04, 0x43, 0x4c, 0x13, 0xaa, 0xe4, 0xe2, 0x77, 0x4f, 0x2d, 0x41, 0x92, 0x28, 0x16,
	0x92, 0xd5, 0x03, 0xdb, 0xa0, 0x87, 0x26, 0x1e, 0x94, 0x5a, 0x58, 0x9a, 0x5a, 0x5c, 0x22, 0x25,
	0x87, 0x4b, 0xba, 0xb8, 0x20, 0x3f, 0xaf, 0x38, 0x55, 0x49, 0xbd, 0xe9, 0xf2, 0x93, 0xc9, 0x4c,
	0x8a, 0x42, 0xf2, 0xfa, 0x65, 0x86, 0xc8, 0xae, 0xd4, 0x47, 0xb7, 0x27, 0x8b, 0x8b, 0x0f, 0x55,
	0x48, 0x48, 0x06, 0xab, 0xd1, 0x30, 0x8b, 0x85, 0xa0, 0xb2, 0x48, 0x52, 0x4a, 0x6a, 0x60, 0xcb,
	0x14, 0x84, 0xe4, 0xf0, 0x5b, 0xe6, 0xa4, 0x13, 0xa5, 0x95, 0x9e, 0x59, 0x92, 0x51, 0x9a, 0xa4,
	0x97, 0x9c, 0x9f, 0xab, 0x5f, 0x95, 0x9f, 0x9f, 0x94, 0x0c, 0x21, 0x75, 0x93, 0xf3, 0x8b, 0x52,
	0xf5, 0x93, 0xf3, 0x73, 0x73, 0xf3, 0xf3, 0xf4, 0xa1, 0x41, 0x9c, 0xc4, 0x06, 0x0e, 0x32, 0x63,
	0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0x1c, 0xa5, 0xa1, 0x1e, 0x8f, 0x01, 0x00, 0x00,
}

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

// TransactionServiceServer is the server API for TransactionService service.
type TransactionServiceServer interface {
	GetTransactions(context.Context, *model.GetTransactionsRequest) (*model.GetTransactionsResponse, error)
	GetTransaction(context.Context, *model.GetTransactionRequest) (*model.Transaction, error)
}

// UnimplementedTransactionServiceServer can be embedded to have forward compatible implementations.
type UnimplementedTransactionServiceServer struct {
}

func (*UnimplementedTransactionServiceServer) GetTransactions(ctx context.Context, req *model.GetTransactionsRequest) (*model.GetTransactionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTransactions not implemented")
}
func (*UnimplementedTransactionServiceServer) GetTransaction(ctx context.Context, req *model.GetTransactionRequest) (*model.Transaction, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTransaction not implemented")
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
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service/transaction.proto",
}
