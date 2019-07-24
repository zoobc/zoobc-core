// Code generated by protoc-gen-go. DO NOT EDIT.
// source: model/host.proto

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

// Host represent
type Host struct {
	Info                 *Node            `protobuf:"bytes,1,opt,name=Info,proto3" json:"Info,omitempty"`
	ResolvedPeers        map[string]*Peer `protobuf:"bytes,2,rep,name=ResolvedPeers,proto3" json:"ResolvedPeers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	KnownPeers           map[string]*Peer `protobuf:"bytes,3,rep,name=KnownPeers,proto3" json:"KnownPeers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	UnresolvedPeers      map[string]*Peer `protobuf:"bytes,4,rep,name=UnresolvedPeers,proto3" json:"UnresolvedPeers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Stopped              bool             `protobuf:"varint,5,opt,name=Stopped,proto3" json:"Stopped,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *Host) Reset()         { *m = Host{} }
func (m *Host) String() string { return proto.CompactTextString(m) }
func (*Host) ProtoMessage()    {}
func (*Host) Descriptor() ([]byte, []int) {
	return fileDescriptor_2105bc76e5d5e738, []int{0}
}

func (m *Host) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Host.Unmarshal(m, b)
}
func (m *Host) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Host.Marshal(b, m, deterministic)
}
func (m *Host) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Host.Merge(m, src)
}
func (m *Host) XXX_Size() int {
	return xxx_messageInfo_Host.Size(m)
}
func (m *Host) XXX_DiscardUnknown() {
	xxx_messageInfo_Host.DiscardUnknown(m)
}

var xxx_messageInfo_Host proto.InternalMessageInfo

func (m *Host) GetInfo() *Node {
	if m != nil {
		return m.Info
	}
	return nil
}

func (m *Host) GetResolvedPeers() map[string]*Peer {
	if m != nil {
		return m.ResolvedPeers
	}
	return nil
}

func (m *Host) GetKnownPeers() map[string]*Peer {
	if m != nil {
		return m.KnownPeers
	}
	return nil
}

func (m *Host) GetUnresolvedPeers() map[string]*Peer {
	if m != nil {
		return m.UnresolvedPeers
	}
	return nil
}

func (m *Host) GetStopped() bool {
	if m != nil {
		return m.Stopped
	}
	return false
}

func init() {
	proto.RegisterType((*Host)(nil), "model.Host")
	proto.RegisterMapType((map[string]*Peer)(nil), "model.Host.KnownPeersEntry")
	proto.RegisterMapType((map[string]*Peer)(nil), "model.Host.ResolvedPeersEntry")
	proto.RegisterMapType((map[string]*Peer)(nil), "model.Host.UnresolvedPeersEntry")
}

func init() { proto.RegisterFile("model/host.proto", fileDescriptor_2105bc76e5d5e738) }

var fileDescriptor_2105bc76e5d5e738 = []byte{
	// 298 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x92, 0xcd, 0x4b, 0xc3, 0x40,
	0x10, 0xc5, 0x49, 0xd3, 0xf8, 0x31, 0x41, 0x5a, 0x16, 0x0f, 0x21, 0x82, 0x46, 0x4f, 0x41, 0x30,
	0x81, 0x7a, 0x11, 0xbd, 0x89, 0x82, 0x56, 0xfc, 0x60, 0xc5, 0x8b, 0x37, 0x93, 0x8c, 0x56, 0x4c,
	0x76, 0xc2, 0x66, 0x5b, 0xa9, 0x7f, 0xbb, 0x07, 0xc9, 0x6e, 0x8b, 0x49, 0xf4, 0xd4, 0x4b, 0xc8,
	0xce, 0x7b, 0xef, 0xc7, 0xbe, 0x61, 0x61, 0x58, 0x50, 0x86, 0x79, 0x3c, 0xa1, 0x4a, 0x45, 0xa5,
	0x24, 0x45, 0xcc, 0xd1, 0x13, 0x7f, 0x21, 0x08, 0xca, 0xd0, 0x08, 0xcb, 0x49, 0x89, 0x28, 0xcd,
	0xe4, 0xe0, 0xdb, 0x86, 0xfe, 0x15, 0x55, 0x8a, 0xed, 0x41, 0xff, 0x5a, 0xbc, 0x92, 0x67, 0x05,
	0x56, 0xe8, 0x8e, 0xdc, 0x48, 0x3b, 0xa3, 0x3b, 0xca, 0x90, 0x6b, 0x81, 0x5d, 0xc0, 0x16, 0xc7,
	0x8a, 0xf2, 0x19, 0x66, 0x0f, 0x88, 0xb2, 0xf2, 0x7a, 0x81, 0x1d, 0xba, 0xa3, 0xdd, 0x85, 0xb3,
	0x86, 0x44, 0x2d, 0xc3, 0xa5, 0x50, 0x72, 0xce, 0xdb, 0x21, 0x76, 0x06, 0x70, 0x23, 0xe8, 0x53,
	0x18, 0x84, 0xad, 0x11, 0x3b, 0x4d, 0xc4, 0xaf, 0x6a, 0xf2, 0x0d, 0x3b, 0x1b, 0xc3, 0xe0, 0x49,
	0xc8, 0xd6, 0x25, 0xfa, 0x9a, 0x10, 0x34, 0x09, 0x1d, 0x8b, 0xc1, 0x74, 0x83, 0xcc, 0x83, 0xf5,
	0x47, 0x45, 0x65, 0x89, 0x99, 0xe7, 0x04, 0x56, 0xb8, 0xc1, 0x97, 0x47, 0xff, 0x16, 0xd8, 0xdf,
	0x1e, 0x6c, 0x08, 0xf6, 0x07, 0xce, 0xf5, 0x7a, 0x36, 0x79, 0xfd, 0xcb, 0xf6, 0xc1, 0x99, 0xbd,
	0xe4, 0x53, 0xf4, 0x7a, 0xad, 0x95, 0xd5, 0x19, 0x6e, 0x94, 0xd3, 0xde, 0x89, 0xe5, 0x8f, 0x61,
	0xd0, 0xe9, 0xb4, 0x3a, 0xeb, 0x1e, 0xb6, 0xff, 0x6b, 0xb7, 0x32, 0xf0, 0xfc, 0xf0, 0x39, 0x7c,
	0x7b, 0x57, 0x93, 0x69, 0x12, 0xa5, 0x54, 0xc4, 0x5f, 0x44, 0x49, 0x6a, 0xbe, 0x47, 0x29, 0x49,
	0x8c, 0x53, 0x2a, 0x0a, 0x12, 0xb1, 0xce, 0x26, 0x6b, 0xfa, 0xc5, 0x1c, 0xff, 0x04, 0x00, 0x00,
	0xff, 0xff, 0x6e, 0x57, 0x69, 0x68, 0x70, 0x02, 0x00, 0x00,
}
