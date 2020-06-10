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
	VoterAddress         string   `protobuf:"bytes,2,opt,name=VoterAddress,proto3" json:"VoterAddress,omitempty"`
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

func (m *FeeVoteCommitmentVote) GetVoterAddress() string {
	if m != nil {
		return m.VoterAddress
	}
	return ""
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
	VoterAddress         string       `protobuf:"bytes,3,opt,name=VoterAddress,proto3" json:"VoterAddress,omitempty"`
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

func (m *FeeVoteRevealVote) GetVoterAddress() string {
	if m != nil {
		return m.VoterAddress
	}
	return ""
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

func init() { proto.RegisterFile("model/feeVote.proto", fileDescriptor_4966f368323c9cf0) }

var fileDescriptor_4966f368323c9cf0 = []byte{
	// 292 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x91, 0x4f, 0x4f, 0xb4, 0x30,
	0x10, 0xc6, 0xd3, 0xe5, 0x7d, 0xfd, 0x33, 0xa0, 0x66, 0x6b, 0x4c, 0x88, 0xf1, 0x40, 0x38, 0x18,
	0x62, 0x14, 0x8c, 0x7e, 0x02, 0x31, 0x31, 0xeb, 0xb5, 0x26, 0x1e, 0xbc, 0x41, 0x99, 0x05, 0x22,
	0x65, 0x0c, 0x74, 0x4d, 0xf4, 0xea, 0x27, 0xf2, 0x1b, 0x1a, 0x0a, 0xae, 0xb8, 0x1c, 0xbc, 0xb4,
	0x33, 0xbf, 0x4e, 0xf2, 0x3c, 0xf3, 0x14, 0x0e, 0x15, 0x65, 0x58, 0x45, 0x4b, 0xc4, 0x47, 0xd2,
	0x18, 0xbe, 0x34, 0xa4, 0x89, 0xff, 0x37, 0xd0, 0x7f, 0x83, 0xa3, 0xbb, 0x9e, 0xdf, 0x92, 0x52,
	0xa5, 0x56, 0x58, 0xeb, 0xae, 0xe3, 0xc7, 0xb0, 0xd3, 0xdd, 0x8b, 0xa4, 0x2d, 0x5c, 0xe6, 0xb1,
	0xc0, 0x11, 0xeb, 0x9e, 0xfb, 0xe0, 0x74, 0x75, 0x73, 0x93, 0x65, 0x0d, 0xb6, 0xad, 0x3b, 0xf3,
	0x58, 0xb0, 0x2b, 0x7e, 0x31, 0xee, 0x81, 0x1d, 0x57, 0x24, 0x9f, 0x17, 0x58, 0xe6, 0x85, 0x76,
	0x2d, 0x8f, 0x05, 0x7b, 0x62, 0x8c, 0xfc, 0x0f, 0x06, 0xf6, 0xa0, 0x7d, 0x5f, 0x2f, 0x89, 0x07,
	0x70, 0x20, 0x50, 0x62, 0xad, 0xfb, 0xa1, 0x1f, 0xe1, 0x4d, 0xcc, 0xcf, 0x61, 0x3e, 0x46, 0xbd,
	0xc2, 0xcc, 0x28, 0x4c, 0x1f, 0xf8, 0x09, 0x6c, 0x0f, 0x32, 0xc6, 0x85, 0x15, 0xcf, 0x2e, 0x99,
	0xf8, 0x46, 0xfe, 0x27, 0x83, 0xf9, 0x50, 0x0b, 0x7c, 0xc5, 0xa4, 0x32, 0xdb, 0x87, 0xfd, 0xf6,
	0x9d, 0x2f, 0x63, 0xc2, 0xbe, 0xe2, 0xa1, 0x09, 0x2c, 0x1c, 0x39, 0x16, 0xeb, 0x19, 0x7e, 0x0a,
	0xfb, 0x66, 0xfb, 0x87, 0x32, 0xaf, 0x13, 0xbd, 0x6a, 0xd0, 0xd8, 0x71, 0xc4, 0x06, 0x9d, 0x24,
	0x67, 0xfd, 0x9d, 0xdc, 0xbf, 0x49, 0x72, 0xf1, 0xd9, 0x53, 0x90, 0x97, 0xba, 0x58, 0xa5, 0xa1,
	0x24, 0x15, 0xbd, 0x13, 0xa5, 0xb2, 0x3f, 0x2f, 0x24, 0x35, 0x18, 0x49, 0x52, 0x8a, 0xea, 0xc8,
	0xf8, 0x4d, 0xb7, 0xcc, 0x77, 0x5f, 0x7f, 0x05, 0x00, 0x00, 0xff, 0xff, 0x40, 0x67, 0xe5, 0x81,
	0x05, 0x02, 0x00, 0x00,
}
