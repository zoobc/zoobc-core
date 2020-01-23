// Code generated by protoc-gen-go. DO NOT EDIT.
// source: model/snapshot.proto

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

// SnapshotFileInfo model to pass data between snapshot and spineBlockManifest interfaces
type SnapshotFileInfo struct {
	SnapshotFileHash           []byte                 `protobuf:"bytes,1,opt,name=SnapshotFileHash,proto3" json:"SnapshotFileHash,omitempty"`
	Height                     uint32                 `protobuf:"varint,2,opt,name=Height,proto3" json:"Height,omitempty"`
	ProcessExpirationTimestamp int64                  `protobuf:"varint,3,opt,name=ProcessExpirationTimestamp,proto3" json:"ProcessExpirationTimestamp,omitempty"`
	ChainType                  int32                  `protobuf:"varint,4,opt,name=ChainType,proto3" json:"ChainType,omitempty"`
	SpineBlockManifestType     SpineBlockManifestType `protobuf:"varint,5,opt,name=SpineBlockManifestType,proto3,enum=model.SpineBlockManifestType" json:"SpineBlockManifestType,omitempty"`
	FileChunksHashes           [][]byte               `protobuf:"bytes,6,rep,name=FileChunksHashes,proto3" json:"FileChunksHashes,omitempty"`
	XXX_NoUnkeyedLiteral       struct{}               `json:"-"`
	XXX_unrecognized           []byte                 `json:"-"`
	XXX_sizecache              int32                  `json:"-"`
}

func (m *SnapshotFileInfo) Reset()         { *m = SnapshotFileInfo{} }
func (m *SnapshotFileInfo) String() string { return proto.CompactTextString(m) }
func (*SnapshotFileInfo) ProtoMessage()    {}
func (*SnapshotFileInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_5d9d8140a8c06fc6, []int{0}
}

func (m *SnapshotFileInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SnapshotFileInfo.Unmarshal(m, b)
}
func (m *SnapshotFileInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SnapshotFileInfo.Marshal(b, m, deterministic)
}
func (m *SnapshotFileInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SnapshotFileInfo.Merge(m, src)
}
func (m *SnapshotFileInfo) XXX_Size() int {
	return xxx_messageInfo_SnapshotFileInfo.Size(m)
}
func (m *SnapshotFileInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_SnapshotFileInfo.DiscardUnknown(m)
}

var xxx_messageInfo_SnapshotFileInfo proto.InternalMessageInfo

func (m *SnapshotFileInfo) GetSnapshotFileHash() []byte {
	if m != nil {
		return m.SnapshotFileHash
	}
	return nil
}

func (m *SnapshotFileInfo) GetHeight() uint32 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *SnapshotFileInfo) GetProcessExpirationTimestamp() int64 {
	if m != nil {
		return m.ProcessExpirationTimestamp
	}
	return 0
}

func (m *SnapshotFileInfo) GetChainType() int32 {
	if m != nil {
		return m.ChainType
	}
	return 0
}

func (m *SnapshotFileInfo) GetSpineBlockManifestType() SpineBlockManifestType {
	if m != nil {
		return m.SpineBlockManifestType
	}
	return SpineBlockManifestType_Snapshot
}

func (m *SnapshotFileInfo) GetFileChunksHashes() [][]byte {
	if m != nil {
		return m.FileChunksHashes
	}
	return nil
}

func init() {
	proto.RegisterType((*SnapshotFileInfo)(nil), "model.SnapshotFileInfo")
}

func init() { proto.RegisterFile("model/snapshot.proto", fileDescriptor_5d9d8140a8c06fc6) }

var fileDescriptor_5d9d8140a8c06fc6 = []byte{
	// 267 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x90, 0x41, 0x4b, 0xc3, 0x30,
	0x14, 0x80, 0xc9, 0x6a, 0x0b, 0x06, 0x15, 0x09, 0x32, 0xca, 0x50, 0x09, 0x9e, 0xc2, 0xc0, 0x16,
	0xf4, 0xee, 0x61, 0x43, 0x99, 0x07, 0x41, 0xba, 0x79, 0xf1, 0x96, 0xc6, 0x6c, 0x09, 0x6b, 0xf2,
	0x42, 0x93, 0x81, 0xfa, 0xc7, 0xfc, 0x7b, 0xd2, 0x58, 0x10, 0x9d, 0xee, 0x12, 0xc8, 0x97, 0xef,
	0x41, 0xbe, 0x87, 0x4f, 0x0c, 0xbc, 0xc8, 0xa6, 0xf4, 0x96, 0x3b, 0xaf, 0x20, 0x14, 0xae, 0x85,
	0x00, 0x24, 0x8d, 0x74, 0x74, 0xde, 0x3f, 0x3a, 0x6d, 0xe5, 0xa4, 0x01, 0xb1, 0x7e, 0xe0, 0x56,
	0x2f, 0xa5, 0xef, 0xb5, 0x8b, 0x8f, 0x01, 0x3e, 0x9e, 0xf7, 0x93, 0x77, 0xba, 0x91, 0xf7, 0x76,
	0x09, 0x64, 0xfc, 0x93, 0xcd, 0xb8, 0x57, 0x39, 0xa2, 0x88, 0x1d, 0x54, 0x5b, 0x9c, 0x0c, 0x71,
	0x36, 0x93, 0x7a, 0xa5, 0x42, 0x3e, 0xa0, 0x88, 0x1d, 0x56, 0xfd, 0x8d, 0xdc, 0xe0, 0xd1, 0x63,
	0x0b, 0x42, 0x7a, 0x7f, 0xfb, 0xea, 0x74, 0xcb, 0x83, 0x06, 0xbb, 0xd0, 0x46, 0xfa, 0xc0, 0x8d,
	0xcb, 0x13, 0x8a, 0x58, 0x52, 0xed, 0x30, 0xc8, 0x29, 0xde, 0x9f, 0x2a, 0xae, 0xed, 0xe2, 0xcd,
	0xc9, 0x7c, 0x8f, 0x22, 0x96, 0x56, 0xdf, 0x80, 0x3c, 0xe1, 0xe1, 0x7c, 0x2b, 0x29, 0xaa, 0x29,
	0x45, 0xec, 0xe8, 0xea, 0xac, 0x88, 0xdd, 0xc5, 0xdf, 0x52, 0xf5, 0xcf, 0x70, 0x17, 0xde, 0x85,
	0x4d, 0xd5, 0xc6, 0xae, 0x7d, 0x97, 0x27, 0x7d, 0x9e, 0xd1, 0xa4, 0x0b, 0xff, 0xcd, 0x27, 0xe3,
	0x67, 0xb6, 0xd2, 0x41, 0x6d, 0xea, 0x42, 0x80, 0x29, 0xdf, 0x01, 0x6a, 0xf1, 0x75, 0x5e, 0x0a,
	0x68, 0x65, 0x29, 0xc0, 0x18, 0xb0, 0x65, 0xfc, 0x46, 0x9d, 0xc5, 0x65, 0x5f, 0x7f, 0x06, 0x00,
	0x00, 0xff, 0xff, 0xbd, 0x9e, 0x75, 0x62, 0xab, 0x01, 0x00, 0x00,
}
