// Code generated by protoc-gen-go. DO NOT EDIT.
// source: service/p2pCommunication.proto

package service

import (
	context "context"
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	model "github.com/zoobc/zoobc-core/common/model"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

func init() { proto.RegisterFile("service/p2pCommunication.proto", fileDescriptor_5d547fbc25d9babc) }

var fileDescriptor_5d547fbc25d9babc = []byte{
	// 434 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x54, 0x4f, 0x6f, 0xd3, 0x30,
	0x14, 0xef, 0x69, 0x08, 0x6f, 0x68, 0x9b, 0x25, 0xe8, 0xc8, 0xb4, 0x1d, 0x0a, 0x4c, 0x80, 0x20,
	0x91, 0x0a, 0x12, 0x07, 0x2e, 0xb0, 0x15, 0x45, 0x15, 0x6a, 0x55, 0x15, 0x4e, 0xdc, 0x12, 0xe7,
	0x95, 0x1a, 0x62, 0xbf, 0x10, 0x3b, 0x15, 0xe5, 0xcb, 0x83, 0x92, 0x38, 0x8e, 0x43, 0x53, 0xd8,
	0xa5, 0x87, 0xdf, 0x7f, 0xbb, 0x49, 0xc8, 0xa5, 0x82, 0x7c, 0xc3, 0x19, 0x04, 0xd9, 0x38, 0xbb,
	0x41, 0x21, 0x0a, 0xc9, 0x59, 0xa4, 0x39, 0x4a, 0x3f, 0xcb, 0x51, 0x23, 0xbd, 0x63, 0x78, 0xef,
	0x44, 0x60, 0x02, 0x69, 0x90, 0x01, 0xe4, 0x35, 0xd5, 0x20, 0x12, 0x13, 0x30, 0xc8, 0x69, 0x8d,
	0x80, 0xc8, 0xf4, 0xb6, 0x0b, 0xc5, 0x29, 0xb2, 0xef, 0x06, 0x7a, 0xe0, 0x40, 0x6c, 0x1d, 0x71,
	0x53, 0xe5, 0x0d, 0x6b, 0x5c, 0xe7, 0x91, 0x54, 0x11, 0x6b, 0x37, 0x8c, 0x7f, 0x1f, 0x90, 0x93,
	0xc5, 0x78, 0xd1, 0x99, 0x47, 0xdf, 0x90, 0xc3, 0x10, 0xf4, 0x02, 0x20, 0x9f, 0xca, 0x15, 0xd2,
	0x87, 0x7e, 0xe5, 0xf6, 0x1d, 0x6c, 0x09, 0x3f, 0x0a, 0x50, 0xda, 0x3b, 0x34, 0xd4, 0x1c, 0x13,
	0x18, 0x0d, 0xe8, 0x5b, 0x72, 0x14, 0x82, 0x9e, 0x61, 0x0e, 0xa5, 0x50, 0xd1, 0x23, 0x43, 0x7f,
	0x28, 0x57, 0x7b, 0xe7, 0x6d, 0x8e, 0x95, 0x2c, 0x41, 0x65, 0x28, 0x55, 0x69, 0x7e, 0x4d, 0xee,
	0x7e, 0x02, 0x99, 0xd4, 0xce, 0xa1, 0xd1, 0x5a, 0xa4, 0x69, 0xec, 0x44, 0x8e, 0x06, 0xf4, 0x5d,
	0xed, 0xba, 0x2e, 0x4f, 0xdc, 0x71, 0x55, 0x48, 0xe3, 0x3a, 0xdb, 0x25, 0x6c, 0xef, 0x92, 0x1c,
	0x97, 0xf0, 0xe7, 0xf6, 0x6e, 0xe8, 0x85, 0x23, 0x77, 0xf0, 0x26, 0xed, 0x72, 0x1f, 0x6d, 0x33,
	0x13, 0x72, 0xdf, 0x56, 0x39, 0x0a, 0x45, 0x1f, 0xfd, 0x3d, 0xc4, 0x65, 0x9b, 0xfc, 0xc7, 0xff,
	0x16, 0x39, 0xcb, 0xcf, 0x8c, 0x65, 0xb7, 0xe8, 0xca, 0x64, 0xec, 0x13, 0xec, 0xbb, 0xcf, 0x6f,
	0x64, 0x18, 0x82, 0xbe, 0x29, 0x44, 0x91, 0x46, 0x9a, 0x6f, 0x60, 0xc2, 0x57, 0x2b, 0xce, 0x8a,
	0x54, 0x6f, 0xe9, 0x93, 0xf6, 0xff, 0xeb, 0xe3, 0x9b, 0xc4, 0xab, 0xff, 0xc9, 0xec, 0x7e, 0x45,
	0xbc, 0x52, 0x84, 0x42, 0xa0, 0x9c, 0xf1, 0x14, 0x94, 0x46, 0x09, 0xd5, 0xd2, 0xe9, 0x44, 0xd1,
	0xa7, 0x4e, 0x4e, 0x9f, 0x24, 0xb1, 0x67, 0x78, 0x76, 0x0b, 0xa5, 0x2d, 0xfd, 0x48, 0x8e, 0x43,
	0xd0, 0x73, 0xf8, 0xa9, 0x6d, 0xd3, 0x45, 0xeb, 0x6f, 0xf1, 0x36, 0xbe, 0x79, 0xaa, 0x7a, 0xc2,
	0xde, 0x93, 0x7b, 0xae, 0x49, 0xd1, 0xf3, 0x9e, 0x28, 0x1b, 0x74, 0xea, 0x06, 0xa9, 0x49, 0xa4,
	0xa3, 0xd1, 0xe0, 0xfa, 0xc5, 0x97, 0xe7, 0x5f, 0xb9, 0x5e, 0x17, 0xb1, 0xcf, 0x50, 0x04, 0xbf,
	0x10, 0x63, 0x56, 0xff, 0xbe, 0x64, 0x98, 0x43, 0xc0, 0xaa, 0x03, 0x05, 0xe6, 0x53, 0x11, 0x1f,
	0x54, 0xaf, 0xed, 0xab, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0xb9, 0xeb, 0x94, 0x68, 0x5c, 0x04,
	0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// P2PCommunicationClient is the client API for P2PCommunication service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type P2PCommunicationClient interface {
	GetPeerInfo(ctx context.Context, in *model.GetPeerInfoRequest, opts ...grpc.CallOption) (*model.Node, error)
	GetMorePeers(ctx context.Context, in *model.Empty, opts ...grpc.CallOption) (*model.GetMorePeersResponse, error)
	SendPeers(ctx context.Context, in *model.SendPeersRequest, opts ...grpc.CallOption) (*model.Empty, error)
	SendBlock(ctx context.Context, in *model.SendBlockRequest, opts ...grpc.CallOption) (*model.SendBlockResponse, error)
	SendTransaction(ctx context.Context, in *model.SendTransactionRequest, opts ...grpc.CallOption) (*model.SendTransactionResponse, error)
	SendBlockTransactions(ctx context.Context, in *model.SendBlockTransactionsRequest, opts ...grpc.CallOption) (*model.SendBlockTransactionsResponse, error)
	RequestBlockTransactions(ctx context.Context, in *model.RequestBlockTransactionsRequest, opts ...grpc.CallOption) (*model.Empty, error)
	GetCumulativeDifficulty(ctx context.Context, in *model.GetCumulativeDifficultyRequest, opts ...grpc.CallOption) (*model.GetCumulativeDifficultyResponse, error)
	GetCommonMilestoneBlockIDs(ctx context.Context, in *model.GetCommonMilestoneBlockIdsRequest, opts ...grpc.CallOption) (*model.GetCommonMilestoneBlockIdsResponse, error)
	GetNextBlockIDs(ctx context.Context, in *model.GetNextBlockIdsRequest, opts ...grpc.CallOption) (*model.BlockIdsResponse, error)
	GetNextBlocks(ctx context.Context, in *model.GetNextBlocksRequest, opts ...grpc.CallOption) (*model.BlocksData, error)
}

type p2PCommunicationClient struct {
	cc *grpc.ClientConn
}

func NewP2PCommunicationClient(cc *grpc.ClientConn) P2PCommunicationClient {
	return &p2PCommunicationClient{cc}
}

func (c *p2PCommunicationClient) GetPeerInfo(ctx context.Context, in *model.GetPeerInfoRequest, opts ...grpc.CallOption) (*model.Node, error) {
	out := new(model.Node)
	err := c.cc.Invoke(ctx, "/service.P2PCommunication/GetPeerInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *p2PCommunicationClient) GetMorePeers(ctx context.Context, in *model.Empty, opts ...grpc.CallOption) (*model.GetMorePeersResponse, error) {
	out := new(model.GetMorePeersResponse)
	err := c.cc.Invoke(ctx, "/service.P2PCommunication/GetMorePeers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *p2PCommunicationClient) SendPeers(ctx context.Context, in *model.SendPeersRequest, opts ...grpc.CallOption) (*model.Empty, error) {
	out := new(model.Empty)
	err := c.cc.Invoke(ctx, "/service.P2PCommunication/SendPeers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *p2PCommunicationClient) SendBlock(ctx context.Context, in *model.SendBlockRequest, opts ...grpc.CallOption) (*model.SendBlockResponse, error) {
	out := new(model.SendBlockResponse)
	err := c.cc.Invoke(ctx, "/service.P2PCommunication/SendBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *p2PCommunicationClient) SendTransaction(ctx context.Context, in *model.SendTransactionRequest, opts ...grpc.CallOption) (*model.SendTransactionResponse, error) {
	out := new(model.SendTransactionResponse)
	err := c.cc.Invoke(ctx, "/service.P2PCommunication/SendTransaction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *p2PCommunicationClient) SendBlockTransactions(ctx context.Context, in *model.SendBlockTransactionsRequest, opts ...grpc.CallOption) (*model.SendBlockTransactionsResponse, error) {
	out := new(model.SendBlockTransactionsResponse)
	err := c.cc.Invoke(ctx, "/service.P2PCommunication/SendBlockTransactions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *p2PCommunicationClient) RequestBlockTransactions(ctx context.Context, in *model.RequestBlockTransactionsRequest, opts ...grpc.CallOption) (*model.Empty, error) {
	out := new(model.Empty)
	err := c.cc.Invoke(ctx, "/service.P2PCommunication/RequestBlockTransactions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *p2PCommunicationClient) GetCumulativeDifficulty(ctx context.Context, in *model.GetCumulativeDifficultyRequest, opts ...grpc.CallOption) (*model.GetCumulativeDifficultyResponse, error) {
	out := new(model.GetCumulativeDifficultyResponse)
	err := c.cc.Invoke(ctx, "/service.P2PCommunication/GetCumulativeDifficulty", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *p2PCommunicationClient) GetCommonMilestoneBlockIDs(ctx context.Context, in *model.GetCommonMilestoneBlockIdsRequest, opts ...grpc.CallOption) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	out := new(model.GetCommonMilestoneBlockIdsResponse)
	err := c.cc.Invoke(ctx, "/service.P2PCommunication/GetCommonMilestoneBlockIDs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *p2PCommunicationClient) GetNextBlockIDs(ctx context.Context, in *model.GetNextBlockIdsRequest, opts ...grpc.CallOption) (*model.BlockIdsResponse, error) {
	out := new(model.BlockIdsResponse)
	err := c.cc.Invoke(ctx, "/service.P2PCommunication/GetNextBlockIDs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *p2PCommunicationClient) GetNextBlocks(ctx context.Context, in *model.GetNextBlocksRequest, opts ...grpc.CallOption) (*model.BlocksData, error) {
	out := new(model.BlocksData)
	err := c.cc.Invoke(ctx, "/service.P2PCommunication/GetNextBlocks", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// P2PCommunicationServer is the server API for P2PCommunication service.
type P2PCommunicationServer interface {
	GetPeerInfo(context.Context, *model.GetPeerInfoRequest) (*model.Node, error)
	GetMorePeers(context.Context, *model.Empty) (*model.GetMorePeersResponse, error)
	SendPeers(context.Context, *model.SendPeersRequest) (*model.Empty, error)
	SendBlock(context.Context, *model.SendBlockRequest) (*model.SendBlockResponse, error)
	SendTransaction(context.Context, *model.SendTransactionRequest) (*model.SendTransactionResponse, error)
	SendBlockTransactions(context.Context, *model.SendBlockTransactionsRequest) (*model.SendBlockTransactionsResponse, error)
	RequestBlockTransactions(context.Context, *model.RequestBlockTransactionsRequest) (*model.Empty, error)
	GetCumulativeDifficulty(context.Context, *model.GetCumulativeDifficultyRequest) (*model.GetCumulativeDifficultyResponse, error)
	GetCommonMilestoneBlockIDs(context.Context, *model.GetCommonMilestoneBlockIdsRequest) (*model.GetCommonMilestoneBlockIdsResponse, error)
	GetNextBlockIDs(context.Context, *model.GetNextBlockIdsRequest) (*model.BlockIdsResponse, error)
	GetNextBlocks(context.Context, *model.GetNextBlocksRequest) (*model.BlocksData, error)
}

// UnimplementedP2PCommunicationServer can be embedded to have forward compatible implementations.
type UnimplementedP2PCommunicationServer struct {
}

func (*UnimplementedP2PCommunicationServer) GetPeerInfo(ctx context.Context, req *model.GetPeerInfoRequest) (*model.Node, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPeerInfo not implemented")
}
func (*UnimplementedP2PCommunicationServer) GetMorePeers(ctx context.Context, req *model.Empty) (*model.GetMorePeersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMorePeers not implemented")
}
func (*UnimplementedP2PCommunicationServer) SendPeers(ctx context.Context, req *model.SendPeersRequest) (*model.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendPeers not implemented")
}
func (*UnimplementedP2PCommunicationServer) SendBlock(ctx context.Context, req *model.SendBlockRequest) (*model.SendBlockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendBlock not implemented")
}
func (*UnimplementedP2PCommunicationServer) SendTransaction(ctx context.Context, req *model.SendTransactionRequest) (*model.SendTransactionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendTransaction not implemented")
}
func (*UnimplementedP2PCommunicationServer) SendBlockTransactions(ctx context.Context, req *model.SendBlockTransactionsRequest) (*model.SendBlockTransactionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendBlockTransactions not implemented")
}
func (*UnimplementedP2PCommunicationServer) RequestBlockTransactions(ctx context.Context, req *model.RequestBlockTransactionsRequest) (*model.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestBlockTransactions not implemented")
}
func (*UnimplementedP2PCommunicationServer) GetCumulativeDifficulty(ctx context.Context, req *model.GetCumulativeDifficultyRequest) (*model.GetCumulativeDifficultyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCumulativeDifficulty not implemented")
}
func (*UnimplementedP2PCommunicationServer) GetCommonMilestoneBlockIDs(ctx context.Context, req *model.GetCommonMilestoneBlockIdsRequest) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCommonMilestoneBlockIDs not implemented")
}
func (*UnimplementedP2PCommunicationServer) GetNextBlockIDs(ctx context.Context, req *model.GetNextBlockIdsRequest) (*model.BlockIdsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNextBlockIDs not implemented")
}
func (*UnimplementedP2PCommunicationServer) GetNextBlocks(ctx context.Context, req *model.GetNextBlocksRequest) (*model.BlocksData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNextBlocks not implemented")
}

func RegisterP2PCommunicationServer(s *grpc.Server, srv P2PCommunicationServer) {
	s.RegisterService(&_P2PCommunication_serviceDesc, srv)
}

func _P2PCommunication_GetPeerInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetPeerInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(P2PCommunicationServer).GetPeerInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.P2PCommunication/GetPeerInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(P2PCommunicationServer).GetPeerInfo(ctx, req.(*model.GetPeerInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _P2PCommunication_GetMorePeers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(P2PCommunicationServer).GetMorePeers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.P2PCommunication/GetMorePeers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(P2PCommunicationServer).GetMorePeers(ctx, req.(*model.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _P2PCommunication_SendPeers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.SendPeersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(P2PCommunicationServer).SendPeers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.P2PCommunication/SendPeers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(P2PCommunicationServer).SendPeers(ctx, req.(*model.SendPeersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _P2PCommunication_SendBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.SendBlockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(P2PCommunicationServer).SendBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.P2PCommunication/SendBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(P2PCommunicationServer).SendBlock(ctx, req.(*model.SendBlockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _P2PCommunication_SendTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.SendTransactionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(P2PCommunicationServer).SendTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.P2PCommunication/SendTransaction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(P2PCommunicationServer).SendTransaction(ctx, req.(*model.SendTransactionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _P2PCommunication_SendBlockTransactions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.SendBlockTransactionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(P2PCommunicationServer).SendBlockTransactions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.P2PCommunication/SendBlockTransactions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(P2PCommunicationServer).SendBlockTransactions(ctx, req.(*model.SendBlockTransactionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _P2PCommunication_RequestBlockTransactions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.RequestBlockTransactionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(P2PCommunicationServer).RequestBlockTransactions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.P2PCommunication/RequestBlockTransactions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(P2PCommunicationServer).RequestBlockTransactions(ctx, req.(*model.RequestBlockTransactionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _P2PCommunication_GetCumulativeDifficulty_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetCumulativeDifficultyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(P2PCommunicationServer).GetCumulativeDifficulty(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.P2PCommunication/GetCumulativeDifficulty",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(P2PCommunicationServer).GetCumulativeDifficulty(ctx, req.(*model.GetCumulativeDifficultyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _P2PCommunication_GetCommonMilestoneBlockIDs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetCommonMilestoneBlockIdsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(P2PCommunicationServer).GetCommonMilestoneBlockIDs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.P2PCommunication/GetCommonMilestoneBlockIDs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(P2PCommunicationServer).GetCommonMilestoneBlockIDs(ctx, req.(*model.GetCommonMilestoneBlockIdsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _P2PCommunication_GetNextBlockIDs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetNextBlockIdsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(P2PCommunicationServer).GetNextBlockIDs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.P2PCommunication/GetNextBlockIDs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(P2PCommunicationServer).GetNextBlockIDs(ctx, req.(*model.GetNextBlockIdsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _P2PCommunication_GetNextBlocks_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetNextBlocksRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(P2PCommunicationServer).GetNextBlocks(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.P2PCommunication/GetNextBlocks",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(P2PCommunicationServer).GetNextBlocks(ctx, req.(*model.GetNextBlocksRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _P2PCommunication_serviceDesc = grpc.ServiceDesc{
	ServiceName: "service.P2PCommunication",
	HandlerType: (*P2PCommunicationServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetPeerInfo",
			Handler:    _P2PCommunication_GetPeerInfo_Handler,
		},
		{
			MethodName: "GetMorePeers",
			Handler:    _P2PCommunication_GetMorePeers_Handler,
		},
		{
			MethodName: "SendPeers",
			Handler:    _P2PCommunication_SendPeers_Handler,
		},
		{
			MethodName: "SendBlock",
			Handler:    _P2PCommunication_SendBlock_Handler,
		},
		{
			MethodName: "SendTransaction",
			Handler:    _P2PCommunication_SendTransaction_Handler,
		},
		{
			MethodName: "SendBlockTransactions",
			Handler:    _P2PCommunication_SendBlockTransactions_Handler,
		},
		{
			MethodName: "RequestBlockTransactions",
			Handler:    _P2PCommunication_RequestBlockTransactions_Handler,
		},
		{
			MethodName: "GetCumulativeDifficulty",
			Handler:    _P2PCommunication_GetCumulativeDifficulty_Handler,
		},
		{
			MethodName: "GetCommonMilestoneBlockIDs",
			Handler:    _P2PCommunication_GetCommonMilestoneBlockIDs_Handler,
		},
		{
			MethodName: "GetNextBlockIDs",
			Handler:    _P2PCommunication_GetNextBlockIDs_Handler,
		},
		{
			MethodName: "GetNextBlocks",
			Handler:    _P2PCommunication_GetNextBlocks_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service/p2pCommunication.proto",
}
