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
	ID                      int64    `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
	FeePerByte              int32    `protobuf:"varint,2,opt,name=FeePerByte,proto3" json:"FeePerByte,omitempty"`
	ArrivalTimestamp        int64    `protobuf:"varint,3,opt,name=ArrivalTimestamp,proto3" json:"ArrivalTimestamp,omitempty"`
	TransactionBytes        []byte   `protobuf:"bytes,4,opt,name=TransactionBytes,proto3" json:"TransactionBytes,omitempty"`
	SenderAccountAddress    string   `protobuf:"bytes,5,opt,name=SenderAccountAddress,proto3" json:"SenderAccountAddress,omitempty"`
	RecipientAccountAddress string   `protobuf:"bytes,6,opt,name=RecipientAccountAddress,proto3" json:"RecipientAccountAddress,omitempty"`
	XXX_NoUnkeyedLiteral    struct{} `json:"-"`
	XXX_unrecognized        []byte   `json:"-"`
	XXX_sizecache           int32    `json:"-"`
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

func (m *MempoolTransaction) GetSenderAccountAddress() string {
	if m != nil {
		return m.SenderAccountAddress
	}
	return ""
}

func (m *MempoolTransaction) GetRecipientAccountAddress() string {
	if m != nil {
		return m.RecipientAccountAddress
	}
	return ""
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
	// 337 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x92, 0x4f, 0x4b, 0xf3, 0x40,
	0x10, 0xc6, 0xd9, 0xf4, 0x0f, 0xbc, 0xd3, 0x57, 0x28, 0x5b, 0xd1, 0x08, 0x5a, 0x42, 0x4e, 0xa1,
	0x60, 0x22, 0xf5, 0xe2, 0xb5, 0xa5, 0x54, 0x6a, 0x11, 0x24, 0xcd, 0xc9, 0x5b, 0xba, 0x19, 0x34,
	0xd0, 0xcd, 0xc4, 0xdd, 0xad, 0x60, 0xbf, 0x84, 0xe0, 0x27, 0x96, 0x6e, 0xab, 0x04, 0x9b, 0x5e,
	0x16, 0xf6, 0x79, 0x98, 0xdf, 0xce, 0x3e, 0x33, 0xd0, 0x93, 0x94, 0xe1, 0x2a, 0x92, 0x28, 0x4b,
	0xa2, 0x55, 0x58, 0x2a, 0x32, 0xc4, 0x5b, 0x56, 0xf4, 0xbf, 0x1c, 0xe0, 0x8f, 0x3b, 0x23, 0x51,
	0x69, 0xa1, 0x53, 0x61, 0x72, 0x2a, 0x38, 0x07, 0x67, 0x36, 0x71, 0x99, 0xc7, 0x82, 0xc6, 0xd8,
	0xb9, 0x61, 0xb1, 0x33, 0x9b, 0xf0, 0x3e, 0xc0, 0x14, 0xf1, 0x09, 0xd5, 0xf8, 0xc3, 0xa0, 0xeb,
	0x78, 0x2c, 0x68, 0xc5, 0x15, 0x85, 0x87, 0xd0, 0x1d, 0x29, 0x95, 0xbf, 0xa7, 0xab, 0x24, 0x97,
	0xa8, 0x4d, 0x2a, 0x4b, 0xb7, 0xf1, 0x4b, 0x38, 0xf0, 0xf8, 0x00, 0xba, 0x95, 0x27, 0xb7, 0x08,
	0xed, 0x36, 0x3d, 0x16, 0xfc, 0x8f, 0x0f, 0x74, 0x3e, 0x84, 0xd3, 0x05, 0x16, 0x19, 0xaa, 0x91,
	0x10, 0xb4, 0x2e, 0xcc, 0x28, 0xcb, 0x14, 0x6a, 0xed, 0xb6, 0x3c, 0x16, 0xfc, 0x8b, 0x6b, 0x3d,
	0x7e, 0x07, 0xe7, 0x31, 0x8a, 0xbc, 0xcc, 0xb1, 0x30, 0x7f, 0xca, 0xda, 0xb6, 0xec, 0x98, 0xed,
	0x3f, 0xc0, 0xe5, 0x3d, 0x9a, 0xc3, 0x58, 0x62, 0x7c, 0x5b, 0xa3, 0x36, 0xb5, 0x9d, 0xb3, 0xfa,
	0xce, 0xfd, 0x39, 0x5c, 0xd5, 0xb2, 0xf4, 0x0f, 0xec, 0x0c, 0x9a, 0x53, 0x45, 0xb2, 0x12, 0xb6,
	0xbd, 0x6f, 0x47, 0x90, 0x90, 0x8d, 0x79, 0x3f, 0x82, 0x84, 0xfc, 0x4f, 0x06, 0xfd, 0x63, 0x34,
	0x5d, 0x52, 0xa1, 0x91, 0x7b, 0xd0, 0xd9, 0xdb, 0x8b, 0x7c, 0x83, 0x96, 0x7a, 0x12, 0x57, 0x25,
	0x3e, 0x87, 0x5e, 0x0d, 0xc0, 0x75, 0xbc, 0x46, 0xd0, 0x19, 0x5e, 0x84, 0x76, 0x2f, 0xc2, 0x9a,
	0xcf, 0xd7, 0x55, 0x8d, 0x07, 0xcf, 0xc1, 0x4b, 0x6e, 0x5e, 0xd7, 0xcb, 0x50, 0x90, 0x8c, 0x36,
	0x44, 0x4b, 0xb1, 0x3b, 0xaf, 0x05, 0x29, 0x8c, 0x04, 0x49, 0x49, 0x45, 0x64, 0x99, 0xcb, 0xb6,
	0xdd, 0xbc, 0xdb, 0xef, 0x00, 0x00, 0x00, 0xff, 0xff, 0x34, 0xa2, 0x86, 0xca, 0x90, 0x02, 0x00,
	0x00,
}
