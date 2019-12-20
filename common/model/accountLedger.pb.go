// Code generated by protoc-gen-go. DO NOT EDIT.
// source: model/accountLedger.proto

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

type AccountLedger struct {
	AccountAddress       string    `protobuf:"bytes,1,opt,name=AccountAddress,proto3" json:"AccountAddress,omitempty"`
	BalanceChange        int64     `protobuf:"varint,2,opt,name=BalanceChange,proto3" json:"BalanceChange,omitempty"`
	BlockHeight          uint32    `protobuf:"varint,3,opt,name=BlockHeight,proto3" json:"BlockHeight,omitempty"`
	TransactionID        int64     `protobuf:"varint,4,opt,name=TransactionID,proto3" json:"TransactionID,omitempty"`
	EventType            EventType `protobuf:"varint,5,opt,name=EventType,proto3,enum=model.EventType" json:"EventType,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *AccountLedger) Reset()         { *m = AccountLedger{} }
func (m *AccountLedger) String() string { return proto.CompactTextString(m) }
func (*AccountLedger) ProtoMessage()    {}
func (*AccountLedger) Descriptor() ([]byte, []int) {
	return fileDescriptor_8b8de9896218a2b4, []int{0}
}

func (m *AccountLedger) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AccountLedger.Unmarshal(m, b)
}
func (m *AccountLedger) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AccountLedger.Marshal(b, m, deterministic)
}
func (m *AccountLedger) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AccountLedger.Merge(m, src)
}
func (m *AccountLedger) XXX_Size() int {
	return xxx_messageInfo_AccountLedger.Size(m)
}
func (m *AccountLedger) XXX_DiscardUnknown() {
	xxx_messageInfo_AccountLedger.DiscardUnknown(m)
}

var xxx_messageInfo_AccountLedger proto.InternalMessageInfo

func (m *AccountLedger) GetAccountAddress() string {
	if m != nil {
		return m.AccountAddress
	}
	return ""
}

func (m *AccountLedger) GetBalanceChange() int64 {
	if m != nil {
		return m.BalanceChange
	}
	return 0
}

func (m *AccountLedger) GetBlockHeight() uint32 {
	if m != nil {
		return m.BlockHeight
	}
	return 0
}

func (m *AccountLedger) GetTransactionID() int64 {
	if m != nil {
		return m.TransactionID
	}
	return 0
}

func (m *AccountLedger) GetEventType() EventType {
	if m != nil {
		return m.EventType
	}
	return EventType_EventSendMoneyTransaction
}

func init() {
	proto.RegisterType((*AccountLedger)(nil), "model.AccountLedger")
}

func init() { proto.RegisterFile("model/accountLedger.proto", fileDescriptor_8b8de9896218a2b4) }

var fileDescriptor_8b8de9896218a2b4 = []byte{
	// 235 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x90, 0xc1, 0x4a, 0xc4, 0x30,
	0x10, 0x86, 0x89, 0xeb, 0x0a, 0x1b, 0xe9, 0xa2, 0x39, 0x45, 0x4f, 0x41, 0x44, 0x82, 0x60, 0x0a,
	0xfa, 0x04, 0x5b, 0x15, 0x14, 0x3c, 0x95, 0x3d, 0x79, 0x4b, 0xa7, 0x43, 0x5b, 0x6c, 0x33, 0x4b,
	0x9a, 0x15, 0xf4, 0x3d, 0x7d, 0x1f, 0xd9, 0x44, 0x74, 0xeb, 0x25, 0x87, 0xef, 0xff, 0xf3, 0x31,
	0x33, 0xfc, 0x6c, 0xa0, 0x1a, 0xfb, 0xdc, 0x02, 0xd0, 0xd6, 0x85, 0x17, 0xac, 0x1b, 0xf4, 0x66,
	0xe3, 0x29, 0x90, 0x98, 0xc7, 0xe8, 0xfc, 0x34, 0x35, 0xf0, 0x1d, 0x5d, 0x48, 0xc9, 0xc5, 0x17,
	0xe3, 0xd9, 0x6a, 0xff, 0x87, 0xb8, 0xe2, 0xcb, 0x1f, 0xb0, 0xaa, 0x6b, 0x8f, 0xe3, 0x28, 0x99,
	0x62, 0x7a, 0x51, 0xfe, 0xa3, 0xe2, 0x92, 0x67, 0x85, 0xed, 0xad, 0x03, 0xbc, 0x6f, 0xad, 0x6b,
	0x50, 0x1e, 0x28, 0xa6, 0x67, 0xe5, 0x14, 0x0a, 0xc5, 0x8f, 0x8b, 0x9e, 0xe0, 0xed, 0x09, 0xbb,
	0xa6, 0x0d, 0x72, 0xa6, 0x98, 0xce, 0xca, 0x7d, 0xb4, 0xf3, 0xac, 0xbd, 0x75, 0xa3, 0x85, 0xd0,
	0x91, 0x7b, 0x7e, 0x90, 0x87, 0xc9, 0x33, 0x81, 0xc2, 0xf0, 0xc5, 0xe3, 0x6e, 0xec, 0xf5, 0xc7,
	0x06, 0xe5, 0x5c, 0x31, 0xbd, 0xbc, 0x3d, 0x31, 0x71, 0x1d, 0xf3, 0xcb, 0xcb, 0xbf, 0x4a, 0x71,
	0xfd, 0xaa, 0x9b, 0x2e, 0xb4, 0xdb, 0xca, 0x00, 0x0d, 0xf9, 0x27, 0x51, 0x05, 0xe9, 0xbd, 0x01,
	0xf2, 0x98, 0x03, 0x0d, 0x03, 0xb9, 0x3c, 0x0a, 0xaa, 0xa3, 0x78, 0x8a, 0xbb, 0xef, 0x00, 0x00,
	0x00, 0xff, 0xff, 0xa2, 0x18, 0x7f, 0x30, 0x41, 0x01, 0x00, 0x00,
}
