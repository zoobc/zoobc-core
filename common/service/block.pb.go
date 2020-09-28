// Code generated by protoc-gen-go. DO NOT EDIT.
// source: service/block.proto

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

func init() { proto.RegisterFile("service/block.proto", fileDescriptor_9727476210275ad5) }

var fileDescriptor_9727476210275ad5 = []byte{
	// 216 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2e, 0x4e, 0x2d, 0x2a,
	0xcb, 0x4c, 0x4e, 0xd5, 0x4f, 0xca, 0xc9, 0x4f, 0xce, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17,
	0x62, 0x87, 0x0a, 0x4a, 0x09, 0xe6, 0xe6, 0xa7, 0xa4, 0xe6, 0x20, 0xcb, 0x49, 0xc9, 0xa4, 0xe7,
	0xe7, 0xa7, 0xe7, 0xa4, 0xea, 0x27, 0x16, 0x64, 0xea, 0x27, 0xe6, 0xe5, 0xe5, 0x97, 0x24, 0x96,
	0x64, 0xe6, 0xe7, 0x15, 0x43, 0x64, 0x8d, 0x8e, 0x30, 0x72, 0xf1, 0x38, 0x81, 0x54, 0x07, 0x43,
	0x4c, 0x10, 0x8a, 0xe6, 0xe2, 0x74, 0x4f, 0x2d, 0x01, 0x0b, 0x15, 0x0b, 0x89, 0xeb, 0x81, 0xcd,
	0xd3, 0x83, 0x8b, 0x04, 0xa5, 0x16, 0x96, 0xa6, 0x16, 0x97, 0x48, 0x49, 0x60, 0x4a, 0x14, 0x17,
	0xe4, 0xe7, 0x15, 0xa7, 0x2a, 0x49, 0x37, 0x5d, 0x7e, 0x32, 0x99, 0x49, 0x54, 0x48, 0x58, 0xbf,
	0xcc, 0x10, 0xe2, 0x0e, 0x7d, 0x84, 0x79, 0xe1, 0x5c, 0x1c, 0x30, 0x8e, 0x90, 0x18, 0x9a, 0x11,
	0x30, 0xa3, 0xc5, 0x31, 0xc4, 0xa1, 0x26, 0x4b, 0x81, 0x4d, 0x16, 0x11, 0x12, 0xc2, 0x34, 0xd9,
	0x49, 0x27, 0x4a, 0x2b, 0x3d, 0xb3, 0x24, 0xa3, 0x34, 0x49, 0x2f, 0x39, 0x3f, 0x57, 0xbf, 0x2a,
	0x3f, 0x3f, 0x29, 0x19, 0x42, 0xea, 0x26, 0xe7, 0x17, 0xa5, 0xea, 0x27, 0xe7, 0xe7, 0xe6, 0xe6,
	0xe7, 0xe9, 0x43, 0x43, 0x29, 0x89, 0x0d, 0xec, 0x77, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff,
	0x8e, 0xa5, 0xc3, 0x7d, 0x4c, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// BlockServiceClient is the client API for BlockService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type BlockServiceClient interface {
	GetBlocks(ctx context.Context, in *model.GetBlocksRequest, opts ...grpc.CallOption) (*model.GetBlocksResponse, error)
	GetBlock(ctx context.Context, in *model.GetBlockRequest, opts ...grpc.CallOption) (*model.GetBlockResponse, error)
}

type blockServiceClient struct {
	cc *grpc.ClientConn
}

func NewBlockServiceClient(cc *grpc.ClientConn) BlockServiceClient {
	return &blockServiceClient{cc}
}

func (c *blockServiceClient) GetBlocks(ctx context.Context, in *model.GetBlocksRequest, opts ...grpc.CallOption) (*model.GetBlocksResponse, error) {
	out := new(model.GetBlocksResponse)
	err := c.cc.Invoke(ctx, "/service.BlockService/GetBlocks", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blockServiceClient) GetBlock(ctx context.Context, in *model.GetBlockRequest, opts ...grpc.CallOption) (*model.GetBlockResponse, error) {
	out := new(model.GetBlockResponse)
	err := c.cc.Invoke(ctx, "/service.BlockService/GetBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BlockServiceServer is the server API for BlockService service.
type BlockServiceServer interface {
	GetBlocks(context.Context, *model.GetBlocksRequest) (*model.GetBlocksResponse, error)
	GetBlock(context.Context, *model.GetBlockRequest) (*model.GetBlockResponse, error)
}

// UnimplementedBlockServiceServer can be embedded to have forward compatible implementations.
type UnimplementedBlockServiceServer struct {
}

func (*UnimplementedBlockServiceServer) GetBlocks(ctx context.Context, req *model.GetBlocksRequest) (*model.GetBlocksResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlocks not implemented")
}
func (*UnimplementedBlockServiceServer) GetBlock(ctx context.Context, req *model.GetBlockRequest) (*model.GetBlockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlock not implemented")
}

func RegisterBlockServiceServer(s *grpc.Server, srv BlockServiceServer) {
	s.RegisterService(&_BlockService_serviceDesc, srv)
}

func _BlockService_GetBlocks_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetBlocksRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockServiceServer).GetBlocks(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.BlockService/GetBlocks",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockServiceServer).GetBlocks(ctx, req.(*model.GetBlocksRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlockService_GetBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetBlockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockServiceServer).GetBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.BlockService/GetBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockServiceServer).GetBlock(ctx, req.(*model.GetBlockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _BlockService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "service.BlockService",
	HandlerType: (*BlockServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetBlocks",
			Handler:    _BlockService_GetBlocks_Handler,
		},
		{
			MethodName: "GetBlock",
			Handler:    _BlockService_GetBlock_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service/block.proto",
}
