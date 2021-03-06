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
// source: model/healthCheck.proto

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

// HealthCheckResponse represent the response body of health check request
type HealthCheckResponse struct {
	Reply                string   `protobuf:"bytes,1,opt,name=Reply,proto3" json:"Reply,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HealthCheckResponse) Reset()         { *m = HealthCheckResponse{} }
func (m *HealthCheckResponse) String() string { return proto.CompactTextString(m) }
func (*HealthCheckResponse) ProtoMessage()    {}
func (*HealthCheckResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_45c88e150f25dc67, []int{0}
}

func (m *HealthCheckResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HealthCheckResponse.Unmarshal(m, b)
}
func (m *HealthCheckResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HealthCheckResponse.Marshal(b, m, deterministic)
}
func (m *HealthCheckResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HealthCheckResponse.Merge(m, src)
}
func (m *HealthCheckResponse) XXX_Size() int {
	return xxx_messageInfo_HealthCheckResponse.Size(m)
}
func (m *HealthCheckResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_HealthCheckResponse.DiscardUnknown(m)
}

var xxx_messageInfo_HealthCheckResponse proto.InternalMessageInfo

func (m *HealthCheckResponse) GetReply() string {
	if m != nil {
		return m.Reply
	}
	return ""
}

// HealthCheckResponse represent the response body of health check request
type GetNodeStatusResponse struct {
	LastMainBlockHeight  uint32   `protobuf:"varint,1,opt,name=LastMainBlockHeight,proto3" json:"LastMainBlockHeight,omitempty"`
	LastMainBlockHash    string   `protobuf:"bytes,2,opt,name=LastMainBlockHash,proto3" json:"LastMainBlockHash,omitempty"`
	LastSpineBlockHeight uint32   `protobuf:"varint,3,opt,name=LastSpineBlockHeight,proto3" json:"LastSpineBlockHeight,omitempty"`
	LastSpineBlockHash   string   `protobuf:"bytes,4,opt,name=LastSpineBlockHash,proto3" json:"LastSpineBlockHash,omitempty"`
	Version              string   `protobuf:"bytes,5,opt,name=Version,proto3" json:"Version,omitempty"`
	NodePublicKey        string   `protobuf:"bytes,6,opt,name=NodePublicKey,proto3" json:"NodePublicKey,omitempty"`
	UnresolvedPeers      uint32   `protobuf:"varint,7,opt,name=UnresolvedPeers,proto3" json:"UnresolvedPeers,omitempty"`
	ResolvedPeers        uint32   `protobuf:"varint,8,opt,name=ResolvedPeers,proto3" json:"ResolvedPeers,omitempty"`
	BlocksmithIndex      int32    `protobuf:"varint,9,opt,name=BlocksmithIndex,proto3" json:"BlocksmithIndex,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetNodeStatusResponse) Reset()         { *m = GetNodeStatusResponse{} }
func (m *GetNodeStatusResponse) String() string { return proto.CompactTextString(m) }
func (*GetNodeStatusResponse) ProtoMessage()    {}
func (*GetNodeStatusResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_45c88e150f25dc67, []int{1}
}

func (m *GetNodeStatusResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetNodeStatusResponse.Unmarshal(m, b)
}
func (m *GetNodeStatusResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetNodeStatusResponse.Marshal(b, m, deterministic)
}
func (m *GetNodeStatusResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetNodeStatusResponse.Merge(m, src)
}
func (m *GetNodeStatusResponse) XXX_Size() int {
	return xxx_messageInfo_GetNodeStatusResponse.Size(m)
}
func (m *GetNodeStatusResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetNodeStatusResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetNodeStatusResponse proto.InternalMessageInfo

func (m *GetNodeStatusResponse) GetLastMainBlockHeight() uint32 {
	if m != nil {
		return m.LastMainBlockHeight
	}
	return 0
}

func (m *GetNodeStatusResponse) GetLastMainBlockHash() string {
	if m != nil {
		return m.LastMainBlockHash
	}
	return ""
}

func (m *GetNodeStatusResponse) GetLastSpineBlockHeight() uint32 {
	if m != nil {
		return m.LastSpineBlockHeight
	}
	return 0
}

func (m *GetNodeStatusResponse) GetLastSpineBlockHash() string {
	if m != nil {
		return m.LastSpineBlockHash
	}
	return ""
}

func (m *GetNodeStatusResponse) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *GetNodeStatusResponse) GetNodePublicKey() string {
	if m != nil {
		return m.NodePublicKey
	}
	return ""
}

func (m *GetNodeStatusResponse) GetUnresolvedPeers() uint32 {
	if m != nil {
		return m.UnresolvedPeers
	}
	return 0
}

func (m *GetNodeStatusResponse) GetResolvedPeers() uint32 {
	if m != nil {
		return m.ResolvedPeers
	}
	return 0
}

func (m *GetNodeStatusResponse) GetBlocksmithIndex() int32 {
	if m != nil {
		return m.BlocksmithIndex
	}
	return 0
}

func init() {
	proto.RegisterType((*HealthCheckResponse)(nil), "model.HealthCheckResponse")
	proto.RegisterType((*GetNodeStatusResponse)(nil), "model.GetNodeStatusResponse")
}

func init() {
	proto.RegisterFile("model/healthCheck.proto", fileDescriptor_45c88e150f25dc67)
}

var fileDescriptor_45c88e150f25dc67 = []byte{
	// 305 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x91, 0x5d, 0x4b, 0xf3, 0x30,
	0x14, 0xc7, 0xd9, 0xb3, 0xa7, 0x9b, 0x0b, 0x0c, 0x31, 0x9b, 0x98, 0xcb, 0x31, 0xbc, 0x28, 0xbe,
	0xac, 0xa2, 0xdf, 0x60, 0x5e, 0x38, 0xf1, 0x85, 0xd1, 0xa1, 0x17, 0xde, 0xb5, 0xe9, 0x61, 0x09,
	0x4b, 0x73, 0x4a, 0x93, 0x89, 0xf3, 0x7b, 0xf8, 0x7d, 0xa5, 0x67, 0xf8, 0xd2, 0xd9, 0x9b, 0x42,
	0x7e, 0xbf, 0x7f, 0xf3, 0x3f, 0xe1, 0xb0, 0xa3, 0x1c, 0x33, 0x30, 0x91, 0x82, 0xc4, 0x78, 0x75,
	0xad, 0x40, 0xae, 0x26, 0x45, 0x89, 0x1e, 0x79, 0x40, 0x62, 0x7c, 0xca, 0x06, 0xb3, 0x1f, 0x17,
	0x83, 0x2b, 0xd0, 0x3a, 0xe0, 0x43, 0x16, 0xc4, 0x50, 0x98, 0x8d, 0x68, 0x8d, 0x5a, 0x61, 0x2f,
	0xde, 0x1e, 0xc6, 0x1f, 0x6d, 0x76, 0x78, 0x03, 0xfe, 0x11, 0x33, 0x58, 0xf8, 0xc4, 0xaf, 0xdd,
	0x77, 0xfe, 0x82, 0x0d, 0xee, 0x13, 0xe7, 0x1f, 0x12, 0x6d, 0xa7, 0x06, 0xe5, 0x6a, 0x06, 0x7a,
	0xa9, 0x3c, 0xfd, 0xdd, 0x8f, 0x9b, 0x14, 0x3f, 0x63, 0x07, 0x75, 0x9c, 0x38, 0x25, 0xfe, 0x51,
	0xdb, 0x5f, 0xc1, 0x2f, 0xd9, 0xb0, 0x82, 0x8b, 0x42, 0x5b, 0xf8, 0x5d, 0xd0, 0xa6, 0x82, 0x46,
	0xc7, 0x27, 0x8c, 0xef, 0xf0, 0xaa, 0xe2, 0x3f, 0x55, 0x34, 0x18, 0x2e, 0x58, 0xf7, 0x19, 0x4a,
	0xa7, 0xd1, 0x8a, 0x80, 0x42, 0x5f, 0x47, 0x7e, 0xcc, 0xfa, 0xd5, 0x9b, 0xe7, 0xeb, 0xd4, 0x68,
	0x79, 0x07, 0x1b, 0xd1, 0x21, 0x5f, 0x87, 0x3c, 0x64, 0xfb, 0x4f, 0xb6, 0x04, 0x87, 0xe6, 0x15,
	0xb2, 0x39, 0x40, 0xe9, 0x44, 0x97, 0xc6, 0xdb, 0xc5, 0xd5, 0x7d, 0x71, 0x2d, 0xb7, 0x47, 0xb9,
	0x3a, 0xac, 0xee, 0xa3, 0xe1, 0x5c, 0xae, 0xbd, 0xba, 0xb5, 0x19, 0xbc, 0x89, 0xde, 0xa8, 0x15,
	0x06, 0xf1, 0x2e, 0x9e, 0x9e, 0xbc, 0x84, 0x4b, 0xed, 0xd5, 0x3a, 0x9d, 0x48, 0xcc, 0xa3, 0x77,
	0xc4, 0x54, 0x6e, 0xbf, 0xe7, 0x12, 0x4b, 0x88, 0x24, 0xe6, 0x39, 0xda, 0x88, 0x16, 0x9e, 0x76,
	0x68, 0xfd, 0x57, 0x9f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x22, 0xde, 0x61, 0xd7, 0x19, 0x02, 0x00,
	0x00,
}
