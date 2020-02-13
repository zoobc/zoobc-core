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

// PostEscrowApprovalRequest message for approval transaction escrow
type PostEscrowApprovalRequest struct {
	// ApprovalBytes format: [uint32 EscrowApproval][int64 TransactionID]
	ApprovalBytes        []byte   `protobuf:"bytes,1,opt,name=ApprovalBytes,proto3" json:"ApprovalBytes,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PostEscrowApprovalRequest) Reset()         { *m = PostEscrowApprovalRequest{} }
func (m *PostEscrowApprovalRequest) String() string { return proto.CompactTextString(m) }
func (*PostEscrowApprovalRequest) ProtoMessage()    {}
func (*PostEscrowApprovalRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_c4ffdfca00fa52ba, []int{1}
}

func (m *PostEscrowApprovalRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PostEscrowApprovalRequest.Unmarshal(m, b)
}
func (m *PostEscrowApprovalRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PostEscrowApprovalRequest.Marshal(b, m, deterministic)
}
func (m *PostEscrowApprovalRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PostEscrowApprovalRequest.Merge(m, src)
}
func (m *PostEscrowApprovalRequest) XXX_Size() int {
	return xxx_messageInfo_PostEscrowApprovalRequest.Size(m)
}
func (m *PostEscrowApprovalRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PostEscrowApprovalRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PostEscrowApprovalRequest proto.InternalMessageInfo

func (m *PostEscrowApprovalRequest) GetApprovalBytes() []byte {
	if m != nil {
		return m.ApprovalBytes
	}
	return nil
}

func init() {
	proto.RegisterEnum("model.EscrowStatus", EscrowStatus_name, EscrowStatus_value)
	proto.RegisterEnum("model.EscrowApproval", EscrowApproval_name, EscrowApproval_value)
	proto.RegisterType((*Escrow)(nil), "model.Escrow")
	proto.RegisterType((*PostEscrowApprovalRequest)(nil), "model.PostEscrowApprovalRequest")
}

func init() { proto.RegisterFile("model/escrow.proto", fileDescriptor_c4ffdfca00fa52ba) }

var fileDescriptor_c4ffdfca00fa52ba = []byte{
	// 397 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x92, 0x4f, 0x6f, 0xd3, 0x40,
	0x10, 0xc5, 0xbb, 0x4e, 0xeb, 0xa4, 0x93, 0xa4, 0x58, 0x83, 0x84, 0x0c, 0xe2, 0x60, 0x45, 0x3d,
	0x98, 0x20, 0x12, 0x04, 0x9f, 0x20, 0x21, 0x95, 0x88, 0xc4, 0xa1, 0xda, 0x72, 0xe2, 0x96, 0xac,
	0x47, 0xe9, 0x42, 0xbc, 0x63, 0x76, 0xd7, 0xfc, 0xfb, 0xf2, 0x20, 0xaf, 0x9d, 0x92, 0xc0, 0xc5,
	0xd2, 0xfc, 0xe6, 0x3d, 0xfb, 0xf9, 0xed, 0x02, 0x96, 0x5c, 0xd0, 0x7e, 0x4e, 0x4e, 0x59, 0xfe,
	0x3e, 0xab, 0x2c, 0x7b, 0xc6, 0x8b, 0xc0, 0x26, 0xbf, 0x23, 0x88, 0x6f, 0x02, 0x47, 0x84, 0x68,
	0xbd, 0x4a, 0x45, 0x26, 0xf2, 0xde, 0x32, 0x7a, 0x2d, 0x64, 0xb4, 0x5e, 0xe1, 0x35, 0x8c, 0xef,
	0xc8, 0x14, 0x64, 0x17, 0x45, 0x61, 0xc9, 0xb9, 0x34, 0xca, 0x44, 0x7e, 0x29, 0x4f, 0x21, 0x4e,
	0x21, 0x91, 0xa4, 0x74, 0xa5, 0xc9, 0xf8, 0x83, 0xb0, 0x17, 0x84, 0xff, 0x71, 0xcc, 0xe1, 0xd1,
	0xa2, 0xaa, 0x2c, 0x7f, 0xfb, 0xfb, 0xce, 0xf3, 0x20, 0xfd, 0x17, 0xe3, 0x33, 0x88, 0x17, 0x25,
	0xd7, 0xc6, 0xa7, 0x17, 0x0f, 0x99, 0x3a, 0x82, 0x13, 0x80, 0x77, 0x5c, 0x96, 0xda, 0x39, 0xcd,
	0x26, 0x8d, 0x1f, 0xf6, 0x47, 0x14, 0x9f, 0x43, 0xff, 0xa3, 0x2e, 0x89, 0x6b, 0x9f, 0xf6, 0x33,
	0x91, 0x9f, 0x07, 0xc1, 0x01, 0xe1, 0x4b, 0x88, 0xef, 0xfc, 0xc6, 0xd7, 0x2e, 0x1d, 0x64, 0x22,
	0xbf, 0x7a, 0xf3, 0x78, 0x16, 0x0a, 0x99, 0xb5, 0x65, 0xb4, 0x2b, 0xd9, 0x49, 0x30, 0x83, 0xe1,
	0x72, 0xcf, 0xea, 0xcb, 0x7b, 0xd2, 0xbb, 0x7b, 0x9f, 0x5e, 0x66, 0x22, 0x1f, 0xcb, 0x63, 0x84,
	0x4f, 0x20, 0xfe, 0xb0, 0xf1, 0xe4, 0x7c, 0x0a, 0x99, 0xc8, 0x07, 0xb2, 0x9b, 0x1a, 0xe7, 0xda,
	0x38, 0x6f, 0x6b, 0xe5, 0x9b, 0xa4, 0xc3, 0xf0, 0xab, 0xc7, 0x68, 0xb2, 0x80, 0xa7, 0xb7, 0xec,
	0x7c, 0xfb, 0xdd, 0xb6, 0x83, 0xcd, 0x5e, 0xd2, 0xd7, 0xba, 0xb1, 0x5f, 0xc3, 0xf8, 0x80, 0x96,
	0x3f, 0x3d, 0xb9, 0x70, 0x3c, 0x23, 0x79, 0x0a, 0xa7, 0x2b, 0x18, 0x1d, 0xc7, 0xc6, 0x21, 0xf4,
	0x6f, 0xc9, 0x14, 0xda, 0xec, 0x92, 0x33, 0x1c, 0xc1, 0xa0, 0x6b, 0xb6, 0x48, 0x44, 0x33, 0x49,
	0xfa, 0x4c, 0xca, 0x53, 0x91, 0x44, 0x8d, 0xf0, 0xe6, 0x47, 0xa5, 0x2d, 0x15, 0x49, 0x6f, 0xfa,
	0x02, 0xae, 0x4e, 0x43, 0x34, 0xeb, 0xce, 0x9a, 0x9c, 0x21, 0x40, 0xdc, 0x3a, 0x13, 0xb1, 0x9c,
	0x7e, 0xca, 0x77, 0xda, 0xdf, 0xd7, 0xdb, 0x99, 0xe2, 0x72, 0xfe, 0x8b, 0x79, 0xab, 0xda, 0xe7,
	0x2b, 0xc5, 0x96, 0xe6, 0x8a, 0xcb, 0x92, 0xcd, 0x3c, 0x14, 0xba, 0x8d, 0xc3, 0x7d, 0x7b, 0xfb,
	0x27, 0x00, 0x00, 0xff, 0xff, 0xd5, 0x6d, 0x1b, 0x24, 0x85, 0x02, 0x00, 0x00,
}
