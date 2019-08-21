// Code generated by protoc-gen-go. DO NOT EDIT.
// source: service/nodeHardware.proto

package service

import (
	context "context"
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	model "github.com/zoobc/zoobc-core/common/model"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

func init() { proto.RegisterFile("service/nodeHardware.proto", fileDescriptor_a9639805248a2100) }

var fileDescriptor_a9639805248a2100 = []byte{
	// 183 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0xcf, 0xbd, 0x0a, 0xc2, 0x30,
	0x10, 0x07, 0x70, 0xbb, 0x28, 0x74, 0x11, 0xea, 0x22, 0x41, 0x1d, 0x9c, 0x44, 0x34, 0x11, 0x7d,
	0x03, 0x17, 0x9d, 0x1c, 0x14, 0x1c, 0xdc, 0xd2, 0xf4, 0xa8, 0x81, 0x26, 0xff, 0x9a, 0xa4, 0x0a,
	0x3e, 0xbd, 0xd0, 0x76, 0x28, 0xa2, 0xcb, 0x0d, 0xf7, 0xbb, 0xcf, 0x98, 0x79, 0x72, 0x4f, 0xad,
	0x48, 0x58, 0x64, 0x74, 0x94, 0x2e, 0x7b, 0x49, 0x47, 0xbc, 0x74, 0x08, 0x48, 0x06, 0xad, 0xb1,
	0xb1, 0x41, 0x46, 0xc5, 0x8f, 0x12, 0x36, 0xc9, 0x81, 0xbc, 0x20, 0x21, 0x4b, 0x2d, 0xa4, 0xb5,
	0x08, 0x32, 0x68, 0x58, 0xdf, 0xe8, 0xd6, 0xc4, 0xa3, 0x53, 0xa7, 0xe7, 0xd2, 0x8c, 0x4b, 0xae,
	0xf1, 0xf0, 0x40, 0xa1, 0x2b, 0xc9, 0x94, 0xd7, 0x2b, 0xf8, 0x57, 0xfe, 0x4c, 0x8f, 0x8a, 0x7c,
	0x60, 0xb3, 0x7f, 0xec, 0x4b, 0x58, 0x4f, 0xf3, 0xde, 0x22, 0xda, 0x44, 0xfb, 0xd5, 0x6d, 0x99,
	0xeb, 0x70, 0xaf, 0x52, 0xae, 0x60, 0xc4, 0x1b, 0x48, 0x55, 0x13, 0xd7, 0x0a, 0x8e, 0x84, 0x82,
	0x31, 0xb0, 0xa2, 0x7d, 0x2a, 0xed, 0xd7, 0x37, 0xee, 0x3e, 0x01, 0x00, 0x00, 0xff, 0xff, 0x68,
	0x6b, 0xe8, 0x5b, 0x02, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// NodeHardwareServiceClient is the client API for NodeHardwareService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type NodeHardwareServiceClient interface {
	GetNodeHardware(ctx context.Context, opts ...grpc.CallOption) (NodeHardwareService_GetNodeHardwareClient, error)
}

type nodeHardwareServiceClient struct {
	cc *grpc.ClientConn
}

func NewNodeHardwareServiceClient(cc *grpc.ClientConn) NodeHardwareServiceClient {
	return &nodeHardwareServiceClient{cc}
}

func (c *nodeHardwareServiceClient) GetNodeHardware(ctx context.Context, opts ...grpc.CallOption) (NodeHardwareService_GetNodeHardwareClient, error) {
	stream, err := c.cc.NewStream(ctx, &_NodeHardwareService_serviceDesc.Streams[0], "/service.NodeHardwareService/GetNodeHardware", opts...)
	if err != nil {
		return nil, err
	}
	x := &nodeHardwareServiceGetNodeHardwareClient{stream}
	return x, nil
}

type NodeHardwareService_GetNodeHardwareClient interface {
	Send(*model.GetNodeHardwareRequest) error
	Recv() (*model.GetNodeHardwareResponse, error)
	grpc.ClientStream
}

type nodeHardwareServiceGetNodeHardwareClient struct {
	grpc.ClientStream
}

func (x *nodeHardwareServiceGetNodeHardwareClient) Send(m *model.GetNodeHardwareRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *nodeHardwareServiceGetNodeHardwareClient) Recv() (*model.GetNodeHardwareResponse, error) {
	m := new(model.GetNodeHardwareResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// NodeHardwareServiceServer is the server API for NodeHardwareService service.
type NodeHardwareServiceServer interface {
	GetNodeHardware(NodeHardwareService_GetNodeHardwareServer) error
}

func RegisterNodeHardwareServiceServer(s *grpc.Server, srv NodeHardwareServiceServer) {
	s.RegisterService(&_NodeHardwareService_serviceDesc, srv)
}

func _NodeHardwareService_GetNodeHardware_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(NodeHardwareServiceServer).GetNodeHardware(&nodeHardwareServiceGetNodeHardwareServer{stream})
}

type NodeHardwareService_GetNodeHardwareServer interface {
	Send(*model.GetNodeHardwareResponse) error
	Recv() (*model.GetNodeHardwareRequest, error)
	grpc.ServerStream
}

type nodeHardwareServiceGetNodeHardwareServer struct {
	grpc.ServerStream
}

func (x *nodeHardwareServiceGetNodeHardwareServer) Send(m *model.GetNodeHardwareResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *nodeHardwareServiceGetNodeHardwareServer) Recv() (*model.GetNodeHardwareRequest, error) {
	m := new(model.GetNodeHardwareRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _NodeHardwareService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "service.NodeHardwareService",
	HandlerType: (*NodeHardwareServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetNodeHardware",
			Handler:       _NodeHardwareService_GetNodeHardware_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "service/nodeHardware.proto",
}
