// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
// Code generated by protoc-gen-go. DO NOT EDIT.
// source: service/skippedBlocksmith.proto

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

func init() {
	proto.RegisterFile("service/skippedBlocksmith.proto", fileDescriptor_027d39273f52889b)
}

var fileDescriptor_027d39273f52889b = []byte{
	// 214 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x2f, 0x4e, 0x2d, 0x2a,
	0xcb, 0x4c, 0x4e, 0xd5, 0x2f, 0xce, 0xce, 0x2c, 0x28, 0x48, 0x4d, 0x71, 0xca, 0xc9, 0x4f, 0xce,
	0x2e, 0xce, 0xcd, 0x2c, 0xc9, 0xd0, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x87, 0x2a, 0x90,
	0x92, 0xcd, 0xcd, 0x4f, 0x49, 0xcd, 0xc1, 0xa5, 0x4e, 0x4a, 0x26, 0x3d, 0x3f, 0x3f, 0x3d, 0x27,
	0x55, 0x3f, 0xb1, 0x20, 0x53, 0x3f, 0x31, 0x2f, 0x2f, 0xbf, 0x24, 0xb1, 0x24, 0x33, 0x3f, 0xaf,
	0x18, 0x22, 0x6b, 0xb4, 0x8d, 0x91, 0x4b, 0x32, 0x18, 0x49, 0x67, 0x30, 0x48, 0x67, 0x71, 0x30,
	0xc4, 0x68, 0xa1, 0x19, 0x8c, 0x5c, 0xa2, 0xee, 0xa9, 0x25, 0x98, 0x0a, 0x84, 0x94, 0xf5, 0xc0,
	0xb6, 0xea, 0xa1, 0xc9, 0x82, 0x2d, 0x2e, 0x0e, 0x4a, 0x2d, 0x2c, 0x4d, 0x2d, 0x2e, 0x91, 0x52,
	0xc1, 0xaf, 0xa8, 0xb8, 0x20, 0x3f, 0xaf, 0x38, 0x55, 0xc9, 0xa4, 0xe9, 0xf2, 0x93, 0xc9, 0x4c,
	0x7a, 0x42, 0x3a, 0xfa, 0x65, 0x86, 0x28, 0xbe, 0x80, 0x58, 0xa5, 0x8f, 0xd5, 0x01, 0x4e, 0x3a,
	0x51, 0x5a, 0xe9, 0x99, 0x25, 0x19, 0xa5, 0x49, 0x7a, 0xc9, 0xf9, 0xb9, 0xfa, 0x55, 0xf9, 0xf9,
	0x49, 0xc9, 0x10, 0x52, 0x37, 0x39, 0xbf, 0x28, 0x55, 0x3f, 0x39, 0x3f, 0x37, 0x37, 0x3f, 0x4f,
	0x1f, 0x1a, 0x46, 0x49, 0x6c, 0x60, 0xdf, 0x1a, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0x0c, 0xb8,
	0xf6, 0x0c, 0x56, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// SkippedBlockSmithsServiceClient is the client API for SkippedBlockSmithsService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type SkippedBlockSmithsServiceClient interface {
	GetSkippedBlockSmiths(ctx context.Context, in *model.GetSkippedBlocksmithsRequest, opts ...grpc.CallOption) (*model.GetSkippedBlocksmithsResponse, error)
}

type skippedBlockSmithsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewSkippedBlockSmithsServiceClient(cc grpc.ClientConnInterface) SkippedBlockSmithsServiceClient {
	return &skippedBlockSmithsServiceClient{cc}
}

func (c *skippedBlockSmithsServiceClient) GetSkippedBlockSmiths(ctx context.Context, in *model.GetSkippedBlocksmithsRequest, opts ...grpc.CallOption) (*model.GetSkippedBlocksmithsResponse, error) {
	out := new(model.GetSkippedBlocksmithsResponse)
	err := c.cc.Invoke(ctx, "/service.SkippedBlockSmithsService/GetSkippedBlockSmiths", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SkippedBlockSmithsServiceServer is the server API for SkippedBlockSmithsService service.
type SkippedBlockSmithsServiceServer interface {
	GetSkippedBlockSmiths(context.Context, *model.GetSkippedBlocksmithsRequest) (*model.GetSkippedBlocksmithsResponse, error)
}

// UnimplementedSkippedBlockSmithsServiceServer can be embedded to have forward compatible implementations.
type UnimplementedSkippedBlockSmithsServiceServer struct {
}

func (*UnimplementedSkippedBlockSmithsServiceServer) GetSkippedBlockSmiths(ctx context.Context, req *model.GetSkippedBlocksmithsRequest) (*model.GetSkippedBlocksmithsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSkippedBlockSmiths not implemented")
}

func RegisterSkippedBlockSmithsServiceServer(s *grpc.Server, srv SkippedBlockSmithsServiceServer) {
	s.RegisterService(&_SkippedBlockSmithsService_serviceDesc, srv)
}

func _SkippedBlockSmithsService_GetSkippedBlockSmiths_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetSkippedBlocksmithsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SkippedBlockSmithsServiceServer).GetSkippedBlockSmiths(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.SkippedBlockSmithsService/GetSkippedBlockSmiths",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SkippedBlockSmithsServiceServer).GetSkippedBlockSmiths(ctx, req.(*model.GetSkippedBlocksmithsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _SkippedBlockSmithsService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "service.SkippedBlockSmithsService",
	HandlerType: (*SkippedBlockSmithsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetSkippedBlockSmiths",
			Handler:    _SkippedBlockSmithsService_GetSkippedBlockSmiths_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service/skippedBlocksmith.proto",
}
