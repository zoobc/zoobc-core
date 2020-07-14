// Code generated by protoc-gen-go. DO NOT EDIT.
// source: service/p2pCommunication.proto

package service

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	model "github.com/zoobc/zoobc-core/common/model"
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

func init() { proto.RegisterFile("service/p2pCommunication.proto", fileDescriptor_5d547fbc25d9babc) }

var fileDescriptor_5d547fbc25d9babc = []byte{
	// 542 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x54, 0x4f, 0x6f, 0xd3, 0x30,
	0x14, 0xef, 0x09, 0x34, 0x33, 0xb4, 0xcd, 0x1b, 0xb4, 0x4b, 0xb5, 0x49, 0x74, 0x30, 0x01, 0x82,
	0x56, 0x2a, 0xdc, 0xb8, 0xb0, 0xad, 0xac, 0x4c, 0x68, 0x5b, 0x55, 0x38, 0x71, 0x4b, 0x9d, 0x97,
	0xcd, 0x90, 0xf8, 0x05, 0xdb, 0x19, 0x8c, 0x6f, 0xcb, 0x37, 0x41, 0x49, 0x6c, 0xc7, 0xd9, 0xd2,
	0xc1, 0xa5, 0x87, 0xdf, 0xbf, 0xf7, 0xf3, 0x6b, 0x6c, 0xb2, 0xab, 0x40, 0x5e, 0x71, 0x06, 0xa3,
	0x6c, 0x9c, 0x1d, 0x61, 0x9a, 0xe6, 0x82, 0xb3, 0x50, 0x73, 0x14, 0xc3, 0x4c, 0xa2, 0x46, 0x7a,
	0xdf, 0xf0, 0xc1, 0x7a, 0x8a, 0x11, 0x24, 0xa3, 0x0c, 0x40, 0x56, 0x54, 0xb0, 0x51, 0x21, 0x90,
	0x66, 0xfa, 0xba, 0x09, 0x2d, 0x12, 0x64, 0xdf, 0x0d, 0xf4, 0xd8, 0x83, 0xd8, 0x65, 0xc8, 0x4d,
	0x70, 0xd0, 0xad, 0x70, 0x2d, 0x43, 0xa1, 0x42, 0x56, 0x4f, 0x0c, 0x7a, 0x15, 0x11, 0xf3, 0x04,
	0x26, 0xf8, 0x53, 0x24, 0x18, 0x46, 0x86, 0xe9, 0x57, 0x8c, 0xc0, 0x08, 0x0e, 0xa2, 0x48, 0x82,
	0x52, 0x27, 0x22, 0x46, 0x43, 0x6e, 0x9b, 0x7e, 0x12, 0x31, 0x3e, 0x8f, 0xcf, 0x25, 0xbf, 0xb0,
	0xa3, 0xc6, 0x7f, 0x56, 0xc8, 0xfa, 0x6c, 0x3c, 0x6b, 0x1c, 0x8f, 0x86, 0x64, 0x6b, 0x0a, 0xfa,
	0xac, 0xce, 0x82, 0x32, 0x8d, 0x0e, 0x86, 0x65, 0xd0, 0xb0, 0x8d, 0x9c, 0xc3, 0x8f, 0x1c, 0x94,
	0x0e, 0xf6, 0xee, 0xd4, 0xa8, 0x0c, 0x85, 0x82, 0x41, 0x87, 0x7e, 0x24, 0x9b, 0x9f, 0x41, 0x44,
	0x67, 0xcd, 0xbe, 0xf4, 0x89, 0x71, 0xb7, 0x70, 0x76, 0xc0, 0xaa, 0x91, 0x7c, 0x28, 0x76, 0x3b,
	0xe8, 0xd0, 0xb9, 0x2b, 0x3b, 0xf3, 0xcf, 0x77, 0xb3, 0x6c, 0x83, 0xb4, 0x59, 0x5b, 0x46, 0xd3,
	0x20, 0x07, 0x1d, 0x7a, 0x4c, 0x1e, 0x4c, 0x41, 0xcf, 0x00, 0x64, 0xd9, 0x6a, 0xbb, 0x8e, 0xb2,
	0x98, 0x4d, 0x08, 0xda, 0x28, 0x77, 0xca, 0x77, 0x64, 0x75, 0x0a, 0xfa, 0x14, 0x25, 0x14, 0xa4,
	0xa2, 0x8d, 0xee, 0x41, 0xbf, 0xf6, 0x3a, 0x89, 0x67, 0x7e, 0x4b, 0x56, 0x8a, 0x35, 0x54, 0xce,
	0xae, 0xb7, 0x18, 0x23, 0x6c, 0x5f, 0xc7, 0xfb, 0xca, 0x75, 0x58, 0x7c, 0x53, 0x0d, 0x57, 0x89,
	0x58, 0x57, 0xef, 0x36, 0xe1, 0xe6, 0xce, 0xc9, 0x5a, 0x01, 0x7f, 0xa9, 0xbf, 0x3e, 0xba, 0xe3,
	0xc9, 0x3d, 0xdc, 0xa6, 0xed, 0x2e, 0xa3, 0x5d, 0x66, 0x44, 0x1e, 0xb9, 0x51, 0x9e, 0x42, 0xd1,
	0xbd, 0x9b, 0x45, 0x7c, 0xd6, 0xe6, 0x3f, 0xbd, 0x5b, 0xe4, 0x35, 0xef, 0x19, 0xcb, 0xed, 0x41,
	0xfb, 0x26, 0x63, 0x99, 0x60, 0xd9, 0x3e, 0xbf, 0x91, 0xee, 0x14, 0xf4, 0x51, 0x9e, 0xe6, 0x49,
	0xa8, 0xf9, 0x15, 0x4c, 0x78, 0x1c, 0x73, 0x96, 0x27, 0xfa, 0x9a, 0x3e, 0xab, 0xff, 0xbf, 0x36,
	0xde, 0x26, 0xee, 0xff, 0x4b, 0xe6, 0xfa, 0x2b, 0x12, 0x14, 0x22, 0x4c, 0x53, 0x14, 0xa7, 0x3c,
	0x01, 0xa5, 0x51, 0x40, 0xd9, 0xf4, 0x64, 0xa2, 0xe8, 0x73, 0x2f, 0xa7, 0x4d, 0x12, 0xb9, 0x33,
	0xbc, 0xf8, 0x0f, 0xa5, 0x1b, 0xfa, 0x89, 0xac, 0x15, 0x57, 0x04, 0x7e, 0x69, 0x37, 0x69, 0xc7,
	0xbb, 0x3a, 0x0e, 0xaf, 0xe3, 0xed, 0x57, 0xd5, 0x12, 0x76, 0x40, 0x1e, 0xfa, 0x26, 0x45, 0xfb,
	0x2d, 0x51, 0x2e, 0x68, 0xc3, 0x0f, 0x52, 0x93, 0x50, 0x87, 0x83, 0x0e, 0x9d, 0x91, 0x4d, 0xc3,
	0x1f, 0x7b, 0xcf, 0x1c, 0xb5, 0x17, 0xcd, 0x07, 0x6d, 0x4e, 0xbf, 0x95, 0xb3, 0xa5, 0x0e, 0x5f,
	0x7d, 0x7d, 0x79, 0xc1, 0xf5, 0x65, 0xbe, 0x18, 0x32, 0x4c, 0x47, 0xbf, 0x11, 0x17, 0xac, 0xfa,
	0x7d, 0xcd, 0x50, 0xc2, 0x88, 0x95, 0x2b, 0x1a, 0x99, 0xc7, 0x7c, 0x71, 0xaf, 0x7c, 0x18, 0xdf,
	0xfc, 0x0d, 0x00, 0x00, 0xff, 0xff, 0x25, 0x3b, 0xb6, 0xe9, 0xfe, 0x05, 0x00, 0x00,
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
	GetNodeAddressesInfo(ctx context.Context, in *model.GetNodeAddressesInfoRequest, opts ...grpc.CallOption) (*model.GetNodeAddressesInfoResponse, error)
	SendNodeAddressInfo(ctx context.Context, in *model.SendNodeAddressInfoRequest, opts ...grpc.CallOption) (*model.Empty, error)
	GetNodeProofOfOrigin(ctx context.Context, in *model.GetNodeProofOfOriginRequest, opts ...grpc.CallOption) (*model.ProofOfOrigin, error)
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

func (c *p2PCommunicationClient) GetNodeAddressesInfo(ctx context.Context, in *model.GetNodeAddressesInfoRequest, opts ...grpc.CallOption) (*model.GetNodeAddressesInfoResponse, error) {
	out := new(model.GetNodeAddressesInfoResponse)
	err := c.cc.Invoke(ctx, "/service.P2PCommunication/GetNodeAddressesInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *p2PCommunicationClient) SendNodeAddressInfo(ctx context.Context, in *model.SendNodeAddressInfoRequest, opts ...grpc.CallOption) (*model.Empty, error) {
	out := new(model.Empty)
	err := c.cc.Invoke(ctx, "/service.P2PCommunication/SendNodeAddressInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *p2PCommunicationClient) GetNodeProofOfOrigin(ctx context.Context, in *model.GetNodeProofOfOriginRequest, opts ...grpc.CallOption) (*model.ProofOfOrigin, error) {
	out := new(model.ProofOfOrigin)
	err := c.cc.Invoke(ctx, "/service.P2PCommunication/GetNodeProofOfOrigin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
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
	GetNodeAddressesInfo(context.Context, *model.GetNodeAddressesInfoRequest) (*model.GetNodeAddressesInfoResponse, error)
	SendNodeAddressInfo(context.Context, *model.SendNodeAddressInfoRequest) (*model.Empty, error)
	GetNodeProofOfOrigin(context.Context, *model.GetNodeProofOfOriginRequest) (*model.ProofOfOrigin, error)
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

func RegisterP2PCommunicationServer(s *grpc.Server, srv P2PCommunicationServer) {
	s.RegisterService(&_P2PCommunication_serviceDesc, srv)
}

func _P2PCommunication_GetNodeAddressesInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetNodeAddressesInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(P2PCommunicationServer).GetNodeAddressesInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.P2PCommunication/GetNodeAddressesInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(P2PCommunicationServer).GetNodeAddressesInfo(ctx, req.(*model.GetNodeAddressesInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _P2PCommunication_SendNodeAddressInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.SendNodeAddressInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(P2PCommunicationServer).SendNodeAddressInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.P2PCommunication/SendNodeAddressInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(P2PCommunicationServer).SendNodeAddressInfo(ctx, req.(*model.SendNodeAddressInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _P2PCommunication_GetNodeProofOfOrigin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetNodeProofOfOriginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(P2PCommunicationServer).GetNodeProofOfOrigin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.P2PCommunication/GetNodeProofOfOrigin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(P2PCommunicationServer).GetNodeProofOfOrigin(ctx, req.(*model.GetNodeProofOfOriginRequest))
	}
	return interceptor(ctx, in, info, handler)
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
			MethodName: "GetNodeAddressesInfo",
			Handler:    _P2PCommunication_GetNodeAddressesInfo_Handler,
		},
		{
			MethodName: "SendNodeAddressInfo",
			Handler:    _P2PCommunication_SendNodeAddressInfo_Handler,
		},
		{
			MethodName: "GetNodeProofOfOrigin",
			Handler:    _P2PCommunication_GetNodeProofOfOrigin_Handler,
		},
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
