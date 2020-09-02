// Code generated by protoc-gen-go. DO NOT EDIT.
// source: model/multiSignature.proto

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

type PendingTransactionStatus int32

const (
	PendingTransactionStatus_PendingTransactionPending  PendingTransactionStatus = 0
	PendingTransactionStatus_PendingTransactionExecuted PendingTransactionStatus = 1
	PendingTransactionStatus_PendingTransactionNoOp     PendingTransactionStatus = 2
	PendingTransactionStatus_PendingTransactionExpired  PendingTransactionStatus = 3
)

var PendingTransactionStatus_name = map[int32]string{
	0: "PendingTransactionPending",
	1: "PendingTransactionExecuted",
	2: "PendingTransactionNoOp",
	3: "PendingTransactionExpired",
}

var PendingTransactionStatus_value = map[string]int32{
	"PendingTransactionPending":  0,
	"PendingTransactionExecuted": 1,
	"PendingTransactionNoOp":     2,
	"PendingTransactionExpired":  3,
}

func (x PendingTransactionStatus) String() string {
	return proto.EnumName(PendingTransactionStatus_name, int32(x))
}

func (PendingTransactionStatus) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_136af44c597c17ae, []int{0}
}

type MultiSignatureInfo struct {
	MinimumSignatures    uint32   `protobuf:"varint,1,opt,name=MinimumSignatures,proto3" json:"MinimumSignatures,omitempty"`
	Nonce                int64    `protobuf:"varint,2,opt,name=Nonce,proto3" json:"Nonce,omitempty"`
	MultisigAddress      string   `protobuf:"bytes,3,opt,name=MultisigAddress,proto3" json:"MultisigAddress,omitempty"`
	BlockHeight          uint32   `protobuf:"varint,4,opt,name=BlockHeight,proto3" json:"BlockHeight,omitempty"`
	Latest               bool     `protobuf:"varint,5,opt,name=Latest,proto3" json:"Latest,omitempty"`
	Addresses            []string `protobuf:"bytes,6,rep,name=Addresses,proto3" json:"Addresses,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MultiSignatureInfo) Reset()         { *m = MultiSignatureInfo{} }
func (m *MultiSignatureInfo) String() string { return proto.CompactTextString(m) }
func (*MultiSignatureInfo) ProtoMessage()    {}
func (*MultiSignatureInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_136af44c597c17ae, []int{0}
}

func (m *MultiSignatureInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MultiSignatureInfo.Unmarshal(m, b)
}
func (m *MultiSignatureInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MultiSignatureInfo.Marshal(b, m, deterministic)
}
func (m *MultiSignatureInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MultiSignatureInfo.Merge(m, src)
}
func (m *MultiSignatureInfo) XXX_Size() int {
	return xxx_messageInfo_MultiSignatureInfo.Size(m)
}
func (m *MultiSignatureInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_MultiSignatureInfo.DiscardUnknown(m)
}

var xxx_messageInfo_MultiSignatureInfo proto.InternalMessageInfo

func (m *MultiSignatureInfo) GetMinimumSignatures() uint32 {
	if m != nil {
		return m.MinimumSignatures
	}
	return 0
}

func (m *MultiSignatureInfo) GetNonce() int64 {
	if m != nil {
		return m.Nonce
	}
	return 0
}

func (m *MultiSignatureInfo) GetMultisigAddress() string {
	if m != nil {
		return m.MultisigAddress
	}
	return ""
}

func (m *MultiSignatureInfo) GetBlockHeight() uint32 {
	if m != nil {
		return m.BlockHeight
	}
	return 0
}

func (m *MultiSignatureInfo) GetLatest() bool {
	if m != nil {
		return m.Latest
	}
	return false
}

func (m *MultiSignatureInfo) GetAddresses() []string {
	if m != nil {
		return m.Addresses
	}
	return nil
}

// represent the signature posted by account
type SignatureInfo struct {
	TransactionHash      []byte            `protobuf:"bytes,1,opt,name=TransactionHash,proto3" json:"TransactionHash,omitempty"`
	Signatures           map[string][]byte `protobuf:"bytes,2,rep,name=Signatures,proto3" json:"Signatures,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *SignatureInfo) Reset()         { *m = SignatureInfo{} }
func (m *SignatureInfo) String() string { return proto.CompactTextString(m) }
func (*SignatureInfo) ProtoMessage()    {}
func (*SignatureInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_136af44c597c17ae, []int{1}
}

func (m *SignatureInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SignatureInfo.Unmarshal(m, b)
}
func (m *SignatureInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SignatureInfo.Marshal(b, m, deterministic)
}
func (m *SignatureInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SignatureInfo.Merge(m, src)
}
func (m *SignatureInfo) XXX_Size() int {
	return xxx_messageInfo_SignatureInfo.Size(m)
}
func (m *SignatureInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_SignatureInfo.DiscardUnknown(m)
}

var xxx_messageInfo_SignatureInfo proto.InternalMessageInfo

func (m *SignatureInfo) GetTransactionHash() []byte {
	if m != nil {
		return m.TransactionHash
	}
	return nil
}

func (m *SignatureInfo) GetSignatures() map[string][]byte {
	if m != nil {
		return m.Signatures
	}
	return nil
}

// represent the multi signature's participant account addresses
type MultiSignatureParticipant struct {
	MultiSignatureAddress string   `protobuf:"bytes,1,opt,name=MultiSignatureAddress,proto3" json:"MultiSignatureAddress,omitempty"`
	AccountAddress        string   `protobuf:"bytes,2,opt,name=AccountAddress,proto3" json:"AccountAddress,omitempty"`
	AccountAddressIndex   uint32   `protobuf:"varint,3,opt,name=AccountAddressIndex,proto3" json:"AccountAddressIndex,omitempty"`
	Latest                bool     `protobuf:"varint,4,opt,name=Latest,proto3" json:"Latest,omitempty"`
	BlockHeight           uint32   `protobuf:"varint,5,opt,name=BlockHeight,proto3" json:"BlockHeight,omitempty"`
	XXX_NoUnkeyedLiteral  struct{} `json:"-"`
	XXX_unrecognized      []byte   `json:"-"`
	XXX_sizecache         int32    `json:"-"`
}

func (m *MultiSignatureParticipant) Reset()         { *m = MultiSignatureParticipant{} }
func (m *MultiSignatureParticipant) String() string { return proto.CompactTextString(m) }
func (*MultiSignatureParticipant) ProtoMessage()    {}
func (*MultiSignatureParticipant) Descriptor() ([]byte, []int) {
	return fileDescriptor_136af44c597c17ae, []int{2}
}

func (m *MultiSignatureParticipant) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MultiSignatureParticipant.Unmarshal(m, b)
}
func (m *MultiSignatureParticipant) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MultiSignatureParticipant.Marshal(b, m, deterministic)
}
func (m *MultiSignatureParticipant) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MultiSignatureParticipant.Merge(m, src)
}
func (m *MultiSignatureParticipant) XXX_Size() int {
	return xxx_messageInfo_MultiSignatureParticipant.Size(m)
}
func (m *MultiSignatureParticipant) XXX_DiscardUnknown() {
	xxx_messageInfo_MultiSignatureParticipant.DiscardUnknown(m)
}

var xxx_messageInfo_MultiSignatureParticipant proto.InternalMessageInfo

func (m *MultiSignatureParticipant) GetMultiSignatureAddress() string {
	if m != nil {
		return m.MultiSignatureAddress
	}
	return ""
}

func (m *MultiSignatureParticipant) GetAccountAddress() string {
	if m != nil {
		return m.AccountAddress
	}
	return ""
}

func (m *MultiSignatureParticipant) GetAccountAddressIndex() uint32 {
	if m != nil {
		return m.AccountAddressIndex
	}
	return 0
}

func (m *MultiSignatureParticipant) GetLatest() bool {
	if m != nil {
		return m.Latest
	}
	return false
}

func (m *MultiSignatureParticipant) GetBlockHeight() uint32 {
	if m != nil {
		return m.BlockHeight
	}
	return 0
}

// represent the pending signature counter stored by node for multi-signature transaction
type PendingSignature struct {
	TransactionHash      []byte   `protobuf:"bytes,1,opt,name=TransactionHash,proto3" json:"TransactionHash,omitempty"`
	AccountAddress       string   `protobuf:"bytes,2,opt,name=AccountAddress,proto3" json:"AccountAddress,omitempty"`
	Signature            []byte   `protobuf:"bytes,3,opt,name=Signature,proto3" json:"Signature,omitempty"`
	BlockHeight          uint32   `protobuf:"varint,4,opt,name=BlockHeight,proto3" json:"BlockHeight,omitempty"`
	Latest               bool     `protobuf:"varint,5,opt,name=Latest,proto3" json:"Latest,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PendingSignature) Reset()         { *m = PendingSignature{} }
func (m *PendingSignature) String() string { return proto.CompactTextString(m) }
func (*PendingSignature) ProtoMessage()    {}
func (*PendingSignature) Descriptor() ([]byte, []int) {
	return fileDescriptor_136af44c597c17ae, []int{3}
}

func (m *PendingSignature) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PendingSignature.Unmarshal(m, b)
}
func (m *PendingSignature) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PendingSignature.Marshal(b, m, deterministic)
}
func (m *PendingSignature) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PendingSignature.Merge(m, src)
}
func (m *PendingSignature) XXX_Size() int {
	return xxx_messageInfo_PendingSignature.Size(m)
}
func (m *PendingSignature) XXX_DiscardUnknown() {
	xxx_messageInfo_PendingSignature.DiscardUnknown(m)
}

var xxx_messageInfo_PendingSignature proto.InternalMessageInfo

func (m *PendingSignature) GetTransactionHash() []byte {
	if m != nil {
		return m.TransactionHash
	}
	return nil
}

func (m *PendingSignature) GetAccountAddress() string {
	if m != nil {
		return m.AccountAddress
	}
	return ""
}

func (m *PendingSignature) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func (m *PendingSignature) GetBlockHeight() uint32 {
	if m != nil {
		return m.BlockHeight
	}
	return 0
}

func (m *PendingSignature) GetLatest() bool {
	if m != nil {
		return m.Latest
	}
	return false
}

// represent transaction inside multisig body
type PendingTransaction struct {
	SenderAddress        string                   `protobuf:"bytes,1,opt,name=SenderAddress,proto3" json:"SenderAddress,omitempty"`
	TransactionHash      []byte                   `protobuf:"bytes,2,opt,name=TransactionHash,proto3" json:"TransactionHash,omitempty"`
	TransactionBytes     []byte                   `protobuf:"bytes,3,opt,name=TransactionBytes,proto3" json:"TransactionBytes,omitempty"`
	Status               PendingTransactionStatus `protobuf:"varint,4,opt,name=Status,proto3,enum=model.PendingTransactionStatus" json:"Status,omitempty"`
	BlockHeight          uint32                   `protobuf:"varint,5,opt,name=BlockHeight,proto3" json:"BlockHeight,omitempty"`
	Latest               bool                     `protobuf:"varint,6,opt,name=Latest,proto3" json:"Latest,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                 `json:"-"`
	XXX_unrecognized     []byte                   `json:"-"`
	XXX_sizecache        int32                    `json:"-"`
}

func (m *PendingTransaction) Reset()         { *m = PendingTransaction{} }
func (m *PendingTransaction) String() string { return proto.CompactTextString(m) }
func (*PendingTransaction) ProtoMessage()    {}
func (*PendingTransaction) Descriptor() ([]byte, []int) {
	return fileDescriptor_136af44c597c17ae, []int{4}
}

func (m *PendingTransaction) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PendingTransaction.Unmarshal(m, b)
}
func (m *PendingTransaction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PendingTransaction.Marshal(b, m, deterministic)
}
func (m *PendingTransaction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PendingTransaction.Merge(m, src)
}
func (m *PendingTransaction) XXX_Size() int {
	return xxx_messageInfo_PendingTransaction.Size(m)
}
func (m *PendingTransaction) XXX_DiscardUnknown() {
	xxx_messageInfo_PendingTransaction.DiscardUnknown(m)
}

var xxx_messageInfo_PendingTransaction proto.InternalMessageInfo

func (m *PendingTransaction) GetSenderAddress() string {
	if m != nil {
		return m.SenderAddress
	}
	return ""
}

func (m *PendingTransaction) GetTransactionHash() []byte {
	if m != nil {
		return m.TransactionHash
	}
	return nil
}

func (m *PendingTransaction) GetTransactionBytes() []byte {
	if m != nil {
		return m.TransactionBytes
	}
	return nil
}

func (m *PendingTransaction) GetStatus() PendingTransactionStatus {
	if m != nil {
		return m.Status
	}
	return PendingTransactionStatus_PendingTransactionPending
}

func (m *PendingTransaction) GetBlockHeight() uint32 {
	if m != nil {
		return m.BlockHeight
	}
	return 0
}

func (m *PendingTransaction) GetLatest() bool {
	if m != nil {
		return m.Latest
	}
	return false
}

type GetPendingTransactionsRequest struct {
	SenderAddress        string                   `protobuf:"bytes,1,opt,name=SenderAddress,proto3" json:"SenderAddress,omitempty"`
	Status               PendingTransactionStatus `protobuf:"varint,2,opt,name=Status,proto3,enum=model.PendingTransactionStatus" json:"Status,omitempty"`
	Pagination           *Pagination              `protobuf:"bytes,3,opt,name=Pagination,proto3" json:"Pagination,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                 `json:"-"`
	XXX_unrecognized     []byte                   `json:"-"`
	XXX_sizecache        int32                    `json:"-"`
}

func (m *GetPendingTransactionsRequest) Reset()         { *m = GetPendingTransactionsRequest{} }
func (m *GetPendingTransactionsRequest) String() string { return proto.CompactTextString(m) }
func (*GetPendingTransactionsRequest) ProtoMessage()    {}
func (*GetPendingTransactionsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_136af44c597c17ae, []int{5}
}

func (m *GetPendingTransactionsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetPendingTransactionsRequest.Unmarshal(m, b)
}
func (m *GetPendingTransactionsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetPendingTransactionsRequest.Marshal(b, m, deterministic)
}
func (m *GetPendingTransactionsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetPendingTransactionsRequest.Merge(m, src)
}
func (m *GetPendingTransactionsRequest) XXX_Size() int {
	return xxx_messageInfo_GetPendingTransactionsRequest.Size(m)
}
func (m *GetPendingTransactionsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetPendingTransactionsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetPendingTransactionsRequest proto.InternalMessageInfo

func (m *GetPendingTransactionsRequest) GetSenderAddress() string {
	if m != nil {
		return m.SenderAddress
	}
	return ""
}

func (m *GetPendingTransactionsRequest) GetStatus() PendingTransactionStatus {
	if m != nil {
		return m.Status
	}
	return PendingTransactionStatus_PendingTransactionPending
}

func (m *GetPendingTransactionsRequest) GetPagination() *Pagination {
	if m != nil {
		return m.Pagination
	}
	return nil
}

type GetPendingTransactionsResponse struct {
	// Number of item in current page
	Count uint32 `protobuf:"varint,1,opt,name=Count,proto3" json:"Count,omitempty"`
	// Starting page
	Page uint32 `protobuf:"varint,2,opt,name=Page,proto3" json:"Page,omitempty"`
	// content of the request
	PendingTransactions  []*PendingTransaction `protobuf:"bytes,3,rep,name=PendingTransactions,proto3" json:"PendingTransactions,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *GetPendingTransactionsResponse) Reset()         { *m = GetPendingTransactionsResponse{} }
func (m *GetPendingTransactionsResponse) String() string { return proto.CompactTextString(m) }
func (*GetPendingTransactionsResponse) ProtoMessage()    {}
func (*GetPendingTransactionsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_136af44c597c17ae, []int{6}
}

func (m *GetPendingTransactionsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetPendingTransactionsResponse.Unmarshal(m, b)
}
func (m *GetPendingTransactionsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetPendingTransactionsResponse.Marshal(b, m, deterministic)
}
func (m *GetPendingTransactionsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetPendingTransactionsResponse.Merge(m, src)
}
func (m *GetPendingTransactionsResponse) XXX_Size() int {
	return xxx_messageInfo_GetPendingTransactionsResponse.Size(m)
}
func (m *GetPendingTransactionsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetPendingTransactionsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetPendingTransactionsResponse proto.InternalMessageInfo

func (m *GetPendingTransactionsResponse) GetCount() uint32 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *GetPendingTransactionsResponse) GetPage() uint32 {
	if m != nil {
		return m.Page
	}
	return 0
}

func (m *GetPendingTransactionsResponse) GetPendingTransactions() []*PendingTransaction {
	if m != nil {
		return m.PendingTransactions
	}
	return nil
}

type GetPendingTransactionDetailByTransactionHashRequest struct {
	// hex of transaction hash
	TransactionHashHex   string   `protobuf:"bytes,1,opt,name=TransactionHashHex,proto3" json:"TransactionHashHex,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetPendingTransactionDetailByTransactionHashRequest) Reset() {
	*m = GetPendingTransactionDetailByTransactionHashRequest{}
}
func (m *GetPendingTransactionDetailByTransactionHashRequest) String() string {
	return proto.CompactTextString(m)
}
func (*GetPendingTransactionDetailByTransactionHashRequest) ProtoMessage() {}
func (*GetPendingTransactionDetailByTransactionHashRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_136af44c597c17ae, []int{7}
}

func (m *GetPendingTransactionDetailByTransactionHashRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetPendingTransactionDetailByTransactionHashRequest.Unmarshal(m, b)
}
func (m *GetPendingTransactionDetailByTransactionHashRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetPendingTransactionDetailByTransactionHashRequest.Marshal(b, m, deterministic)
}
func (m *GetPendingTransactionDetailByTransactionHashRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetPendingTransactionDetailByTransactionHashRequest.Merge(m, src)
}
func (m *GetPendingTransactionDetailByTransactionHashRequest) XXX_Size() int {
	return xxx_messageInfo_GetPendingTransactionDetailByTransactionHashRequest.Size(m)
}
func (m *GetPendingTransactionDetailByTransactionHashRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetPendingTransactionDetailByTransactionHashRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetPendingTransactionDetailByTransactionHashRequest proto.InternalMessageInfo

func (m *GetPendingTransactionDetailByTransactionHashRequest) GetTransactionHashHex() string {
	if m != nil {
		return m.TransactionHashHex
	}
	return ""
}

type GetPendingTransactionDetailByTransactionHashResponse struct {
	PendingTransaction   *PendingTransaction `protobuf:"bytes,1,opt,name=PendingTransaction,proto3" json:"PendingTransaction,omitempty"`
	PendingSignatures    []*PendingSignature `protobuf:"bytes,2,rep,name=PendingSignatures,proto3" json:"PendingSignatures,omitempty"`
	MultiSignatureInfo   *MultiSignatureInfo `protobuf:"bytes,3,opt,name=MultiSignatureInfo,proto3" json:"MultiSignatureInfo,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *GetPendingTransactionDetailByTransactionHashResponse) Reset() {
	*m = GetPendingTransactionDetailByTransactionHashResponse{}
}
func (m *GetPendingTransactionDetailByTransactionHashResponse) String() string {
	return proto.CompactTextString(m)
}
func (*GetPendingTransactionDetailByTransactionHashResponse) ProtoMessage() {}
func (*GetPendingTransactionDetailByTransactionHashResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_136af44c597c17ae, []int{8}
}

func (m *GetPendingTransactionDetailByTransactionHashResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetPendingTransactionDetailByTransactionHashResponse.Unmarshal(m, b)
}
func (m *GetPendingTransactionDetailByTransactionHashResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetPendingTransactionDetailByTransactionHashResponse.Marshal(b, m, deterministic)
}
func (m *GetPendingTransactionDetailByTransactionHashResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetPendingTransactionDetailByTransactionHashResponse.Merge(m, src)
}
func (m *GetPendingTransactionDetailByTransactionHashResponse) XXX_Size() int {
	return xxx_messageInfo_GetPendingTransactionDetailByTransactionHashResponse.Size(m)
}
func (m *GetPendingTransactionDetailByTransactionHashResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetPendingTransactionDetailByTransactionHashResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetPendingTransactionDetailByTransactionHashResponse proto.InternalMessageInfo

func (m *GetPendingTransactionDetailByTransactionHashResponse) GetPendingTransaction() *PendingTransaction {
	if m != nil {
		return m.PendingTransaction
	}
	return nil
}

func (m *GetPendingTransactionDetailByTransactionHashResponse) GetPendingSignatures() []*PendingSignature {
	if m != nil {
		return m.PendingSignatures
	}
	return nil
}

func (m *GetPendingTransactionDetailByTransactionHashResponse) GetMultiSignatureInfo() *MultiSignatureInfo {
	if m != nil {
		return m.MultiSignatureInfo
	}
	return nil
}

type GetMultisignatureInfoRequest struct {
	MultisigAddress      string      `protobuf:"bytes,1,opt,name=MultisigAddress,proto3" json:"MultisigAddress,omitempty"`
	Pagination           *Pagination `protobuf:"bytes,2,opt,name=Pagination,proto3" json:"Pagination,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *GetMultisignatureInfoRequest) Reset()         { *m = GetMultisignatureInfoRequest{} }
func (m *GetMultisignatureInfoRequest) String() string { return proto.CompactTextString(m) }
func (*GetMultisignatureInfoRequest) ProtoMessage()    {}
func (*GetMultisignatureInfoRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_136af44c597c17ae, []int{9}
}

func (m *GetMultisignatureInfoRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetMultisignatureInfoRequest.Unmarshal(m, b)
}
func (m *GetMultisignatureInfoRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetMultisignatureInfoRequest.Marshal(b, m, deterministic)
}
func (m *GetMultisignatureInfoRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetMultisignatureInfoRequest.Merge(m, src)
}
func (m *GetMultisignatureInfoRequest) XXX_Size() int {
	return xxx_messageInfo_GetMultisignatureInfoRequest.Size(m)
}
func (m *GetMultisignatureInfoRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetMultisignatureInfoRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetMultisignatureInfoRequest proto.InternalMessageInfo

func (m *GetMultisignatureInfoRequest) GetMultisigAddress() string {
	if m != nil {
		return m.MultisigAddress
	}
	return ""
}

func (m *GetMultisignatureInfoRequest) GetPagination() *Pagination {
	if m != nil {
		return m.Pagination
	}
	return nil
}

type GetMultisignatureInfoResponse struct {
	// Number of item in current page
	Count uint32 `protobuf:"varint,1,opt,name=Count,proto3" json:"Count,omitempty"`
	// Starting page
	Page uint32 `protobuf:"varint,2,opt,name=Page,proto3" json:"Page,omitempty"`
	// content of the request
	MultisignatureInfo   []*MultiSignatureInfo `protobuf:"bytes,3,rep,name=MultisignatureInfo,proto3" json:"MultisignatureInfo,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *GetMultisignatureInfoResponse) Reset()         { *m = GetMultisignatureInfoResponse{} }
func (m *GetMultisignatureInfoResponse) String() string { return proto.CompactTextString(m) }
func (*GetMultisignatureInfoResponse) ProtoMessage()    {}
func (*GetMultisignatureInfoResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_136af44c597c17ae, []int{10}
}

func (m *GetMultisignatureInfoResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetMultisignatureInfoResponse.Unmarshal(m, b)
}
func (m *GetMultisignatureInfoResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetMultisignatureInfoResponse.Marshal(b, m, deterministic)
}
func (m *GetMultisignatureInfoResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetMultisignatureInfoResponse.Merge(m, src)
}
func (m *GetMultisignatureInfoResponse) XXX_Size() int {
	return xxx_messageInfo_GetMultisignatureInfoResponse.Size(m)
}
func (m *GetMultisignatureInfoResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetMultisignatureInfoResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetMultisignatureInfoResponse proto.InternalMessageInfo

func (m *GetMultisignatureInfoResponse) GetCount() uint32 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *GetMultisignatureInfoResponse) GetPage() uint32 {
	if m != nil {
		return m.Page
	}
	return 0
}

func (m *GetMultisignatureInfoResponse) GetMultisignatureInfo() []*MultiSignatureInfo {
	if m != nil {
		return m.MultisignatureInfo
	}
	return nil
}

type GetMultisigAddressByParticipantAddressesRequest struct {
	Addresses            []string    `protobuf:"bytes,1,rep,name=addresses,proto3" json:"addresses,omitempty"`
	Pagination           *Pagination `protobuf:"bytes,2,opt,name=Pagination,proto3" json:"Pagination,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *GetMultisigAddressByParticipantAddressesRequest) Reset() {
	*m = GetMultisigAddressByParticipantAddressesRequest{}
}
func (m *GetMultisigAddressByParticipantAddressesRequest) String() string {
	return proto.CompactTextString(m)
}
func (*GetMultisigAddressByParticipantAddressesRequest) ProtoMessage() {}
func (*GetMultisigAddressByParticipantAddressesRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_136af44c597c17ae, []int{11}
}

func (m *GetMultisigAddressByParticipantAddressesRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetMultisigAddressByParticipantAddressesRequest.Unmarshal(m, b)
}
func (m *GetMultisigAddressByParticipantAddressesRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetMultisigAddressByParticipantAddressesRequest.Marshal(b, m, deterministic)
}
func (m *GetMultisigAddressByParticipantAddressesRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetMultisigAddressByParticipantAddressesRequest.Merge(m, src)
}
func (m *GetMultisigAddressByParticipantAddressesRequest) XXX_Size() int {
	return xxx_messageInfo_GetMultisigAddressByParticipantAddressesRequest.Size(m)
}
func (m *GetMultisigAddressByParticipantAddressesRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetMultisigAddressByParticipantAddressesRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetMultisigAddressByParticipantAddressesRequest proto.InternalMessageInfo

func (m *GetMultisigAddressByParticipantAddressesRequest) GetAddresses() []string {
	if m != nil {
		return m.Addresses
	}
	return nil
}

func (m *GetMultisigAddressByParticipantAddressesRequest) GetPagination() *Pagination {
	if m != nil {
		return m.Pagination
	}
	return nil
}

type Addresses struct {
	Addresses            []string `protobuf:"bytes,1,rep,name=addresses,proto3" json:"addresses,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Addresses) Reset()         { *m = Addresses{} }
func (m *Addresses) String() string { return proto.CompactTextString(m) }
func (*Addresses) ProtoMessage()    {}
func (*Addresses) Descriptor() ([]byte, []int) {
	return fileDescriptor_136af44c597c17ae, []int{12}
}

func (m *Addresses) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Addresses.Unmarshal(m, b)
}
func (m *Addresses) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Addresses.Marshal(b, m, deterministic)
}
func (m *Addresses) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Addresses.Merge(m, src)
}
func (m *Addresses) XXX_Size() int {
	return xxx_messageInfo_Addresses.Size(m)
}
func (m *Addresses) XXX_DiscardUnknown() {
	xxx_messageInfo_Addresses.DiscardUnknown(m)
}

var xxx_messageInfo_Addresses proto.InternalMessageInfo

func (m *Addresses) GetAddresses() []string {
	if m != nil {
		return m.Addresses
	}
	return nil
}

type GetMultisigAddressByParticipantAddressesResponse struct {
	// Total of participant address
	Total uint32 `protobuf:"varint,1,opt,name=Total,proto3" json:"Total,omitempty"`
	// content of the request
	MultiSignatureAddresses map[string]*Addresses `protobuf:"bytes,2,rep,name=MultiSignatureAddresses,proto3" json:"MultiSignatureAddresses,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral    struct{}              `json:"-"`
	XXX_unrecognized        []byte                `json:"-"`
	XXX_sizecache           int32                 `json:"-"`
}

func (m *GetMultisigAddressByParticipantAddressesResponse) Reset() {
	*m = GetMultisigAddressByParticipantAddressesResponse{}
}
func (m *GetMultisigAddressByParticipantAddressesResponse) String() string {
	return proto.CompactTextString(m)
}
func (*GetMultisigAddressByParticipantAddressesResponse) ProtoMessage() {}
func (*GetMultisigAddressByParticipantAddressesResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_136af44c597c17ae, []int{13}
}

func (m *GetMultisigAddressByParticipantAddressesResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetMultisigAddressByParticipantAddressesResponse.Unmarshal(m, b)
}
func (m *GetMultisigAddressByParticipantAddressesResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetMultisigAddressByParticipantAddressesResponse.Marshal(b, m, deterministic)
}
func (m *GetMultisigAddressByParticipantAddressesResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetMultisigAddressByParticipantAddressesResponse.Merge(m, src)
}
func (m *GetMultisigAddressByParticipantAddressesResponse) XXX_Size() int {
	return xxx_messageInfo_GetMultisigAddressByParticipantAddressesResponse.Size(m)
}
func (m *GetMultisigAddressByParticipantAddressesResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetMultisigAddressByParticipantAddressesResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetMultisigAddressByParticipantAddressesResponse proto.InternalMessageInfo

func (m *GetMultisigAddressByParticipantAddressesResponse) GetTotal() uint32 {
	if m != nil {
		return m.Total
	}
	return 0
}

func (m *GetMultisigAddressByParticipantAddressesResponse) GetMultiSignatureAddresses() map[string]*Addresses {
	if m != nil {
		return m.MultiSignatureAddresses
	}
	return nil
}

func init() {
	proto.RegisterEnum("model.PendingTransactionStatus", PendingTransactionStatus_name, PendingTransactionStatus_value)
	proto.RegisterType((*MultiSignatureInfo)(nil), "model.MultiSignatureInfo")
	proto.RegisterType((*SignatureInfo)(nil), "model.SignatureInfo")
	proto.RegisterMapType((map[string][]byte)(nil), "model.SignatureInfo.SignaturesEntry")
	proto.RegisterType((*MultiSignatureParticipant)(nil), "model.MultiSignatureParticipant")
	proto.RegisterType((*PendingSignature)(nil), "model.PendingSignature")
	proto.RegisterType((*PendingTransaction)(nil), "model.PendingTransaction")
	proto.RegisterType((*GetPendingTransactionsRequest)(nil), "model.GetPendingTransactionsRequest")
	proto.RegisterType((*GetPendingTransactionsResponse)(nil), "model.GetPendingTransactionsResponse")
	proto.RegisterType((*GetPendingTransactionDetailByTransactionHashRequest)(nil), "model.GetPendingTransactionDetailByTransactionHashRequest")
	proto.RegisterType((*GetPendingTransactionDetailByTransactionHashResponse)(nil), "model.GetPendingTransactionDetailByTransactionHashResponse")
	proto.RegisterType((*GetMultisignatureInfoRequest)(nil), "model.GetMultisignatureInfoRequest")
	proto.RegisterType((*GetMultisignatureInfoResponse)(nil), "model.GetMultisignatureInfoResponse")
	proto.RegisterType((*GetMultisigAddressByParticipantAddressesRequest)(nil), "model.GetMultisigAddressByParticipantAddressesRequest")
	proto.RegisterType((*Addresses)(nil), "model.addresses")
	proto.RegisterType((*GetMultisigAddressByParticipantAddressesResponse)(nil), "model.GetMultisigAddressByParticipantAddressesResponse")
	proto.RegisterMapType((map[string]*Addresses)(nil), "model.GetMultisigAddressByParticipantAddressesResponse.MultiSignatureAddressesEntry")
}

func init() { proto.RegisterFile("model/multiSignature.proto", fileDescriptor_136af44c597c17ae) }

var fileDescriptor_136af44c597c17ae = []byte{
	// 872 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x56, 0x4f, 0x6f, 0x1b, 0x45,
	0x14, 0x67, 0x76, 0x63, 0x0b, 0x3f, 0xc7, 0xcd, 0x66, 0x0a, 0xe9, 0xc6, 0x72, 0x8b, 0xb5, 0xaa,
	0xaa, 0x25, 0x02, 0x3b, 0xb8, 0x95, 0x40, 0x48, 0x1c, 0x6a, 0x6a, 0x35, 0x11, 0xb4, 0x58, 0x93,
	0x9c, 0x10, 0x97, 0xcd, 0xee, 0xe0, 0xac, 0x6a, 0xcf, 0x98, 0xdd, 0x59, 0x64, 0xc3, 0x0d, 0xee,
	0x5c, 0x88, 0xc4, 0xe7, 0xe0, 0xc4, 0x05, 0xf1, 0x51, 0xb8, 0xf1, 0x3d, 0x90, 0x67, 0xff, 0x78,
	0xff, 0x8c, 0x53, 0x3b, 0x17, 0xcb, 0xfb, 0xfe, 0xfe, 0xde, 0xef, 0xbd, 0x99, 0x37, 0xd0, 0x9e,
	0x71, 0x8f, 0x4e, 0xfb, 0xb3, 0x68, 0x2a, 0xfc, 0x0b, 0x7f, 0xc2, 0x1c, 0x11, 0x05, 0xb4, 0x37,
	0x0f, 0xb8, 0xe0, 0xb8, 0x26, 0x75, 0xed, 0xa3, 0xd8, 0x64, 0xee, 0x4c, 0x7c, 0xe6, 0x08, 0x9f,
	0xb3, 0x58, 0x6d, 0xfd, 0x8b, 0x00, 0xbf, 0x2a, 0xf8, 0x9d, 0xb3, 0xef, 0x39, 0xfe, 0x08, 0x0e,
	0x5f, 0xf9, 0xcc, 0x9f, 0x45, 0xb3, 0x4c, 0x1e, 0x9a, 0xa8, 0x8b, 0xec, 0x16, 0xa9, 0x2a, 0xb0,
	0x09, 0xb5, 0xd7, 0x9c, 0xb9, 0xd4, 0xd4, 0xba, 0xc8, 0xd6, 0x87, 0xda, 0x29, 0x22, 0xb1, 0x00,
	0xdb, 0x70, 0x20, 0xa3, 0x87, 0xfe, 0xe4, 0xb9, 0xe7, 0x05, 0x34, 0x0c, 0x4d, 0xbd, 0x8b, 0xec,
	0x06, 0x29, 0x8b, 0x71, 0x17, 0x9a, 0xc3, 0x29, 0x77, 0xdf, 0x9c, 0x51, 0x7f, 0x72, 0x2d, 0xcc,
	0x3d, 0x99, 0x2b, 0x2f, 0xc2, 0x47, 0x50, 0xff, 0xda, 0x11, 0x34, 0x14, 0x66, 0xad, 0x8b, 0xec,
	0x77, 0x49, 0xf2, 0x85, 0x3b, 0xd0, 0x48, 0x82, 0xd0, 0xd0, 0xac, 0x77, 0x75, 0xbb, 0x41, 0xd6,
	0x02, 0xeb, 0x1f, 0x04, 0xad, 0x62, 0x6d, 0x36, 0x1c, 0x5c, 0x06, 0x0e, 0x0b, 0x1d, 0x77, 0xc5,
	0xc3, 0x99, 0x13, 0x5e, 0xcb, 0xca, 0xf6, 0x49, 0x59, 0x8c, 0x5f, 0x00, 0xe4, 0xca, 0xd7, 0xba,
	0xba, 0xdd, 0x1c, 0x3c, 0xee, 0x49, 0x26, 0x7b, 0x85, 0x98, 0xeb, 0xaf, 0x70, 0xc4, 0x44, 0xb0,
	0x24, 0x39, 0xbf, 0xf6, 0x17, 0x70, 0x50, 0x52, 0x63, 0x03, 0xf4, 0x37, 0x74, 0x29, 0xd3, 0x36,
	0xc8, 0xea, 0x2f, 0x7e, 0x0f, 0x6a, 0x3f, 0x3a, 0xd3, 0x28, 0xa6, 0x70, 0x9f, 0xc4, 0x1f, 0x9f,
	0x6b, 0x9f, 0x21, 0xeb, 0x3f, 0x04, 0xc7, 0xc5, 0x0e, 0x8d, 0x9d, 0x40, 0xf8, 0xae, 0x3f, 0x77,
	0x98, 0xc0, 0xcf, 0xe0, 0xfd, 0xa2, 0x32, 0xa5, 0x39, 0x8e, 0xad, 0x56, 0xe2, 0x27, 0x70, 0xef,
	0xb9, 0xeb, 0xf2, 0x88, 0x89, 0xd4, 0x5c, 0x93, 0xe6, 0x25, 0x29, 0x3e, 0x85, 0xfb, 0x45, 0xc9,
	0x39, 0xf3, 0xe8, 0x42, 0xb6, 0xb0, 0x45, 0x54, 0xaa, 0x5c, 0x93, 0xf6, 0x0a, 0x4d, 0x2a, 0xb5,
	0xb7, 0x56, 0x69, 0xaf, 0xf5, 0x37, 0x02, 0x63, 0x4c, 0x99, 0xe7, 0xb3, 0x49, 0x86, 0x77, 0x87,
	0x5e, 0x6d, 0x5b, 0x52, 0x07, 0x1a, 0x59, 0x78, 0x59, 0xc8, 0x3e, 0x59, 0x0b, 0xee, 0x3e, 0x85,
	0xd6, 0xaf, 0x1a, 0xe0, 0x04, 0x7e, 0x0e, 0x1a, 0x7e, 0x0c, 0xad, 0x0b, 0xca, 0x3c, 0x1a, 0x14,
	0xfb, 0x52, 0x14, 0xaa, 0xca, 0xd4, 0xd4, 0x65, 0x9e, 0x80, 0x91, 0x13, 0x0d, 0x97, 0x82, 0x86,
	0x49, 0x15, 0x15, 0x39, 0xfe, 0x14, 0xea, 0x17, 0xc2, 0x11, 0x51, 0x28, 0xeb, 0xb8, 0x37, 0xf8,
	0x20, 0x19, 0xdd, 0x2a, 0xcc, 0xd8, 0x8c, 0x24, 0xe6, 0x6f, 0x6f, 0x56, 0x8e, 0x85, 0x7a, 0x81,
	0x85, 0x3f, 0x11, 0x3c, 0x7c, 0x49, 0x45, 0x35, 0x43, 0x48, 0xe8, 0x0f, 0xd1, 0x6a, 0x10, 0xb6,
	0x23, 0x64, 0x0d, 0x5d, 0xdb, 0x0d, 0xfa, 0x27, 0x00, 0xe3, 0xec, 0x8e, 0x93, 0xcc, 0x34, 0x07,
	0x87, 0xa9, 0x73, 0xa6, 0x20, 0x39, 0x23, 0xeb, 0x0f, 0x04, 0x8f, 0x36, 0x61, 0x0e, 0xe7, 0x9c,
	0x85, 0x74, 0x75, 0x3a, 0xbf, 0x5c, 0x0d, 0x51, 0x72, 0x05, 0xc6, 0x1f, 0x18, 0xc3, 0xde, 0xd8,
	0x99, 0xc4, 0x47, 0xb6, 0x45, 0xe4, 0x7f, 0xfc, 0x15, 0xdc, 0x57, 0x04, 0x32, 0x75, 0x79, 0x77,
	0x1c, 0x6f, 0xac, 0x82, 0xa8, 0xbc, 0x2c, 0x0a, 0x4f, 0x95, 0xc0, 0x5e, 0x50, 0xe1, 0xf8, 0xd3,
	0xe1, 0xb2, 0x34, 0x1c, 0x29, 0xc5, 0x3d, 0xc0, 0x25, 0xcd, 0x19, 0x5d, 0x24, 0x3c, 0x2b, 0x34,
	0xd6, 0xef, 0x1a, 0x3c, 0xdb, 0x2d, 0x4f, 0x42, 0xcb, 0xb9, 0x6a, 0xe4, 0x65, 0xa2, 0x5b, 0x6b,
	0x55, 0x9d, 0x93, 0x11, 0x1c, 0x96, 0x0f, 0x7f, 0x7a, 0xe3, 0x3e, 0x28, 0x46, 0xca, 0xf4, 0xa4,
	0xea, 0xb1, 0x42, 0x54, 0xdd, 0x66, 0xc9, 0x18, 0xa4, 0x88, 0xaa, 0x06, 0x44, 0xe1, 0x64, 0xfd,
	0x0c, 0x9d, 0x97, 0x54, 0xa4, 0x6b, 0x2a, 0x67, 0x9c, 0xb0, 0xac, 0x58, 0x6d, 0x48, 0xbd, 0xda,
	0x8a, 0x33, 0xa9, 0x6d, 0x33, 0x93, 0x37, 0xf1, 0x39, 0x52, 0x65, 0xdf, 0x79, 0x24, 0x53, 0x4e,
	0xc2, 0x12, 0x27, 0xfa, 0x36, 0x9c, 0x14, 0x9c, 0xac, 0x5f, 0x10, 0xf4, 0x73, 0xb0, 0x92, 0x02,
	0x87, 0xcb, 0xdc, 0x46, 0xca, 0x36, 0x6f, 0xca, 0x53, 0x07, 0x1a, 0x4e, 0xb6, 0x9e, 0x51, 0xbc,
	0x9e, 0x33, 0xc1, 0x5d, 0xb8, 0xf9, 0x30, 0x17, 0xf0, 0xf6, 0xe8, 0xd6, 0x5f, 0x1a, 0x9c, 0x6e,
	0x8f, 0x77, 0xcd, 0xec, 0x25, 0x17, 0xce, 0x34, 0x65, 0x56, 0x7e, 0xe0, 0xdf, 0x10, 0x3c, 0x50,
	0x2e, 0xd3, 0x6c, 0x4e, 0x2f, 0x13, 0xd8, 0xbb, 0x26, 0xec, 0x6d, 0x08, 0x1b, 0xbf, 0x24, 0x36,
	0x25, 0x6d, 0x7f, 0x07, 0x9d, 0xdb, 0x1c, 0x15, 0x6f, 0x8c, 0x27, 0xf9, 0x37, 0x46, 0x73, 0x60,
	0x24, 0x78, 0x33, 0xba, 0x72, 0xaf, 0x8e, 0x93, 0x1b, 0x04, 0xe6, 0xa6, 0xcb, 0x16, 0x3f, 0x84,
	0xe3, 0xaa, 0x2e, 0x91, 0x18, 0xef, 0xe0, 0x47, 0xd0, 0xae, 0xaa, 0x47, 0x0b, 0xea, 0x46, 0x82,
	0x7a, 0x06, 0xc2, 0x6d, 0x38, 0xaa, 0xea, 0x5f, 0xf3, 0x6f, 0xe6, 0x86, 0xa6, 0x0e, 0x3d, 0x5a,
	0xcc, 0xfd, 0x80, 0x7a, 0x86, 0x3e, 0x3c, 0xf9, 0xd6, 0x9e, 0xf8, 0xe2, 0x3a, 0xba, 0xea, 0xb9,
	0x7c, 0xd6, 0xff, 0x89, 0xf3, 0x2b, 0x37, 0xfe, 0xfd, 0xd8, 0xe5, 0x01, 0xed, 0xbb, 0x7c, 0x36,
	0xe3, 0xac, 0x2f, 0xeb, 0xba, 0xaa, 0xcb, 0x17, 0xee, 0xd3, 0xff, 0x03, 0x00, 0x00, 0xff, 0xff,
	0xcd, 0xdb, 0x14, 0x31, 0x1e, 0x0b, 0x00, 0x00,
}
