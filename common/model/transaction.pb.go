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
	Version                       uint32 `protobuf:"varint,1,opt,name=Version,proto3" json:"Version,omitempty"`
	ID                            int64  `protobuf:"varint,2,opt,name=ID,proto3" json:"ID,omitempty"`
	BlockID                       int64  `protobuf:"varint,3,opt,name=BlockID,proto3" json:"BlockID,omitempty"`
	Height                        uint32 `protobuf:"varint,4,opt,name=Height,proto3" json:"Height,omitempty"`
	SenderAccountAddressLength    uint32 `protobuf:"varint,5,opt,name=SenderAccountAddressLength,proto3" json:"SenderAccountAddressLength,omitempty"`
	SenderAccountAddress          string `protobuf:"bytes,6,opt,name=SenderAccountAddress,proto3" json:"SenderAccountAddress,omitempty"`
	RecipientAccountAddressLength uint32 `protobuf:"varint,7,opt,name=RecipientAccountAddressLength,proto3" json:"RecipientAccountAddressLength,omitempty"`
	RecipientAccountAddress       string `protobuf:"bytes,8,opt,name=RecipientAccountAddress,proto3" json:"RecipientAccountAddress,omitempty"`
	TransactionType               uint32 `protobuf:"varint,9,opt,name=TransactionType,proto3" json:"TransactionType,omitempty"`
	Fee                           int64  `protobuf:"varint,10,opt,name=Fee,proto3" json:"Fee,omitempty"`
	Timestamp                     int64  `protobuf:"varint,11,opt,name=Timestamp,proto3" json:"Timestamp,omitempty"`
	TransactionHash               []byte `protobuf:"bytes,12,opt,name=TransactionHash,proto3" json:"TransactionHash,omitempty"`
	TransactionBodyLength         uint32 `protobuf:"varint,13,opt,name=TransactionBodyLength,proto3" json:"TransactionBodyLength,omitempty"`
	TransactionBodyBytes          []byte `protobuf:"bytes,14,opt,name=TransactionBodyBytes,proto3" json:"TransactionBodyBytes,omitempty"`
	// TransactionBody
	//
	// Types that are valid to be assigned to TransactionBody:
	//	*Transaction_EmptyTransactionBody
	//	*Transaction_SendMoneyTransactionBody
	//	*Transaction_NodeRegistrationTransactionBody
	//	*Transaction_UpdateNodeRegistrationTransactionBody
	TransactionBody      isTransaction_TransactionBody `protobuf_oneof:"TransactionBody"`
	Signature            []byte                        `protobuf:"bytes,19,opt,name=Signature,proto3" json:"Signature,omitempty"`
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

func (m *Transaction) GetSenderAccountAddressLength() uint32 {
	if m != nil {
		return m.SenderAccountAddressLength
	}
	return 0
}

func (m *Transaction) GetSenderAccountAddress() string {
	if m != nil {
		return m.SenderAccountAddress
	}
	return ""
}

func (m *Transaction) GetRecipientAccountAddressLength() uint32 {
	if m != nil {
		return m.RecipientAccountAddressLength
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

type Transaction_NodeRegistrationTransactionBody struct {
	NodeRegistrationTransactionBody *NodeRegistrationTransactionBody `protobuf:"bytes,17,opt,name=nodeRegistrationTransactionBody,proto3,oneof"`
}

type Transaction_UpdateNodeRegistrationTransactionBody struct {
	UpdateNodeRegistrationTransactionBody *UpdateNodeRegistrationTransactionBody `protobuf:"bytes,18,opt,name=updateNodeRegistrationTransactionBody,proto3,oneof"`
}

func (*Transaction_EmptyTransactionBody) isTransaction_TransactionBody() {}

func (*Transaction_SendMoneyTransactionBody) isTransaction_TransactionBody() {}

func (*Transaction_NodeRegistrationTransactionBody) isTransaction_TransactionBody() {}

func (*Transaction_UpdateNodeRegistrationTransactionBody) isTransaction_TransactionBody() {}

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

func (m *Transaction) GetNodeRegistrationTransactionBody() *NodeRegistrationTransactionBody {
	if x, ok := m.GetTransactionBody().(*Transaction_NodeRegistrationTransactionBody); ok {
		return x.NodeRegistrationTransactionBody
	}
	return nil
}

func (m *Transaction) GetUpdateNodeRegistrationTransactionBody() *UpdateNodeRegistrationTransactionBody {
	if x, ok := m.GetTransactionBody().(*Transaction_UpdateNodeRegistrationTransactionBody); ok {
		return x.UpdateNodeRegistrationTransactionBody
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
		(*Transaction_NodeRegistrationTransactionBody)(nil),
		(*Transaction_UpdateNodeRegistrationTransactionBody)(nil),
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

type NodeRegistrationTransactionBody struct {
	NodeID               int64             `protobuf:"varint,1,opt,name=NodeID,proto3" json:"NodeID,omitempty"`
	NodePublicKey        []byte            `protobuf:"bytes,2,opt,name=NodePublicKey,proto3" json:"NodePublicKey,omitempty"`
	AccountAddressLength uint32            `protobuf:"varint,3,opt,name=AccountAddressLength,proto3" json:"AccountAddressLength,omitempty"`
	AccountAddress       string            `protobuf:"bytes,4,opt,name=AccountAddress,proto3" json:"AccountAddress,omitempty"`
	RegistrationHeight   uint32            `protobuf:"varint,5,opt,name=RegistrationHeight,proto3" json:"RegistrationHeight,omitempty"`
	NodeAddressLength    uint32            `protobuf:"varint,6,opt,name=NodeAddressLength,proto3" json:"NodeAddressLength,omitempty"`
	NodeAddress          string            `protobuf:"bytes,7,opt,name=NodeAddress,proto3" json:"NodeAddress,omitempty"`
	LockedBalance        int64             `protobuf:"varint,8,opt,name=LockedBalance,proto3" json:"LockedBalance,omitempty"`
	Poown                *ProofOfOwnership `protobuf:"bytes,9,opt,name=Poown,proto3" json:"Poown,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *NodeRegistrationTransactionBody) Reset()         { *m = NodeRegistrationTransactionBody{} }
func (m *NodeRegistrationTransactionBody) String() string { return proto.CompactTextString(m) }
func (*NodeRegistrationTransactionBody) ProtoMessage()    {}
func (*NodeRegistrationTransactionBody) Descriptor() ([]byte, []int) {
	return fileDescriptor_8333001f09b34082, []int{3}
}

func (m *NodeRegistrationTransactionBody) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NodeRegistrationTransactionBody.Unmarshal(m, b)
}
func (m *NodeRegistrationTransactionBody) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NodeRegistrationTransactionBody.Marshal(b, m, deterministic)
}
func (m *NodeRegistrationTransactionBody) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NodeRegistrationTransactionBody.Merge(m, src)
}
func (m *NodeRegistrationTransactionBody) XXX_Size() int {
	return xxx_messageInfo_NodeRegistrationTransactionBody.Size(m)
}
func (m *NodeRegistrationTransactionBody) XXX_DiscardUnknown() {
	xxx_messageInfo_NodeRegistrationTransactionBody.DiscardUnknown(m)
}

var xxx_messageInfo_NodeRegistrationTransactionBody proto.InternalMessageInfo

func (m *NodeRegistrationTransactionBody) GetNodeID() int64 {
	if m != nil {
		return m.NodeID
	}
	return 0
}

func (m *NodeRegistrationTransactionBody) GetNodePublicKey() []byte {
	if m != nil {
		return m.NodePublicKey
	}
	return nil
}

func (m *NodeRegistrationTransactionBody) GetAccountAddressLength() uint32 {
	if m != nil {
		return m.AccountAddressLength
	}
	return 0
}

func (m *NodeRegistrationTransactionBody) GetAccountAddress() string {
	if m != nil {
		return m.AccountAddress
	}
	return ""
}

func (m *NodeRegistrationTransactionBody) GetRegistrationHeight() uint32 {
	if m != nil {
		return m.RegistrationHeight
	}
	return 0
}

func (m *NodeRegistrationTransactionBody) GetNodeAddressLength() uint32 {
	if m != nil {
		return m.NodeAddressLength
	}
	return 0
}

func (m *NodeRegistrationTransactionBody) GetNodeAddress() string {
	if m != nil {
		return m.NodeAddress
	}
	return ""
}

func (m *NodeRegistrationTransactionBody) GetLockedBalance() int64 {
	if m != nil {
		return m.LockedBalance
	}
	return 0
}

func (m *NodeRegistrationTransactionBody) GetPoown() *ProofOfOwnership {
	if m != nil {
		return m.Poown
	}
	return nil
}

type UpdateNodeRegistrationTransactionBody struct {
	NodeID               int64             `protobuf:"varint,1,opt,name=NodeID,proto3" json:"NodeID,omitempty"`
	NodePublicKey        []byte            `protobuf:"bytes,2,opt,name=NodePublicKey,proto3" json:"NodePublicKey,omitempty"`
	NodeAddressLength    uint32            `protobuf:"varint,3,opt,name=NodeAddressLength,proto3" json:"NodeAddressLength,omitempty"`
	NodeAddress          string            `protobuf:"bytes,4,opt,name=NodeAddress,proto3" json:"NodeAddress,omitempty"`
	LockedBalance        int64             `protobuf:"varint,5,opt,name=LockedBalance,proto3" json:"LockedBalance,omitempty"`
	Poown                *ProofOfOwnership `protobuf:"bytes,6,opt,name=Poown,proto3" json:"Poown,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *UpdateNodeRegistrationTransactionBody) Reset()         { *m = UpdateNodeRegistrationTransactionBody{} }
func (m *UpdateNodeRegistrationTransactionBody) String() string { return proto.CompactTextString(m) }
func (*UpdateNodeRegistrationTransactionBody) ProtoMessage()    {}
func (*UpdateNodeRegistrationTransactionBody) Descriptor() ([]byte, []int) {
	return fileDescriptor_8333001f09b34082, []int{4}
}

func (m *UpdateNodeRegistrationTransactionBody) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdateNodeRegistrationTransactionBody.Unmarshal(m, b)
}
func (m *UpdateNodeRegistrationTransactionBody) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdateNodeRegistrationTransactionBody.Marshal(b, m, deterministic)
}
func (m *UpdateNodeRegistrationTransactionBody) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateNodeRegistrationTransactionBody.Merge(m, src)
}
func (m *UpdateNodeRegistrationTransactionBody) XXX_Size() int {
	return xxx_messageInfo_UpdateNodeRegistrationTransactionBody.Size(m)
}
func (m *UpdateNodeRegistrationTransactionBody) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateNodeRegistrationTransactionBody.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateNodeRegistrationTransactionBody proto.InternalMessageInfo

func (m *UpdateNodeRegistrationTransactionBody) GetNodeID() int64 {
	if m != nil {
		return m.NodeID
	}
	return 0
}

func (m *UpdateNodeRegistrationTransactionBody) GetNodePublicKey() []byte {
	if m != nil {
		return m.NodePublicKey
	}
	return nil
}

func (m *UpdateNodeRegistrationTransactionBody) GetNodeAddressLength() uint32 {
	if m != nil {
		return m.NodeAddressLength
	}
	return 0
}

func (m *UpdateNodeRegistrationTransactionBody) GetNodeAddress() string {
	if m != nil {
		return m.NodeAddress
	}
	return ""
}

func (m *UpdateNodeRegistrationTransactionBody) GetLockedBalance() int64 {
	if m != nil {
		return m.LockedBalance
	}
	return 0
}

func (m *UpdateNodeRegistrationTransactionBody) GetPoown() *ProofOfOwnership {
	if m != nil {
		return m.Poown
	}
	return nil
}

//TODO: shall we move this to a different file?
type ProofOfOwnership struct {
	MessageBytes         []byte   `protobuf:"bytes,1,opt,name=MessageBytes,proto3" json:"MessageBytes,omitempty"`
	Signature            []byte   `protobuf:"bytes,2,opt,name=Signature,proto3" json:"Signature,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ProofOfOwnership) Reset()         { *m = ProofOfOwnership{} }
func (m *ProofOfOwnership) String() string { return proto.CompactTextString(m) }
func (*ProofOfOwnership) ProtoMessage()    {}
func (*ProofOfOwnership) Descriptor() ([]byte, []int) {
	return fileDescriptor_8333001f09b34082, []int{5}
}

func (m *ProofOfOwnership) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProofOfOwnership.Unmarshal(m, b)
}
func (m *ProofOfOwnership) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProofOfOwnership.Marshal(b, m, deterministic)
}
func (m *ProofOfOwnership) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProofOfOwnership.Merge(m, src)
}
func (m *ProofOfOwnership) XXX_Size() int {
	return xxx_messageInfo_ProofOfOwnership.Size(m)
}
func (m *ProofOfOwnership) XXX_DiscardUnknown() {
	xxx_messageInfo_ProofOfOwnership.DiscardUnknown(m)
}

var xxx_messageInfo_ProofOfOwnership proto.InternalMessageInfo

func (m *ProofOfOwnership) GetMessageBytes() []byte {
	if m != nil {
		return m.MessageBytes
	}
	return nil
}

func (m *ProofOfOwnership) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

//TODO: shall we move this to a different file?
type ProofOfOwnershipMessage struct {
	AccountType          uint32   `protobuf:"varint,1,opt,name=AccountType,proto3" json:"AccountType,omitempty"`
	AccountAddress       string   `protobuf:"bytes,2,opt,name=AccountAddress,proto3" json:"AccountAddress,omitempty"`
	BlockHash            []byte   `protobuf:"bytes,3,opt,name=BlockHash,proto3" json:"BlockHash,omitempty"`
	BlockHeight          uint32   `protobuf:"varint,4,opt,name=BlockHeight,proto3" json:"BlockHeight,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ProofOfOwnershipMessage) Reset()         { *m = ProofOfOwnershipMessage{} }
func (m *ProofOfOwnershipMessage) String() string { return proto.CompactTextString(m) }
func (*ProofOfOwnershipMessage) ProtoMessage()    {}
func (*ProofOfOwnershipMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_8333001f09b34082, []int{6}
}

func (m *ProofOfOwnershipMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProofOfOwnershipMessage.Unmarshal(m, b)
}
func (m *ProofOfOwnershipMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProofOfOwnershipMessage.Marshal(b, m, deterministic)
}
func (m *ProofOfOwnershipMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProofOfOwnershipMessage.Merge(m, src)
}
func (m *ProofOfOwnershipMessage) XXX_Size() int {
	return xxx_messageInfo_ProofOfOwnershipMessage.Size(m)
}
func (m *ProofOfOwnershipMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_ProofOfOwnershipMessage.DiscardUnknown(m)
}

var xxx_messageInfo_ProofOfOwnershipMessage proto.InternalMessageInfo

func (m *ProofOfOwnershipMessage) GetAccountType() uint32 {
	if m != nil {
		return m.AccountType
	}
	return 0
}

func (m *ProofOfOwnershipMessage) GetAccountAddress() string {
	if m != nil {
		return m.AccountAddress
	}
	return ""
}

func (m *ProofOfOwnershipMessage) GetBlockHash() []byte {
	if m != nil {
		return m.BlockHash
	}
	return nil
}

func (m *ProofOfOwnershipMessage) GetBlockHeight() uint32 {
	if m != nil {
		return m.BlockHeight
	}
	return 0
}

// GetTransactionRequest return model.Transaction
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
	return fileDescriptor_8333001f09b34082, []int{7}
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

// GetTransactions return GetTransactionsResponse
type GetTransactionsRequest struct {
	// Transactions set limit to be fetched
	Limit uint32 `protobuf:"varint,1,opt,name=Limit,proto3" json:"Limit,omitempty"`
	// Transactions set offset to be fetched
	Offset               uint64   `protobuf:"varint,2,opt,name=Offset,proto3" json:"Offset,omitempty"`
	AccountAddress       string   `protobuf:"bytes,3,opt,name=AccountAddress,proto3" json:"AccountAddress,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetTransactionsRequest) Reset()         { *m = GetTransactionsRequest{} }
func (m *GetTransactionsRequest) String() string { return proto.CompactTextString(m) }
func (*GetTransactionsRequest) ProtoMessage()    {}
func (*GetTransactionsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_8333001f09b34082, []int{8}
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

func (m *GetTransactionsRequest) GetAccountAddress() string {
	if m != nil {
		return m.AccountAddress
	}
	return ""
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
	return fileDescriptor_8333001f09b34082, []int{9}
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

// PostTransactionRequest return PostTransactionResponse
type PostTransactionRequest struct {
	// Signed transaction bytes
	TransactionBytes     []byte   `protobuf:"bytes,1,opt,name=TransactionBytes,proto3" json:"TransactionBytes,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PostTransactionRequest) Reset()         { *m = PostTransactionRequest{} }
func (m *PostTransactionRequest) String() string { return proto.CompactTextString(m) }
func (*PostTransactionRequest) ProtoMessage()    {}
func (*PostTransactionRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_8333001f09b34082, []int{10}
}

func (m *PostTransactionRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PostTransactionRequest.Unmarshal(m, b)
}
func (m *PostTransactionRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PostTransactionRequest.Marshal(b, m, deterministic)
}
func (m *PostTransactionRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PostTransactionRequest.Merge(m, src)
}
func (m *PostTransactionRequest) XXX_Size() int {
	return xxx_messageInfo_PostTransactionRequest.Size(m)
}
func (m *PostTransactionRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PostTransactionRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PostTransactionRequest proto.InternalMessageInfo

func (m *PostTransactionRequest) GetTransactionBytes() []byte {
	if m != nil {
		return m.TransactionBytes
	}
	return nil
}

type PostTransactionResponse struct {
	Transaction          *Transaction `protobuf:"bytes,1,opt,name=Transaction,proto3" json:"Transaction,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *PostTransactionResponse) Reset()         { *m = PostTransactionResponse{} }
func (m *PostTransactionResponse) String() string { return proto.CompactTextString(m) }
func (*PostTransactionResponse) ProtoMessage()    {}
func (*PostTransactionResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_8333001f09b34082, []int{11}
}

func (m *PostTransactionResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PostTransactionResponse.Unmarshal(m, b)
}
func (m *PostTransactionResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PostTransactionResponse.Marshal(b, m, deterministic)
}
func (m *PostTransactionResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PostTransactionResponse.Merge(m, src)
}
func (m *PostTransactionResponse) XXX_Size() int {
	return xxx_messageInfo_PostTransactionResponse.Size(m)
}
func (m *PostTransactionResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PostTransactionResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PostTransactionResponse proto.InternalMessageInfo

func (m *PostTransactionResponse) GetTransaction() *Transaction {
	if m != nil {
		return m.Transaction
	}
	return nil
}

// SendTransactionRequest request in p2pCommunication service
type SendTransactionRequest struct {
	TransactionBytes     []byte   `protobuf:"bytes,1,opt,name=TransactionBytes,proto3" json:"TransactionBytes,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SendTransactionRequest) Reset()         { *m = SendTransactionRequest{} }
func (m *SendTransactionRequest) String() string { return proto.CompactTextString(m) }
func (*SendTransactionRequest) ProtoMessage()    {}
func (*SendTransactionRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_8333001f09b34082, []int{12}
}

func (m *SendTransactionRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SendTransactionRequest.Unmarshal(m, b)
}
func (m *SendTransactionRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SendTransactionRequest.Marshal(b, m, deterministic)
}
func (m *SendTransactionRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SendTransactionRequest.Merge(m, src)
}
func (m *SendTransactionRequest) XXX_Size() int {
	return xxx_messageInfo_SendTransactionRequest.Size(m)
}
func (m *SendTransactionRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SendTransactionRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SendTransactionRequest proto.InternalMessageInfo

func (m *SendTransactionRequest) GetTransactionBytes() []byte {
	if m != nil {
		return m.TransactionBytes
	}
	return nil
}

func init() {
	proto.RegisterType((*Transaction)(nil), "model.Transaction")
	proto.RegisterType((*EmptyTransactionBody)(nil), "model.EmptyTransactionBody")
	proto.RegisterType((*SendMoneyTransactionBody)(nil), "model.SendMoneyTransactionBody")
	proto.RegisterType((*NodeRegistrationTransactionBody)(nil), "model.NodeRegistrationTransactionBody")
	proto.RegisterType((*UpdateNodeRegistrationTransactionBody)(nil), "model.UpdateNodeRegistrationTransactionBody")
	proto.RegisterType((*ProofOfOwnership)(nil), "model.ProofOfOwnership")
	proto.RegisterType((*ProofOfOwnershipMessage)(nil), "model.ProofOfOwnershipMessage")
	proto.RegisterType((*GetTransactionRequest)(nil), "model.GetTransactionRequest")
	proto.RegisterType((*GetTransactionsRequest)(nil), "model.GetTransactionsRequest")
	proto.RegisterType((*GetTransactionsResponse)(nil), "model.GetTransactionsResponse")
	proto.RegisterType((*PostTransactionRequest)(nil), "model.PostTransactionRequest")
	proto.RegisterType((*PostTransactionResponse)(nil), "model.PostTransactionResponse")
	proto.RegisterType((*SendTransactionRequest)(nil), "model.SendTransactionRequest")
}

func init() { proto.RegisterFile("model/transaction.proto", fileDescriptor_8333001f09b34082) }

var fileDescriptor_8333001f09b34082 = []byte{
	// 881 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x56, 0x4d, 0x6f, 0xdb, 0x46,
	0x13, 0x36, 0x45, 0x49, 0x89, 0x46, 0xb2, 0x63, 0xef, 0xeb, 0x48, 0x8b, 0xb7, 0x29, 0x2c, 0x10,
	0x4d, 0x2a, 0x04, 0x89, 0x0c, 0xa8, 0x41, 0xd1, 0x53, 0x01, 0xab, 0x6a, 0x2b, 0xa3, 0x4e, 0xed,
	0xae, 0xd5, 0x1e, 0x0a, 0xf4, 0x40, 0x93, 0x6b, 0x89, 0x88, 0xb8, 0xab, 0x72, 0x57, 0x08, 0x54,
	0xa0, 0xb7, 0x9e, 0x7a, 0xea, 0x7f, 0xe8, 0xa9, 0xff, 0xb2, 0xd8, 0x21, 0x55, 0x93, 0xd4, 0xca,
	0x16, 0x90, 0x8b, 0xa0, 0x79, 0xe6, 0xd9, 0xf9, 0xda, 0x99, 0x59, 0x42, 0x27, 0x96, 0x21, 0x9f,
	0x9f, 0xea, 0xc4, 0x17, 0xca, 0x0f, 0x74, 0x24, 0x45, 0x7f, 0x91, 0x48, 0x2d, 0x49, 0x0d, 0x15,
	0xde, 0x3f, 0x8f, 0xa1, 0x39, 0xb9, 0x53, 0x12, 0x0a, 0x8f, 0x7e, 0xe2, 0x89, 0x8a, 0xa4, 0xa0,
	0x4e, 0xd7, 0xe9, 0xed, 0xb3, 0xb5, 0x48, 0x0e, 0xa0, 0x72, 0x3e, 0xa2, 0x95, 0xae, 0xd3, 0x73,
	0x59, 0xe5, 0x7c, 0x64, 0x98, 0xc3, 0xb9, 0x0c, 0xde, 0x9d, 0x8f, 0xa8, 0x8b, 0xe0, 0x5a, 0x24,
	0x6d, 0xa8, 0x8f, 0x79, 0x34, 0x9d, 0x69, 0x5a, 0x45, 0x13, 0x99, 0x44, 0xbe, 0x84, 0xff, 0x5f,
	0x73, 0x11, 0xf2, 0xe4, 0x2c, 0x08, 0xe4, 0x52, 0xe8, 0xb3, 0x30, 0x4c, 0xb8, 0x52, 0x17, 0x5c,
	0x4c, 0xf5, 0x8c, 0xd6, 0x90, 0x7b, 0x0f, 0x83, 0x0c, 0xe0, 0xd8, 0xa6, 0xa5, 0xf5, 0xae, 0xd3,
	0x6b, 0x30, 0xab, 0x8e, 0x8c, 0xe0, 0x63, 0xc6, 0x83, 0x68, 0x11, 0x71, 0xa1, 0xad, 0x6e, 0x1f,
	0xa1, 0xdb, 0xfb, 0x49, 0xe4, 0x0b, 0xe8, 0x6c, 0x21, 0xd0, 0xc7, 0xe8, 0x7c, 0x9b, 0x9a, 0xf4,
	0xe0, 0x49, 0xae, 0xbc, 0x93, 0xd5, 0x82, 0xd3, 0x06, 0x7a, 0x2c, 0xc3, 0xe4, 0x10, 0xdc, 0x6f,
	0x38, 0xa7, 0x80, 0xb5, 0x34, 0x7f, 0xc9, 0x33, 0x68, 0x4c, 0xa2, 0x98, 0x2b, 0xed, 0xc7, 0x0b,
	0xda, 0x44, 0xfc, 0x0e, 0x28, 0x59, 0x1e, 0xfb, 0x6a, 0x46, 0x5b, 0x5d, 0xa7, 0xd7, 0x62, 0x65,
	0x98, 0xbc, 0x81, 0xa7, 0x39, 0x68, 0x28, 0xc3, 0x55, 0x96, 0xfb, 0x3e, 0x46, 0x62, 0x57, 0x9a,
	0x6a, 0x97, 0x14, 0xc3, 0x95, 0xe6, 0x8a, 0x1e, 0xa0, 0x13, 0xab, 0x8e, 0xfc, 0x00, 0xc7, 0x3c,
	0x5e, 0xe8, 0x55, 0x49, 0x49, 0x9f, 0x74, 0x9d, 0x5e, 0x73, 0xf0, 0x51, 0x1f, 0x7b, 0xae, 0xff,
	0xb5, 0x85, 0x32, 0xde, 0x63, 0xd6, 0xa3, 0xe4, 0x17, 0xa0, 0x8a, 0x8b, 0xf0, 0xad, 0x14, 0x7c,
	0xc3, 0xec, 0x21, 0x9a, 0x3d, 0xc9, 0xcc, 0x5e, 0x6f, 0xa1, 0x8d, 0xf7, 0xd8, 0x56, 0x13, 0x24,
	0x81, 0x13, 0x21, 0x43, 0xce, 0xf8, 0x34, 0x52, 0x3a, 0xf1, 0xf1, 0x36, 0x4a, 0x5e, 0x8e, 0xd0,
	0xcb, 0x8b, 0xcc, 0xcb, 0xf7, 0xf7, 0xb3, 0xc7, 0x7b, 0xec, 0x21, 0x83, 0xe4, 0x0f, 0x07, 0x9e,
	0x2f, 0x17, 0xa1, 0xaf, 0xf9, 0x03, 0xc6, 0x28, 0x41, 0xd7, 0xaf, 0x32, 0xd7, 0x3f, 0xee, 0x72,
	0x66, 0xbc, 0xc7, 0x76, 0x33, 0x6e, 0xda, 0xeb, 0x3a, 0x9a, 0x0a, 0x5f, 0x2f, 0x13, 0x4e, 0xff,
	0x87, 0xb7, 0x7a, 0x07, 0x0c, 0x8f, 0x0a, 0xed, 0x65, 0x0e, 0x78, 0x6d, 0x38, 0xb6, 0x5d, 0x9d,
	0x37, 0x00, 0xba, 0xad, 0xf6, 0x66, 0x17, 0x9c, 0xc5, 0x66, 0x20, 0x70, 0x9d, 0xb8, 0x2c, 0x93,
	0xbc, 0xbf, 0x5c, 0x38, 0x79, 0x28, 0xc0, 0x36, 0xd4, 0x0d, 0xe5, 0x7c, 0xb4, 0x3e, 0x9b, 0x4a,
	0xe4, 0x13, 0xd8, 0x37, 0xff, 0xae, 0x96, 0x37, 0xf3, 0x28, 0xf8, 0x8e, 0xaf, 0x70, 0x29, 0xb5,
	0x58, 0x11, 0x34, 0xfd, 0x6b, 0x1d, 0x78, 0x17, 0x9b, 0xde, 0xaa, 0x23, 0x2f, 0xe0, 0xa0, 0x34,
	0xde, 0x55, 0x1c, 0xef, 0x12, 0x4a, 0xfa, 0x40, 0xf2, 0x81, 0x67, 0xdb, 0x2e, 0xdd, 0x60, 0x16,
	0x0d, 0x79, 0x05, 0x47, 0x26, 0xb8, 0x62, 0x20, 0x75, 0xa4, 0x6f, 0x2a, 0x48, 0x17, 0x9a, 0x39,
	0x10, 0x37, 0x54, 0x83, 0xe5, 0x21, 0x53, 0x81, 0x0b, 0x19, 0xbc, 0xe3, 0xe1, 0xd0, 0x9f, 0xfb,
	0x22, 0xe0, 0xb8, 0x85, 0x5c, 0x56, 0x04, 0xc9, 0x6b, 0xa8, 0x5d, 0x49, 0xf9, 0x5e, 0xe0, 0xc6,
	0x69, 0x0e, 0x3a, 0x59, 0x1b, 0x5d, 0x25, 0x52, 0xde, 0x5e, 0xde, 0x5e, 0xbe, 0x17, 0x3c, 0x51,
	0xb3, 0x68, 0xc1, 0x52, 0x96, 0xf7, 0x67, 0x05, 0x9e, 0xef, 0xd4, 0x62, 0x1f, 0x78, 0x31, 0xd6,
	0x62, 0xb8, 0x3b, 0x16, 0xa3, 0xba, 0x43, 0x31, 0x6a, 0xf7, 0x16, 0xa3, 0xbe, 0x53, 0x31, 0x26,
	0x70, 0x58, 0x56, 0x11, 0x0f, 0x5a, 0x6f, 0xb9, 0x52, 0xfe, 0x94, 0xa7, 0x9b, 0xd0, 0xc1, 0xec,
	0x0a, 0x58, 0x71, 0xa8, 0x2a, 0xa5, 0xa1, 0xf2, 0xfe, 0x76, 0xa0, 0x53, 0x36, 0x9b, 0x1d, 0x37,
	0x89, 0x66, 0x5d, 0x86, 0xaf, 0x44, 0xfa, 0xfa, 0xe6, 0x21, 0x4b, 0x77, 0x56, 0xac, 0xdd, 0xf9,
	0x0c, 0x1a, 0xf8, 0x14, 0xe3, 0x9b, 0xe0, 0xa6, 0x31, 0xfc, 0x07, 0x18, 0x3f, 0xa9, 0x90, 0x7f,
	0xa2, 0xf3, 0x90, 0xf7, 0x29, 0x3c, 0xfd, 0x96, 0xeb, 0xdc, 0xa5, 0x33, 0xfe, 0xeb, 0x92, 0x2b,
	0x9d, 0x7d, 0x02, 0x38, 0xeb, 0x4f, 0x00, 0x4f, 0x40, 0xbb, 0x48, 0x54, 0x6b, 0xe6, 0x31, 0xd4,
	0x2e, 0xa2, 0x38, 0xd2, 0x59, 0x1a, 0xa9, 0x60, 0xfa, 0xe6, 0xf2, 0xf6, 0x56, 0x71, 0x8d, 0x81,
	0x57, 0x59, 0x26, 0x59, 0x12, 0x73, 0x6d, 0x89, 0x79, 0xbf, 0x43, 0x67, 0xc3, 0x9f, 0x5a, 0x48,
	0xa1, 0xb8, 0x71, 0x38, 0x91, 0xda, 0x9f, 0xa3, 0xc3, 0x2a, 0x4b, 0x05, 0x83, 0x7e, 0x85, 0xcb,
	0xa7, 0x92, 0x86, 0x81, 0x02, 0xf9, 0x1c, 0x5a, 0x79, 0x1b, 0xd4, 0xed, 0xba, 0xbd, 0xe6, 0x80,
	0x64, 0x1d, 0x91, 0xcf, 0xbb, 0xc0, 0xf3, 0x46, 0xd0, 0xbe, 0x92, 0xca, 0x56, 0x98, 0x97, 0x70,
	0x98, 0x9f, 0x91, 0x5c, 0x77, 0x6c, 0xe0, 0xde, 0x25, 0x74, 0x36, 0xac, 0x64, 0x49, 0xbc, 0x29,
	0x7c, 0x8b, 0xa1, 0x05, 0x7b, 0x5c, 0x79, 0x9a, 0x09, 0xcb, 0xac, 0xdf, 0x0f, 0x0b, 0x6b, 0xf8,
	0xf2, 0xe7, 0xde, 0x34, 0xd2, 0xb3, 0xe5, 0x4d, 0x3f, 0x90, 0xf1, 0xe9, 0x6f, 0x52, 0xde, 0x04,
	0xe9, 0xef, 0xeb, 0x40, 0x26, 0xfc, 0x34, 0x90, 0x71, 0x2c, 0xc5, 0x29, 0x86, 0x72, 0x53, 0xc7,
	0x4f, 0xc8, 0xcf, 0xfe, 0x0d, 0x00, 0x00, 0xff, 0xff, 0x5c, 0x9b, 0x41, 0xda, 0x5d, 0x0a, 0x00,
	0x00,
}
