// Code generated by protoc-gen-go. DO NOT EDIT.
// source: service/multiSignature.proto

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

func init() { proto.RegisterFile("service/multiSignature.proto", fileDescriptor_c7c370ee2b80617f) }

var fileDescriptor_c7c370ee2b80617f = []byte{
	// 354 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x93, 0xbd, 0x4a, 0x03, 0x41,
	0x14, 0x85, 0x59, 0x41, 0x85, 0x6d, 0x84, 0x01, 0x2d, 0x96, 0x54, 0xf9, 0xb1, 0x88, 0x71, 0x07,
	0x8d, 0x29, 0x34, 0x55, 0x82, 0x10, 0x2d, 0x84, 0x60, 0xac, 0xec, 0x26, 0xbb, 0xd7, 0xc9, 0xc0,
	0xee, 0xdc, 0x75, 0x66, 0x36, 0x10, 0x4b, 0x1f, 0xc0, 0x42, 0xdf, 0x4c, 0x5f, 0xc1, 0xca, 0x37,
	0xb0, 0x13, 0xf7, 0x07, 0x83, 0xd9, 0xdd, 0xa0, 0xcd, 0x14, 0x73, 0xee, 0x99, 0xf3, 0x1d, 0x86,
	0x6b, 0xd7, 0x34, 0xa8, 0xb9, 0xf0, 0x80, 0x86, 0x71, 0x60, 0xc4, 0x44, 0x70, 0xc9, 0x4c, 0xac,
	0xc0, 0x8d, 0x14, 0x1a, 0x24, 0xdb, 0x99, 0xea, 0x38, 0x21, 0xfa, 0x10, 0x14, 0x0e, 0x39, 0x35,
	0x8e, 0xc8, 0x03, 0xa0, 0x2c, 0x12, 0x94, 0x49, 0x89, 0x86, 0x19, 0x81, 0x52, 0xa7, 0xea, 0xf1,
	0xe7, 0xa6, 0xbd, 0x73, 0xf5, 0x6d, 0xd3, 0x82, 0x4f, 0xd2, 0xd7, 0xc8, 0xb3, 0x65, 0xef, 0x8d,
	0xc0, 0x8c, 0x41, 0xfa, 0x42, 0xf2, 0x1b, 0xc5, 0xa4, 0x66, 0x5e, 0x62, 0x22, 0x4d, 0x37, 0x49,
	0x72, 0x8b, 0xe5, 0x6b, 0xb8, 0x8f, 0x41, 0x1b, 0xa7, 0xb5, 0x66, 0x4a, 0x47, 0x28, 0x35, 0xd4,
	0x0f, 0x1e, 0xdf, 0xde, 0x5f, 0x36, 0x5a, 0xa4, 0x41, 0xe7, 0x47, 0x29, 0xbb, 0x16, 0x9c, 0x96,
	0x24, 0x7f, 0x58, 0x76, 0xa7, 0x50, 0x3a, 0x07, 0xc3, 0x44, 0x30, 0x5c, 0x2c, 0x5d, 0x5d, 0x30,
	0x3d, 0x23, 0x67, 0x55, 0x10, 0x25, 0xa6, 0xbc, 0x40, 0xff, 0x5f, 0xde, 0xac, 0xd6, 0x20, 0xa9,
	0xd5, 0x27, 0xa7, 0xeb, 0x6b, 0x95, 0xb1, 0x3f, 0x59, 0xf6, 0xee, 0x08, 0x4c, 0xfe, 0x31, 0xe9,
	0x7f, 0x5e, 0xca, 0x3b, 0x24, 0x8d, 0x1f, 0xb2, 0x55, 0x35, 0xc7, 0x6f, 0x56, 0x0f, 0x65, 0x9c,
	0xed, 0x84, 0xb3, 0x49, 0xea, 0xbf, 0x39, 0x0b, 0x62, 0x5f, 0x2d, 0x7b, 0x7f, 0x49, 0x19, 0xf8,
	0xbe, 0x02, 0xad, 0x87, 0x8b, 0x31, 0x53, 0x46, 0x78, 0x22, 0x62, 0xd2, 0x64, 0x77, 0xe4, 0x64,
	0x35, 0xbc, 0x62, 0x3c, 0x47, 0xee, 0xfd, 0xd1, 0x95, 0x75, 0xe8, 0x27, 0x1d, 0x7a, 0xa4, 0x5b,
	0xd6, 0xa1, 0xe2, 0x91, 0x61, 0xe7, 0xb6, 0xcd, 0x85, 0x99, 0xc5, 0x53, 0xd7, 0xc3, 0x90, 0x3e,
	0x20, 0x4e, 0xbd, 0xf4, 0x3c, 0xf4, 0x50, 0x01, 0xf5, 0x30, 0x0c, 0x51, 0xd2, 0x6c, 0xc7, 0xa6,
	0x5b, 0xc9, 0xc2, 0x74, 0xbf, 0x02, 0x00, 0x00, 0xff, 0xff, 0x30, 0x22, 0xd9, 0x70, 0x93, 0x03,
	0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MultisigServiceClient is the client API for MultisigService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MultisigServiceClient interface {
	GetPendingTransactions(ctx context.Context, in *model.GetPendingTransactionsRequest, opts ...grpc.CallOption) (*model.GetPendingTransactionsResponse, error)
	GetPendingTransactionDetailByTransactionHash(ctx context.Context, in *model.GetPendingTransactionDetailByTransactionHashRequest, opts ...grpc.CallOption) (*model.GetPendingTransactionDetailByTransactionHashResponse, error)
	GetMultisignatureInfo(ctx context.Context, in *model.GetMultisignatureInfoRequest, opts ...grpc.CallOption) (*model.GetMultisignatureInfoResponse, error)
	GetMultisigAddressByParticipantAddress(ctx context.Context, in *model.GetMultisigAddressByParticipantAddressRequest, opts ...grpc.CallOption) (*model.GetMultisigAddressByParticipantAddressResponse, error)
}

type multisigServiceClient struct {
	cc *grpc.ClientConn
}

func NewMultisigServiceClient(cc *grpc.ClientConn) MultisigServiceClient {
	return &multisigServiceClient{cc}
}

func (c *multisigServiceClient) GetPendingTransactions(ctx context.Context, in *model.GetPendingTransactionsRequest, opts ...grpc.CallOption) (*model.GetPendingTransactionsResponse, error) {
	out := new(model.GetPendingTransactionsResponse)
	err := c.cc.Invoke(ctx, "/service.MultisigService/GetPendingTransactions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *multisigServiceClient) GetPendingTransactionDetailByTransactionHash(ctx context.Context, in *model.GetPendingTransactionDetailByTransactionHashRequest, opts ...grpc.CallOption) (*model.GetPendingTransactionDetailByTransactionHashResponse, error) {
	out := new(model.GetPendingTransactionDetailByTransactionHashResponse)
	err := c.cc.Invoke(ctx, "/service.MultisigService/GetPendingTransactionDetailByTransactionHash", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *multisigServiceClient) GetMultisignatureInfo(ctx context.Context, in *model.GetMultisignatureInfoRequest, opts ...grpc.CallOption) (*model.GetMultisignatureInfoResponse, error) {
	out := new(model.GetMultisignatureInfoResponse)
	err := c.cc.Invoke(ctx, "/service.MultisigService/GetMultisignatureInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *multisigServiceClient) GetMultisigAddressByParticipantAddress(ctx context.Context, in *model.GetMultisigAddressByParticipantAddressRequest, opts ...grpc.CallOption) (*model.GetMultisigAddressByParticipantAddressResponse, error) {
	out := new(model.GetMultisigAddressByParticipantAddressResponse)
	err := c.cc.Invoke(ctx, "/service.MultisigService/GetMultisigAddressByParticipantAddress", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MultisigServiceServer is the server API for MultisigService service.
type MultisigServiceServer interface {
	GetPendingTransactions(context.Context, *model.GetPendingTransactionsRequest) (*model.GetPendingTransactionsResponse, error)
	GetPendingTransactionDetailByTransactionHash(context.Context, *model.GetPendingTransactionDetailByTransactionHashRequest) (*model.GetPendingTransactionDetailByTransactionHashResponse, error)
	GetMultisignatureInfo(context.Context, *model.GetMultisignatureInfoRequest) (*model.GetMultisignatureInfoResponse, error)
	GetMultisigAddressByParticipantAddress(context.Context, *model.GetMultisigAddressByParticipantAddressRequest) (*model.GetMultisigAddressByParticipantAddressResponse, error)
}

// UnimplementedMultisigServiceServer can be embedded to have forward compatible implementations.
type UnimplementedMultisigServiceServer struct {
}

func (*UnimplementedMultisigServiceServer) GetPendingTransactions(ctx context.Context, req *model.GetPendingTransactionsRequest) (*model.GetPendingTransactionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPendingTransactions not implemented")
}
func (*UnimplementedMultisigServiceServer) GetPendingTransactionDetailByTransactionHash(ctx context.Context, req *model.GetPendingTransactionDetailByTransactionHashRequest) (*model.GetPendingTransactionDetailByTransactionHashResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPendingTransactionDetailByTransactionHash not implemented")
}
func (*UnimplementedMultisigServiceServer) GetMultisignatureInfo(ctx context.Context, req *model.GetMultisignatureInfoRequest) (*model.GetMultisignatureInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMultisignatureInfo not implemented")
}
func (*UnimplementedMultisigServiceServer) GetMultisigAddressByParticipantAddress(ctx context.Context, req *model.GetMultisigAddressByParticipantAddressRequest) (*model.GetMultisigAddressByParticipantAddressResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMultisigAddressByParticipantAddress not implemented")
}

func RegisterMultisigServiceServer(s *grpc.Server, srv MultisigServiceServer) {
	s.RegisterService(&_MultisigService_serviceDesc, srv)
}

func _MultisigService_GetPendingTransactions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetPendingTransactionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MultisigServiceServer).GetPendingTransactions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.MultisigService/GetPendingTransactions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MultisigServiceServer).GetPendingTransactions(ctx, req.(*model.GetPendingTransactionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MultisigService_GetPendingTransactionDetailByTransactionHash_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetPendingTransactionDetailByTransactionHashRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MultisigServiceServer).GetPendingTransactionDetailByTransactionHash(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.MultisigService/GetPendingTransactionDetailByTransactionHash",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MultisigServiceServer).GetPendingTransactionDetailByTransactionHash(ctx, req.(*model.GetPendingTransactionDetailByTransactionHashRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MultisigService_GetMultisignatureInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetMultisignatureInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MultisigServiceServer).GetMultisignatureInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.MultisigService/GetMultisignatureInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MultisigServiceServer).GetMultisignatureInfo(ctx, req.(*model.GetMultisignatureInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MultisigService_GetMultisigAddressByParticipantAddress_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetMultisigAddressByParticipantAddressRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MultisigServiceServer).GetMultisigAddressByParticipantAddress(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.MultisigService/GetMultisigAddressByParticipantAddress",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MultisigServiceServer).GetMultisigAddressByParticipantAddress(ctx, req.(*model.GetMultisigAddressByParticipantAddressRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _MultisigService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "service.MultisigService",
	HandlerType: (*MultisigServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetPendingTransactions",
			Handler:    _MultisigService_GetPendingTransactions_Handler,
		},
		{
			MethodName: "GetPendingTransactionDetailByTransactionHash",
			Handler:    _MultisigService_GetPendingTransactionDetailByTransactionHash_Handler,
		},
		{
			MethodName: "GetMultisignatureInfo",
			Handler:    _MultisigService_GetMultisignatureInfo_Handler,
		},
		{
			MethodName: "GetMultisigAddressByParticipantAddress",
			Handler:    _MultisigService_GetMultisigAddressByParticipantAddress_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service/multiSignature.proto",
}
