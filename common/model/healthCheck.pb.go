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
	LastBlockHeight      uint32   `protobuf:"varint,1,opt,name=LastBlockHeight,proto3" json:"LastBlockHeight,omitempty"`
	LastBlockHash        []byte   `protobuf:"bytes,2,opt,name=LastBlockHash,proto3" json:"LastBlockHash,omitempty"`
	Version              string   `protobuf:"bytes,3,opt,name=Version,proto3" json:"Version,omitempty"`
	NodePublicKey        string   `protobuf:"bytes,4,opt,name=NodePublicKey,proto3" json:"NodePublicKey,omitempty"`
	UnresolvedPeers      uint32   `protobuf:"varint,5,opt,name=UnresolvedPeers,proto3" json:"UnresolvedPeers,omitempty"`
	ResolvedPeers        uint64   `protobuf:"varint,6,opt,name=ResolvedPeers,proto3" json:"ResolvedPeers,omitempty"`
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

func (m *GetNodeStatusResponse) GetLastBlockHeight() uint32 {
	if m != nil {
		return m.LastBlockHeight
	}
	return 0
}

func (m *GetNodeStatusResponse) GetLastBlockHash() []byte {
	if m != nil {
		return m.LastBlockHash
	}
	return nil
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

func (m *GetNodeStatusResponse) GetResolvedPeers() uint64 {
	if m != nil {
		return m.ResolvedPeers
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
	// 254 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x90, 0xc1, 0x4b, 0xc3, 0x30,
	0x14, 0xc6, 0x89, 0xae, 0x13, 0x83, 0x43, 0xa8, 0x8a, 0x39, 0x96, 0xe1, 0x21, 0x28, 0xae, 0x07,
	0xff, 0x83, 0x79, 0x70, 0xa0, 0xc8, 0x88, 0xe8, 0xc1, 0x5b, 0x9b, 0x3e, 0x96, 0xb2, 0xb4, 0xaf,
	0x24, 0xaf, 0xc2, 0xfc, 0xc7, 0xbd, 0x4a, 0x53, 0x74, 0x64, 0x97, 0xc0, 0xfb, 0x7d, 0xef, 0xcb,
	0xfb, 0xf8, 0xf8, 0x75, 0x83, 0x15, 0xd8, 0xdc, 0x40, 0x61, 0xc9, 0x3c, 0x1a, 0xd0, 0xdb, 0x45,
	0xe7, 0x90, 0x30, 0x4d, 0x82, 0x30, 0xbf, 0xe3, 0x17, 0xab, 0xbd, 0xa6, 0xc0, 0x77, 0xd8, 0x7a,
	0x48, 0x2f, 0x79, 0xa2, 0xa0, 0xb3, 0x3b, 0xc1, 0x32, 0x26, 0x4f, 0xd5, 0x38, 0xcc, 0x7f, 0x18,
	0xbf, 0x7a, 0x02, 0x7a, 0xc5, 0x0a, 0xde, 0xa8, 0xa0, 0xde, 0xff, 0xef, 0x4b, 0x7e, 0xfe, 0x52,
	0x78, 0x5a, 0x5a, 0xd4, 0xdb, 0x15, 0xd4, 0x1b, 0x43, 0xc1, 0x39, 0x53, 0x87, 0x38, 0xbd, 0xe1,
	0xb3, 0x3d, 0x2a, 0xbc, 0x11, 0x47, 0x19, 0x93, 0x67, 0x2a, 0x86, 0xa9, 0xe0, 0x27, 0x1f, 0xe0,
	0x7c, 0x8d, 0xad, 0x38, 0x0e, 0x09, 0xfe, 0xc6, 0xc1, 0x3f, 0xdc, 0x5f, 0xf7, 0xa5, 0xad, 0xf5,
	0x33, 0xec, 0xc4, 0x24, 0xe8, 0x31, 0x1c, 0xf2, 0xbc, 0xb7, 0x0e, 0x3c, 0xda, 0x2f, 0xa8, 0xd6,
	0x00, 0xce, 0x8b, 0x64, 0xcc, 0x73, 0x80, 0x87, 0xff, 0x54, 0xb4, 0x37, 0xcd, 0x98, 0x9c, 0xa8,
	0x18, 0x2e, 0x6f, 0x3f, 0xe5, 0xa6, 0x26, 0xd3, 0x97, 0x0b, 0x8d, 0x4d, 0xfe, 0x8d, 0x58, 0xea,
	0xf1, 0xbd, 0xd7, 0xe8, 0x20, 0xd7, 0xd8, 0x34, 0xd8, 0xe6, 0xa1, 0xd2, 0x72, 0x1a, 0x0a, 0x7e,
	0xf8, 0x0d, 0x00, 0x00, 0xff, 0xff, 0xe6, 0xd2, 0xb3, 0xbf, 0x7b, 0x01, 0x00, 0x00,
}
