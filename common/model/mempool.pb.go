// Code generated by protoc-gen-go. DO NOT EDIT.
// source: model/mempool.proto

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

// Mempool represent the mempool data structure stored in the database
type MempoolTransaction struct {
	ID                   int64    `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
	FeePerByte           int32    `protobuf:"varint,2,opt,name=FeePerByte,proto3" json:"FeePerByte,omitempty"`
	ArrivalTimestamp     int64    `protobuf:"varint,3,opt,name=ArrivalTimestamp,proto3" json:"ArrivalTimestamp,omitempty"`
	TransactionBytes     []byte   `protobuf:"bytes,4,opt,name=TransactionBytes,proto3" json:"TransactionBytes,omitempty"`
	SenderAccountID      []byte   `protobuf:"bytes,5,opt,name=SenderAccountID,proto3" json:"SenderAccountID,omitempty"`
	RecipientAccountID   []byte   `protobuf:"bytes,6,opt,name=RecipientAccountID,proto3" json:"RecipientAccountID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MempoolTransaction) Reset()         { *m = MempoolTransaction{} }
func (m *MempoolTransaction) String() string { return proto.CompactTextString(m) }
func (*MempoolTransaction) ProtoMessage()    {}
func (*MempoolTransaction) Descriptor() ([]byte, []int) {
	return fileDescriptor_22ea31ac6d427b7b, []int{0}
}

func (m *MempoolTransaction) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MempoolTransaction.Unmarshal(m, b)
}
func (m *MempoolTransaction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MempoolTransaction.Marshal(b, m, deterministic)
}
func (m *MempoolTransaction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MempoolTransaction.Merge(m, src)
}
func (m *MempoolTransaction) XXX_Size() int {
	return xxx_messageInfo_MempoolTransaction.Size(m)
}
func (m *MempoolTransaction) XXX_DiscardUnknown() {
	xxx_messageInfo_MempoolTransaction.DiscardUnknown(m)
}

var xxx_messageInfo_MempoolTransaction proto.InternalMessageInfo

func (m *MempoolTransaction) GetID() int64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *MempoolTransaction) GetFeePerByte() int32 {
	if m != nil {
		return m.FeePerByte
	}
	return 0
}

func (m *MempoolTransaction) GetArrivalTimestamp() int64 {
	if m != nil {
		return m.ArrivalTimestamp
	}
	return 0
}

func (m *MempoolTransaction) GetTransactionBytes() []byte {
	if m != nil {
		return m.TransactionBytes
	}
	return nil
}

func (m *MempoolTransaction) GetSenderAccountID() []byte {
	if m != nil {
		return m.SenderAccountID
	}
	return nil
}

func (m *MempoolTransaction) GetRecipientAccountID() []byte {
	if m != nil {
		return m.RecipientAccountID
	}
	return nil
}

type GetMempoolTransactionRequest struct {
	// Fetch Mempool Transaction by its TransactionBytes
	TransactionBytes     []byte   `protobuf:"bytes,1,opt,name=TransactionBytes,proto3" json:"TransactionBytes,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetMempoolTransactionRequest) Reset()         { *m = GetMempoolTransactionRequest{} }
func (m *GetMempoolTransactionRequest) String() string { return proto.CompactTextString(m) }
func (*GetMempoolTransactionRequest) ProtoMessage()    {}
func (*GetMempoolTransactionRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_22ea31ac6d427b7b, []int{1}
}

func (m *GetMempoolTransactionRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetMempoolTransactionRequest.Unmarshal(m, b)
}
func (m *GetMempoolTransactionRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetMempoolTransactionRequest.Marshal(b, m, deterministic)
}
func (m *GetMempoolTransactionRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetMempoolTransactionRequest.Merge(m, src)
}
func (m *GetMempoolTransactionRequest) XXX_Size() int {
	return xxx_messageInfo_GetMempoolTransactionRequest.Size(m)
}
func (m *GetMempoolTransactionRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetMempoolTransactionRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetMempoolTransactionRequest proto.InternalMessageInfo

func (m *GetMempoolTransactionRequest) GetTransactionBytes() []byte {
	if m != nil {
		return m.TransactionBytes
	}
	return nil
}

type GetMempoolTransactionsRequest struct {
	// Fetch Mempool transactions from arrival timestamp
	From int64 `protobuf:"varint,1,opt,name=From,proto3" json:"From,omitempty"`
	// Fetch Mempool transactions to arrival timestamp
	To                   int64    `protobuf:"varint,2,opt,name=To,proto3" json:"To,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetMempoolTransactionsRequest) Reset()         { *m = GetMempoolTransactionsRequest{} }
func (m *GetMempoolTransactionsRequest) String() string { return proto.CompactTextString(m) }
func (*GetMempoolTransactionsRequest) ProtoMessage()    {}
func (*GetMempoolTransactionsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_22ea31ac6d427b7b, []int{2}
}

func (m *GetMempoolTransactionsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetMempoolTransactionsRequest.Unmarshal(m, b)
}
func (m *GetMempoolTransactionsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetMempoolTransactionsRequest.Marshal(b, m, deterministic)
}
func (m *GetMempoolTransactionsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetMempoolTransactionsRequest.Merge(m, src)
}
func (m *GetMempoolTransactionsRequest) XXX_Size() int {
	return xxx_messageInfo_GetMempoolTransactionsRequest.Size(m)
}
func (m *GetMempoolTransactionsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetMempoolTransactionsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetMempoolTransactionsRequest proto.InternalMessageInfo

func (m *GetMempoolTransactionsRequest) GetFrom() int64 {
	if m != nil {
		return m.From
	}
	return 0
}

func (m *GetMempoolTransactionsRequest) GetTo() int64 {
	if m != nil {
		return m.To
	}
	return 0
}

type GetMempoolTransactionsResponse struct {
	// Number of transactions returned
	MempoolSize uint32 `protobuf:"varint,1,opt,name=MempoolSize,proto3" json:"MempoolSize,omitempty"`
	// Mempool transactions returned
	MempoolTransactions  []*MempoolTransaction `protobuf:"bytes,2,rep,name=MempoolTransactions,proto3" json:"MempoolTransactions,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *GetMempoolTransactionsResponse) Reset()         { *m = GetMempoolTransactionsResponse{} }
func (m *GetMempoolTransactionsResponse) String() string { return proto.CompactTextString(m) }
func (*GetMempoolTransactionsResponse) ProtoMessage()    {}
func (*GetMempoolTransactionsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_22ea31ac6d427b7b, []int{3}
}

func (m *GetMempoolTransactionsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetMempoolTransactionsResponse.Unmarshal(m, b)
}
func (m *GetMempoolTransactionsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetMempoolTransactionsResponse.Marshal(b, m, deterministic)
}
func (m *GetMempoolTransactionsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetMempoolTransactionsResponse.Merge(m, src)
}
func (m *GetMempoolTransactionsResponse) XXX_Size() int {
	return xxx_messageInfo_GetMempoolTransactionsResponse.Size(m)
}
func (m *GetMempoolTransactionsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetMempoolTransactionsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetMempoolTransactionsResponse proto.InternalMessageInfo

func (m *GetMempoolTransactionsResponse) GetMempoolSize() uint32 {
	if m != nil {
		return m.MempoolSize
	}
	return 0
}

func (m *GetMempoolTransactionsResponse) GetMempoolTransactions() []*MempoolTransaction {
	if m != nil {
		return m.MempoolTransactions
	}
	return nil
}

func init() {
	proto.RegisterType((*MempoolTransaction)(nil), "model.MempoolTransaction")
	proto.RegisterType((*GetMempoolTransactionRequest)(nil), "model.GetMempoolTransactionRequest")
	proto.RegisterType((*GetMempoolTransactionsRequest)(nil), "model.GetMempoolTransactionsRequest")
	proto.RegisterType((*GetMempoolTransactionsResponse)(nil), "model.GetMempoolTransactionsResponse")
}

func init() { proto.RegisterFile("model/mempool.proto", fileDescriptor_22ea31ac6d427b7b) }

var fileDescriptor_22ea31ac6d427b7b = []byte{
	// 324 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x92, 0xcf, 0x4a, 0xc3, 0x40,
	0x10, 0xc6, 0x49, 0xd2, 0xf6, 0x30, 0xf5, 0x1f, 0xdb, 0x4b, 0x04, 0x2d, 0x21, 0xa7, 0x50, 0x30,
	0x01, 0x7d, 0x82, 0xd6, 0x52, 0xa9, 0x22, 0x48, 0x9a, 0x93, 0xb7, 0x74, 0x3b, 0xe8, 0x42, 0x77,
	0x27, 0xee, 0x6e, 0x05, 0xfb, 0x12, 0xbe, 0xad, 0x67, 0xe9, 0xb6, 0x62, 0xb1, 0xeb, 0x25, 0x84,
	0x6f, 0xbe, 0xf9, 0xed, 0xcc, 0xb7, 0x0b, 0x3d, 0x49, 0x0b, 0x5c, 0x16, 0x12, 0x65, 0x43, 0xb4,
	0xcc, 0x1b, 0x4d, 0x96, 0x58, 0xdb, 0x89, 0xe9, 0x57, 0x00, 0xec, 0x71, 0x5b, 0xa8, 0x74, 0xad,
	0x4c, 0xcd, 0xad, 0x20, 0xc5, 0x4e, 0x20, 0x9c, 0x8e, 0xe3, 0x20, 0x09, 0xb2, 0xa8, 0x0c, 0xa7,
	0x63, 0xd6, 0x07, 0x98, 0x20, 0x3e, 0xa1, 0x1e, 0x7d, 0x58, 0x8c, 0xc3, 0x24, 0xc8, 0xda, 0xe5,
	0x9e, 0xc2, 0x06, 0x70, 0x36, 0xd4, 0x5a, 0xbc, 0xd7, 0xcb, 0x4a, 0x48, 0x34, 0xb6, 0x96, 0x4d,
	0x1c, 0xb9, 0xee, 0x03, 0x7d, 0xe3, 0xdd, 0x3b, 0x6a, 0xd3, 0x6e, 0xe2, 0x56, 0x12, 0x64, 0x47,
	0xe5, 0x81, 0xce, 0x32, 0x38, 0x9d, 0xa1, 0x5a, 0xa0, 0x1e, 0x72, 0x4e, 0x2b, 0x65, 0xa7, 0xe3,
	0xb8, 0xed, 0xac, 0x7f, 0x65, 0x96, 0x03, 0x2b, 0x91, 0x8b, 0x46, 0xa0, 0xb2, 0xbf, 0xe6, 0x8e,
	0x33, 0x7b, 0x2a, 0xe9, 0x3d, 0x5c, 0xdc, 0xa1, 0x3d, 0x5c, 0xbd, 0xc4, 0xb7, 0x15, 0x1a, 0xeb,
	0x9d, 0x32, 0xf0, 0x4f, 0x99, 0xde, 0xc2, 0xa5, 0x97, 0x65, 0x7e, 0x60, 0x0c, 0x5a, 0x13, 0x4d,
	0x72, 0x17, 0xa8, 0xfb, 0xdf, 0x44, 0x5c, 0x91, 0x8b, 0x32, 0x2a, 0xc3, 0x8a, 0xd2, 0xcf, 0x00,
	0xfa, 0xff, 0x51, 0x4c, 0x43, 0xca, 0x20, 0x4b, 0xa0, 0xbb, 0x2b, 0xcf, 0xc4, 0x1a, 0x1d, 0xed,
	0xb8, 0xdc, 0x97, 0xd8, 0x03, 0xf4, 0x3c, 0x80, 0x38, 0x4c, 0xa2, 0xac, 0x7b, 0x7d, 0x9e, 0xbb,
	0x3b, 0xcf, 0x3d, 0x4b, 0xfb, 0xba, 0x46, 0x83, 0xe7, 0xec, 0x45, 0xd8, 0xd7, 0xd5, 0x3c, 0xe7,
	0x24, 0x8b, 0x35, 0xd1, 0x9c, 0x6f, 0xbf, 0x57, 0x9c, 0x34, 0x16, 0x9c, 0xa4, 0x24, 0x55, 0x38,
	0xe6, 0xbc, 0xe3, 0x5e, 0xd5, 0xcd, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0x80, 0x97, 0x8a, 0xb5,
	0x6c, 0x02, 0x00, 0x00,
}
