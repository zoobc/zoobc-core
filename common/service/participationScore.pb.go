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
// source: service/participationScore.proto

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
	proto.RegisterFile("service/participationScore.proto", fileDescriptor_a9b7e07dcf4c0206)
}

var fileDescriptor_a9b7e07dcf4c0206 = []byte{
	// 259 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x28, 0x4e, 0x2d, 0x2a,
	0xcb, 0x4c, 0x4e, 0xd5, 0x2f, 0x48, 0x2c, 0x2a, 0xc9, 0x4c, 0xce, 0x2c, 0x48, 0x2c, 0xc9, 0xcc,
	0xcf, 0x0b, 0x4e, 0xce, 0x2f, 0x4a, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x87, 0xaa,
	0x90, 0x92, 0xcb, 0xcd, 0x4f, 0x49, 0xcd, 0xc1, 0xa9, 0x50, 0x4a, 0x26, 0x3d, 0x3f, 0x3f, 0x3d,
	0x27, 0x55, 0x3f, 0xb1, 0x20, 0x53, 0x3f, 0x31, 0x2f, 0x2f, 0xbf, 0x04, 0xac, 0xa2, 0x18, 0x22,
	0x6b, 0x74, 0x86, 0x89, 0x4b, 0x32, 0x00, 0x43, 0x6b, 0x30, 0xc4, 0x6c, 0xa1, 0x89, 0x8c, 0x5c,
	0x62, 0xee, 0xa9, 0x25, 0x98, 0x0a, 0x8a, 0x85, 0x54, 0xf4, 0xc0, 0xf6, 0xea, 0x61, 0x97, 0x0e,
	0x4a, 0x2d, 0x2c, 0x4d, 0x2d, 0x2e, 0x91, 0x52, 0x25, 0xa0, 0xaa, 0xb8, 0x20, 0x3f, 0xaf, 0x38,
	0x55, 0x49, 0xab, 0xe9, 0xf2, 0x93, 0xc9, 0x4c, 0x2a, 0x42, 0x4a, 0xfa, 0x65, 0x86, 0xfa, 0x89,
	0xc9, 0xc9, 0xf9, 0xa5, 0x79, 0x25, 0xfa, 0x38, 0x2c, 0x9e, 0xc5, 0xc8, 0xa5, 0xec, 0x9e, 0x5a,
	0xe2, 0x93, 0x58, 0x92, 0x5a, 0x8c, 0x45, 0x81, 0x53, 0xa5, 0x5f, 0x7e, 0x4a, 0xaa, 0xa7, 0x8b,
	0x90, 0x21, 0xc2, 0x6a, 0x42, 0x6a, 0x61, 0xae, 0x95, 0x84, 0x6a, 0xc1, 0x54, 0xa9, 0xa4, 0x0a,
	0x76, 0xa1, 0xbc, 0x90, 0x2c, 0xc8, 0x85, 0x98, 0x61, 0xad, 0x9f, 0x03, 0xb6, 0xc2, 0x49, 0x27,
	0x4a, 0x2b, 0x3d, 0xb3, 0x24, 0xa3, 0x34, 0x49, 0x2f, 0x39, 0x3f, 0x57, 0xbf, 0x2a, 0x3f, 0x3f,
	0x29, 0x19, 0x42, 0xea, 0x82, 0x55, 0x25, 0xe7, 0xe7, 0xe6, 0xe6, 0xe7, 0xe9, 0x43, 0xa3, 0x2e,
	0x89, 0x0d, 0x1c, 0x07, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0xbf, 0xca, 0xc4, 0x4e, 0xee,
	0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// ParticipationScoreServiceClient is the client API for ParticipationScoreService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ParticipationScoreServiceClient interface {
	// GetParticipationScores return list of participation score history recorded by node
	GetParticipationScores(ctx context.Context, in *model.GetParticipationScoresRequest, opts ...grpc.CallOption) (*model.GetParticipationScoresResponse, error)
	// GetLatestParticipationScoreByNodeID returns the latest participation score accumulated by node
	GetLatestParticipationScoreByNodeID(ctx context.Context, in *model.GetLatestParticipationScoreByNodeIDRequest, opts ...grpc.CallOption) (*model.ParticipationScore, error)
}

type participationScoreServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewParticipationScoreServiceClient(cc grpc.ClientConnInterface) ParticipationScoreServiceClient {
	return &participationScoreServiceClient{cc}
}

func (c *participationScoreServiceClient) GetParticipationScores(ctx context.Context, in *model.GetParticipationScoresRequest, opts ...grpc.CallOption) (*model.GetParticipationScoresResponse, error) {
	out := new(model.GetParticipationScoresResponse)
	err := c.cc.Invoke(ctx, "/service.ParticipationScoreService/GetParticipationScores", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *participationScoreServiceClient) GetLatestParticipationScoreByNodeID(ctx context.Context, in *model.GetLatestParticipationScoreByNodeIDRequest, opts ...grpc.CallOption) (*model.ParticipationScore, error) {
	out := new(model.ParticipationScore)
	err := c.cc.Invoke(ctx, "/service.ParticipationScoreService/GetLatestParticipationScoreByNodeID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ParticipationScoreServiceServer is the server API for ParticipationScoreService service.
type ParticipationScoreServiceServer interface {
	// GetParticipationScores return list of participation score history recorded by node
	GetParticipationScores(context.Context, *model.GetParticipationScoresRequest) (*model.GetParticipationScoresResponse, error)
	// GetLatestParticipationScoreByNodeID returns the latest participation score accumulated by node
	GetLatestParticipationScoreByNodeID(context.Context, *model.GetLatestParticipationScoreByNodeIDRequest) (*model.ParticipationScore, error)
}

// UnimplementedParticipationScoreServiceServer can be embedded to have forward compatible implementations.
type UnimplementedParticipationScoreServiceServer struct {
}

func (*UnimplementedParticipationScoreServiceServer) GetParticipationScores(ctx context.Context, req *model.GetParticipationScoresRequest) (*model.GetParticipationScoresResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetParticipationScores not implemented")
}
func (*UnimplementedParticipationScoreServiceServer) GetLatestParticipationScoreByNodeID(ctx context.Context, req *model.GetLatestParticipationScoreByNodeIDRequest) (*model.ParticipationScore, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLatestParticipationScoreByNodeID not implemented")
}

func RegisterParticipationScoreServiceServer(s *grpc.Server, srv ParticipationScoreServiceServer) {
	s.RegisterService(&_ParticipationScoreService_serviceDesc, srv)
}

func _ParticipationScoreService_GetParticipationScores_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetParticipationScoresRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ParticipationScoreServiceServer).GetParticipationScores(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.ParticipationScoreService/GetParticipationScores",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ParticipationScoreServiceServer).GetParticipationScores(ctx, req.(*model.GetParticipationScoresRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ParticipationScoreService_GetLatestParticipationScoreByNodeID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(model.GetLatestParticipationScoreByNodeIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ParticipationScoreServiceServer).GetLatestParticipationScoreByNodeID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.ParticipationScoreService/GetLatestParticipationScoreByNodeID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ParticipationScoreServiceServer).GetLatestParticipationScoreByNodeID(ctx, req.(*model.GetLatestParticipationScoreByNodeIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ParticipationScoreService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "service.ParticipationScoreService",
	HandlerType: (*ParticipationScoreServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetParticipationScores",
			Handler:    _ParticipationScoreService_GetParticipationScores_Handler,
		},
		{
			MethodName: "GetLatestParticipationScoreByNodeID",
			Handler:    _ParticipationScoreService_GetLatestParticipationScoreByNodeID_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service/participationScore.proto",
}
