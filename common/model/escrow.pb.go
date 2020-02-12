// Code generated by protoc-gen-go. DO NOT EDIT.
// source: model/escrow.proto

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

type EscrowStatus int32

const (
	EscrowStatus_Pending  EscrowStatus = 0
	EscrowStatus_Approved EscrowStatus = 1
	EscrowStatus_Rejected EscrowStatus = 2
	EscrowStatus_Expired  EscrowStatus = 3
)

var EscrowStatus_name = map[int32]string{
	0: "Pending",
	1: "Approved",
	2: "Rejected",
	3: "Expired",
}

var EscrowStatus_value = map[string]int32{
	"Pending":  0,
	"Approved": 1,
	"Rejected": 2,
	"Expired":  3,
}

func (x EscrowStatus) String() string {
	return proto.EnumName(EscrowStatus_name, int32(x))
}

func (EscrowStatus) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_c4ffdfca00fa52ba, []int{0}
}

type EscrowApproval int32

const (
	EscrowApproval_Approve EscrowApproval = 0
	EscrowApproval_Reject  EscrowApproval = 1
)

var EscrowApproval_name = map[int32]string{
	0: "Approve",
	1: "Reject",
}

var EscrowApproval_value = map[string]int32{
	"Approve": 0,
	"Reject":  1,
}

func (x EscrowApproval) String() string {
	return proto.EnumName(EscrowApproval_name, int32(x))
}

func (EscrowApproval) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_c4ffdfca00fa52ba, []int{1}
}

type Escrow struct {
	ID               int64  `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
	SenderAddress    string `protobuf:"bytes,2,opt,name=SenderAddress,proto3" json:"SenderAddress,omitempty"`
	RecipientAddress string `protobuf:"bytes,3,opt,name=RecipientAddress,proto3" json:"RecipientAddress,omitempty"`
	ApproverAddress  string `protobuf:"bytes,4,opt,name=ApproverAddress,proto3" json:"ApproverAddress,omitempty"`
	Amount           int64  `protobuf:"varint,5,opt,name=Amount,proto3" json:"Amount,omitempty"`
	Commission       int64  `protobuf:"varint,6,opt,name=Commission,proto3" json:"Commission,omitempty"`
	// Timeout is BlockHeight gap
	Timeout              uint64       `protobuf:"varint,7,opt,name=Timeout,proto3" json:"Timeout,omitempty"`
	Status               EscrowStatus `protobuf:"varint,8,opt,name=Status,proto3,enum=model.EscrowStatus" json:"Status,omitempty"`
	BlockHeight          uint32       `protobuf:"varint,9,opt,name=BlockHeight,proto3" json:"BlockHeight,omitempty"`
	Latest               bool         `protobuf:"varint,10,opt,name=Latest,proto3" json:"Latest,omitempty"`
	Instruction          string       `protobuf:"bytes,11,opt,name=Instruction,proto3" json:"Instruction,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *Escrow) Reset()         { *m = Escrow{} }
func (m *Escrow) String() string { return proto.CompactTextString(m) }
func (*Escrow) ProtoMessage()    {}
func (*Escrow) Descriptor() ([]byte, []int) {
	return fileDescriptor_c4ffdfca00fa52ba, []int{0}
}

func (m *Escrow) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Escrow.Unmarshal(m, b)
}
func (m *Escrow) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Escrow.Marshal(b, m, deterministic)
}
func (m *Escrow) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Escrow.Merge(m, src)
}
func (m *Escrow) XXX_Size() int {
	return xxx_messageInfo_Escrow.Size(m)
}
func (m *Escrow) XXX_DiscardUnknown() {
	xxx_messageInfo_Escrow.DiscardUnknown(m)
}

var xxx_messageInfo_Escrow proto.InternalMessageInfo

func (m *Escrow) GetID() int64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *Escrow) GetSenderAddress() string {
	if m != nil {
		return m.SenderAddress
	}
	return ""
}

func (m *Escrow) GetRecipientAddress() string {
	if m != nil {
		return m.RecipientAddress
	}
	return ""
}

func (m *Escrow) GetApproverAddress() string {
	if m != nil {
		return m.ApproverAddress
	}
	return ""
}

func (m *Escrow) GetAmount() int64 {
	if m != nil {
		return m.Amount
	}
	return 0
}

func (m *Escrow) GetCommission() int64 {
	if m != nil {
		return m.Commission
	}
	return 0
}

func (m *Escrow) GetTimeout() uint64 {
	if m != nil {
		return m.Timeout
	}
	return 0
}

func (m *Escrow) GetStatus() EscrowStatus {
	if m != nil {
		return m.Status
	}
	return EscrowStatus_Pending
}

func (m *Escrow) GetBlockHeight() uint32 {
	if m != nil {
		return m.BlockHeight
	}
	return 0
}

func (m *Escrow) GetLatest() bool {
	if m != nil {
		return m.Latest
	}
	return false
}

func (m *Escrow) GetInstruction() string {
	if m != nil {
		return m.Instruction
	}
	return ""
}

// GetEscrowTransactionsRequest message for get escrow transactions
type GetEscrowTransactionsRequest struct {
	ApproverAddress      string         `protobuf:"bytes,1,opt,name=ApproverAddress,proto3" json:"ApproverAddress,omitempty"`
	Statuses             []EscrowStatus `protobuf:"varint,2,rep,packed,name=Statuses,proto3,enum=model.EscrowStatus" json:"Statuses,omitempty"`
	BlockHeightStart     uint32         `protobuf:"varint,3,opt,name=BlockHeightStart,proto3" json:"BlockHeightStart,omitempty"`
	BlockHeightEnd       uint32         `protobuf:"varint,4,opt,name=BlockHeightEnd,proto3" json:"BlockHeightEnd,omitempty"`
	Pagination           *Pagination    `protobuf:"bytes,5,opt,name=Pagination,proto3" json:"Pagination,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *GetEscrowTransactionsRequest) Reset()         { *m = GetEscrowTransactionsRequest{} }
func (m *GetEscrowTransactionsRequest) String() string { return proto.CompactTextString(m) }
func (*GetEscrowTransactionsRequest) ProtoMessage()    {}
func (*GetEscrowTransactionsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_c4ffdfca00fa52ba, []int{1}
}

func (m *GetEscrowTransactionsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetEscrowTransactionsRequest.Unmarshal(m, b)
}
func (m *GetEscrowTransactionsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetEscrowTransactionsRequest.Marshal(b, m, deterministic)
}
func (m *GetEscrowTransactionsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetEscrowTransactionsRequest.Merge(m, src)
}
func (m *GetEscrowTransactionsRequest) XXX_Size() int {
	return xxx_messageInfo_GetEscrowTransactionsRequest.Size(m)
}
func (m *GetEscrowTransactionsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetEscrowTransactionsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetEscrowTransactionsRequest proto.InternalMessageInfo

func (m *GetEscrowTransactionsRequest) GetApproverAddress() string {
	if m != nil {
		return m.ApproverAddress
	}
	return ""
}

func (m *GetEscrowTransactionsRequest) GetStatuses() []EscrowStatus {
	if m != nil {
		return m.Statuses
	}
	return nil
}

func (m *GetEscrowTransactionsRequest) GetBlockHeightStart() uint32 {
	if m != nil {
		return m.BlockHeightStart
	}
	return 0
}

func (m *GetEscrowTransactionsRequest) GetBlockHeightEnd() uint32 {
	if m != nil {
		return m.BlockHeightEnd
	}
	return 0
}

func (m *GetEscrowTransactionsRequest) GetPagination() *Pagination {
	if m != nil {
		return m.Pagination
	}
	return nil
}

// GetEscrowTransactionsResponse returns fields of GetEscrowTransactionsRequest
type GetEscrowTransactionsResponse struct {
	// Number of transactions in total
	Total uint64 `protobuf:"varint,1,opt,name=Total,proto3" json:"Total,omitempty"`
	// Transaction transactions returned
	Escrows              []*Escrow `protobuf:"bytes,2,rep,name=Escrows,proto3" json:"Escrows,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *GetEscrowTransactionsResponse) Reset()         { *m = GetEscrowTransactionsResponse{} }
func (m *GetEscrowTransactionsResponse) String() string { return proto.CompactTextString(m) }
func (*GetEscrowTransactionsResponse) ProtoMessage()    {}
func (*GetEscrowTransactionsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_c4ffdfca00fa52ba, []int{2}
}

func (m *GetEscrowTransactionsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetEscrowTransactionsResponse.Unmarshal(m, b)
}
func (m *GetEscrowTransactionsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetEscrowTransactionsResponse.Marshal(b, m, deterministic)
}
func (m *GetEscrowTransactionsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetEscrowTransactionsResponse.Merge(m, src)
}
func (m *GetEscrowTransactionsResponse) XXX_Size() int {
	return xxx_messageInfo_GetEscrowTransactionsResponse.Size(m)
}
func (m *GetEscrowTransactionsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetEscrowTransactionsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetEscrowTransactionsResponse proto.InternalMessageInfo

func (m *GetEscrowTransactionsResponse) GetTotal() uint64 {
	if m != nil {
		return m.Total
	}
	return 0
}

func (m *GetEscrowTransactionsResponse) GetEscrows() []*Escrow {
	if m != nil {
		return m.Escrows
	}
	return nil
}

func init() {
	proto.RegisterEnum("model.EscrowStatus", EscrowStatus_name, EscrowStatus_value)
	proto.RegisterEnum("model.EscrowApproval", EscrowApproval_name, EscrowApproval_value)
	proto.RegisterType((*Escrow)(nil), "model.Escrow")
	proto.RegisterType((*GetEscrowTransactionsRequest)(nil), "model.GetEscrowTransactionsRequest")
	proto.RegisterType((*GetEscrowTransactionsResponse)(nil), "model.GetEscrowTransactionsResponse")
}

func init() { proto.RegisterFile("model/escrow.proto", fileDescriptor_c4ffdfca00fa52ba) }

var fileDescriptor_c4ffdfca00fa52ba = []byte{
	// 503 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x93, 0x5f, 0x6f, 0xd3, 0x30,
	0x14, 0xc5, 0xe7, 0x74, 0x4b, 0xbb, 0xdb, 0xb5, 0x04, 0x23, 0x4d, 0xd1, 0x34, 0xa4, 0xa8, 0x42,
	0x10, 0x8a, 0x68, 0xa1, 0x7c, 0x82, 0x96, 0x56, 0x50, 0x89, 0x87, 0xc9, 0xed, 0x13, 0x6f, 0x69,
	0x72, 0xd5, 0x19, 0x1a, 0x3b, 0xd8, 0x0e, 0x20, 0x9e, 0xf8, 0xe6, 0xa0, 0xd8, 0x69, 0x97, 0xfd,
	0x7b, 0x89, 0xe4, 0xdf, 0x39, 0xd7, 0xb9, 0xf7, 0xdc, 0x04, 0x68, 0x2e, 0x33, 0xdc, 0x8d, 0x51,
	0xa7, 0x4a, 0xfe, 0x1a, 0x15, 0x4a, 0x1a, 0x49, 0x4f, 0x2c, 0xbb, 0x38, 0x77, 0x52, 0x91, 0x6c,
	0xb9, 0x48, 0x0c, 0x97, 0xc2, 0xc9, 0x83, 0x7f, 0x1e, 0xf8, 0x0b, 0xeb, 0xa7, 0x14, 0xbc, 0xe5,
	0x3c, 0x24, 0x11, 0x89, 0x5b, 0x33, 0xef, 0x1d, 0x61, 0xde, 0x72, 0x4e, 0x5f, 0x40, 0x6f, 0x85,
	0x22, 0x43, 0x35, 0xcd, 0x32, 0x85, 0x5a, 0x87, 0x5e, 0x44, 0xe2, 0x53, 0x76, 0x1b, 0xd2, 0x21,
	0x04, 0x0c, 0x53, 0x5e, 0x70, 0x14, 0x66, 0x6f, 0x6c, 0x59, 0xe3, 0x3d, 0x4e, 0x63, 0x78, 0x32,
	0x2d, 0x0a, 0x25, 0x7f, 0xde, 0xdc, 0x79, 0x6c, 0xad, 0x77, 0x31, 0xbd, 0x00, 0x7f, 0x9a, 0xcb,
	0x52, 0x98, 0xf0, 0xe4, 0xd0, 0x53, 0x4d, 0xe8, 0x00, 0xe0, 0xa3, 0xcc, 0x73, 0xae, 0x35, 0x97,
	0x22, 0xf4, 0x0f, 0x7a, 0x83, 0xd2, 0x4b, 0x68, 0xaf, 0x79, 0x8e, 0xb2, 0x34, 0x61, 0x3b, 0x22,
	0xf1, 0xb1, 0x35, 0xec, 0x11, 0x7d, 0x03, 0xfe, 0xca, 0x24, 0xa6, 0xd4, 0x61, 0x27, 0x22, 0x71,
	0x7f, 0xf2, 0x6c, 0x64, 0x13, 0x1a, 0xb9, 0x30, 0x9c, 0xc4, 0x6a, 0x0b, 0x8d, 0xa0, 0x3b, 0xdb,
	0xc9, 0xf4, 0xfb, 0x67, 0xe4, 0xdb, 0x6b, 0x13, 0x9e, 0x46, 0x24, 0xee, 0xb1, 0x26, 0xa2, 0xe7,
	0xe0, 0x7f, 0x49, 0x0c, 0x6a, 0x13, 0x42, 0x44, 0xe2, 0x0e, 0xab, 0x4f, 0x55, 0xe5, 0x52, 0x68,
	0xa3, 0xca, 0xb4, 0x0a, 0x3d, 0xec, 0xda, 0x51, 0x9b, 0x68, 0xf0, 0xd7, 0x83, 0xcb, 0x4f, 0x68,
	0xdc, 0x7b, 0xd7, 0x2a, 0x11, 0x3a, 0xb1, 0x82, 0x66, 0xf8, 0xa3, 0xac, 0xae, 0x78, 0x20, 0x31,
	0xf2, 0x70, 0x62, 0x63, 0xe8, 0xb8, 0x86, 0xb1, 0x5a, 0x54, 0xeb, 0xb1, 0xa9, 0x0e, 0xa6, 0x6a,
	0x71, 0x8d, 0x21, 0x56, 0x26, 0x51, 0xc6, 0x2e, 0xae, 0xc7, 0xee, 0x71, 0xfa, 0x12, 0xfa, 0x0d,
	0xb6, 0x10, 0x99, 0xdd, 0x5b, 0x8f, 0xdd, 0xa1, 0xf4, 0x3d, 0xc0, 0xd5, 0xe1, 0x2b, 0xb3, 0xab,
	0xeb, 0x4e, 0x9e, 0xd6, 0x6d, 0xdc, 0x08, 0xac, 0x61, 0x1a, 0x6c, 0xe0, 0xf9, 0x23, 0x09, 0xe8,
	0x42, 0x0a, 0x8d, 0x34, 0x84, 0x93, 0xb5, 0x34, 0xc9, 0xce, 0x0e, 0xee, 0x16, 0xe9, 0x00, 0x7d,
	0x05, 0x6d, 0x57, 0xe7, 0x26, 0xee, 0x4e, 0x7a, 0xb7, 0x26, 0x66, 0x7b, 0x75, 0x38, 0x87, 0xb3,
	0x66, 0x08, 0xb4, 0x0b, 0xed, 0x2b, 0x14, 0x19, 0x17, 0xdb, 0xe0, 0x88, 0x9e, 0x41, 0xa7, 0xce,
	0x32, 0x0b, 0x48, 0x75, 0x62, 0xf8, 0x0d, 0x53, 0x83, 0x59, 0xe0, 0x55, 0xc6, 0xc5, 0xef, 0x82,
	0x2b, 0xcc, 0x82, 0xd6, 0xf0, 0x35, 0xf4, 0xdd, 0x2d, 0xce, 0x9e, 0xec, 0x2a, 0xb9, 0x2e, 0x0d,
	0x8e, 0x28, 0x80, 0xef, 0x2a, 0x03, 0x32, 0x1b, 0x7e, 0x8d, 0xb7, 0xdc, 0x5c, 0x97, 0x9b, 0x51,
	0x2a, 0xf3, 0xf1, 0x1f, 0x29, 0x37, 0xa9, 0x7b, 0xbe, 0x4d, 0xa5, 0xc2, 0x71, 0x2a, 0xf3, 0x5c,
	0x8a, 0xb1, 0x6d, 0x76, 0xe3, 0xdb, 0x9f, 0xf1, 0xc3, 0xff, 0x00, 0x00, 0x00, 0xff, 0xff, 0xdf,
	0xa7, 0x56, 0x67, 0xc1, 0x03, 0x00, 0x00,
}
