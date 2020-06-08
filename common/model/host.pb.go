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

// Host represent data sructure node and listed peers in node it's self
type Host struct {
	Info                 *Node            `protobuf:"bytes,1,opt,name=Info,proto3" json:"Info,omitempty"`
	ResolvedPeers        map[string]*Peer `protobuf:"bytes,2,rep,name=ResolvedPeers,proto3" json:"ResolvedPeers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	UnresolvedPeers      map[string]*Peer `protobuf:"bytes,3,rep,name=UnresolvedPeers,proto3" json:"UnresolvedPeers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	KnownPeers           map[string]*Peer `protobuf:"bytes,4,rep,name=KnownPeers,proto3" json:"KnownPeers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	BlacklistedPeers     map[string]*Peer `protobuf:"bytes,5,rep,name=BlacklistedPeers,proto3" json:"BlacklistedPeers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Stopped              bool             `protobuf:"varint,6,opt,name=Stopped,proto3" json:"Stopped,omitempty"`
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

func (m *Host) GetUnresolvedPeers() map[string]*Peer {
	if m != nil {
		return m.UnresolvedPeers
	}
	return nil
}

func (m *Host) GetKnownPeers() map[string]*Peer {
	if m != nil {
		return m.KnownPeers
	}
	return nil
}

func (m *Host) GetBlacklistedPeers() map[string]*Peer {
	if m != nil {
		return m.BlacklistedPeers
	}
	return nil
}

func (m *Host) GetStopped() bool {
	if m != nil {
		return m.Stopped
	}
	return false
}

type HostInfo struct {
	Host                 *Host            `protobuf:"bytes,1,opt,name=Host,proto3" json:"Host,omitempty"`
	ChainStatuses        []*ChainStatus   `protobuf:"bytes,2,rep,name=ChainStatuses,proto3" json:"ChainStatuses,omitempty"`
	ScrambledNodes       []*Peer          `protobuf:"bytes,3,rep,name=ScrambledNodes,proto3" json:"ScrambledNodes,omitempty"`
	ScrambledNodesHeight uint32           `protobuf:"varint,4,opt,name=ScrambledNodesHeight,proto3" json:"ScrambledNodesHeight,omitempty"`
	PriorityPeers        map[string]*Peer `protobuf:"bytes,5,rep,name=PriorityPeers,proto3" json:"PriorityPeers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *HostInfo) Reset()         { *m = HostInfo{} }
func (m *HostInfo) String() string { return proto.CompactTextString(m) }
func (*HostInfo) ProtoMessage()    {}
func (*HostInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_2105bc76e5d5e738, []int{1}
}

func (m *HostInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HostInfo.Unmarshal(m, b)
}
func (m *HostInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HostInfo.Marshal(b, m, deterministic)
}
func (m *HostInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HostInfo.Merge(m, src)
}
func (m *HostInfo) XXX_Size() int {
	return xxx_messageInfo_HostInfo.Size(m)
}
func (m *HostInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_HostInfo.DiscardUnknown(m)
}

var xxx_messageInfo_HostInfo proto.InternalMessageInfo

func (m *HostInfo) GetHost() *Host {
	if m != nil {
		return m.Host
	}
	return nil
}

func (m *HostInfo) GetChainStatuses() []*ChainStatus {
	if m != nil {
		return m.ChainStatuses
	}
	return nil
}

func (m *HostInfo) GetScrambledNodes() []*Peer {
	if m != nil {
		return m.ScrambledNodes
	}
	return nil
}

func (m *HostInfo) GetScrambledNodesHeight() uint32 {
	if m != nil {
		return m.ScrambledNodesHeight
	}
	return 0
}

func (m *HostInfo) GetPriorityPeers() map[string]*Peer {
	if m != nil {
		return m.PriorityPeers
	}
	return nil
}

type GetHostPeersResponse struct {
	ResolvedPeers        map[string]*Peer `protobuf:"bytes,1,rep,name=ResolvedPeers,proto3" json:"ResolvedPeers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	UnresolvedPeers      map[string]*Peer `protobuf:"bytes,2,rep,name=UnresolvedPeers,proto3" json:"UnresolvedPeers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *GetHostPeersResponse) Reset()         { *m = GetHostPeersResponse{} }
func (m *GetHostPeersResponse) String() string { return proto.CompactTextString(m) }
func (*GetHostPeersResponse) ProtoMessage()    {}
func (*GetHostPeersResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_2105bc76e5d5e738, []int{2}
}

func (m *GetHostPeersResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetHostPeersResponse.Unmarshal(m, b)
}
func (m *GetHostPeersResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetHostPeersResponse.Marshal(b, m, deterministic)
}
func (m *GetHostPeersResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetHostPeersResponse.Merge(m, src)
}
func (m *GetHostPeersResponse) XXX_Size() int {
	return xxx_messageInfo_GetHostPeersResponse.Size(m)
}
func (m *GetHostPeersResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetHostPeersResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetHostPeersResponse proto.InternalMessageInfo

func (m *GetHostPeersResponse) GetResolvedPeers() map[string]*Peer {
	if m != nil {
		return m.ResolvedPeers
	}
	return nil
}

func (m *GetHostPeersResponse) GetUnresolvedPeers() map[string]*Peer {
	if m != nil {
		return m.UnresolvedPeers
	}
	return nil
}

func init() {
	proto.RegisterType((*Host)(nil), "model.Host")
	proto.RegisterMapType((map[string]*Peer)(nil), "model.Host.BlacklistedPeersEntry")
	proto.RegisterMapType((map[string]*Peer)(nil), "model.Host.KnownPeersEntry")
	proto.RegisterMapType((map[string]*Peer)(nil), "model.Host.ResolvedPeersEntry")
	proto.RegisterMapType((map[string]*Peer)(nil), "model.Host.UnresolvedPeersEntry")
	proto.RegisterType((*HostInfo)(nil), "model.HostInfo")
	proto.RegisterMapType((map[string]*Peer)(nil), "model.HostInfo.PriorityPeersEntry")
	proto.RegisterType((*GetHostPeersResponse)(nil), "model.GetHostPeersResponse")
	proto.RegisterMapType((map[string]*Peer)(nil), "model.GetHostPeersResponse.ResolvedPeersEntry")
	proto.RegisterMapType((map[string]*Peer)(nil), "model.GetHostPeersResponse.UnresolvedPeersEntry")
}

func init() {
	proto.RegisterFile("model/host.proto", fileDescriptor_2105bc76e5d5e738)
}

var fileDescriptor_2105bc76e5d5e738 = []byte{
	// 500 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xcc, 0x94, 0xcf, 0x6e, 0xd3, 0x40,
	0x10, 0xc6, 0x65, 0x27, 0x29, 0x65, 0xa2, 0xd0, 0x68, 0x15, 0x90, 0x15, 0x24, 0xea, 0xe6, 0x64,
	0x21, 0xe1, 0xa0, 0xf4, 0x52, 0xc1, 0xad, 0x80, 0x08, 0x45, 0x85, 0x68, 0x03, 0x97, 0xde, 0xfc,
	0x67, 0x68, 0xac, 0xd8, 0x1e, 0xcb, 0xde, 0x14, 0x85, 0x77, 0xe1, 0xf9, 0xe0, 0x31, 0xd0, 0xee,
	0xda, 0xc2, 0x76, 0x22, 0x54, 0xe5, 0xc4, 0x25, 0x8a, 0xbf, 0xf9, 0xe6, 0xa7, 0xdd, 0xd9, 0x6f,
	0x17, 0x86, 0x09, 0x85, 0x18, 0x4f, 0x57, 0x54, 0x08, 0x37, 0xcb, 0x49, 0x10, 0xeb, 0x29, 0x65,
	0x5c, 0x16, 0x52, 0x0a, 0x51, 0x17, 0x2a, 0x25, 0x43, 0xcc, 0x4b, 0xe5, 0x89, 0x56, 0xfc, 0x98,
	0x82, 0x75, 0xb0, 0xf2, 0xa2, 0x54, 0xeb, 0x93, 0x9f, 0x3d, 0xe8, 0xce, 0xa9, 0x10, 0xec, 0x14,
	0xba, 0x1f, 0xd2, 0x6f, 0x64, 0x19, 0xb6, 0xe1, 0xf4, 0x67, 0x7d, 0x57, 0xf9, 0xdd, 0x4f, 0x14,
	0x22, 0x57, 0x05, 0xf6, 0x16, 0x06, 0x1c, 0x0b, 0x8a, 0xef, 0x30, 0x5c, 0x20, 0xe6, 0x85, 0x65,
	0xda, 0x1d, 0xa7, 0x3f, 0x7b, 0x56, 0x3a, 0x25, 0xc4, 0x6d, 0x18, 0xde, 0xa5, 0x22, 0xdf, 0xf2,
	0x66, 0x13, 0xbb, 0x82, 0x93, 0xaf, 0x69, 0xde, 0xe0, 0x74, 0x14, 0xc7, 0xae, 0x73, 0x5a, 0x16,
	0x4d, 0x6a, 0x37, 0xb2, 0xd7, 0x00, 0x1f, 0x53, 0xfa, 0x9e, 0x6a, 0x4c, 0x57, 0x61, 0x9e, 0xd6,
	0x31, 0x7f, 0xab, 0x9a, 0x50, 0xb3, 0xb3, 0x6b, 0x18, 0x5e, 0xc6, 0x5e, 0xb0, 0x8e, 0xa3, 0x42,
	0x54, 0x2b, 0xe9, 0x29, 0xc4, 0x59, 0x1d, 0xd1, 0xf6, 0x68, 0xd0, 0x4e, 0x2b, 0xb3, 0xe0, 0xc1,
	0x52, 0x50, 0x96, 0x61, 0x68, 0x1d, 0xd9, 0x86, 0x73, 0xcc, 0xab, 0xcf, 0xf1, 0x35, 0xb0, 0xdd,
	0xb1, 0xb0, 0x21, 0x74, 0xd6, 0xb8, 0x55, 0xd3, 0x7e, 0xc8, 0xe5, 0x5f, 0x76, 0x06, 0xbd, 0x3b,
	0x2f, 0xde, 0xa0, 0x65, 0x36, 0x4e, 0x40, 0xf6, 0x70, 0x5d, 0x79, 0x65, 0x5e, 0x18, 0xe3, 0xcf,
	0x30, 0xda, 0x37, 0x9d, 0xc3, 0x81, 0x57, 0x70, 0xd2, 0x9a, 0xd3, 0xe1, 0xac, 0x05, 0x3c, 0xde,
	0x3b, 0xb0, 0x83, 0x89, 0x93, 0x5f, 0x26, 0x1c, 0xcb, 0x83, 0x50, 0x11, 0x3c, 0xd5, 0x59, 0x6d,
	0x65, 0x54, 0x4a, 0x5c, 0x87, 0xf8, 0x02, 0x06, 0x6f, 0x64, 0xb8, 0x97, 0xc2, 0x13, 0x9b, 0x02,
	0xab, 0x8c, 0xb2, 0xd2, 0x59, 0xab, 0xf1, 0xa6, 0x91, 0x9d, 0xc3, 0xa3, 0x65, 0x90, 0x7b, 0x89,
	0x1f, 0x63, 0x28, 0x43, 0x5f, 0xc5, 0xb2, 0xb1, 0xae, 0x96, 0x85, 0xcd, 0x60, 0xd4, 0x54, 0xe6,
	0x18, 0xdd, 0xae, 0x84, 0xd5, 0xb5, 0x0d, 0x67, 0xc0, 0xf7, 0xd6, 0xd8, 0x1c, 0x06, 0x8b, 0x3c,
	0xa2, 0x3c, 0x12, 0xdb, 0x7a, 0xe8, 0x26, 0xb5, 0xcd, 0xc8, 0xbd, 0xba, 0x0d, 0x53, 0x79, 0x95,
	0x1a, 0x9a, 0x0c, 0xd6, 0xae, 0xe9, 0xf0, 0x49, 0xff, 0x36, 0x61, 0xf4, 0x1e, 0x85, 0x5c, 0x80,
	0xc2, 0x71, 0x2c, 0x32, 0x4a, 0x0b, 0x64, 0x5f, 0xda, 0x17, 0xdf, 0x50, 0x2b, 0x76, 0x4b, 0xce,
	0xbe, 0x9e, 0x7b, 0x3c, 0x04, 0x37, 0xbb, 0x0f, 0x81, 0x3e, 0xac, 0x97, 0xff, 0xe2, 0xde, 0xeb,
	0x61, 0xf8, 0xdf, 0xaf, 0xdc, 0xe5, 0xf3, 0x1b, 0xe7, 0x36, 0x12, 0xab, 0x8d, 0xef, 0x06, 0x94,
	0x4c, 0x7f, 0x10, 0xf9, 0x81, 0xfe, 0x7d, 0x11, 0x50, 0x8e, 0xd3, 0x80, 0x92, 0x84, 0xd2, 0xa9,
	0xea, 0xf5, 0x8f, 0xd4, 0x3b, 0x7d, 0xfe, 0x27, 0x00, 0x00, 0xff, 0xff, 0xbb, 0x33, 0x2e, 0xd8,
	0xfe, 0x05, 0x00, 0x00,
}
