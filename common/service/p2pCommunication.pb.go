// Code generated by protoc-gen-go. DO NOT EDIT.
// source: service/p2pCommunication.proto

package service

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	model "github.com/zoobc/zoobc-core/common/model"
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

func init() { proto.RegisterFile("service/p2pCommunication.proto", fileDescriptor_5d547fbc25d9babc) }

var fileDescriptor_5d547fbc25d9babc = []byte{
	// 470 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x54, 0x4f, 0x6f, 0xd3, 0x30,
	0x14, 0xef, 0x89, 0x09, 0x6f, 0x68, 0x9b, 0x11, 0xb4, 0xa4, 0xda, 0x0e, 0x05, 0x26, 0x40, 0xd0,
	0x48, 0x05, 0x89, 0x03, 0x17, 0xd8, 0x0a, 0xd5, 0x84, 0x36, 0x55, 0x85, 0x13, 0x37, 0xd7, 0x79,
	0x65, 0x86, 0xd8, 0x2f, 0xc4, 0xce, 0x60, 0x7c, 0x4b, 0xbe, 0x11, 0x4a, 0xfc, 0x27, 0x2e, 0x4b,
	0x81, 0x4b, 0x0e, 0xbf, 0x7f, 0xef, 0xe7, 0x97, 0xc4, 0xe4, 0x50, 0x43, 0x79, 0x29, 0x38, 0xa4,
	0xc5, 0xa4, 0x38, 0x41, 0x29, 0x2b, 0x25, 0x38, 0x33, 0x02, 0xd5, 0xb8, 0x28, 0xd1, 0x20, 0xdd,
	0x72, 0x7c, 0xb2, 0x27, 0x31, 0x83, 0x3c, 0x2d, 0x00, 0x4a, 0x4b, 0x79, 0x44, 0x61, 0x06, 0x0e,
	0xd9, 0xb7, 0x08, 0xc8, 0xc2, 0x5c, 0xad, 0x43, 0xcb, 0x1c, 0xf9, 0x57, 0x07, 0xdd, 0x8d, 0x20,
	0x7e, 0xc1, 0x84, 0x1b, 0x95, 0xf4, 0x2d, 0x6e, 0x4a, 0xa6, 0x34, 0xe3, 0x6d, 0x87, 0x64, 0x60,
	0x89, 0x95, 0xc8, 0x61, 0x8a, 0xdf, 0x55, 0x8e, 0x2c, 0xb3, 0xcc, 0xe4, 0xd7, 0x16, 0xd9, 0x9b,
	0x4f, 0xe6, 0x6b, 0xc5, 0xe9, 0x4b, 0xb2, 0x3d, 0x03, 0x33, 0x07, 0x28, 0x4f, 0xd5, 0x0a, 0xe9,
	0xbd, 0x71, 0x63, 0x1f, 0x47, 0xd8, 0x02, 0xbe, 0x55, 0xa0, 0x4d, 0xb2, 0xed, 0xa8, 0x73, 0xcc,
	0x60, 0xd4, 0xa3, 0xaf, 0xc8, 0xce, 0x0c, 0xcc, 0x19, 0x96, 0x50, 0x0b, 0x35, 0xdd, 0x71, 0xf4,
	0xdb, 0xfa, 0x3c, 0xc9, 0xb0, 0xcd, 0x09, 0x92, 0x05, 0xe8, 0x02, 0x95, 0xae, 0xcd, 0x2f, 0xc8,
	0xcd, 0x0f, 0xa0, 0x32, 0xeb, 0xec, 0x3b, 0x6d, 0x40, 0xfc, 0xc4, 0xb5, 0xc8, 0x51, 0x8f, 0xbe,
	0xb6, 0xae, 0xe3, 0x7a, 0x17, 0x6b, 0xae, 0x06, 0xf1, 0xae, 0xc1, 0x75, 0x22, 0xcc, 0x5d, 0x90,
	0xdd, 0x1a, 0xfe, 0xd8, 0x6e, 0x8d, 0x1e, 0x44, 0xf2, 0x08, 0xf7, 0x69, 0x87, 0x9b, 0xe8, 0x90,
	0x99, 0x91, 0x3b, 0x61, 0x54, 0xa4, 0xd0, 0xf4, 0xfe, 0x9f, 0x45, 0x62, 0xd6, 0xe7, 0x3f, 0xf8,
	0xbb, 0x28, 0x6a, 0x3e, 0x70, 0x96, 0xeb, 0x83, 0x8e, 0x5c, 0xc6, 0x26, 0xc1, 0xa6, 0x7d, 0x7e,
	0x21, 0xfd, 0x19, 0x98, 0x93, 0x4a, 0x56, 0x39, 0x33, 0xe2, 0x12, 0xa6, 0x62, 0xb5, 0x12, 0xbc,
	0xca, 0xcd, 0x15, 0x7d, 0xd8, 0xbe, 0xbf, 0x2e, 0xde, 0x27, 0x1e, 0xfd, 0x4b, 0x16, 0xfa, 0x6b,
	0x92, 0xd4, 0x22, 0x94, 0x12, 0xd5, 0x99, 0xc8, 0x41, 0x1b, 0x54, 0xd0, 0x34, 0x3d, 0x9d, 0x6a,
	0xfa, 0x28, 0xca, 0xe9, 0x92, 0x64, 0xe1, 0x0c, 0x8f, 0xff, 0x43, 0x19, 0x86, 0xbe, 0x27, 0xbb,
	0x33, 0x30, 0xe7, 0xf0, 0xc3, 0x84, 0x49, 0x07, 0xad, 0xbf, 0xc5, 0xdb, 0x78, 0xff, 0x55, 0x75,
	0x84, 0xbd, 0x21, 0xb7, 0x62, 0x93, 0xa6, 0xc3, 0x8e, 0xa8, 0x10, 0xb4, 0x1f, 0x07, 0xe9, 0x29,
	0x33, 0x6c, 0xd4, 0xa3, 0x73, 0x72, 0xdb, 0xf1, 0xef, 0xa2, 0xdf, 0x93, 0x26, 0x4e, 0x1b, 0x83,
	0x3e, 0x67, 0xd8, 0xc9, 0xf9, 0x52, 0xc7, 0x4f, 0x3f, 0x3d, 0xf9, 0x2c, 0xcc, 0x45, 0xb5, 0x1c,
	0x73, 0x94, 0xe9, 0x4f, 0xc4, 0x25, 0xb7, 0xcf, 0x67, 0x1c, 0x4b, 0x48, 0x79, 0xb3, 0xa2, 0xd4,
	0x5d, 0x4b, 0xcb, 0x1b, 0xcd, 0x45, 0xf0, 0xfc, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0x77, 0x97,
	0x9d, 0xeb, 0xc8, 0x04, 0x00, 0x00,
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
	GetPeerInfo(ctx context.Context, in *model.GetPeerInfoRequest, opts ...grpc.CallOption) (*model.GetPeerInfoResponse, error)
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
	RequestFileDownload(ctx context.Context, in *model.FileDownloadRequest, opts ...grpc.CallOption) (*model.FileDownloadResponse, error)
}

type p2PCommunicationClient struct {
	cc *grpc.ClientConn
}

func NewP2PCommunicationClient(cc *grpc.ClientConn) P2PCommunicationClient {
	return &p2PCommunicationClient{cc}
}

func (c *p2PCommunicationClient) GetPeerInfo(ctx context.Context, in *model.GetPeerInfoRequest, opts ...grpc.CallOption) (*model.GetPeerInfoResponse, error) {
	out := new(model.GetPeerInfoResponse)
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

func (c *p2PCommunicationClient) RequestFileDownload(ctx context.Context, in *model.FileDownloadRequest, opts ...grpc.CallOption) (*model.FileDownloadResponse, error) {
	out := new(model.FileDownloadResponse)
	err := c.cc.Invoke(ctx, "/service.P2PCommunication/RequestFileDownload", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// P2PCommunicationServer is the server API for P2PCommunication service.
type P2PCommunicationServer interface {
	GetPeerInfo(context.Context, *model.GetPeerInfoRequest) (*model.GetPeerInfoResponse, error)
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
	RequestFileDownload(context.Context, *model.FileDownloadRequest) (*model.FileDownloadResponse, error)
}

// UnimplementedP2PCommunicationServer can be embedded to have forward compatible implementations.
type UnimplementedP2PCommunicationServer struct {
}

func (*UnimplementedP2PCommunicationServer) GetPeerInfo(ctx context.Context, req *model.GetPeerInfoRequest) (*model.GetPeerInfoResponse, error) {
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
func (*UnimplementedP2PCommunicationServer) RequestFileDownload(ctx context.Context, req *model.FileDownloadRequest) (*model.FileDownloadResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestFileDownload not implemented")
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

func _P2PCommunication_RequestFileDownload_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.FileDownloadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(P2PCommunicationServer).RequestFileDownload(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.P2PCommunication/RequestFileDownload",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(P2PCommunicationServer).RequestFileDownload(ctx, req.(*model.FileDownloadRequest))
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
		{
			MethodName: "RequestFileDownload",
			Handler:    _P2PCommunication_RequestFileDownload_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service/p2pCommunication.proto",
}
