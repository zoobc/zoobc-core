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
// source: model/participationScore.proto

package model

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

type ParticipationScore struct {
	NodeID               int64    `protobuf:"varint,1,opt,name=NodeID,proto3" json:"NodeID,omitempty"`
	Score                int64    `protobuf:"varint,2,opt,name=Score,proto3" json:"Score,omitempty"`
	Latest               bool     `protobuf:"varint,3,opt,name=Latest,proto3" json:"Latest,omitempty"`
	Height               uint32   `protobuf:"varint,4,opt,name=Height,proto3" json:"Height,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ParticipationScore) Reset()         { *m = ParticipationScore{} }
func (m *ParticipationScore) String() string { return proto.CompactTextString(m) }
func (*ParticipationScore) ProtoMessage()    {}
func (*ParticipationScore) Descriptor() ([]byte, []int) {
	return fileDescriptor_4eacd5198105d193, []int{0}
}

func (m *ParticipationScore) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ParticipationScore.Unmarshal(m, b)
}
func (m *ParticipationScore) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ParticipationScore.Marshal(b, m, deterministic)
}
func (m *ParticipationScore) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ParticipationScore.Merge(m, src)
}
func (m *ParticipationScore) XXX_Size() int {
	return xxx_messageInfo_ParticipationScore.Size(m)
}
func (m *ParticipationScore) XXX_DiscardUnknown() {
	xxx_messageInfo_ParticipationScore.DiscardUnknown(m)
}

var xxx_messageInfo_ParticipationScore proto.InternalMessageInfo

func (m *ParticipationScore) GetNodeID() int64 {
	if m != nil {
		return m.NodeID
	}
	return 0
}

func (m *ParticipationScore) GetScore() int64 {
	if m != nil {
		return m.Score
	}
	return 0
}

func (m *ParticipationScore) GetLatest() bool {
	if m != nil {
		return m.Latest
	}
	return false
}

func (m *ParticipationScore) GetHeight() uint32 {
	if m != nil {
		return m.Height
	}
	return 0
}

type GetParticipationScoresRequest struct {
	FromHeight           uint32   `protobuf:"varint,1,opt,name=FromHeight,proto3" json:"FromHeight,omitempty"`
	ToHeight             uint32   `protobuf:"varint,2,opt,name=ToHeight,proto3" json:"ToHeight,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetParticipationScoresRequest) Reset()         { *m = GetParticipationScoresRequest{} }
func (m *GetParticipationScoresRequest) String() string { return proto.CompactTextString(m) }
func (*GetParticipationScoresRequest) ProtoMessage()    {}
func (*GetParticipationScoresRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_4eacd5198105d193, []int{1}
}

func (m *GetParticipationScoresRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetParticipationScoresRequest.Unmarshal(m, b)
}
func (m *GetParticipationScoresRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetParticipationScoresRequest.Marshal(b, m, deterministic)
}
func (m *GetParticipationScoresRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetParticipationScoresRequest.Merge(m, src)
}
func (m *GetParticipationScoresRequest) XXX_Size() int {
	return xxx_messageInfo_GetParticipationScoresRequest.Size(m)
}
func (m *GetParticipationScoresRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetParticipationScoresRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetParticipationScoresRequest proto.InternalMessageInfo

func (m *GetParticipationScoresRequest) GetFromHeight() uint32 {
	if m != nil {
		return m.FromHeight
	}
	return 0
}

func (m *GetParticipationScoresRequest) GetToHeight() uint32 {
	if m != nil {
		return m.ToHeight
	}
	return 0
}

type GetParticipationScoresResponse struct {
	ParticipationScores  []*ParticipationScore `protobuf:"bytes,1,rep,name=ParticipationScores,proto3" json:"ParticipationScores,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *GetParticipationScoresResponse) Reset()         { *m = GetParticipationScoresResponse{} }
func (m *GetParticipationScoresResponse) String() string { return proto.CompactTextString(m) }
func (*GetParticipationScoresResponse) ProtoMessage()    {}
func (*GetParticipationScoresResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_4eacd5198105d193, []int{2}
}

func (m *GetParticipationScoresResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetParticipationScoresResponse.Unmarshal(m, b)
}
func (m *GetParticipationScoresResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetParticipationScoresResponse.Marshal(b, m, deterministic)
}
func (m *GetParticipationScoresResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetParticipationScoresResponse.Merge(m, src)
}
func (m *GetParticipationScoresResponse) XXX_Size() int {
	return xxx_messageInfo_GetParticipationScoresResponse.Size(m)
}
func (m *GetParticipationScoresResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetParticipationScoresResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetParticipationScoresResponse proto.InternalMessageInfo

func (m *GetParticipationScoresResponse) GetParticipationScores() []*ParticipationScore {
	if m != nil {
		return m.ParticipationScores
	}
	return nil
}

type GetLatestParticipationScoreByNodeIDRequest struct {
	NodeID               int64    `protobuf:"varint,1,opt,name=NodeID,proto3" json:"NodeID,omitempty"`
	NodeAddress          string   `protobuf:"bytes,2,opt,name=NodeAddress,proto3" json:"NodeAddress,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetLatestParticipationScoreByNodeIDRequest) Reset() {
	*m = GetLatestParticipationScoreByNodeIDRequest{}
}
func (m *GetLatestParticipationScoreByNodeIDRequest) String() string {
	return proto.CompactTextString(m)
}
func (*GetLatestParticipationScoreByNodeIDRequest) ProtoMessage() {}
func (*GetLatestParticipationScoreByNodeIDRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_4eacd5198105d193, []int{3}
}

func (m *GetLatestParticipationScoreByNodeIDRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetLatestParticipationScoreByNodeIDRequest.Unmarshal(m, b)
}
func (m *GetLatestParticipationScoreByNodeIDRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetLatestParticipationScoreByNodeIDRequest.Marshal(b, m, deterministic)
}
func (m *GetLatestParticipationScoreByNodeIDRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetLatestParticipationScoreByNodeIDRequest.Merge(m, src)
}
func (m *GetLatestParticipationScoreByNodeIDRequest) XXX_Size() int {
	return xxx_messageInfo_GetLatestParticipationScoreByNodeIDRequest.Size(m)
}
func (m *GetLatestParticipationScoreByNodeIDRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetLatestParticipationScoreByNodeIDRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetLatestParticipationScoreByNodeIDRequest proto.InternalMessageInfo

func (m *GetLatestParticipationScoreByNodeIDRequest) GetNodeID() int64 {
	if m != nil {
		return m.NodeID
	}
	return 0
}

func (m *GetLatestParticipationScoreByNodeIDRequest) GetNodeAddress() string {
	if m != nil {
		return m.NodeAddress
	}
	return ""
}

func init() {
	proto.RegisterType((*ParticipationScore)(nil), "model.ParticipationScore")
	proto.RegisterType((*GetParticipationScoresRequest)(nil), "model.GetParticipationScoresRequest")
	proto.RegisterType((*GetParticipationScoresResponse)(nil), "model.GetParticipationScoresResponse")
	proto.RegisterType((*GetLatestParticipationScoreByNodeIDRequest)(nil), "model.GetLatestParticipationScoreByNodeIDRequest")
}

func init() {
	proto.RegisterFile("model/participationScore.proto", fileDescriptor_4eacd5198105d193)
}

var fileDescriptor_4eacd5198105d193 = []byte{
	// 286 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x91, 0x4f, 0x4b, 0xc3, 0x40,
	0x10, 0xc5, 0xd9, 0xc6, 0x86, 0x3a, 0xc5, 0xcb, 0x0a, 0xb2, 0x16, 0x0c, 0x21, 0xa7, 0x50, 0x30,
	0x11, 0xfd, 0x04, 0x06, 0xb1, 0x8a, 0x22, 0xb2, 0x7a, 0xd2, 0x53, 0xfe, 0x0c, 0x6d, 0xc4, 0xcd,
	0xc4, 0xec, 0xf6, 0x60, 0x3f, 0xbd, 0x64, 0x13, 0x4b, 0x21, 0xd5, 0x4b, 0xc8, 0xbc, 0xdf, 0xbc,
	0x9d, 0x9d, 0xb7, 0xe0, 0x29, 0x2a, 0xf0, 0x33, 0xae, 0xd3, 0xc6, 0x94, 0x79, 0x59, 0xa7, 0xa6,
	0xa4, 0xea, 0x25, 0xa7, 0x06, 0xa3, 0xba, 0x21, 0x43, 0x7c, 0x6c, 0x79, 0xb0, 0x01, 0xfe, 0x3c,
	0x68, 0xe1, 0x33, 0x70, 0x9f, 0xa8, 0xc0, 0xfb, 0x1b, 0xc1, 0x7c, 0x16, 0x3a, 0xc9, 0xe8, 0x82,
	0xc9, 0x5e, 0xe1, 0x02, 0xc6, 0xb6, 0x49, 0x8c, 0xb6, 0xa8, 0x13, 0xf8, 0x09, 0xb8, 0x8f, 0xa9,
	0x41, 0x6d, 0x84, 0xe3, 0xb3, 0x70, 0x22, 0xfb, 0xaa, 0xd5, 0xef, 0xb0, 0x5c, 0xae, 0x8c, 0x38,
	0xf0, 0x59, 0x78, 0x24, 0xfb, 0x2a, 0x78, 0x87, 0xb3, 0x05, 0x9a, 0xe1, 0x78, 0x2d, 0xf1, 0x6b,
	0xdd, 0x1a, 0x3d, 0x80, 0xdb, 0x86, 0x54, 0x6f, 0x66, 0xd6, 0xbc, 0xa3, 0xf0, 0x19, 0x4c, 0x5e,
	0xa9, 0xa7, 0x23, 0x4b, 0xb7, 0x75, 0xa0, 0xc0, 0xfb, 0xeb, 0x70, 0x5d, 0x53, 0xa5, 0x91, 0x3f,
	0xc0, 0xf1, 0x1e, 0x2c, 0x98, 0xef, 0x84, 0xd3, 0xcb, 0xd3, 0xc8, 0xe6, 0x13, 0x0d, 0x3b, 0xe4,
	0x3e, 0x57, 0xf0, 0x01, 0xf3, 0x05, 0x9a, 0x6e, 0xe1, 0x21, 0x4f, 0xbe, 0xbb, 0xf0, 0x7e, 0x17,
	0xfb, 0x2f, 0x5f, 0x1f, 0xa6, 0xed, 0xdf, 0x75, 0x51, 0x34, 0xa8, 0xb5, 0xdd, 0xeb, 0x50, 0xee,
	0x4a, 0xc9, 0xfc, 0x2d, 0x5c, 0x96, 0x66, 0xb5, 0xce, 0xa2, 0x9c, 0x54, 0xbc, 0x21, 0xca, 0xf2,
	0xee, 0x7b, 0xde, 0xce, 0x8a, 0x73, 0x52, 0x8a, 0xaa, 0xd8, 0xde, 0x3f, 0x73, 0xed, 0x6b, 0x5f,
	0xfd, 0x04, 0x00, 0x00, 0xff, 0xff, 0x69, 0xe7, 0x89, 0x1b, 0x0f, 0x02, 0x00, 0x00,
}
