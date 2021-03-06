// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
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
	AccountAddress       []byte   `protobuf:"bytes,1,opt,name=AccountAddress,proto3" json:"AccountAddress,omitempty"`
	BlockHeight          uint32   `protobuf:"varint,2,opt,name=BlockHeight,proto3" json:"BlockHeight,omitempty"`
	SpendableBalance     int64    `protobuf:"varint,3,opt,name=SpendableBalance,proto3" json:"SpendableBalance,omitempty"`
	Balance              int64    `protobuf:"varint,4,opt,name=Balance,proto3" json:"Balance,omitempty"`
	PopRevenue           int64    `protobuf:"varint,5,opt,name=PopRevenue,proto3" json:"PopRevenue,omitempty"`
	Latest               bool     `protobuf:"varint,6,opt,name=Latest,proto3" json:"Latest,omitempty"`
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
	AccountAddress       []byte   `protobuf:"bytes,1,opt,name=AccountAddress,proto3" json:"AccountAddress,omitempty"`
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
	AccountAddresses     [][]byte `protobuf:"bytes,1,rep,name=AccountAddresses,proto3" json:"AccountAddresses,omitempty"`
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

func (m *GetAccountBalancesRequest) GetAccountAddresses() [][]byte {
	if m != nil {
		return m.AccountAddresses
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
	// 325 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0x4f, 0x4b, 0xc3, 0x40,
	0x10, 0xc5, 0xd9, 0xc6, 0x56, 0x99, 0xb6, 0x5a, 0x16, 0x94, 0xb5, 0x78, 0x08, 0x39, 0x48, 0x28,
	0x98, 0x88, 0x9e, 0x45, 0x9a, 0x4b, 0x3d, 0x78, 0x90, 0xed, 0xad, 0xb7, 0x64, 0x33, 0xb4, 0xc5,
	0x24, 0x13, 0xbb, 0x1b, 0x0f, 0xfd, 0x0e, 0x7e, 0x4b, 0x3f, 0x88, 0x98, 0xfe, 0xa1, 0x49, 0x23,
	0x78, 0x59, 0xd8, 0xf7, 0xde, 0xfe, 0x78, 0x33, 0x2c, 0x0c, 0x53, 0x8a, 0x31, 0xf1, 0x43, 0xa5,
	0xa8, 0xc8, 0x4c, 0x10, 0x26, 0x61, 0xa6, 0xd0, 0xcb, 0x57, 0x64, 0x88, 0xb7, 0x4b, 0xcf, 0xf9,
	0x66, 0x70, 0x3e, 0xae, 0xf8, 0xfc, 0x76, 0xaf, 0x8c, 0xe3, 0x78, 0x85, 0x5a, 0x0b, 0x66, 0x33,
	0xb7, 0x27, 0x6b, 0x2a, 0xb7, 0xa1, 0x1b, 0x24, 0xa4, 0xde, 0x5f, 0x70, 0x39, 0x5f, 0x18, 0xd1,
	0xb2, 0x99, 0xdb, 0x97, 0x87, 0x12, 0xf7, 0x60, 0x30, 0xcd, 0x31, 0x8b, 0xc3, 0x28, 0xc1, 0x2d,
	0x5d, 0x58, 0x36, 0x73, 0xad, 0xa0, 0x75, 0xcf, 0xe4, 0x91, 0xc7, 0x6f, 0xe0, 0x74, 0x17, 0x3b,
	0xd9, 0xc7, 0x76, 0x12, 0x77, 0x00, 0xde, 0x28, 0x97, 0xf8, 0x89, 0x59, 0x81, 0xa2, 0xbd, 0x0f,
	0x1c, 0xa8, 0xfc, 0x0a, 0x3a, 0xaf, 0xa1, 0x41, 0x6d, 0x44, 0xc7, 0x66, 0xee, 0x99, 0xdc, 0xde,
	0x9c, 0x00, 0xc4, 0x04, 0x4d, 0x75, 0x50, 0x89, 0x1f, 0x05, 0x6a, 0xf3, 0xdf, 0x79, 0x9d, 0x19,
	0x5c, 0x37, 0x30, 0x74, 0x4e, 0x99, 0x46, 0xfe, 0x54, 0x5f, 0x63, 0x09, 0xe9, 0x3e, 0x5c, 0x7a,
	0xe5, 0x9e, 0xbd, 0xda, 0xb3, 0x5a, 0xd8, 0x99, 0x34, 0xb0, 0xf5, 0xae, 0xe0, 0x08, 0x06, 0xd5,
	0x2a, 0xf8, 0x5b, 0xd1, 0x72, 0x7b, 0xf2, 0x48, 0x77, 0xbe, 0x18, 0x0c, 0x9b, 0x48, 0xdb, 0x9a,
	0x1e, 0xf0, 0xaa, 0x35, 0x5d, 0xae, 0x37, 0x55, 0xfb, 0xb2, 0xc1, 0xe1, 0xcf, 0x70, 0x51, 0x43,
	0x89, 0x96, 0x6d, 0xfd, 0x3d, 0x57, 0x3d, 0x1d, 0x8c, 0x66, 0xee, 0x7c, 0x69, 0x16, 0x45, 0xe4,
	0x29, 0x4a, 0xfd, 0x35, 0x51, 0xa4, 0x36, 0xe7, 0x9d, 0xa2, 0x15, 0xfa, 0x8a, 0xd2, 0x94, 0x32,
	0xbf, 0x64, 0x45, 0x9d, 0xf2, 0x67, 0x3e, 0xfe, 0x04, 0x00, 0x00, 0xff, 0xff, 0x46, 0x03, 0x0b,
	0x09, 0xb7, 0x02, 0x00, 0x00,
}
