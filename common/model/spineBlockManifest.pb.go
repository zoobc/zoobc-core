// Code generated by protoc-gen-go. DO NOT EDIT.
// source: model/spineBlockManifest.proto

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

// SpineBlockManifestType type of spineBlockManifest (as of now only snapshot)
type SpineBlockManifestType int32

const (
	SpineBlockManifestType_Snapshot SpineBlockManifestType = 0
)

var SpineBlockManifestType_name = map[int32]string{
	0: "Snapshot",
}

var SpineBlockManifestType_value = map[string]int32{
	"Snapshot": 0,
}

func (x SpineBlockManifestType) String() string {
	return proto.EnumName(SpineBlockManifestType_name, int32(x))
}

func (SpineBlockManifestType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_28f5b9e6a17937ec, []int{0}
}

// SpineBlockManifest represent the spineBlockManifest data structure stored in the database
type SpineBlockManifest struct {
	// ID computed as the little endian of the spineBlockManifest hash (hash of all spineBlockManifest fields but the ID)
	ID int64 `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
	// FullFileHash hash of the full - completed (snapshot) file
	FullFileHash []byte `protobuf:"bytes,2,opt,name=FullFileHash,proto3" json:"FullFileHash,omitempty"`
	// FileChunkHashes sequence of hashes (sha256 = 32 byte) of file chunks (sorted) referenced by the spineBlockManifest
	FileChunkHashes []byte `protobuf:"bytes,3,opt,name=FileChunkHashes,proto3" json:"FileChunkHashes,omitempty"`
	// ManifestReferenceHeight height (on the mainchain) at which the (snapshot) file started been computed
	// Note: this is not the last mainchain height contained in the snapshot file (that one should be = SpineBlockManifestHeight - MinRollbackBlocks)
	ManifestReferenceHeight uint32 `protobuf:"varint,4,opt,name=ManifestReferenceHeight,proto3" json:"ManifestReferenceHeight,omitempty"`
	// ManifestSpineBlockHeight (on spinechain) at which the (manifest) got included in the block, this data
	// is tightly coupled to the spine block.
	ManifestSpineBlockHeight uint32 `protobuf:"varint,5,opt,name=ManifestSpineBlockHeight,proto3" json:"ManifestSpineBlockHeight,omitempty"`
	// Number indicating chaintype (at the moment it can only be mainchain, but in future could be others)
	ChainType int32 `protobuf:"varint,6,opt,name=ChainType,proto3" json:"ChainType,omitempty"`
	// SpineBlockManifestType type of spineBlockManifest
	SpineBlockManifestType SpineBlockManifestType `protobuf:"varint,7,opt,name=SpineBlockManifestType,proto3,enum=model.SpineBlockManifestType" json:"SpineBlockManifestType,omitempty"`
	// ExpirationTimestamp timestamp that marks the end of spineBlockManifest processing
	ExpirationTimestamp  int64    `protobuf:"varint,8,opt,name=ExpirationTimestamp,proto3" json:"ExpirationTimestamp,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SpineBlockManifest) Reset()         { *m = SpineBlockManifest{} }
func (m *SpineBlockManifest) String() string { return proto.CompactTextString(m) }
func (*SpineBlockManifest) ProtoMessage()    {}
func (*SpineBlockManifest) Descriptor() ([]byte, []int) {
	return fileDescriptor_28f5b9e6a17937ec, []int{0}
}

func (m *SpineBlockManifest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SpineBlockManifest.Unmarshal(m, b)
}
func (m *SpineBlockManifest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SpineBlockManifest.Marshal(b, m, deterministic)
}
func (m *SpineBlockManifest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SpineBlockManifest.Merge(m, src)
}
func (m *SpineBlockManifest) XXX_Size() int {
	return xxx_messageInfo_SpineBlockManifest.Size(m)
}
func (m *SpineBlockManifest) XXX_DiscardUnknown() {
	xxx_messageInfo_SpineBlockManifest.DiscardUnknown(m)
}

var xxx_messageInfo_SpineBlockManifest proto.InternalMessageInfo

func (m *SpineBlockManifest) GetID() int64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *SpineBlockManifest) GetFullFileHash() []byte {
	if m != nil {
		return m.FullFileHash
	}
	return nil
}

func (m *SpineBlockManifest) GetFileChunkHashes() []byte {
	if m != nil {
		return m.FileChunkHashes
	}
	return nil
}

func (m *SpineBlockManifest) GetManifestReferenceHeight() uint32 {
	if m != nil {
		return m.ManifestReferenceHeight
	}
	return 0
}

func (m *SpineBlockManifest) GetManifestSpineBlockHeight() uint32 {
	if m != nil {
		return m.ManifestSpineBlockHeight
	}
	return 0
}

func (m *SpineBlockManifest) GetChainType() int32 {
	if m != nil {
		return m.ChainType
	}
	return 0
}

func (m *SpineBlockManifest) GetSpineBlockManifestType() SpineBlockManifestType {
	if m != nil {
		return m.SpineBlockManifestType
	}
	return SpineBlockManifestType_Snapshot
}

func (m *SpineBlockManifest) GetExpirationTimestamp() int64 {
	if m != nil {
		return m.ExpirationTimestamp
	}
	return 0
}

func init() {
	proto.RegisterEnum("model.SpineBlockManifestType", SpineBlockManifestType_name, SpineBlockManifestType_value)
	proto.RegisterType((*SpineBlockManifest)(nil), "model.SpineBlockManifest")
}

func init() { proto.RegisterFile("model/spineBlockManifest.proto", fileDescriptor_28f5b9e6a17937ec) }

var fileDescriptor_28f5b9e6a17937ec = []byte{
	// 306 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x91, 0xcf, 0x4f, 0xc2, 0x30,
	0x1c, 0xc5, 0xed, 0x10, 0xc4, 0x06, 0x7f, 0xa4, 0x26, 0xda, 0x83, 0x9a, 0x85, 0x83, 0x69, 0x48,
	0xdc, 0x88, 0x5e, 0x8c, 0x47, 0x40, 0x02, 0x07, 0x2f, 0x05, 0x2f, 0xde, 0xc6, 0xfc, 0xc2, 0x1a,
	0xb6, 0x7e, 0x97, 0xb5, 0x24, 0xea, 0xbf, 0xe1, 0x3f, 0x6c, 0x56, 0x25, 0x44, 0x61, 0x97, 0xa6,
	0x79, 0x9f, 0xf7, 0xd2, 0x97, 0x57, 0x7a, 0x9d, 0xe1, 0x1b, 0xa4, 0xa1, 0xc9, 0x95, 0x86, 0x5e,
	0x8a, 0xf1, 0xf2, 0x39, 0xd2, 0x6a, 0x0e, 0xc6, 0x06, 0x79, 0x81, 0x16, 0x59, 0xdd, 0xf1, 0xf6,
	0x57, 0x8d, 0xb2, 0xc9, 0x96, 0x87, 0x31, 0xea, 0x8d, 0x07, 0x9c, 0xf8, 0x44, 0xd4, 0x7a, 0x5e,
	0x97, 0x48, 0x6f, 0x3c, 0x60, 0x6d, 0xda, 0x1a, 0xae, 0xd2, 0x74, 0xa8, 0x52, 0x18, 0x45, 0x26,
	0xe1, 0x9e, 0x4f, 0x44, 0x4b, 0xfe, 0xd1, 0x98, 0xa0, 0x27, 0xe5, 0xbd, 0x9f, 0xac, 0xf4, 0xb2,
	0x14, 0xc0, 0xf0, 0x9a, 0xb3, 0xfd, 0x97, 0xd9, 0x03, 0xbd, 0x58, 0xbf, 0x26, 0x61, 0x0e, 0x05,
	0xe8, 0x18, 0x46, 0xa0, 0x16, 0x89, 0xe5, 0xfb, 0x3e, 0x11, 0x47, 0xb2, 0x0a, 0xb3, 0x47, 0xca,
	0xd7, 0x68, 0xd3, 0xfc, 0x37, 0x5a, 0x77, 0xd1, 0x4a, 0xce, 0x2e, 0xe9, 0x61, 0x3f, 0x89, 0x94,
	0x9e, 0x7e, 0xe4, 0xc0, 0x1b, 0x3e, 0x11, 0x75, 0xb9, 0x11, 0xd8, 0x0b, 0x3d, 0xdf, 0xde, 0xc2,
	0x59, 0x0f, 0x7c, 0x22, 0x8e, 0xef, 0xae, 0x02, 0x37, 0x5a, 0xb0, 0xdb, 0x24, 0x2b, 0xc2, 0xac,
	0x4b, 0xcf, 0x9e, 0xde, 0x73, 0x55, 0x44, 0x56, 0xa1, 0x9e, 0xaa, 0x0c, 0x8c, 0x8d, 0xb2, 0x9c,
	0x37, 0xcb, 0x75, 0xe5, 0x2e, 0xd4, 0xb9, 0xa9, 0x2a, 0xc2, 0x5a, 0xb4, 0x39, 0xd1, 0x51, 0x6e,
	0x12, 0xb4, 0xa7, 0x7b, 0xbd, 0xce, 0xab, 0x58, 0x28, 0x9b, 0xac, 0x66, 0x41, 0x8c, 0x59, 0xf8,
	0x89, 0x38, 0x8b, 0x7f, 0xce, 0xdb, 0x18, 0x0b, 0x08, 0x63, 0xcc, 0x32, 0xd4, 0xa1, 0x2b, 0x3d,
	0x6b, 0xb8, 0x7f, 0xbf, 0xff, 0x0e, 0x00, 0x00, 0xff, 0xff, 0xe9, 0x9f, 0x65, 0x53, 0x19, 0x02,
	0x00, 0x00,
}
