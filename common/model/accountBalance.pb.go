// Code generated by protoc-gen-go. DO NOT EDIT.
// source: model/accountBalance.proto

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

// AccountBalance represent the transaction data structure stored in the database
type AccountBalance struct {
	AccountAddressType   uint32   `protobuf:"varint,1,opt,name=AccountAddressType,proto3" json:"AccountAddressType,omitempty"`
	AccountAddress       []byte   `protobuf:"bytes,2,opt,name=AccountAddress,proto3" json:"AccountAddress,omitempty"`
	BlockHeight          uint32   `protobuf:"varint,3,opt,name=BlockHeight,proto3" json:"BlockHeight,omitempty"`
	SpendableBalance     int64    `protobuf:"varint,4,opt,name=SpendableBalance,proto3" json:"SpendableBalance,omitempty"`
	Balance              int64    `protobuf:"varint,5,opt,name=Balance,proto3" json:"Balance,omitempty"`
	PopRevenue           int64    `protobuf:"varint,6,opt,name=PopRevenue,proto3" json:"PopRevenue,omitempty"`
	Latest               bool     `protobuf:"varint,7,opt,name=Latest,proto3" json:"Latest,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AccountBalance) Reset()         { *m = AccountBalance{} }
func (m *AccountBalance) String() string { return proto.CompactTextString(m) }
func (*AccountBalance) ProtoMessage()    {}
func (*AccountBalance) Descriptor() ([]byte, []int) {
	return fileDescriptor_44b9b1c521a5bcaa, []int{0}
}

func (m *AccountBalance) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AccountBalance.Unmarshal(m, b)
}
func (m *AccountBalance) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AccountBalance.Marshal(b, m, deterministic)
}
func (m *AccountBalance) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AccountBalance.Merge(m, src)
}
func (m *AccountBalance) XXX_Size() int {
	return xxx_messageInfo_AccountBalance.Size(m)
}
func (m *AccountBalance) XXX_DiscardUnknown() {
	xxx_messageInfo_AccountBalance.DiscardUnknown(m)
}

var xxx_messageInfo_AccountBalance proto.InternalMessageInfo

func (m *AccountBalance) GetAccountAddressType() uint32 {
	if m != nil {
		return m.AccountAddressType
	}
	return 0
}

func (m *AccountBalance) GetAccountAddress() []byte {
	if m != nil {
		return m.AccountAddress
	}
	return nil
}

func (m *AccountBalance) GetBlockHeight() uint32 {
	if m != nil {
		return m.BlockHeight
	}
	return 0
}

func (m *AccountBalance) GetSpendableBalance() int64 {
	if m != nil {
		return m.SpendableBalance
	}
	return 0
}

func (m *AccountBalance) GetBalance() int64 {
	if m != nil {
		return m.Balance
	}
	return 0
}

func (m *AccountBalance) GetPopRevenue() int64 {
	if m != nil {
		return m.PopRevenue
	}
	return 0
}

func (m *AccountBalance) GetLatest() bool {
	if m != nil {
		return m.Latest
	}
	return false
}

type GetAccountBalanceRequest struct {
	// Fetch AccountBalance by type/address
	AccountAddressType   uint32   `protobuf:"varint,1,opt,name=AccountAddressType,proto3" json:"AccountAddressType,omitempty"`
	AccountAddress       []byte   `protobuf:"bytes,2,opt,name=AccountAddress,proto3" json:"AccountAddress,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetAccountBalanceRequest) Reset()         { *m = GetAccountBalanceRequest{} }
func (m *GetAccountBalanceRequest) String() string { return proto.CompactTextString(m) }
func (*GetAccountBalanceRequest) ProtoMessage()    {}
func (*GetAccountBalanceRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_44b9b1c521a5bcaa, []int{1}
}

func (m *GetAccountBalanceRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetAccountBalanceRequest.Unmarshal(m, b)
}
func (m *GetAccountBalanceRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetAccountBalanceRequest.Marshal(b, m, deterministic)
}
func (m *GetAccountBalanceRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetAccountBalanceRequest.Merge(m, src)
}
func (m *GetAccountBalanceRequest) XXX_Size() int {
	return xxx_messageInfo_GetAccountBalanceRequest.Size(m)
}
func (m *GetAccountBalanceRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetAccountBalanceRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetAccountBalanceRequest proto.InternalMessageInfo

func (m *GetAccountBalanceRequest) GetAccountAddressType() uint32 {
	if m != nil {
		return m.AccountAddressType
	}
	return 0
}

func (m *GetAccountBalanceRequest) GetAccountAddress() []byte {
	if m != nil {
		return m.AccountAddress
	}
	return nil
}

type GetAccountBalanceResponse struct {
	AccountBalance       *AccountBalance `protobuf:"bytes,1,opt,name=AccountBalance,proto3" json:"AccountBalance,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *GetAccountBalanceResponse) Reset()         { *m = GetAccountBalanceResponse{} }
func (m *GetAccountBalanceResponse) String() string { return proto.CompactTextString(m) }
func (*GetAccountBalanceResponse) ProtoMessage()    {}
func (*GetAccountBalanceResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_44b9b1c521a5bcaa, []int{2}
}

func (m *GetAccountBalanceResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetAccountBalanceResponse.Unmarshal(m, b)
}
func (m *GetAccountBalanceResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetAccountBalanceResponse.Marshal(b, m, deterministic)
}
func (m *GetAccountBalanceResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetAccountBalanceResponse.Merge(m, src)
}
func (m *GetAccountBalanceResponse) XXX_Size() int {
	return xxx_messageInfo_GetAccountBalanceResponse.Size(m)
}
func (m *GetAccountBalanceResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetAccountBalanceResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetAccountBalanceResponse proto.InternalMessageInfo

func (m *GetAccountBalanceResponse) GetAccountBalance() *AccountBalance {
	if m != nil {
		return m.AccountBalance
	}
	return nil
}

type GetAccountBalancesRequest struct {
	// Fetch AccountBalances by type/addresses
	AccountAddressType   uint32   `protobuf:"varint,1,opt,name=AccountAddressType,proto3" json:"AccountAddressType,omitempty"`
	AccountAddress       [][]byte `protobuf:"bytes,2,rep,name=AccountAddress,proto3" json:"AccountAddress,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetAccountBalancesRequest) Reset()         { *m = GetAccountBalancesRequest{} }
func (m *GetAccountBalancesRequest) String() string { return proto.CompactTextString(m) }
func (*GetAccountBalancesRequest) ProtoMessage()    {}
func (*GetAccountBalancesRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_44b9b1c521a5bcaa, []int{3}
}

func (m *GetAccountBalancesRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetAccountBalancesRequest.Unmarshal(m, b)
}
func (m *GetAccountBalancesRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetAccountBalancesRequest.Marshal(b, m, deterministic)
}
func (m *GetAccountBalancesRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetAccountBalancesRequest.Merge(m, src)
}
func (m *GetAccountBalancesRequest) XXX_Size() int {
	return xxx_messageInfo_GetAccountBalancesRequest.Size(m)
}
func (m *GetAccountBalancesRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetAccountBalancesRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetAccountBalancesRequest proto.InternalMessageInfo

func (m *GetAccountBalancesRequest) GetAccountAddressType() uint32 {
	if m != nil {
		return m.AccountAddressType
	}
	return 0
}

func (m *GetAccountBalancesRequest) GetAccountAddress() [][]byte {
	if m != nil {
		return m.AccountAddress
	}
	return nil
}

type GetAccountBalancesResponse struct {
	// Number of accounts returned
	AccountBalanceSize uint32 `protobuf:"varint,1,opt,name=AccountBalanceSize,proto3" json:"AccountBalanceSize,omitempty"`
	// AccountBalances returned
	AccountBalances      []*AccountBalance `protobuf:"bytes,2,rep,name=AccountBalances,proto3" json:"AccountBalances,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *GetAccountBalancesResponse) Reset()         { *m = GetAccountBalancesResponse{} }
func (m *GetAccountBalancesResponse) String() string { return proto.CompactTextString(m) }
func (*GetAccountBalancesResponse) ProtoMessage()    {}
func (*GetAccountBalancesResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_44b9b1c521a5bcaa, []int{4}
}

func (m *GetAccountBalancesResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetAccountBalancesResponse.Unmarshal(m, b)
}
func (m *GetAccountBalancesResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetAccountBalancesResponse.Marshal(b, m, deterministic)
}
func (m *GetAccountBalancesResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetAccountBalancesResponse.Merge(m, src)
}
func (m *GetAccountBalancesResponse) XXX_Size() int {
	return xxx_messageInfo_GetAccountBalancesResponse.Size(m)
}
func (m *GetAccountBalancesResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetAccountBalancesResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetAccountBalancesResponse proto.InternalMessageInfo

func (m *GetAccountBalancesResponse) GetAccountBalanceSize() uint32 {
	if m != nil {
		return m.AccountBalanceSize
	}
	return 0
}

func (m *GetAccountBalancesResponse) GetAccountBalances() []*AccountBalance {
	if m != nil {
		return m.AccountBalances
	}
	return nil
}

func init() {
	proto.RegisterType((*AccountBalance)(nil), "model.AccountBalance")
	proto.RegisterType((*GetAccountBalanceRequest)(nil), "model.GetAccountBalanceRequest")
	proto.RegisterType((*GetAccountBalanceResponse)(nil), "model.GetAccountBalanceResponse")
	proto.RegisterType((*GetAccountBalancesRequest)(nil), "model.GetAccountBalancesRequest")
	proto.RegisterType((*GetAccountBalancesResponse)(nil), "model.GetAccountBalancesResponse")
}

func init() {
	proto.RegisterFile("model/accountBalance.proto", fileDescriptor_44b9b1c521a5bcaa)
}

var fileDescriptor_44b9b1c521a5bcaa = []byte{
	// 329 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x53, 0x4f, 0x4b, 0xfb, 0x40,
	0x14, 0x64, 0xdb, 0x5f, 0xdb, 0x1f, 0xaf, 0xfe, 0x63, 0x41, 0x59, 0x8b, 0x87, 0x90, 0x83, 0x04,
	0xc1, 0x44, 0xf4, 0x2c, 0xd2, 0x5c, 0xf4, 0xe0, 0x41, 0x52, 0x4f, 0xbd, 0x25, 0x9b, 0x47, 0x5b,
	0x4c, 0xf2, 0x62, 0x76, 0x23, 0xd8, 0xef, 0xe0, 0xcd, 0x0f, 0x2c, 0xac, 0x69, 0x6d, 0xd7, 0x78,
	0xd3, 0x4b, 0x20, 0x33, 0xb3, 0x33, 0xfb, 0xe6, 0xb1, 0x30, 0xca, 0x29, 0xc5, 0x2c, 0x88, 0xa5,
	0xa4, 0xba, 0xd0, 0x61, 0x9c, 0xc5, 0x85, 0x44, 0xbf, 0xac, 0x48, 0x13, 0xef, 0x19, 0xce, 0x7d,
	0xef, 0xc0, 0xde, 0x78, 0x8b, 0xe7, 0x3e, 0xf0, 0x06, 0x19, 0xa7, 0x69, 0x85, 0x4a, 0x3d, 0xbe,
	0x96, 0x28, 0x98, 0xc3, 0xbc, 0xdd, 0xa8, 0x85, 0xe1, 0xa7, 0x6b, 0x87, 0x06, 0x15, 0x1d, 0x87,
	0x79, 0x3b, 0x91, 0x85, 0x72, 0x07, 0x86, 0x61, 0x46, 0xf2, 0xe9, 0x0e, 0x17, 0xb3, 0xb9, 0x16,
	0x5d, 0x63, 0xb8, 0x09, 0x71, 0x1f, 0x0e, 0x26, 0x25, 0x16, 0x69, 0x9c, 0x64, 0xd8, 0xdc, 0x46,
	0xfc, 0x73, 0x98, 0xd7, 0x0d, 0x3b, 0x17, 0x2c, 0xfa, 0xc6, 0xf1, 0x13, 0x18, 0xac, 0x64, 0xbd,
	0xb5, 0x6c, 0x05, 0x71, 0x17, 0xe0, 0x81, 0xca, 0x08, 0x5f, 0xb0, 0xa8, 0x51, 0xf4, 0xd7, 0x82,
	0x0d, 0x94, 0x1f, 0x41, 0xff, 0x3e, 0xd6, 0xa8, 0xb4, 0x18, 0x38, 0xcc, 0xfb, 0x1f, 0x35, 0x7f,
	0x6e, 0x05, 0xe2, 0x16, 0xf5, 0x76, 0x31, 0x11, 0x3e, 0xd7, 0xa8, 0xf4, 0x5f, 0xf5, 0xe3, 0x4e,
	0xe1, 0xb8, 0x25, 0x53, 0x95, 0x54, 0x28, 0xe4, 0xd7, 0xf6, 0x9a, 0x4c, 0xe0, 0xf0, 0xf2, 0xd0,
	0x37, 0x7b, 0xf4, 0xad, 0x63, 0x96, 0xd8, 0x55, 0x2d, 0xde, 0xea, 0x37, 0x07, 0xea, 0xb6, 0x0c,
	0xf4, 0xc6, 0x60, 0xd4, 0x96, 0xda, 0x8c, 0xf4, 0x15, 0xdb, 0x50, 0x93, 0xc5, 0xd2, 0x8e, 0xdd,
	0x60, 0xf8, 0x0d, 0xec, 0x5b, 0x56, 0x26, 0xf7, 0xc7, 0x0e, 0x6c, 0x75, 0x78, 0x36, 0xf5, 0x66,
	0x0b, 0x3d, 0xaf, 0x13, 0x5f, 0x52, 0x1e, 0x2c, 0x89, 0x12, 0xf9, 0xf9, 0x3d, 0x97, 0x54, 0x61,
	0x20, 0x29, 0xcf, 0xa9, 0x08, 0x8c, 0x57, 0xd2, 0x37, 0xaf, 0xe4, 0xea, 0x23, 0x00, 0x00, 0xff,
	0xff, 0x25, 0x28, 0x85, 0x53, 0x43, 0x03, 0x00, 0x00,
}
