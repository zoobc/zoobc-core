// Code generated by protoc-gen-go. DO NOT EDIT.
// source: model/megablock.proto

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

// Megablock represent the megablock data structure stored in the database
type Megablock struct {
	FullSnapshotHash     []byte   `protobuf:"bytes,1,opt,name=FullSnapshotHash,proto3" json:"FullSnapshotHash,omitempty"`
	SpineBlockHeight     uint32   `protobuf:"varint,2,opt,name=SpineBlockHeight,proto3" json:"SpineBlockHeight,omitempty"`
	MainBlockHeight      uint32   `protobuf:"varint,3,opt,name=MainBlockHeight,proto3" json:"MainBlockHeight,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Megablock) Reset()         { *m = Megablock{} }
func (m *Megablock) String() string { return proto.CompactTextString(m) }
func (*Megablock) ProtoMessage()    {}
func (*Megablock) Descriptor() ([]byte, []int) {
	return fileDescriptor_2b40a10b4e757ea5, []int{0}
}

func (m *Megablock) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Megablock.Unmarshal(m, b)
}
func (m *Megablock) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Megablock.Marshal(b, m, deterministic)
}
func (m *Megablock) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Megablock.Merge(m, src)
}
func (m *Megablock) XXX_Size() int {
	return xxx_messageInfo_Megablock.Size(m)
}
func (m *Megablock) XXX_DiscardUnknown() {
	xxx_messageInfo_Megablock.DiscardUnknown(m)
}

var xxx_messageInfo_Megablock proto.InternalMessageInfo

func (m *Megablock) GetFullSnapshotHash() []byte {
	if m != nil {
		return m.FullSnapshotHash
	}
	return nil
}

func (m *Megablock) GetSpineBlockHeight() uint32 {
	if m != nil {
		return m.SpineBlockHeight
	}
	return 0
}

func (m *Megablock) GetMainBlockHeight() uint32 {
	if m != nil {
		return m.MainBlockHeight
	}
	return 0
}

func init() {
	proto.RegisterType((*Megablock)(nil), "model.Megablock")
}

func init() { proto.RegisterFile("model/megablock.proto", fileDescriptor_2b40a10b4e757ea5) }

var fileDescriptor_2b40a10b4e757ea5 = []byte{
	// 170 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0xcd, 0xcd, 0x4f, 0x49,
	0xcd, 0xd1, 0xcf, 0x4d, 0x4d, 0x4f, 0x4c, 0xca, 0xc9, 0x4f, 0xce, 0xd6, 0x2b, 0x28, 0xca, 0x2f,
	0xc9, 0x17, 0x62, 0x05, 0x0b, 0x2b, 0xf5, 0x32, 0x72, 0x71, 0xfa, 0xc2, 0xa4, 0x84, 0xb4, 0xb8,
	0x04, 0xdc, 0x4a, 0x73, 0x72, 0x82, 0xf3, 0x12, 0x0b, 0x8a, 0x33, 0xf2, 0x4b, 0x3c, 0x12, 0x8b,
	0x33, 0x24, 0x18, 0x15, 0x18, 0x35, 0x78, 0x82, 0x30, 0xc4, 0x41, 0x6a, 0x83, 0x0b, 0x32, 0xf3,
	0x52, 0x9d, 0x40, 0x3a, 0x3d, 0x52, 0x33, 0xd3, 0x33, 0x4a, 0x24, 0x98, 0x14, 0x18, 0x35, 0x78,
	0x83, 0x30, 0xc4, 0x85, 0x34, 0xb8, 0xf8, 0x7d, 0x13, 0x33, 0xf3, 0x90, 0x95, 0x32, 0x83, 0x95,
	0xa2, 0x0b, 0x3b, 0x69, 0x45, 0x69, 0xa4, 0x67, 0x96, 0x64, 0x94, 0x26, 0xe9, 0x25, 0xe7, 0xe7,
	0xea, 0x57, 0xe5, 0xe7, 0x27, 0x25, 0x43, 0x48, 0xdd, 0xe4, 0xfc, 0xa2, 0x54, 0xfd, 0xe4, 0xfc,
	0xdc, 0xdc, 0xfc, 0x3c, 0x7d, 0xb0, 0xdb, 0x93, 0xd8, 0xc0, 0x3e, 0x31, 0x06, 0x04, 0x00, 0x00,
	0xff, 0xff, 0x4a, 0xfb, 0x9f, 0x9b, 0xe2, 0x00, 0x00, 0x00,
}
