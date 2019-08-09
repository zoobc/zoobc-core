// Code generated by protoc-gen-go. DO NOT EDIT.
// source: model/blockchain.proto

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

type ChainStatus struct {
	// Integer indicating chaintype
	ChainType            int32    `protobuf:"varint,1,opt,name=ChainType,proto3" json:"ChainType,omitempty"`
	Height               uint32   `protobuf:"varint,2,opt,name=Height,proto3" json:"Height,omitempty"`
	LastBlock            *Block   `protobuf:"bytes,3,opt,name=LastBlock,proto3" json:"LastBlock,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ChainStatus) Reset()         { *m = ChainStatus{} }
func (m *ChainStatus) String() string { return proto.CompactTextString(m) }
func (*ChainStatus) ProtoMessage()    {}
func (*ChainStatus) Descriptor() ([]byte, []int) {
	return fileDescriptor_c2ffcbd1121992b1, []int{0}
}

func (m *ChainStatus) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChainStatus.Unmarshal(m, b)
}
func (m *ChainStatus) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChainStatus.Marshal(b, m, deterministic)
}
func (m *ChainStatus) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChainStatus.Merge(m, src)
}
func (m *ChainStatus) XXX_Size() int {
	return xxx_messageInfo_ChainStatus.Size(m)
}
func (m *ChainStatus) XXX_DiscardUnknown() {
	xxx_messageInfo_ChainStatus.DiscardUnknown(m)
}

var xxx_messageInfo_ChainStatus proto.InternalMessageInfo

func (m *ChainStatus) GetChainType() int32 {
	if m != nil {
		return m.ChainType
	}
	return 0
}

func (m *ChainStatus) GetHeight() uint32 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *ChainStatus) GetLastBlock() *Block {
	if m != nil {
		return m.LastBlock
	}
	return nil
}

type GetCumulativeDifficultyResponse struct {
	CumulativeDifficulty string   `protobuf:"bytes,1,opt,name=CumulativeDifficulty,proto3" json:"CumulativeDifficulty,omitempty"`
	Height               uint32   `protobuf:"varint,2,opt,name=Height,proto3" json:"Height,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetCumulativeDifficultyResponse) Reset()         { *m = GetCumulativeDifficultyResponse{} }
func (m *GetCumulativeDifficultyResponse) String() string { return proto.CompactTextString(m) }
func (*GetCumulativeDifficultyResponse) ProtoMessage()    {}
func (*GetCumulativeDifficultyResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_c2ffcbd1121992b1, []int{1}
}

func (m *GetCumulativeDifficultyResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetCumulativeDifficultyResponse.Unmarshal(m, b)
}
func (m *GetCumulativeDifficultyResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetCumulativeDifficultyResponse.Marshal(b, m, deterministic)
}
func (m *GetCumulativeDifficultyResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetCumulativeDifficultyResponse.Merge(m, src)
}
func (m *GetCumulativeDifficultyResponse) XXX_Size() int {
	return xxx_messageInfo_GetCumulativeDifficultyResponse.Size(m)
}
func (m *GetCumulativeDifficultyResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetCumulativeDifficultyResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetCumulativeDifficultyResponse proto.InternalMessageInfo

func (m *GetCumulativeDifficultyResponse) GetCumulativeDifficulty() string {
	if m != nil {
		return m.CumulativeDifficulty
	}
	return ""
}

func (m *GetCumulativeDifficultyResponse) GetHeight() uint32 {
	if m != nil {
		return m.Height
	}
	return 0
}

type GetCumulativeDifficultyRequest struct {
	// Integer indicating chaintype
	ChainType            int32    `protobuf:"varint,1,opt,name=ChainType,proto3" json:"ChainType,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetCumulativeDifficultyRequest) Reset()         { *m = GetCumulativeDifficultyRequest{} }
func (m *GetCumulativeDifficultyRequest) String() string { return proto.CompactTextString(m) }
func (*GetCumulativeDifficultyRequest) ProtoMessage()    {}
func (*GetCumulativeDifficultyRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_c2ffcbd1121992b1, []int{2}
}

func (m *GetCumulativeDifficultyRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetCumulativeDifficultyRequest.Unmarshal(m, b)
}
func (m *GetCumulativeDifficultyRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetCumulativeDifficultyRequest.Marshal(b, m, deterministic)
}
func (m *GetCumulativeDifficultyRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetCumulativeDifficultyRequest.Merge(m, src)
}
func (m *GetCumulativeDifficultyRequest) XXX_Size() int {
	return xxx_messageInfo_GetCumulativeDifficultyRequest.Size(m)
}
func (m *GetCumulativeDifficultyRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetCumulativeDifficultyRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetCumulativeDifficultyRequest proto.InternalMessageInfo

func (m *GetCumulativeDifficultyRequest) GetChainType() int32 {
	if m != nil {
		return m.ChainType
	}
	return 0
}

type GetCommonMilestoneBlockIdsRequest struct {
	// Integer indicating chaintype
	ChainType            int32    `protobuf:"varint,1,opt,name=ChainType,proto3" json:"ChainType,omitempty"`
	LastBlockId          int64    `protobuf:"varint,2,opt,name=lastBlockId,proto3" json:"lastBlockId,omitempty"`
	LastMilestoneBlockId int64    `protobuf:"varint,3,opt,name=lastMilestoneBlockId,proto3" json:"lastMilestoneBlockId,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetCommonMilestoneBlockIdsRequest) Reset()         { *m = GetCommonMilestoneBlockIdsRequest{} }
func (m *GetCommonMilestoneBlockIdsRequest) String() string { return proto.CompactTextString(m) }
func (*GetCommonMilestoneBlockIdsRequest) ProtoMessage()    {}
func (*GetCommonMilestoneBlockIdsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_c2ffcbd1121992b1, []int{3}
}

func (m *GetCommonMilestoneBlockIdsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetCommonMilestoneBlockIdsRequest.Unmarshal(m, b)
}
func (m *GetCommonMilestoneBlockIdsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetCommonMilestoneBlockIdsRequest.Marshal(b, m, deterministic)
}
func (m *GetCommonMilestoneBlockIdsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetCommonMilestoneBlockIdsRequest.Merge(m, src)
}
func (m *GetCommonMilestoneBlockIdsRequest) XXX_Size() int {
	return xxx_messageInfo_GetCommonMilestoneBlockIdsRequest.Size(m)
}
func (m *GetCommonMilestoneBlockIdsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetCommonMilestoneBlockIdsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetCommonMilestoneBlockIdsRequest proto.InternalMessageInfo

func (m *GetCommonMilestoneBlockIdsRequest) GetChainType() int32 {
	if m != nil {
		return m.ChainType
	}
	return 0
}

func (m *GetCommonMilestoneBlockIdsRequest) GetLastBlockId() int64 {
	if m != nil {
		return m.LastBlockId
	}
	return 0
}

func (m *GetCommonMilestoneBlockIdsRequest) GetLastMilestoneBlockId() int64 {
	if m != nil {
		return m.LastMilestoneBlockId
	}
	return 0
}

type GetCommonMilestoneBlockIdsResponse struct {
	BlockIds             []int64  `protobuf:"varint,1,rep,packed,name=blockIds,proto3" json:"blockIds,omitempty"`
	Last                 bool     `protobuf:"varint,2,opt,name=last,proto3" json:"last,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetCommonMilestoneBlockIdsResponse) Reset()         { *m = GetCommonMilestoneBlockIdsResponse{} }
func (m *GetCommonMilestoneBlockIdsResponse) String() string { return proto.CompactTextString(m) }
func (*GetCommonMilestoneBlockIdsResponse) ProtoMessage()    {}
func (*GetCommonMilestoneBlockIdsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_c2ffcbd1121992b1, []int{4}
}

func (m *GetCommonMilestoneBlockIdsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetCommonMilestoneBlockIdsResponse.Unmarshal(m, b)
}
func (m *GetCommonMilestoneBlockIdsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetCommonMilestoneBlockIdsResponse.Marshal(b, m, deterministic)
}
func (m *GetCommonMilestoneBlockIdsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetCommonMilestoneBlockIdsResponse.Merge(m, src)
}
func (m *GetCommonMilestoneBlockIdsResponse) XXX_Size() int {
	return xxx_messageInfo_GetCommonMilestoneBlockIdsResponse.Size(m)
}
func (m *GetCommonMilestoneBlockIdsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetCommonMilestoneBlockIdsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetCommonMilestoneBlockIdsResponse proto.InternalMessageInfo

func (m *GetCommonMilestoneBlockIdsResponse) GetBlockIds() []int64 {
	if m != nil {
		return m.BlockIds
	}
	return nil
}

func (m *GetCommonMilestoneBlockIdsResponse) GetLast() bool {
	if m != nil {
		return m.Last
	}
	return false
}

func init() {
	proto.RegisterType((*ChainStatus)(nil), "model.ChainStatus")
	proto.RegisterType((*GetCumulativeDifficultyResponse)(nil), "model.GetCumulativeDifficultyResponse")
	proto.RegisterType((*GetCumulativeDifficultyRequest)(nil), "model.GetCumulativeDifficultyRequest")
	proto.RegisterType((*GetCommonMilestoneBlockIdsRequest)(nil), "model.GetCommonMilestoneBlockIdsRequest")
	proto.RegisterType((*GetCommonMilestoneBlockIdsResponse)(nil), "model.GetCommonMilestoneBlockIdsResponse")
}

func init() { proto.RegisterFile("model/blockchain.proto", fileDescriptor_c2ffcbd1121992b1) }

var fileDescriptor_c2ffcbd1121992b1 = []byte{
	// 328 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x52, 0x41, 0x4b, 0xf3, 0x40,
	0x10, 0x65, 0xdb, 0xaf, 0xa5, 0x9d, 0x7e, 0x1e, 0x5c, 0x4a, 0x09, 0x22, 0x35, 0x06, 0x0f, 0xa1,
	0x60, 0x22, 0x15, 0x3c, 0x7a, 0x68, 0x05, 0x2d, 0xe8, 0x65, 0xf5, 0x20, 0xde, 0x92, 0xed, 0xb6,
	0x5d, 0x4c, 0x32, 0xb5, 0xbb, 0x11, 0xea, 0x9f, 0xf1, 0xaf, 0x4a, 0xa6, 0xd1, 0x16, 0x69, 0x8b,
	0x97, 0x65, 0xe7, 0xbd, 0xe1, 0xbd, 0x37, 0xb3, 0x0b, 0x9d, 0x14, 0xc7, 0x2a, 0x09, 0xe3, 0x04,
	0xe5, 0xab, 0x9c, 0x45, 0x3a, 0x0b, 0xe6, 0x0b, 0xb4, 0xc8, 0x6b, 0x84, 0x1f, 0x1d, 0x6e, 0xd0,
	0x2b, 0xc6, 0x43, 0x68, 0x0d, 0x8b, 0xc6, 0x47, 0x1b, 0xd9, 0xdc, 0xf0, 0x63, 0x68, 0x52, 0xf9,
	0xb4, 0x9c, 0x2b, 0x87, 0xb9, 0xcc, 0xaf, 0x89, 0x35, 0xc0, 0x3b, 0x50, 0xbf, 0x53, 0x7a, 0x3a,
	0xb3, 0x4e, 0xc5, 0x65, 0xfe, 0x81, 0x28, 0x2b, 0xde, 0x83, 0xe6, 0x7d, 0x64, 0xec, 0xa0, 0xd0,
	0x75, 0xaa, 0x2e, 0xf3, 0x5b, 0xfd, 0xff, 0x01, 0x79, 0x05, 0x84, 0x89, 0x35, 0xed, 0xa5, 0x70,
	0x72, 0xab, 0xec, 0x30, 0x4f, 0xf3, 0x24, 0xb2, 0xfa, 0x5d, 0xdd, 0xe8, 0xc9, 0x44, 0xcb, 0x3c,
	0xb1, 0x4b, 0xa1, 0xcc, 0x1c, 0x33, 0xa3, 0x78, 0x1f, 0xda, 0xdb, 0x78, 0xca, 0xd3, 0x14, 0x5b,
	0xb9, 0x5d, 0xd1, 0xbc, 0x6b, 0xe8, 0xee, 0xb4, 0x7b, 0xcb, 0x95, 0xb1, 0xfb, 0x47, 0xf6, 0x3e,
	0x19, 0x9c, 0x16, 0x02, 0x98, 0xa6, 0x98, 0x3d, 0xe8, 0x44, 0x19, 0x8b, 0x99, 0xa2, 0x51, 0x46,
	0x63, 0xf3, 0x27, 0x0d, 0x7e, 0x06, 0xad, 0xe4, 0x7b, 0xfe, 0xd1, 0x98, 0x02, 0x56, 0x07, 0x95,
	0x0b, 0x26, 0x36, 0x61, 0x7e, 0x05, 0xed, 0xa2, 0xfc, 0xed, 0x41, 0xfb, 0x5c, 0xb5, 0x6f, 0xe5,
	0xbd, 0x67, 0xf0, 0xf6, 0x05, 0x2c, 0x77, 0xda, 0x85, 0x46, 0x5c, 0x62, 0x0e, 0x73, 0xab, 0xa5,
	0xe2, 0x0f, 0xc6, 0x39, 0xfc, 0x2b, 0xd4, 0x29, 0x5c, 0x43, 0xd0, 0x7d, 0xd0, 0x7b, 0xf1, 0xa7,
	0xda, 0xce, 0xf2, 0x38, 0x90, 0x98, 0x86, 0x1f, 0x88, 0xb1, 0x5c, 0x9d, 0xe7, 0x12, 0x17, 0x2a,
	0x94, 0x64, 0x19, 0xd2, 0x3b, 0xc7, 0x75, 0xfa, 0x4e, 0x97, 0x5f, 0x01, 0x00, 0x00, 0xff, 0xff,
	0xc1, 0x1e, 0x38, 0xb2, 0x82, 0x02, 0x00, 0x00,
}
