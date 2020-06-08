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
	Timestamp            uint64    `protobuf:"varint,5,opt,name=Timestamp,proto3" json:"Timestamp,omitempty"`
	EventType            EventType `protobuf:"varint,6,opt,name=EventType,proto3,enum=model.EventType" json:"EventType,omitempty"`
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

func (m *AccountLedger) GetTimestamp() uint64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *AccountLedger) GetEventType() EventType {
	if m != nil {
		return m.EventType
	}
	return EventType_EventAny
}

type GetAccountLedgersRequest struct {
	AccountAddress       string      `protobuf:"bytes,1,opt,name=AccountAddress,proto3" json:"AccountAddress,omitempty"`
	EventType            EventType   `protobuf:"varint,2,opt,name=EventType,proto3,enum=model.EventType" json:"EventType,omitempty"`
	TransactionID        int64       `protobuf:"varint,3,opt,name=TransactionID,proto3" json:"TransactionID,omitempty"`
	TimestampStart       uint64      `protobuf:"varint,4,opt,name=TimestampStart,proto3" json:"TimestampStart,omitempty"`
	TimestampEnd         uint32      `protobuf:"varint,5,opt,name=TimestampEnd,proto3" json:"TimestampEnd,omitempty"`
	Pagination           *Pagination `protobuf:"bytes,6,opt,name=Pagination,proto3" json:"Pagination,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *GetAccountLedgersRequest) Reset()         { *m = GetAccountLedgersRequest{} }
func (m *GetAccountLedgersRequest) String() string { return proto.CompactTextString(m) }
func (*GetAccountLedgersRequest) ProtoMessage()    {}
func (*GetAccountLedgersRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_8b8de9896218a2b4, []int{1}
}

func (m *GetAccountLedgersRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetAccountLedgersRequest.Unmarshal(m, b)
}
func (m *GetAccountLedgersRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetAccountLedgersRequest.Marshal(b, m, deterministic)
}
func (m *GetAccountLedgersRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetAccountLedgersRequest.Merge(m, src)
}
func (m *GetAccountLedgersRequest) XXX_Size() int {
	return xxx_messageInfo_GetAccountLedgersRequest.Size(m)
}
func (m *GetAccountLedgersRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetAccountLedgersRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetAccountLedgersRequest proto.InternalMessageInfo

func (m *GetAccountLedgersRequest) GetAccountAddress() string {
	if m != nil {
		return m.AccountAddress
	}
	return ""
}

func (m *GetAccountLedgersRequest) GetEventType() EventType {
	if m != nil {
		return m.EventType
	}
	return EventType_EventAny
}

func (m *GetAccountLedgersRequest) GetTransactionID() int64 {
	if m != nil {
		return m.TransactionID
	}
	return 0
}

func (m *GetAccountLedgersRequest) GetTimestampStart() uint64 {
	if m != nil {
		return m.TimestampStart
	}
	return 0
}

func (m *GetAccountLedgersRequest) GetTimestampEnd() uint32 {
	if m != nil {
		return m.TimestampEnd
	}
	return 0
}

func (m *GetAccountLedgersRequest) GetPagination() *Pagination {
	if m != nil {
		return m.Pagination
	}
	return nil
}

type GetAccountLedgersResponse struct {
	// Number of transactions in total
	Total uint64 `protobuf:"varint,1,opt,name=Total,proto3" json:"Total,omitempty"`
	// Transaction transactions returned
	AccountLedgers       []*AccountLedger `protobuf:"bytes,2,rep,name=AccountLedgers,proto3" json:"AccountLedgers,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *GetAccountLedgersResponse) Reset()         { *m = GetAccountLedgersResponse{} }
func (m *GetAccountLedgersResponse) String() string { return proto.CompactTextString(m) }
func (*GetAccountLedgersResponse) ProtoMessage()    {}
func (*GetAccountLedgersResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_8b8de9896218a2b4, []int{2}
}

func (m *GetAccountLedgersResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetAccountLedgersResponse.Unmarshal(m, b)
}
func (m *GetAccountLedgersResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetAccountLedgersResponse.Marshal(b, m, deterministic)
}
func (m *GetAccountLedgersResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetAccountLedgersResponse.Merge(m, src)
}
func (m *GetAccountLedgersResponse) XXX_Size() int {
	return xxx_messageInfo_GetAccountLedgersResponse.Size(m)
}
func (m *GetAccountLedgersResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetAccountLedgersResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetAccountLedgersResponse proto.InternalMessageInfo

func (m *GetAccountLedgersResponse) GetTotal() uint64 {
	if m != nil {
		return m.Total
	}
	return 0
}

func (m *GetAccountLedgersResponse) GetAccountLedgers() []*AccountLedger {
	if m != nil {
		return m.AccountLedgers
	}
	return nil
}

func init() {
	proto.RegisterType((*AccountLedger)(nil), "model.AccountLedger")
	proto.RegisterType((*GetAccountLedgersRequest)(nil), "model.GetAccountLedgersRequest")
	proto.RegisterType((*GetAccountLedgersResponse)(nil), "model.GetAccountLedgersResponse")
}

func init() {
	proto.RegisterFile("model/accountLedger.proto", fileDescriptor_8b8de9896218a2b4)
}

var fileDescriptor_8b8de9896218a2b4 = []byte{
	// 389 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0x41, 0xab, 0xd3, 0x40,
	0x14, 0x85, 0x99, 0x24, 0x7d, 0xd0, 0xa9, 0x29, 0xbe, 0x41, 0x64, 0x5e, 0x71, 0x11, 0x82, 0x94,
	0x41, 0x30, 0xd1, 0xba, 0x75, 0xd3, 0x68, 0x51, 0xc1, 0x85, 0x8c, 0x59, 0xb9, 0x9b, 0x4e, 0x2e,
	0x69, 0x30, 0x99, 0x89, 0x99, 0xa9, 0xa0, 0x7f, 0xc5, 0x1f, 0xea, 0x56, 0x3a, 0x09, 0x69, 0x53,
	0x15, 0xdd, 0x64, 0xf1, 0x9d, 0x93, 0x7b, 0xef, 0x39, 0x09, 0xbe, 0x6b, 0x74, 0x01, 0x75, 0x2a,
	0xa4, 0xd4, 0x47, 0x65, 0xdf, 0x43, 0x51, 0x42, 0x97, 0xb4, 0x9d, 0xb6, 0x9a, 0xcc, 0x9c, 0xb4,
	0xba, 0xed, 0x1d, 0xf0, 0x15, 0x94, 0xed, 0x95, 0xd5, 0xc3, 0x1e, 0xb5, 0xa2, 0xac, 0x94, 0xb0,
	0x95, 0x56, 0x3d, 0x8f, 0x7f, 0x22, 0x1c, 0x6e, 0x2f, 0x27, 0x91, 0x35, 0x5e, 0x0e, 0x60, 0x5b,
	0x14, 0x1d, 0x18, 0x43, 0x51, 0x84, 0xd8, 0x9c, 0x5f, 0x51, 0xf2, 0x18, 0x87, 0x99, 0xa8, 0x85,
	0x92, 0xf0, 0xea, 0x20, 0x54, 0x09, 0xd4, 0x8b, 0x10, 0xf3, 0xf9, 0x14, 0x92, 0x08, 0x2f, 0xb2,
	0x5a, 0xcb, 0xcf, 0x6f, 0xa1, 0x2a, 0x0f, 0x96, 0xfa, 0x11, 0x62, 0x21, 0xbf, 0x44, 0x84, 0xe1,
	0x30, 0xef, 0x84, 0x32, 0x42, 0x9e, 0xce, 0x7a, 0xf7, 0x9a, 0x06, 0xa7, 0x39, 0x99, 0xf7, 0x0c,
	0xf1, 0xa9, 0x40, 0x1e, 0xe1, 0x79, 0x5e, 0x35, 0x60, 0xac, 0x68, 0x5a, 0x3a, 0x8b, 0x10, 0x0b,
	0xf8, 0x19, 0x90, 0x04, 0xcf, 0x77, 0xa7, 0xc0, 0xf9, 0xb7, 0x16, 0xe8, 0x4d, 0x84, 0xd8, 0x72,
	0x73, 0x3f, 0x71, 0xa9, 0x93, 0x91, 0xf3, 0xb3, 0x25, 0xfe, 0xe1, 0x61, 0xfa, 0x06, 0xec, 0x24,
	0xbc, 0xe1, 0xf0, 0xe5, 0x08, 0xc6, 0xfe, 0x77, 0x09, 0x93, 0xa5, 0xde, 0x3f, 0x97, 0xfe, 0x1e,
	0xd6, 0xff, 0x5b, 0xd8, 0x35, 0x5e, 0x8e, 0xd9, 0x3e, 0x5a, 0xd1, 0x59, 0xd7, 0x4b, 0xc0, 0xaf,
	0x28, 0x89, 0xf1, 0xbd, 0x91, 0xec, 0x54, 0xe1, 0x7a, 0x09, 0xf9, 0x84, 0x91, 0xe7, 0x18, 0x7f,
	0x18, 0x3f, 0xbc, 0xeb, 0x66, 0xb1, 0xb9, 0x1d, 0xce, 0x3c, 0x0b, 0xfc, 0xc2, 0x14, 0x1b, 0x7c,
	0xf7, 0x87, 0x72, 0x4c, 0xab, 0x95, 0x01, 0x42, 0xf1, 0x2c, 0xd7, 0x56, 0xd4, 0xae, 0x94, 0xc0,
	0x5d, 0xdf, 0x03, 0xf2, 0x72, 0xec, 0x6d, 0x78, 0x87, 0x7a, 0x91, 0xcf, 0x16, 0x9b, 0x07, 0xc3,
	0xb6, 0x89, 0xc8, 0xaf, 0xbc, 0xd9, 0x93, 0x4f, 0xac, 0xac, 0xec, 0xe1, 0xb8, 0x4f, 0xa4, 0x6e,
	0xd2, 0xef, 0x5a, 0xef, 0x65, 0xff, 0x7c, 0x2a, 0x75, 0x07, 0xa9, 0xd4, 0x4d, 0xa3, 0x55, 0xea,
	0x26, 0xed, 0x6f, 0xdc, 0xff, 0xfb, 0xe2, 0x57, 0x00, 0x00, 0x00, 0xff, 0xff, 0x54, 0x9e, 0xe9,
	0xf1, 0x0e, 0x03, 0x00, 0x00,
}
