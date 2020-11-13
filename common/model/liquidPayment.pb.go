// Code generated by protoc-gen-go. DO NOT EDIT.
// source: model/liquidPayment.proto

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

type LiquidPaymentStatus int32

const (
	LiquidPaymentStatus_LiquidPaymentPending   LiquidPaymentStatus = 0
	LiquidPaymentStatus_LiquidPaymentCompleted LiquidPaymentStatus = 1
)

var LiquidPaymentStatus_name = map[int32]string{
	0: "LiquidPaymentPending",
	1: "LiquidPaymentCompleted",
}

var LiquidPaymentStatus_value = map[string]int32{
	"LiquidPaymentPending":   0,
	"LiquidPaymentCompleted": 1,
}

func (x LiquidPaymentStatus) String() string {
	return proto.EnumName(LiquidPaymentStatus_name, int32(x))
}

func (LiquidPaymentStatus) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_d0147bdf7fdaeca5, []int{0}
}

type LiquidPayment struct {
	ID                   int64               `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
	SenderAddress        []byte              `protobuf:"bytes,2,opt,name=SenderAddress,proto3" json:"SenderAddress,omitempty"`
	RecipientAddress     []byte              `protobuf:"bytes,3,opt,name=RecipientAddress,proto3" json:"RecipientAddress,omitempty"`
	Amount               int64               `protobuf:"varint,4,opt,name=Amount,proto3" json:"Amount,omitempty"`
	AppliedTime          int64               `protobuf:"varint,5,opt,name=AppliedTime,proto3" json:"AppliedTime,omitempty"`
	CompleteMinutes      uint64              `protobuf:"varint,6,opt,name=CompleteMinutes,proto3" json:"CompleteMinutes,omitempty"`
	Status               LiquidPaymentStatus `protobuf:"varint,7,opt,name=Status,proto3,enum=model.LiquidPaymentStatus" json:"Status,omitempty"`
	BlockHeight          uint32              `protobuf:"varint,8,opt,name=BlockHeight,proto3" json:"BlockHeight,omitempty"`
	Latest               bool                `protobuf:"varint,9,opt,name=Latest,proto3" json:"Latest,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *LiquidPayment) Reset()         { *m = LiquidPayment{} }
func (m *LiquidPayment) String() string { return proto.CompactTextString(m) }
func (*LiquidPayment) ProtoMessage()    {}
func (*LiquidPayment) Descriptor() ([]byte, []int) {
	return fileDescriptor_d0147bdf7fdaeca5, []int{0}
}

func (m *LiquidPayment) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LiquidPayment.Unmarshal(m, b)
}
func (m *LiquidPayment) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LiquidPayment.Marshal(b, m, deterministic)
}
func (m *LiquidPayment) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LiquidPayment.Merge(m, src)
}
func (m *LiquidPayment) XXX_Size() int {
	return xxx_messageInfo_LiquidPayment.Size(m)
}
func (m *LiquidPayment) XXX_DiscardUnknown() {
	xxx_messageInfo_LiquidPayment.DiscardUnknown(m)
}

var xxx_messageInfo_LiquidPayment proto.InternalMessageInfo

func (m *LiquidPayment) GetID() int64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *LiquidPayment) GetSenderAddress() []byte {
	if m != nil {
		return m.SenderAddress
	}
	return nil
}

func (m *LiquidPayment) GetRecipientAddress() []byte {
	if m != nil {
		return m.RecipientAddress
	}
	return nil
}

func (m *LiquidPayment) GetAmount() int64 {
	if m != nil {
		return m.Amount
	}
	return 0
}

func (m *LiquidPayment) GetAppliedTime() int64 {
	if m != nil {
		return m.AppliedTime
	}
	return 0
}

func (m *LiquidPayment) GetCompleteMinutes() uint64 {
	if m != nil {
		return m.CompleteMinutes
	}
	return 0
}

func (m *LiquidPayment) GetStatus() LiquidPaymentStatus {
	if m != nil {
		return m.Status
	}
	return LiquidPaymentStatus_LiquidPaymentPending
}

func (m *LiquidPayment) GetBlockHeight() uint32 {
	if m != nil {
		return m.BlockHeight
	}
	return 0
}

func (m *LiquidPayment) GetLatest() bool {
	if m != nil {
		return m.Latest
	}
	return false
}

func init() {
	proto.RegisterEnum("model.LiquidPaymentStatus", LiquidPaymentStatus_name, LiquidPaymentStatus_value)
	proto.RegisterType((*LiquidPayment)(nil), "model.LiquidPayment")
}

func init() {
	proto.RegisterFile("model/liquidPayment.proto", fileDescriptor_d0147bdf7fdaeca5)
}

var fileDescriptor_d0147bdf7fdaeca5 = []byte{
	// 321 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x91, 0x5f, 0x4f, 0xf2, 0x30,
	0x14, 0xc6, 0xdf, 0x0e, 0xd8, 0x8b, 0x07, 0x51, 0x52, 0x0d, 0xa9, 0x5c, 0x35, 0x86, 0x8b, 0x86,
	0xe8, 0x66, 0xf0, 0x13, 0x80, 0x5c, 0x48, 0xc4, 0x84, 0x0c, 0xaf, 0xbc, 0x83, 0xf5, 0x04, 0x1a,
	0xd7, 0x76, 0x6e, 0xdd, 0x85, 0x7e, 0x4d, 0xbf, 0x90, 0xc9, 0xf8, 0x13, 0xa6, 0xde, 0x34, 0xe9,
	0xef, 0xf9, 0xa5, 0x3d, 0x4f, 0x0e, 0x5c, 0x69, 0x2b, 0x31, 0x09, 0x13, 0xf5, 0x5e, 0x28, 0x39,
	0x5f, 0x7e, 0x68, 0x34, 0x2e, 0x48, 0x33, 0xeb, 0x2c, 0x6d, 0x94, 0xd1, 0xf5, 0x97, 0x07, 0xed,
	0xd9, 0x71, 0x4c, 0x29, 0x78, 0xd3, 0x09, 0x23, 0x9c, 0x88, 0xda, 0xd8, 0xbb, 0x23, 0x91, 0x37,
	0x9d, 0xd0, 0x3e, 0xb4, 0x17, 0x68, 0x24, 0x66, 0x23, 0x29, 0x33, 0xcc, 0x73, 0xe6, 0x71, 0x22,
	0x4e, 0xa3, 0x2a, 0xa4, 0x03, 0xe8, 0x44, 0x18, 0xab, 0x54, 0xa1, 0x71, 0x7b, 0xb1, 0x56, 0x8a,
	0xbf, 0x38, 0xed, 0x81, 0x3f, 0xd2, 0xb6, 0x30, 0x8e, 0xd5, 0x0f, 0x3f, 0xed, 0x08, 0xed, 0x43,
	0x6b, 0x94, 0xa6, 0x89, 0x42, 0xf9, 0xa2, 0x34, 0xb2, 0xc6, 0x41, 0x38, 0xc6, 0xf4, 0x06, 0xce,
	0x1f, 0xac, 0x4e, 0x13, 0x74, 0xf8, 0xac, 0x4c, 0xe1, 0x30, 0x67, 0x3e, 0x27, 0xa2, 0x5e, 0x9a,
	0x3f, 0x23, 0x3a, 0x04, 0x7f, 0xe1, 0x96, 0xae, 0xc8, 0xd9, 0x7f, 0x4e, 0xc4, 0xd9, 0xb0, 0x17,
	0x94, 0xfd, 0x83, 0x4a, 0xf7, 0xad, 0x11, 0xed, 0x4c, 0xca, 0xa1, 0x35, 0x4e, 0x6c, 0xfc, 0xf6,
	0x88, 0x6a, 0xbd, 0x71, 0xac, 0xc9, 0x89, 0x68, 0x47, 0xc7, 0x88, 0x76, 0xc1, 0x9f, 0x2d, 0x1d,
	0xe6, 0x8e, 0x9d, 0x70, 0x22, 0x9a, 0xd1, 0xee, 0x36, 0x78, 0x82, 0x8b, 0x3f, 0x1e, 0xa6, 0x0c,
	0x2e, 0x2b, 0x78, 0x8e, 0x46, 0x2a, 0xb3, 0xee, 0xfc, 0xa3, 0x3d, 0xe8, 0x56, 0x92, 0xfd, 0xf8,
	0xb2, 0x43, 0xc6, 0x83, 0x57, 0xb1, 0x56, 0x6e, 0x53, 0xac, 0x82, 0xd8, 0xea, 0xf0, 0xd3, 0xda,
	0x55, 0xbc, 0x3d, 0x6f, 0x63, 0x9b, 0x61, 0x18, 0x5b, 0xad, 0xad, 0x09, 0xcb, 0x3a, 0x2b, 0xbf,
	0x5c, 0xee, 0xfd, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0x58, 0x75, 0x1c, 0xab, 0xf9, 0x01, 0x00,
	0x00,
}
