// Code generated by protoc-gen-go. DO NOT EDIT.
// source: model/peer.proto

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

type Peer struct {
	Info                 *Node    `protobuf:"bytes,1,opt,name=Info,proto3" json:"Info,omitempty"`
	LastInboundRequest   uint32   `protobuf:"varint,2,opt,name=LastInboundRequest,proto3" json:"LastInboundRequest,omitempty"`
	BlacklistingCause    string   `protobuf:"bytes,3,opt,name=BlacklistingCause,proto3" json:"BlacklistingCause,omitempty"`
	BlacklistingTime     uint64   `protobuf:"varint,4,opt,name=BlacklistingTime,proto3" json:"BlacklistingTime,omitempty"`
	LastUpdated          int64    `protobuf:"varint,5,opt,name=LastUpdated,proto3" json:"LastUpdated,omitempty"`
	ConnectionAttempted  int32    `protobuf:"varint,6,opt,name=connectionAttempted,proto3" json:"connectionAttempted,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Peer) Reset()         { *m = Peer{} }
func (m *Peer) String() string { return proto.CompactTextString(m) }
func (*Peer) ProtoMessage()    {}
func (*Peer) Descriptor() ([]byte, []int) {
	return fileDescriptor_10c2876293c17304, []int{0}
}

func (m *Peer) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Peer.Unmarshal(m, b)
}
func (m *Peer) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Peer.Marshal(b, m, deterministic)
}
func (m *Peer) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Peer.Merge(m, src)
}
func (m *Peer) XXX_Size() int {
	return xxx_messageInfo_Peer.Size(m)
}
func (m *Peer) XXX_DiscardUnknown() {
	xxx_messageInfo_Peer.DiscardUnknown(m)
}

var xxx_messageInfo_Peer proto.InternalMessageInfo

func (m *Peer) GetInfo() *Node {
	if m != nil {
		return m.Info
	}
	return nil
}

func (m *Peer) GetLastInboundRequest() uint32 {
	if m != nil {
		return m.LastInboundRequest
	}
	return 0
}

func (m *Peer) GetBlacklistingCause() string {
	if m != nil {
		return m.BlacklistingCause
	}
	return ""
}

func (m *Peer) GetBlacklistingTime() uint64 {
	if m != nil {
		return m.BlacklistingTime
	}
	return 0
}

func (m *Peer) GetLastUpdated() int64 {
	if m != nil {
		return m.LastUpdated
	}
	return 0
}

func (m *Peer) GetConnectionAttempted() int32 {
	if m != nil {
		return m.ConnectionAttempted
	}
	return 0
}

type PeerBasicResponse struct {
	Success              bool     `protobuf:"varint,1,opt,name=Success,proto3" json:"Success,omitempty"`
	Error                string   `protobuf:"bytes,2,opt,name=Error,proto3" json:"Error,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PeerBasicResponse) Reset()         { *m = PeerBasicResponse{} }
func (m *PeerBasicResponse) String() string { return proto.CompactTextString(m) }
func (*PeerBasicResponse) ProtoMessage()    {}
func (*PeerBasicResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_10c2876293c17304, []int{1}
}

func (m *PeerBasicResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PeerBasicResponse.Unmarshal(m, b)
}
func (m *PeerBasicResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PeerBasicResponse.Marshal(b, m, deterministic)
}
func (m *PeerBasicResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PeerBasicResponse.Merge(m, src)
}
func (m *PeerBasicResponse) XXX_Size() int {
	return xxx_messageInfo_PeerBasicResponse.Size(m)
}
func (m *PeerBasicResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PeerBasicResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PeerBasicResponse proto.InternalMessageInfo

func (m *PeerBasicResponse) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *PeerBasicResponse) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

type GetPeerInfoRequest struct {
	Version              string   `protobuf:"bytes,1,opt,name=Version,proto3" json:"Version,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetPeerInfoRequest) Reset()         { *m = GetPeerInfoRequest{} }
func (m *GetPeerInfoRequest) String() string { return proto.CompactTextString(m) }
func (*GetPeerInfoRequest) ProtoMessage()    {}
func (*GetPeerInfoRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_10c2876293c17304, []int{2}
}

func (m *GetPeerInfoRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetPeerInfoRequest.Unmarshal(m, b)
}
func (m *GetPeerInfoRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetPeerInfoRequest.Marshal(b, m, deterministic)
}
func (m *GetPeerInfoRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetPeerInfoRequest.Merge(m, src)
}
func (m *GetPeerInfoRequest) XXX_Size() int {
	return xxx_messageInfo_GetPeerInfoRequest.Size(m)
}
func (m *GetPeerInfoRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetPeerInfoRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetPeerInfoRequest proto.InternalMessageInfo

func (m *GetPeerInfoRequest) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

type GetMorePeersResponse struct {
	Peers                []*Node  `protobuf:"bytes,1,rep,name=Peers,proto3" json:"Peers,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetMorePeersResponse) Reset()         { *m = GetMorePeersResponse{} }
func (m *GetMorePeersResponse) String() string { return proto.CompactTextString(m) }
func (*GetMorePeersResponse) ProtoMessage()    {}
func (*GetMorePeersResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_10c2876293c17304, []int{3}
}

func (m *GetMorePeersResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetMorePeersResponse.Unmarshal(m, b)
}
func (m *GetMorePeersResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetMorePeersResponse.Marshal(b, m, deterministic)
}
func (m *GetMorePeersResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetMorePeersResponse.Merge(m, src)
}
func (m *GetMorePeersResponse) XXX_Size() int {
	return xxx_messageInfo_GetMorePeersResponse.Size(m)
}
func (m *GetMorePeersResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetMorePeersResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetMorePeersResponse proto.InternalMessageInfo

func (m *GetMorePeersResponse) GetPeers() []*Node {
	if m != nil {
		return m.Peers
	}
	return nil
}

type SendPeersRequest struct {
	Peers                []*Node  `protobuf:"bytes,1,rep,name=Peers,proto3" json:"Peers,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SendPeersRequest) Reset()         { *m = SendPeersRequest{} }
func (m *SendPeersRequest) String() string { return proto.CompactTextString(m) }
func (*SendPeersRequest) ProtoMessage()    {}
func (*SendPeersRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_10c2876293c17304, []int{4}
}

func (m *SendPeersRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SendPeersRequest.Unmarshal(m, b)
}
func (m *SendPeersRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SendPeersRequest.Marshal(b, m, deterministic)
}
func (m *SendPeersRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SendPeersRequest.Merge(m, src)
}
func (m *SendPeersRequest) XXX_Size() int {
	return xxx_messageInfo_SendPeersRequest.Size(m)
}
func (m *SendPeersRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SendPeersRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SendPeersRequest proto.InternalMessageInfo

func (m *SendPeersRequest) GetPeers() []*Node {
	if m != nil {
		return m.Peers
	}
	return nil
}

func init() {
	proto.RegisterType((*Peer)(nil), "model.Peer")
	proto.RegisterType((*PeerBasicResponse)(nil), "model.PeerBasicResponse")
	proto.RegisterType((*GetPeerInfoRequest)(nil), "model.GetPeerInfoRequest")
	proto.RegisterType((*GetMorePeersResponse)(nil), "model.GetMorePeersResponse")
	proto.RegisterType((*SendPeersRequest)(nil), "model.SendPeersRequest")
}

func init() { proto.RegisterFile("model/peer.proto", fileDescriptor_10c2876293c17304) }

var fileDescriptor_10c2876293c17304 = []byte{
	// 357 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x92, 0xc1, 0xeb, 0xd3, 0x30,
	0x14, 0xc7, 0xc9, 0x6f, 0xed, 0xe6, 0xde, 0x10, 0xb6, 0xb8, 0x43, 0xf0, 0x62, 0xec, 0xa9, 0x0c,
	0x6d, 0x45, 0xf1, 0xe0, 0xd1, 0x0d, 0x19, 0x03, 0x15, 0xc9, 0xd4, 0x83, 0xb7, 0x36, 0x7d, 0xce,
	0xe2, 0x9a, 0x57, 0x93, 0xf4, 0xe2, 0xdd, 0xff, 0x5b, 0x9a, 0x6e, 0x32, 0xdc, 0xe0, 0x77, 0x09,
	0xbc, 0xef, 0xe7, 0x7d, 0xdf, 0x23, 0xdf, 0x04, 0xe6, 0x0d, 0x55, 0x78, 0xcc, 0x5b, 0x44, 0x9b,
	0xb5, 0x96, 0x3c, 0xf1, 0x38, 0x28, 0x8f, 0x4f, 0xc0, 0x50, 0x85, 0x03, 0x48, 0xfe, 0xdc, 0x41,
	0xf4, 0x09, 0xd1, 0xf2, 0x27, 0x10, 0xed, 0xcc, 0x77, 0x12, 0x4c, 0xb2, 0x74, 0xf6, 0x72, 0x96,
	0x85, 0xce, 0xec, 0x23, 0x55, 0xa8, 0x02, 0xe0, 0x19, 0xf0, 0xf7, 0x85, 0xf3, 0x3b, 0x53, 0x52,
	0x67, 0x2a, 0x85, 0xbf, 0x3a, 0x74, 0x5e, 0xdc, 0x49, 0x96, 0x3e, 0x54, 0x37, 0x08, 0x7f, 0x06,
	0x8b, 0xf5, 0xb1, 0xd0, 0x3f, 0x8f, 0xb5, 0xf3, 0xb5, 0x39, 0x6c, 0x8a, 0xce, 0xa1, 0x18, 0x49,
	0x96, 0x4e, 0xd5, 0x35, 0xe0, 0x2b, 0x98, 0x5f, 0x8a, 0x9f, 0xeb, 0x06, 0x45, 0x24, 0x59, 0x1a,
	0xa9, 0x2b, 0x9d, 0x4b, 0x98, 0xf5, 0xfb, 0xbe, 0xb4, 0x55, 0xe1, 0xb1, 0x12, 0xb1, 0x64, 0xe9,
	0x48, 0x5d, 0x4a, 0xfc, 0x05, 0x3c, 0xd2, 0x64, 0x0c, 0x6a, 0x5f, 0x93, 0x79, 0xeb, 0x3d, 0x36,
	0x6d, 0xdf, 0x39, 0x96, 0x2c, 0x8d, 0xd5, 0x2d, 0x94, 0x6c, 0x60, 0xd1, 0xc7, 0xb0, 0x2e, 0x5c,
	0xad, 0x15, 0xba, 0x96, 0x8c, 0x43, 0x2e, 0x60, 0xb2, 0xef, 0xb4, 0x46, 0xe7, 0x42, 0x2c, 0x0f,
	0xd4, 0xb9, 0xe4, 0x4b, 0x88, 0xdf, 0x59, 0x4b, 0x36, 0xdc, 0x7f, 0xaa, 0x86, 0x22, 0xc9, 0x80,
	0x6f, 0xd1, 0xf7, 0x73, 0xfa, 0xc4, 0xce, 0x41, 0x08, 0x98, 0x7c, 0x45, 0xeb, 0x6a, 0x32, 0x61,
	0xca, 0x54, 0x9d, 0xcb, 0xe4, 0x0d, 0x2c, 0xb7, 0xe8, 0x3f, 0x90, 0xc5, 0xde, 0xe3, 0xfe, 0xed,
	0x7d, 0x0a, 0x71, 0x10, 0x04, 0x93, 0xa3, 0xff, 0x1f, 0x63, 0x20, 0xc9, 0x6b, 0x98, 0xef, 0xd1,
	0x54, 0x27, 0xdf, 0xb0, 0xe8, 0x7e, 0xdb, 0x7a, 0xf5, 0x2d, 0x3d, 0xd4, 0xfe, 0x47, 0x57, 0x66,
	0x9a, 0x9a, 0xfc, 0x37, 0x51, 0xa9, 0x87, 0xf3, 0xb9, 0x26, 0x8b, 0xb9, 0xa6, 0xa6, 0x21, 0x93,
	0x07, 0x5f, 0x39, 0x0e, 0x3f, 0xe4, 0xd5, 0xdf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x6c, 0x2b, 0x21,
	0x42, 0x4e, 0x02, 0x00, 0x00,
}
