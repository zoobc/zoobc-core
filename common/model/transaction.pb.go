// Code generated by protoc-gen-go. DO NOT EDIT.
// source: model/transaction.proto

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

// Transaction represent the transaction data structure stored in the database
type Transaction struct {
	Version                 uint32 `protobuf:"varint,1,opt,name=Version,proto3" json:"Version,omitempty"`
	ID                      int64  `protobuf:"varint,2,opt,name=ID,proto3" json:"ID,omitempty"`
	BlockID                 int64  `protobuf:"varint,3,opt,name=BlockID,proto3" json:"BlockID,omitempty"`
	Height                  uint32 `protobuf:"varint,4,opt,name=Height,proto3" json:"Height,omitempty"`
	SenderAccountType       uint32 `protobuf:"varint,5,opt,name=SenderAccountType,proto3" json:"SenderAccountType,omitempty"`
	SenderAccountAddress    string `protobuf:"bytes,6,opt,name=SenderAccountAddress,proto3" json:"SenderAccountAddress,omitempty"`
	RecipientAccountType    uint32 `protobuf:"varint,7,opt,name=RecipientAccountType,proto3" json:"RecipientAccountType,omitempty"`
	RecipientAccountAddress string `protobuf:"bytes,8,opt,name=RecipientAccountAddress,proto3" json:"RecipientAccountAddress,omitempty"`
	TransactionType         uint32 `protobuf:"varint,9,opt,name=TransactionType,proto3" json:"TransactionType,omitempty"`
	Fee                     int64  `protobuf:"varint,10,opt,name=Fee,proto3" json:"Fee,omitempty"`
	Timestamp               int64  `protobuf:"varint,11,opt,name=Timestamp,proto3" json:"Timestamp,omitempty"`
	TransactionHash         []byte `protobuf:"bytes,12,opt,name=TransactionHash,proto3" json:"TransactionHash,omitempty"`
	TransactionBodyLength   uint32 `protobuf:"varint,13,opt,name=TransactionBodyLength,proto3" json:"TransactionBodyLength,omitempty"`
	TransactionBodyBytes    []byte `protobuf:"bytes,14,opt,name=TransactionBodyBytes,proto3" json:"TransactionBodyBytes,omitempty"`
	// TransactionBody
	//
	// Types that are valid to be assigned to TransactionBody:
	//	*Transaction_EmptyTransactionBody
	//	*Transaction_SendMoneyTransactionBody
	TransactionBody      isTransaction_TransactionBody `protobuf_oneof:"TransactionBody"`
	Signature            []byte                        `protobuf:"bytes,17,opt,name=Signature,proto3" json:"Signature,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                      `json:"-"`
	XXX_unrecognized     []byte                        `json:"-"`
	XXX_sizecache        int32                         `json:"-"`
}

func (m *Transaction) Reset()         { *m = Transaction{} }
func (m *Transaction) String() string { return proto.CompactTextString(m) }
func (*Transaction) ProtoMessage()    {}
func (*Transaction) Descriptor() ([]byte, []int) {
	return fileDescriptor_8333001f09b34082, []int{0}
}

func (m *Transaction) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Transaction.Unmarshal(m, b)
}
func (m *Transaction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Transaction.Marshal(b, m, deterministic)
}
func (m *Transaction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Transaction.Merge(m, src)
}
func (m *Transaction) XXX_Size() int {
	return xxx_messageInfo_Transaction.Size(m)
}
func (m *Transaction) XXX_DiscardUnknown() {
	xxx_messageInfo_Transaction.DiscardUnknown(m)
}

var xxx_messageInfo_Transaction proto.InternalMessageInfo

func (m *Transaction) GetVersion() uint32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *Transaction) GetID() int64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *Transaction) GetBlockID() int64 {
	if m != nil {
		return m.BlockID
	}
	return 0
}

func (m *Transaction) GetHeight() uint32 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *Transaction) GetSenderAccountType() uint32 {
	if m != nil {
		return m.SenderAccountType
	}
	return 0
}

func (m *Transaction) GetSenderAccountAddress() string {
	if m != nil {
		return m.SenderAccountAddress
	}
	return ""
}

func (m *Transaction) GetRecipientAccountType() uint32 {
	if m != nil {
		return m.RecipientAccountType
	}
	return 0
}

func (m *Transaction) GetRecipientAccountAddress() string {
	if m != nil {
		return m.RecipientAccountAddress
	}
	return ""
}

func (m *Transaction) GetTransactionType() uint32 {
	if m != nil {
		return m.TransactionType
	}
	return 0
}

func (m *Transaction) GetFee() int64 {
	if m != nil {
		return m.Fee
	}
	return 0
}

func (m *Transaction) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *Transaction) GetTransactionHash() []byte {
	if m != nil {
		return m.TransactionHash
	}
	return nil
}

func (m *Transaction) GetTransactionBodyLength() uint32 {
	if m != nil {
		return m.TransactionBodyLength
	}
	return 0
}

func (m *Transaction) GetTransactionBodyBytes() []byte {
	if m != nil {
		return m.TransactionBodyBytes
	}
	return nil
}

type isTransaction_TransactionBody interface {
	isTransaction_TransactionBody()
}

type Transaction_EmptyTransactionBody struct {
	EmptyTransactionBody *EmptyTransactionBody `protobuf:"bytes,15,opt,name=emptyTransactionBody,proto3,oneof"`
}

type Transaction_SendMoneyTransactionBody struct {
	SendMoneyTransactionBody *SendMoneyTransactionBody `protobuf:"bytes,16,opt,name=sendMoneyTransactionBody,proto3,oneof"`
}

func (*Transaction_EmptyTransactionBody) isTransaction_TransactionBody() {}

func (*Transaction_SendMoneyTransactionBody) isTransaction_TransactionBody() {}

func (m *Transaction) GetTransactionBody() isTransaction_TransactionBody {
	if m != nil {
		return m.TransactionBody
	}
	return nil
}

func (m *Transaction) GetEmptyTransactionBody() *EmptyTransactionBody {
	if x, ok := m.GetTransactionBody().(*Transaction_EmptyTransactionBody); ok {
		return x.EmptyTransactionBody
	}
	return nil
}

func (m *Transaction) GetSendMoneyTransactionBody() *SendMoneyTransactionBody {
	if x, ok := m.GetTransactionBody().(*Transaction_SendMoneyTransactionBody); ok {
		return x.SendMoneyTransactionBody
	}
	return nil
}

func (m *Transaction) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Transaction) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Transaction_EmptyTransactionBody)(nil),
		(*Transaction_SendMoneyTransactionBody)(nil),
	}
}

type EmptyTransactionBody struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *EmptyTransactionBody) Reset()         { *m = EmptyTransactionBody{} }
func (m *EmptyTransactionBody) String() string { return proto.CompactTextString(m) }
func (*EmptyTransactionBody) ProtoMessage()    {}
func (*EmptyTransactionBody) Descriptor() ([]byte, []int) {
	return fileDescriptor_8333001f09b34082, []int{1}
}

func (m *EmptyTransactionBody) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EmptyTransactionBody.Unmarshal(m, b)
}
func (m *EmptyTransactionBody) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EmptyTransactionBody.Marshal(b, m, deterministic)
}
func (m *EmptyTransactionBody) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EmptyTransactionBody.Merge(m, src)
}
func (m *EmptyTransactionBody) XXX_Size() int {
	return xxx_messageInfo_EmptyTransactionBody.Size(m)
}
func (m *EmptyTransactionBody) XXX_DiscardUnknown() {
	xxx_messageInfo_EmptyTransactionBody.DiscardUnknown(m)
}

var xxx_messageInfo_EmptyTransactionBody proto.InternalMessageInfo

type SendMoneyTransactionBody struct {
	Amount               int64    `protobuf:"varint,1,opt,name=Amount,proto3" json:"Amount,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SendMoneyTransactionBody) Reset()         { *m = SendMoneyTransactionBody{} }
func (m *SendMoneyTransactionBody) String() string { return proto.CompactTextString(m) }
func (*SendMoneyTransactionBody) ProtoMessage()    {}
func (*SendMoneyTransactionBody) Descriptor() ([]byte, []int) {
	return fileDescriptor_8333001f09b34082, []int{2}
}

func (m *SendMoneyTransactionBody) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SendMoneyTransactionBody.Unmarshal(m, b)
}
func (m *SendMoneyTransactionBody) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SendMoneyTransactionBody.Marshal(b, m, deterministic)
}
func (m *SendMoneyTransactionBody) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SendMoneyTransactionBody.Merge(m, src)
}
func (m *SendMoneyTransactionBody) XXX_Size() int {
	return xxx_messageInfo_SendMoneyTransactionBody.Size(m)
}
func (m *SendMoneyTransactionBody) XXX_DiscardUnknown() {
	xxx_messageInfo_SendMoneyTransactionBody.DiscardUnknown(m)
}

var xxx_messageInfo_SendMoneyTransactionBody proto.InternalMessageInfo

func (m *SendMoneyTransactionBody) GetAmount() int64 {
	if m != nil {
		return m.Amount
	}
	return 0
}

type GetTransactionRequest struct {
	// Fetch Transaction by its ID
	ID                   int64    `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetTransactionRequest) Reset()         { *m = GetTransactionRequest{} }
func (m *GetTransactionRequest) String() string { return proto.CompactTextString(m) }
func (*GetTransactionRequest) ProtoMessage()    {}
func (*GetTransactionRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_8333001f09b34082, []int{3}
}

func (m *GetTransactionRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetTransactionRequest.Unmarshal(m, b)
}
func (m *GetTransactionRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetTransactionRequest.Marshal(b, m, deterministic)
}
func (m *GetTransactionRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetTransactionRequest.Merge(m, src)
}
func (m *GetTransactionRequest) XXX_Size() int {
	return xxx_messageInfo_GetTransactionRequest.Size(m)
}
func (m *GetTransactionRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetTransactionRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetTransactionRequest proto.InternalMessageInfo

func (m *GetTransactionRequest) GetID() int64 {
	if m != nil {
		return m.ID
	}
	return 0
}

type GetTransactionsRequest struct {
	// Transactions set limit to be fetched
	Limit uint32 `protobuf:"varint,1,opt,name=Limit,proto3" json:"Limit,omitempty"`
	// Transactions set offset to be fetched
	Offset               uint64   `protobuf:"varint,2,opt,name=Offset,proto3" json:"Offset,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetTransactionsRequest) Reset()         { *m = GetTransactionsRequest{} }
func (m *GetTransactionsRequest) String() string { return proto.CompactTextString(m) }
func (*GetTransactionsRequest) ProtoMessage()    {}
func (*GetTransactionsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_8333001f09b34082, []int{4}
}

func (m *GetTransactionsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetTransactionsRequest.Unmarshal(m, b)
}
func (m *GetTransactionsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetTransactionsRequest.Marshal(b, m, deterministic)
}
func (m *GetTransactionsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetTransactionsRequest.Merge(m, src)
}
func (m *GetTransactionsRequest) XXX_Size() int {
	return xxx_messageInfo_GetTransactionsRequest.Size(m)
}
func (m *GetTransactionsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetTransactionsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetTransactionsRequest proto.InternalMessageInfo

func (m *GetTransactionsRequest) GetLimit() uint32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

func (m *GetTransactionsRequest) GetOffset() uint64 {
	if m != nil {
		return m.Offset
	}
	return 0
}

type GetTransactionsResponse struct {
	// Number of transactions in total
	Total uint64 `protobuf:"varint,1,opt,name=Total,proto3" json:"Total,omitempty"`
	// Number of transactions returned
	Count uint32 `protobuf:"varint,2,opt,name=Count,proto3" json:"Count,omitempty"`
	// Transaction transactions returned
	Transactions         []*Transaction `protobuf:"bytes,3,rep,name=Transactions,proto3" json:"Transactions,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *GetTransactionsResponse) Reset()         { *m = GetTransactionsResponse{} }
func (m *GetTransactionsResponse) String() string { return proto.CompactTextString(m) }
func (*GetTransactionsResponse) ProtoMessage()    {}
func (*GetTransactionsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_8333001f09b34082, []int{5}
}

func (m *GetTransactionsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetTransactionsResponse.Unmarshal(m, b)
}
func (m *GetTransactionsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetTransactionsResponse.Marshal(b, m, deterministic)
}
func (m *GetTransactionsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetTransactionsResponse.Merge(m, src)
}
func (m *GetTransactionsResponse) XXX_Size() int {
	return xxx_messageInfo_GetTransactionsResponse.Size(m)
}
func (m *GetTransactionsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetTransactionsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetTransactionsResponse proto.InternalMessageInfo

func (m *GetTransactionsResponse) GetTotal() uint64 {
	if m != nil {
		return m.Total
	}
	return 0
}

func (m *GetTransactionsResponse) GetCount() uint32 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *GetTransactionsResponse) GetTransactions() []*Transaction {
	if m != nil {
		return m.Transactions
	}
	return nil
}

type PostUnconfirmedTransactionRequest struct {
	Transaction          *Transaction `protobuf:"bytes,1,opt,name=Transaction,proto3" json:"Transaction,omitempty"`
	ArrivalTimestamp     uint32       `protobuf:"varint,2,opt,name=ArrivalTimestamp,proto3" json:"ArrivalTimestamp,omitempty"`
	FeePerByte           int64        `protobuf:"varint,3,opt,name=FeePerByte,proto3" json:"FeePerByte,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *PostUnconfirmedTransactionRequest) Reset()         { *m = PostUnconfirmedTransactionRequest{} }
func (m *PostUnconfirmedTransactionRequest) String() string { return proto.CompactTextString(m) }
func (*PostUnconfirmedTransactionRequest) ProtoMessage()    {}
func (*PostUnconfirmedTransactionRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_8333001f09b34082, []int{6}
}

func (m *PostUnconfirmedTransactionRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PostUnconfirmedTransactionRequest.Unmarshal(m, b)
}
func (m *PostUnconfirmedTransactionRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PostUnconfirmedTransactionRequest.Marshal(b, m, deterministic)
}
func (m *PostUnconfirmedTransactionRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PostUnconfirmedTransactionRequest.Merge(m, src)
}
func (m *PostUnconfirmedTransactionRequest) XXX_Size() int {
	return xxx_messageInfo_PostUnconfirmedTransactionRequest.Size(m)
}
func (m *PostUnconfirmedTransactionRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PostUnconfirmedTransactionRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PostUnconfirmedTransactionRequest proto.InternalMessageInfo

func (m *PostUnconfirmedTransactionRequest) GetTransaction() *Transaction {
	if m != nil {
		return m.Transaction
	}
	return nil
}

func (m *PostUnconfirmedTransactionRequest) GetArrivalTimestamp() uint32 {
	if m != nil {
		return m.ArrivalTimestamp
	}
	return 0
}

func (m *PostUnconfirmedTransactionRequest) GetFeePerByte() int64 {
	if m != nil {
		return m.FeePerByte
	}
	return 0
}

func init() {
	proto.RegisterType((*Transaction)(nil), "model.Transaction")
	proto.RegisterType((*EmptyTransactionBody)(nil), "model.EmptyTransactionBody")
	proto.RegisterType((*SendMoneyTransactionBody)(nil), "model.SendMoneyTransactionBody")
	proto.RegisterType((*GetTransactionRequest)(nil), "model.GetTransactionRequest")
	proto.RegisterType((*GetTransactionsRequest)(nil), "model.GetTransactionsRequest")
	proto.RegisterType((*GetTransactionsResponse)(nil), "model.GetTransactionsResponse")
	proto.RegisterType((*PostUnconfirmedTransactionRequest)(nil), "model.PostUnconfirmedTransactionRequest")
}

func init() { proto.RegisterFile("model/transaction.proto", fileDescriptor_8333001f09b34082) }

var fileDescriptor_8333001f09b34082 = []byte{
	// 585 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x54, 0x4d, 0x6f, 0xd3, 0x40,
	0x10, 0xad, 0xe3, 0x36, 0xa5, 0x93, 0x7e, 0xae, 0xd2, 0x76, 0x25, 0x10, 0x18, 0x5f, 0xb0, 0x2a,
	0x48, 0xa4, 0x50, 0x21, 0xae, 0x09, 0xa5, 0xa4, 0x52, 0x11, 0xc5, 0x09, 0x1c, 0x90, 0x38, 0x38,
	0xf6, 0x24, 0x59, 0x11, 0xef, 0x1a, 0xef, 0x06, 0x29, 0x48, 0xfc, 0x01, 0xfe, 0x07, 0xff, 0x13,
	0xed, 0xc6, 0x51, 0x6c, 0xc7, 0xb9, 0x44, 0x99, 0xf7, 0x66, 0xdf, 0x9b, 0x1d, 0xef, 0x0c, 0x5c,
	0xc6, 0x22, 0xc2, 0x59, 0x5b, 0xa5, 0x01, 0x97, 0x41, 0xa8, 0x98, 0xe0, 0xad, 0x24, 0x15, 0x4a,
	0x90, 0x3d, 0x43, 0xb8, 0x7f, 0xeb, 0xd0, 0x18, 0xae, 0x49, 0x42, 0x61, 0xff, 0x2b, 0xa6, 0x92,
	0x09, 0x4e, 0x2d, 0xc7, 0xf2, 0x8e, 0xfc, 0x55, 0x48, 0x8e, 0xa1, 0x76, 0x77, 0x43, 0x6b, 0x8e,
	0xe5, 0xd9, 0x7e, 0xed, 0xee, 0x46, 0x67, 0xf6, 0x66, 0x22, 0xfc, 0x71, 0x77, 0x43, 0x6d, 0x03,
	0xae, 0x42, 0x72, 0x01, 0xf5, 0x3e, 0xb2, 0xc9, 0x54, 0xd1, 0x5d, 0x23, 0x91, 0x45, 0xe4, 0x25,
	0x9c, 0x0d, 0x90, 0x47, 0x98, 0x76, 0xc3, 0x50, 0xcc, 0xb9, 0x1a, 0x2e, 0x12, 0xa4, 0x7b, 0x26,
	0x65, 0x93, 0x20, 0x1d, 0x68, 0x16, 0xc0, 0x6e, 0x14, 0xa5, 0x28, 0x25, 0xad, 0x3b, 0x96, 0x77,
	0xe0, 0x57, 0x72, 0xfa, 0x8c, 0x8f, 0x21, 0x4b, 0x18, 0x72, 0x95, 0x37, 0xd9, 0x37, 0x26, 0x95,
	0x1c, 0x79, 0x0b, 0x97, 0x65, 0x7c, 0x65, 0xf5, 0xc8, 0x58, 0x6d, 0xa3, 0x89, 0x07, 0x27, 0xb9,
	0xd6, 0x19, 0xa3, 0x03, 0x63, 0x54, 0x86, 0xc9, 0x29, 0xd8, 0xb7, 0x88, 0x14, 0x4c, 0x9f, 0xf4,
	0x5f, 0xf2, 0x04, 0x0e, 0x86, 0x2c, 0x46, 0xa9, 0x82, 0x38, 0xa1, 0x0d, 0x83, 0xaf, 0x81, 0x92,
	0x72, 0x3f, 0x90, 0x53, 0x7a, 0xe8, 0x58, 0xde, 0xa1, 0x5f, 0x86, 0xc9, 0x35, 0x9c, 0xe7, 0xa0,
	0x9e, 0x88, 0x16, 0xf7, 0xc8, 0x27, 0x6a, 0x4a, 0x8f, 0x4c, 0x25, 0xd5, 0xa4, 0xee, 0x53, 0x89,
	0xe8, 0x2d, 0x14, 0x4a, 0x7a, 0x6c, 0x4c, 0x2a, 0x39, 0xf2, 0x19, 0x9a, 0x18, 0x27, 0x6a, 0x51,
	0x22, 0xe9, 0x89, 0x63, 0x79, 0x8d, 0xce, 0xe3, 0x96, 0x79, 0x4f, 0xad, 0xf7, 0x15, 0x29, 0xfd,
	0x1d, 0xbf, 0xf2, 0x28, 0xf9, 0x0e, 0x54, 0x22, 0x8f, 0x3e, 0x0a, 0x8e, 0x1b, 0xb2, 0xa7, 0x46,
	0xf6, 0x59, 0x26, 0x3b, 0xd8, 0x92, 0xd6, 0xdf, 0xf1, 0xb7, 0x4a, 0xe8, 0x1e, 0x0f, 0xd8, 0x84,
	0x07, 0x6a, 0x9e, 0x22, 0x3d, 0x33, 0x57, 0x5b, 0x03, 0xbd, 0xb3, 0x42, 0x8f, 0xf5, 0x01, 0xf7,
	0x02, 0x9a, 0x55, 0xf5, 0xbb, 0x1d, 0xa0, 0xdb, 0x0a, 0xd0, 0x8f, 0xbd, 0x1b, 0xeb, 0x57, 0x61,
	0xe6, 0xc5, 0xf6, 0xb3, 0xc8, 0x7d, 0x01, 0xe7, 0x1f, 0x50, 0xe5, 0xb2, 0x7d, 0xfc, 0x39, 0x47,
	0xa9, 0xb2, 0x39, 0xb2, 0x56, 0x73, 0xe4, 0xde, 0xc2, 0x45, 0x31, 0x51, 0xae, 0x32, 0x9b, 0xb0,
	0x77, 0xcf, 0x62, 0xa6, 0xb2, 0x49, 0x5c, 0x06, 0xda, 0xf0, 0xd3, 0x78, 0x2c, 0x51, 0x99, 0x59,
	0xdc, 0xf5, 0xb3, 0xc8, 0xfd, 0x03, 0x97, 0x1b, 0x3a, 0x32, 0x11, 0x5c, 0xa2, 0x16, 0x1a, 0x0a,
	0x15, 0xcc, 0x8c, 0xd0, 0xae, 0xbf, 0x0c, 0x34, 0xfa, 0xce, 0x14, 0x5e, 0x5b, 0xca, 0x9b, 0x80,
	0xbc, 0x81, 0xc3, 0xbc, 0x06, 0xb5, 0x1d, 0xdb, 0x6b, 0x74, 0x48, 0xf6, 0x1d, 0xf2, 0xf7, 0x29,
	0xe4, 0xb9, 0xff, 0x2c, 0x78, 0xfe, 0x20, 0xa4, 0xfa, 0xc2, 0x43, 0xc1, 0xc7, 0x2c, 0x8d, 0x31,
	0xaa, 0xb8, 0xfc, 0x75, 0x61, 0xdb, 0x98, 0x7a, 0xaa, 0xc5, 0x0b, 0x4b, 0xe9, 0x0a, 0x4e, 0xbb,
	0x69, 0xca, 0x7e, 0x05, 0xb3, 0xf5, 0xcc, 0x2c, 0x8b, 0xde, 0xc0, 0xc9, 0x53, 0x80, 0x5b, 0xc4,
	0x07, 0x4c, 0xf5, 0xab, 0xcd, 0x36, 0x53, 0x0e, 0xe9, 0x5d, 0x7d, 0xf3, 0x26, 0x4c, 0x4d, 0xe7,
	0xa3, 0x56, 0x28, 0xe2, 0xf6, 0x6f, 0x21, 0x46, 0xe1, 0xf2, 0xf7, 0x55, 0x28, 0x52, 0x6c, 0x87,
	0x22, 0x8e, 0x05, 0x6f, 0x9b, 0x82, 0x46, 0x75, 0xb3, 0x2a, 0x5f, 0xff, 0x0f, 0x00, 0x00, 0xff,
	0xff, 0xf1, 0x9f, 0x84, 0x8b, 0x45, 0x05, 0x00, 0x00,
}
