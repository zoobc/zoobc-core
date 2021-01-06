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
	AccountAddress       []byte    `protobuf:"bytes,1,opt,name=AccountAddress,proto3" json:"AccountAddress,omitempty"`
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

func (m *AccountLedger) GetAccountAddress() []byte {
	if m != nil {
		return m.AccountAddress
	}
	return nil
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
	AccountAddress       []byte      `protobuf:"bytes,1,opt,name=AccountAddress,proto3" json:"AccountAddress,omitempty"`
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

func (m *GetAccountLedgersRequest) GetAccountAddress() []byte {
	if m != nil {
		return m.AccountAddress
	}
	return nil
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
	// 390 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0x41, 0xab, 0xd3, 0x40,
	0x14, 0x85, 0x99, 0x24, 0x7d, 0xf0, 0xa6, 0x2f, 0xc5, 0x37, 0x88, 0xcc, 0x2b, 0x2e, 0x42, 0x90,
	0x32, 0x08, 0x26, 0x5a, 0xb7, 0x6e, 0x1a, 0x2d, 0x2a, 0xb8, 0x90, 0x31, 0x2b, 0x77, 0xd3, 0xc9,
	0x25, 0x0d, 0x26, 0x33, 0x31, 0x33, 0x15, 0xf4, 0xaf, 0xf8, 0x43, 0xdd, 0x4a, 0x27, 0x21, 0x6d,
	0xaa, 0xe2, 0xdb, 0x64, 0xf1, 0x9d, 0x93, 0x7b, 0xef, 0x39, 0x09, 0xbe, 0x6b, 0x74, 0x01, 0x75,
	0x2a, 0xa4, 0xd4, 0x07, 0x65, 0x3f, 0x40, 0x51, 0x42, 0x97, 0xb4, 0x9d, 0xb6, 0x9a, 0xcc, 0x9c,
	0xb4, 0xbc, 0xed, 0x1d, 0xf0, 0x0d, 0x94, 0xed, 0x95, 0xe5, 0xa3, 0x1e, 0xb5, 0xa2, 0xac, 0x94,
	0xb0, 0x95, 0x56, 0x3d, 0x8f, 0x7f, 0x21, 0x1c, 0x6e, 0xce, 0x27, 0x91, 0x15, 0x5e, 0x0c, 0x60,
	0x53, 0x14, 0x1d, 0x18, 0x43, 0x51, 0x84, 0xd8, 0x0d, 0xbf, 0xa0, 0xe4, 0x09, 0x0e, 0x33, 0x51,
	0x0b, 0x25, 0xe1, 0xf5, 0x5e, 0xa8, 0x12, 0xa8, 0x17, 0x21, 0xe6, 0xf3, 0x29, 0x24, 0x11, 0x9e,
	0x67, 0xb5, 0x96, 0x5f, 0xde, 0x41, 0x55, 0xee, 0x2d, 0xf5, 0x23, 0xc4, 0x42, 0x7e, 0x8e, 0x08,
	0xc3, 0x61, 0xde, 0x09, 0x65, 0x84, 0x3c, 0x9e, 0xf5, 0xfe, 0x0d, 0x0d, 0x8e, 0x73, 0x32, 0xef,
	0x39, 0xe2, 0x53, 0x81, 0x3c, 0xc6, 0xd7, 0x79, 0xd5, 0x80, 0xb1, 0xa2, 0x69, 0xe9, 0x2c, 0x42,
	0x2c, 0xe0, 0x27, 0x40, 0x12, 0x7c, 0xbd, 0x3d, 0x06, 0xce, 0xbf, 0xb7, 0x40, 0xaf, 0x22, 0xc4,
	0x16, 0xeb, 0x07, 0x89, 0x4b, 0x9d, 0x8c, 0x9c, 0x9f, 0x2c, 0xf1, 0x4f, 0x0f, 0xd3, 0xb7, 0x60,
	0x27, 0xe1, 0x0d, 0x87, 0xaf, 0x07, 0x30, 0xf6, 0xde, 0x25, 0x4c, 0x96, 0x7a, 0xff, 0x5d, 0xfa,
	0x67, 0x58, 0xff, 0x5f, 0x61, 0x57, 0x78, 0x31, 0x66, 0xfb, 0x64, 0x45, 0x67, 0x5d, 0x2f, 0x01,
	0xbf, 0xa0, 0x24, 0xc6, 0x37, 0x23, 0xd9, 0xaa, 0xc2, 0xf5, 0x12, 0xf2, 0x09, 0x23, 0x2f, 0x30,
	0xfe, 0x38, 0x7e, 0x78, 0xd7, 0xcd, 0x7c, 0x7d, 0x3b, 0x9c, 0x79, 0x12, 0xf8, 0x99, 0x29, 0x36,
	0xf8, 0xee, 0x2f, 0xe5, 0x98, 0x56, 0x2b, 0x03, 0x84, 0xe2, 0x59, 0xae, 0xad, 0xa8, 0x5d, 0x29,
	0x81, 0xbb, 0xbe, 0x07, 0xe4, 0xd5, 0xd8, 0xdb, 0xf0, 0x0e, 0xf5, 0x22, 0x9f, 0xcd, 0xd7, 0x0f,
	0x87, 0x6d, 0x13, 0x91, 0x5f, 0x78, 0xb3, 0xa7, 0x9f, 0x59, 0x59, 0xd9, 0xfd, 0x61, 0x97, 0x48,
	0xdd, 0xa4, 0x3f, 0xb4, 0xde, 0xc9, 0xfe, 0xf9, 0x4c, 0xea, 0x0e, 0x52, 0xa9, 0x9b, 0x46, 0xab,
	0xd4, 0x4d, 0xda, 0x5d, 0xb9, 0xff, 0xf7, 0xe5, 0xef, 0x00, 0x00, 0x00, 0xff, 0xff, 0x10, 0x18,
	0x4a, 0x38, 0x0e, 0x03, 0x00, 0x00,
}
