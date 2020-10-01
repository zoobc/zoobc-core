// Code generated by protoc-gen-go. DO NOT EDIT.
// source: model/accountType.proto

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

type AccountType int32

const (
	AccountType_ZbcAccountType AccountType = 0
	AccountType_BTCAccountType AccountType = 1
)

var AccountType_name = map[int32]string{
	0: "ZbcAccountType",
	1: "BTCAccountType",
}

var AccountType_value = map[string]int32{
	"ZbcAccountType": 0,
	"BTCAccountType": 1,
}

func (x AccountType) String() string {
	return proto.EnumName(AccountType_name, int32(x))
}

func (AccountType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_7d04e165ba99c2b5, []int{0}
}

// AccountType represent transcoding table for account addresses
// It transcode a full account address (bytes) into its components and keeps a reference to the encoded (string) representation of the account
type AccountAddress struct {
	AccountAddress       []byte   `protobuf:"bytes,1,opt,name=AccountAddress,proto3" json:"AccountAddress,omitempty"`
	AccountType          int32    `protobuf:"varint,2,opt,name=AccountType,proto3" json:"AccountType,omitempty"`
	AccountPublicKey     []byte   `protobuf:"bytes,3,opt,name=AccountPublicKey,proto3" json:"AccountPublicKey,omitempty"`
	EncodedAccount       string   `protobuf:"bytes,4,opt,name=EncodedAccount,proto3" json:"EncodedAccount,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AccountAddress) Reset()         { *m = AccountAddress{} }
func (m *AccountAddress) String() string { return proto.CompactTextString(m) }
func (*AccountAddress) ProtoMessage()    {}
func (*AccountAddress) Descriptor() ([]byte, []int) {
	return fileDescriptor_7d04e165ba99c2b5, []int{0}
}

func (m *AccountAddress) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AccountAddress.Unmarshal(m, b)
}
func (m *AccountAddress) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AccountAddress.Marshal(b, m, deterministic)
}
func (m *AccountAddress) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AccountAddress.Merge(m, src)
}
func (m *AccountAddress) XXX_Size() int {
	return xxx_messageInfo_AccountAddress.Size(m)
}
func (m *AccountAddress) XXX_DiscardUnknown() {
	xxx_messageInfo_AccountAddress.DiscardUnknown(m)
}

var xxx_messageInfo_AccountAddress proto.InternalMessageInfo

func (m *AccountAddress) GetAccountAddress() []byte {
	if m != nil {
		return m.AccountAddress
	}
	return nil
}

func (m *AccountAddress) GetAccountType() int32 {
	if m != nil {
		return m.AccountType
	}
	return 0
}

func (m *AccountAddress) GetAccountPublicKey() []byte {
	if m != nil {
		return m.AccountPublicKey
	}
	return nil
}

func (m *AccountAddress) GetEncodedAccount() string {
	if m != nil {
		return m.EncodedAccount
	}
	return ""
}

func init() {
	proto.RegisterEnum("model.AccountType", AccountType_name, AccountType_value)
	proto.RegisterType((*AccountAddress)(nil), "model.AccountAddress")
}

func init() {
	proto.RegisterFile("model/accountType.proto", fileDescriptor_7d04e165ba99c2b5)
}

var fileDescriptor_7d04e165ba99c2b5 = []byte{
	// 203 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0xcf, 0xcd, 0x4f, 0x49,
	0xcd, 0xd1, 0x4f, 0x4c, 0x4e, 0xce, 0x2f, 0xcd, 0x2b, 0x09, 0xa9, 0x2c, 0x48, 0xd5, 0x2b, 0x28,
	0xca, 0x2f, 0xc9, 0x17, 0x62, 0x05, 0x4b, 0x28, 0xad, 0x63, 0xe4, 0xe2, 0x73, 0x84, 0x48, 0x3a,
	0xa6, 0xa4, 0x14, 0xa5, 0x16, 0x17, 0x0b, 0xa9, 0xa1, 0x8b, 0x48, 0x30, 0x2a, 0x30, 0x6a, 0xf0,
	0x04, 0xa1, 0xab, 0x53, 0xe0, 0xe2, 0x76, 0x44, 0x18, 0x2b, 0xc1, 0xa4, 0xc0, 0xa8, 0xc1, 0x1a,
	0x84, 0x2c, 0x24, 0xa4, 0xc5, 0x25, 0x00, 0xe5, 0x06, 0x94, 0x26, 0xe5, 0x64, 0x26, 0x7b, 0xa7,
	0x56, 0x4a, 0x30, 0x83, 0xcd, 0xc2, 0x10, 0x07, 0xd9, 0xea, 0x9a, 0x97, 0x9c, 0x9f, 0x92, 0x9a,
	0x02, 0x95, 0x92, 0x60, 0x51, 0x60, 0xd4, 0xe0, 0x0c, 0x42, 0x13, 0xd5, 0x32, 0x45, 0xb1, 0x55,
	0x48, 0x88, 0x8b, 0x2f, 0x2a, 0x29, 0x19, 0x49, 0x44, 0x80, 0x01, 0x24, 0xe6, 0x14, 0xe2, 0x8c,
	0x2c, 0xc6, 0xe8, 0xa4, 0x15, 0xa5, 0x91, 0x9e, 0x59, 0x92, 0x51, 0x9a, 0xa4, 0x97, 0x9c, 0x9f,
	0xab, 0x5f, 0x95, 0x9f, 0x9f, 0x94, 0x0c, 0x21, 0x75, 0x93, 0xf3, 0x8b, 0x52, 0xf5, 0x93, 0xf3,
	0x73, 0x73, 0xf3, 0xf3, 0xf4, 0xc1, 0x61, 0x92, 0xc4, 0x06, 0x0e, 0x21, 0x63, 0x40, 0x00, 0x00,
	0x00, 0xff, 0xff, 0x42, 0xb4, 0x8f, 0x53, 0x3c, 0x01, 0x00, 0x00,
}
