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
// source: model/feeVote.proto

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

// FeeVoteCommitmentVote represent the commitment vote of fee vote structure stored in the database
type FeeVoteCommitmentVote struct {
	VoteHash             []byte   `protobuf:"bytes,1,opt,name=VoteHash,proto3" json:"VoteHash,omitempty"`
	VoterAddress         []byte   `protobuf:"bytes,2,opt,name=VoterAddress,proto3" json:"VoterAddress,omitempty"`
	BlockHeight          uint32   `protobuf:"varint,3,opt,name=BlockHeight,proto3" json:"BlockHeight,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FeeVoteCommitmentVote) Reset()         { *m = FeeVoteCommitmentVote{} }
func (m *FeeVoteCommitmentVote) String() string { return proto.CompactTextString(m) }
func (*FeeVoteCommitmentVote) ProtoMessage()    {}
func (*FeeVoteCommitmentVote) Descriptor() ([]byte, []int) {
	return fileDescriptor_4966f368323c9cf0, []int{0}
}

func (m *FeeVoteCommitmentVote) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FeeVoteCommitmentVote.Unmarshal(m, b)
}
func (m *FeeVoteCommitmentVote) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FeeVoteCommitmentVote.Marshal(b, m, deterministic)
}
func (m *FeeVoteCommitmentVote) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FeeVoteCommitmentVote.Merge(m, src)
}
func (m *FeeVoteCommitmentVote) XXX_Size() int {
	return xxx_messageInfo_FeeVoteCommitmentVote.Size(m)
}
func (m *FeeVoteCommitmentVote) XXX_DiscardUnknown() {
	xxx_messageInfo_FeeVoteCommitmentVote.DiscardUnknown(m)
}

var xxx_messageInfo_FeeVoteCommitmentVote proto.InternalMessageInfo

func (m *FeeVoteCommitmentVote) GetVoteHash() []byte {
	if m != nil {
		return m.VoteHash
	}
	return nil
}

func (m *FeeVoteCommitmentVote) GetVoterAddress() []byte {
	if m != nil {
		return m.VoterAddress
	}
	return nil
}

func (m *FeeVoteCommitmentVote) GetBlockHeight() uint32 {
	if m != nil {
		return m.BlockHeight
	}
	return 0
}

// FeeVoteInfo represents the fields might be contains what previous commitmentVote submitted
type FeeVoteInfo struct {
	RecentBlockHash      []byte   `protobuf:"bytes,1,opt,name=RecentBlockHash,proto3" json:"RecentBlockHash,omitempty"`
	RecentBlockHeight    uint32   `protobuf:"varint,2,opt,name=RecentBlockHeight,proto3" json:"RecentBlockHeight,omitempty"`
	FeeVote              int64    `protobuf:"varint,3,opt,name=FeeVote,proto3" json:"FeeVote,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FeeVoteInfo) Reset()         { *m = FeeVoteInfo{} }
func (m *FeeVoteInfo) String() string { return proto.CompactTextString(m) }
func (*FeeVoteInfo) ProtoMessage()    {}
func (*FeeVoteInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_4966f368323c9cf0, []int{1}
}

func (m *FeeVoteInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FeeVoteInfo.Unmarshal(m, b)
}
func (m *FeeVoteInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FeeVoteInfo.Marshal(b, m, deterministic)
}
func (m *FeeVoteInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FeeVoteInfo.Merge(m, src)
}
func (m *FeeVoteInfo) XXX_Size() int {
	return xxx_messageInfo_FeeVoteInfo.Size(m)
}
func (m *FeeVoteInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_FeeVoteInfo.DiscardUnknown(m)
}

var xxx_messageInfo_FeeVoteInfo proto.InternalMessageInfo

func (m *FeeVoteInfo) GetRecentBlockHash() []byte {
	if m != nil {
		return m.RecentBlockHash
	}
	return nil
}

func (m *FeeVoteInfo) GetRecentBlockHeight() uint32 {
	if m != nil {
		return m.RecentBlockHeight
	}
	return 0
}

func (m *FeeVoteInfo) GetFeeVote() int64 {
	if m != nil {
		return m.FeeVote
	}
	return 0
}

// FeeVoteRevealVote represents the fields might be contains what previous commitmentVote submitted
type FeeVoteRevealVote struct {
	VoteInfo             *FeeVoteInfo `protobuf:"bytes,1,opt,name=VoteInfo,proto3" json:"VoteInfo,omitempty"`
	VoterSignature       []byte       `protobuf:"bytes,2,opt,name=VoterSignature,proto3" json:"VoterSignature,omitempty"`
	VoterAddress         []byte       `protobuf:"bytes,3,opt,name=VoterAddress,proto3" json:"VoterAddress,omitempty"`
	BlockHeight          uint32       `protobuf:"varint,4,opt,name=BlockHeight,proto3" json:"BlockHeight,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *FeeVoteRevealVote) Reset()         { *m = FeeVoteRevealVote{} }
func (m *FeeVoteRevealVote) String() string { return proto.CompactTextString(m) }
func (*FeeVoteRevealVote) ProtoMessage()    {}
func (*FeeVoteRevealVote) Descriptor() ([]byte, []int) {
	return fileDescriptor_4966f368323c9cf0, []int{2}
}

func (m *FeeVoteRevealVote) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FeeVoteRevealVote.Unmarshal(m, b)
}
func (m *FeeVoteRevealVote) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FeeVoteRevealVote.Marshal(b, m, deterministic)
}
func (m *FeeVoteRevealVote) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FeeVoteRevealVote.Merge(m, src)
}
func (m *FeeVoteRevealVote) XXX_Size() int {
	return xxx_messageInfo_FeeVoteRevealVote.Size(m)
}
func (m *FeeVoteRevealVote) XXX_DiscardUnknown() {
	xxx_messageInfo_FeeVoteRevealVote.DiscardUnknown(m)
}

var xxx_messageInfo_FeeVoteRevealVote proto.InternalMessageInfo

func (m *FeeVoteRevealVote) GetVoteInfo() *FeeVoteInfo {
	if m != nil {
		return m.VoteInfo
	}
	return nil
}

func (m *FeeVoteRevealVote) GetVoterSignature() []byte {
	if m != nil {
		return m.VoterSignature
	}
	return nil
}

func (m *FeeVoteRevealVote) GetVoterAddress() []byte {
	if m != nil {
		return m.VoterAddress
	}
	return nil
}

func (m *FeeVoteRevealVote) GetBlockHeight() uint32 {
	if m != nil {
		return m.BlockHeight
	}
	return 0
}

func init() {
	proto.RegisterType((*FeeVoteCommitmentVote)(nil), "model.FeeVoteCommitmentVote")
	proto.RegisterType((*FeeVoteInfo)(nil), "model.FeeVoteInfo")
	proto.RegisterType((*FeeVoteRevealVote)(nil), "model.FeeVoteRevealVote")
}

func init() {
	proto.RegisterFile("model/feeVote.proto", fileDescriptor_4966f368323c9cf0)
}

var fileDescriptor_4966f368323c9cf0 = []byte{
	// 287 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x91, 0xcf, 0x4a, 0x85, 0x40,
	0x14, 0xc6, 0x19, 0xed, 0x1f, 0xc7, 0x5b, 0x71, 0x27, 0x02, 0x89, 0x16, 0xe2, 0x22, 0x24, 0x4a,
	0xa3, 0x9e, 0x20, 0x83, 0xb8, 0x6d, 0x27, 0x68, 0xd1, 0x4e, 0xc7, 0x73, 0x55, 0x72, 0x3c, 0xa1,
	0x73, 0x83, 0xda, 0xf6, 0x44, 0xbd, 0x61, 0x38, 0xda, 0xcd, 0xae, 0x8b, 0x36, 0x7a, 0xce, 0xef,
	0x1c, 0xf8, 0xbe, 0xf3, 0x0d, 0x1c, 0x29, 0xca, 0xb0, 0x8a, 0x96, 0x88, 0x4f, 0xa4, 0x31, 0x7c,
	0x6d, 0x48, 0x13, 0xdf, 0x36, 0xd0, 0x7f, 0x87, 0xe3, 0xfb, 0x9e, 0xdf, 0x91, 0x52, 0xa5, 0x56,
	0x58, 0xeb, 0xae, 0xe3, 0x27, 0xb0, 0xd7, 0xfd, 0x17, 0x49, 0x5b, 0xb8, 0xcc, 0x63, 0xc1, 0x4c,
	0xac, 0x7b, 0xee, 0xc3, 0xac, 0xab, 0x9b, 0xdb, 0x2c, 0x6b, 0xb0, 0x6d, 0x5d, 0xcb, 0xcc, 0xff,
	0x30, 0xee, 0x81, 0x13, 0x57, 0x24, 0x5f, 0x16, 0x58, 0xe6, 0x85, 0x76, 0x6d, 0x8f, 0x05, 0xfb,
	0x62, 0x8c, 0xfc, 0x4f, 0x06, 0xce, 0xa0, 0xfd, 0x50, 0x2f, 0x89, 0x07, 0x70, 0x28, 0x50, 0x62,
	0xad, 0xfb, 0xa5, 0x5f, 0xe1, 0x4d, 0xcc, 0x2f, 0x60, 0x3e, 0x46, 0xbd, 0x82, 0x65, 0x14, 0xa6,
	0x03, 0x7e, 0x0a, 0xbb, 0x83, 0x8c, 0x71, 0x61, 0xc7, 0xd6, 0x15, 0x13, 0x3f, 0xc8, 0xff, 0x62,
	0x30, 0x1f, 0x6a, 0x81, 0x6f, 0x98, 0x54, 0xe6, 0xfa, 0xb0, 0xbf, 0xbe, 0xf3, 0x65, 0x4c, 0x38,
	0xd7, 0x3c, 0x34, 0x81, 0x85, 0x23, 0xc7, 0x62, 0xbd, 0xc3, 0xcf, 0xe0, 0xc0, 0x5c, 0xff, 0x58,
	0xe6, 0x75, 0xa2, 0x57, 0x0d, 0x0e, 0x99, 0x6c, 0xd0, 0x49, 0x72, 0xf6, 0xff, 0xc9, 0x6d, 0x4d,
	0x92, 0x8b, 0xcf, 0x9f, 0x83, 0xbc, 0xd4, 0xc5, 0x2a, 0x0d, 0x25, 0xa9, 0xe8, 0x83, 0x28, 0x95,
	0xfd, 0xf7, 0x52, 0x52, 0x83, 0x91, 0x24, 0xa5, 0xa8, 0x8e, 0x8c, 0xdf, 0x74, 0xc7, 0x3c, 0xf7,
	0xcd, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0xb0, 0x88, 0x44, 0x25, 0x05, 0x02, 0x00, 0x00,
}
