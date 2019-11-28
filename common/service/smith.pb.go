// Code generated by protoc-gen-go. DO NOT EDIT.
// source: service/smith.proto

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

func init() { proto.RegisterFile("service/smith.proto", fileDescriptor_277a41428c81ce5f) }

var fileDescriptor_277a41428c81ce5f = []byte{
	// 206 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2e, 0x4e, 0x2d, 0x2a,
	0xcb, 0x4c, 0x4e, 0xd5, 0x2f, 0xce, 0xcd, 0x2c, 0xc9, 0xd0, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17,
	0x62, 0x87, 0x0a, 0x4a, 0x09, 0xe6, 0xe6, 0xa7, 0xa4, 0xe6, 0xe8, 0xa7, 0xe6, 0x16, 0x94, 0x54,
	0x42, 0xe4, 0x60, 0x42, 0x48, 0xca, 0xa5, 0x64, 0xd2, 0xf3, 0xf3, 0xd3, 0x73, 0x52, 0xf5, 0x13,
	0x0b, 0x32, 0xf5, 0x13, 0xf3, 0xf2, 0xf2, 0x4b, 0x12, 0x4b, 0x32, 0xf3, 0xf3, 0x8a, 0x21, 0xb2,
	0x46, 0x05, 0x5c, 0xfc, 0xc1, 0x20, 0xc5, 0x99, 0x79, 0xe9, 0xc1, 0x10, 0x63, 0x85, 0x62, 0xb9,
	0x04, 0xdd, 0x53, 0x4b, 0xe0, 0xa2, 0x25, 0x89, 0x25, 0xa5, 0xc5, 0x42, 0x3c, 0x7a, 0x60, 0x93,
	0xf5, 0x5c, 0x41, 0x96, 0x49, 0x29, 0x40, 0x79, 0x18, 0xea, 0x82, 0x52, 0x8b, 0x0b, 0xf2, 0xf3,
	0x8a, 0x53, 0x95, 0x24, 0x9a, 0x2e, 0x3f, 0x99, 0xcc, 0x24, 0x24, 0x24, 0xa0, 0x5f, 0x66, 0x08,
	0x71, 0x8e, 0x7e, 0x31, 0x58, 0x85, 0x93, 0x4e, 0x94, 0x56, 0x7a, 0x66, 0x49, 0x46, 0x69, 0x92,
	0x5e, 0x72, 0x7e, 0xae, 0x7e, 0x55, 0x7e, 0x7e, 0x52, 0x32, 0x84, 0xd4, 0x4d, 0xce, 0x2f, 0x4a,
	0xd5, 0x4f, 0xce, 0xcf, 0xcd, 0xcd, 0xcf, 0xd3, 0x87, 0xfa, 0x31, 0x89, 0x0d, 0xec, 0x4c, 0x63,
	0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0xfb, 0xa3, 0x0e, 0x9d, 0x0a, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// SmithingServiceClient is the client API for SmithingService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type SmithingServiceClient interface {
	GetSmithingStatus(ctx context.Context, in *model.Empty, opts ...grpc.CallOption) (*model.GetSmithingStatusResponse, error)
}

type smithingServiceClient struct {
	cc *grpc.ClientConn
}

func NewSmithingServiceClient(cc *grpc.ClientConn) SmithingServiceClient {
	return &smithingServiceClient{cc}
}

func (c *smithingServiceClient) GetSmithingStatus(ctx context.Context, in *model.Empty, opts ...grpc.CallOption) (*model.GetSmithingStatusResponse, error) {
	out := new(model.GetSmithingStatusResponse)
	err := c.cc.Invoke(ctx, "/service.SmithingService/GetSmithingStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SmithingServiceServer is the server API for SmithingService service.
type SmithingServiceServer interface {
	GetSmithingStatus(context.Context, *model.Empty) (*model.GetSmithingStatusResponse, error)
}

func RegisterSmithingServiceServer(s *grpc.Server, srv SmithingServiceServer) {
	s.RegisterService(&_SmithingService_serviceDesc, srv)
}

func _SmithingService_GetSmithingStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SmithingServiceServer).GetSmithingStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.SmithingService/GetSmithingStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SmithingServiceServer).GetSmithingStatus(ctx, req.(*model.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _SmithingService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "service.SmithingService",
	HandlerType: (*SmithingServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetSmithingStatus",
			Handler:    _SmithingService_GetSmithingStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service/smith.proto",
}
