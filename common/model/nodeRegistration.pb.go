// Code generated by protoc-gen-go. DO NOT EDIT.
// source: model/nodeRegistration.proto

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

type NodeRegistrationState int32

const (
	// NodeRegistered 'registred' node status (= 0): a node in node registry with this status is registered
	NodeRegistrationState_NodeRegistered NodeRegistrationState = 0
	// NodeQueued 'queued' node status (= 1): a node in node registry with this status is queued, or 'pending registered'
	NodeRegistrationState_NodeQueued NodeRegistrationState = 1
	// NodeDeleted 'deleted' node status (= 2): a node in node registry with this status is marked as deleted
	NodeRegistrationState_NodeDeleted NodeRegistrationState = 2
)

var NodeRegistrationState_name = map[int32]string{
	0: "NodeRegistered",
	1: "NodeQueued",
	2: "NodeDeleted",
}

var NodeRegistrationState_value = map[string]int32{
	"NodeRegistered": 0,
	"NodeQueued":     1,
	"NodeDeleted":    2,
}

func (x NodeRegistrationState) String() string {
	return proto.EnumName(NodeRegistrationState_name, int32(x))
}

func (NodeRegistrationState) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_df1af0ec89e31788, []int{0}
}

type NodeRegistration struct {
	NodeID               int64            `protobuf:"varint,1,opt,name=NodeID,proto3" json:"NodeID,omitempty"`
	NodePublicKey        []byte           `protobuf:"bytes,2,opt,name=NodePublicKey,proto3" json:"NodePublicKey,omitempty"`
	AccountAddressType   uint32           `protobuf:"varint,3,opt,name=AccountAddressType,proto3" json:"AccountAddressType,omitempty"`
	AccountAddress       []byte           `protobuf:"bytes,4,opt,name=AccountAddress,proto3" json:"AccountAddress,omitempty"`
	RegistrationHeight   uint32           `protobuf:"varint,5,opt,name=RegistrationHeight,proto3" json:"RegistrationHeight,omitempty"`
	LockedBalance        int64            `protobuf:"varint,6,opt,name=LockedBalance,proto3" json:"LockedBalance,omitempty"`
	RegistrationStatus   uint32           `protobuf:"varint,7,opt,name=RegistrationStatus,proto3" json:"RegistrationStatus,omitempty"`
	Latest               bool             `protobuf:"varint,8,opt,name=Latest,proto3" json:"Latest,omitempty"`
	Height               uint32           `protobuf:"varint,9,opt,name=Height,proto3" json:"Height,omitempty"`
	NodeAddressInfo      *NodeAddressInfo `protobuf:"bytes,10,opt,name=NodeAddressInfo,proto3" json:"NodeAddressInfo,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *NodeRegistration) Reset()         { *m = NodeRegistration{} }
func (m *NodeRegistration) String() string { return proto.CompactTextString(m) }
func (*NodeRegistration) ProtoMessage()    {}
func (*NodeRegistration) Descriptor() ([]byte, []int) {
	return fileDescriptor_df1af0ec89e31788, []int{0}
}

func (m *NodeRegistration) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NodeRegistration.Unmarshal(m, b)
}
func (m *NodeRegistration) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NodeRegistration.Marshal(b, m, deterministic)
}
func (m *NodeRegistration) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NodeRegistration.Merge(m, src)
}
func (m *NodeRegistration) XXX_Size() int {
	return xxx_messageInfo_NodeRegistration.Size(m)
}
func (m *NodeRegistration) XXX_DiscardUnknown() {
	xxx_messageInfo_NodeRegistration.DiscardUnknown(m)
}

var xxx_messageInfo_NodeRegistration proto.InternalMessageInfo

func (m *NodeRegistration) GetNodeID() int64 {
	if m != nil {
		return m.NodeID
	}
	return 0
}

func (m *NodeRegistration) GetNodePublicKey() []byte {
	if m != nil {
		return m.NodePublicKey
	}
	return nil
}

func (m *NodeRegistration) GetAccountAddressType() uint32 {
	if m != nil {
		return m.AccountAddressType
	}
	return 0
}

func (m *NodeRegistration) GetAccountAddress() []byte {
	if m != nil {
		return m.AccountAddress
	}
	return nil
}

func (m *NodeRegistration) GetRegistrationHeight() uint32 {
	if m != nil {
		return m.RegistrationHeight
	}
	return 0
}

func (m *NodeRegistration) GetLockedBalance() int64 {
	if m != nil {
		return m.LockedBalance
	}
	return 0
}

func (m *NodeRegistration) GetRegistrationStatus() uint32 {
	if m != nil {
		return m.RegistrationStatus
	}
	return 0
}

func (m *NodeRegistration) GetLatest() bool {
	if m != nil {
		return m.Latest
	}
	return false
}

func (m *NodeRegistration) GetHeight() uint32 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *NodeRegistration) GetNodeAddressInfo() *NodeAddressInfo {
	if m != nil {
		return m.NodeAddressInfo
	}
	return nil
}

// GetNodeRegisterRequest create request to get a list of node registry
type GetNodeRegistrationsRequest struct {
	// Fetch Node Registries based on queue status
	//2 : will retrieve node registries that have been deleted
	//1 : will retrieve node registries that still pending
	//0: will retrieve node registries that already registered
	RegistrationStatuses []uint32 `protobuf:"varint,1,rep,packed,name=RegistrationStatuses,proto3" json:"RegistrationStatuses,omitempty"`
	// Fetch Node Registries when registration height is greater than or equal to
	MinRegistrationHeight uint32 `protobuf:"varint,2,opt,name=MinRegistrationHeight,proto3" json:"MinRegistrationHeight,omitempty"`
	// Fetch Node Registries when registration height is less than or equal to
	MaxRegistrationHeight uint32 `protobuf:"varint,3,opt,name=MaxRegistrationHeight,proto3" json:"MaxRegistrationHeight,omitempty"`
	// Fetch Node Registries based on Pagination field
	Pagination           *Pagination `protobuf:"bytes,4,opt,name=Pagination,proto3" json:"Pagination,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *GetNodeRegistrationsRequest) Reset()         { *m = GetNodeRegistrationsRequest{} }
func (m *GetNodeRegistrationsRequest) String() string { return proto.CompactTextString(m) }
func (*GetNodeRegistrationsRequest) ProtoMessage()    {}
func (*GetNodeRegistrationsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_df1af0ec89e31788, []int{1}
}

func (m *GetNodeRegistrationsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetNodeRegistrationsRequest.Unmarshal(m, b)
}
func (m *GetNodeRegistrationsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetNodeRegistrationsRequest.Marshal(b, m, deterministic)
}
func (m *GetNodeRegistrationsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetNodeRegistrationsRequest.Merge(m, src)
}
func (m *GetNodeRegistrationsRequest) XXX_Size() int {
	return xxx_messageInfo_GetNodeRegistrationsRequest.Size(m)
}
func (m *GetNodeRegistrationsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetNodeRegistrationsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetNodeRegistrationsRequest proto.InternalMessageInfo

func (m *GetNodeRegistrationsRequest) GetRegistrationStatuses() []uint32 {
	if m != nil {
		return m.RegistrationStatuses
	}
	return nil
}

func (m *GetNodeRegistrationsRequest) GetMinRegistrationHeight() uint32 {
	if m != nil {
		return m.MinRegistrationHeight
	}
	return 0
}

func (m *GetNodeRegistrationsRequest) GetMaxRegistrationHeight() uint32 {
	if m != nil {
		return m.MaxRegistrationHeight
	}
	return 0
}

func (m *GetNodeRegistrationsRequest) GetPagination() *Pagination {
	if m != nil {
		return m.Pagination
	}
	return nil
}

type GetNodeRegistrationsResponse struct {
	// Number of node registry in total
	Total uint64 `protobuf:"varint,1,opt,name=Total,proto3" json:"Total,omitempty"`
	// NodeRegistrations list of NodeRegistration
	NodeRegistrations    []*NodeRegistration `protobuf:"bytes,2,rep,name=NodeRegistrations,proto3" json:"NodeRegistrations,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *GetNodeRegistrationsResponse) Reset()         { *m = GetNodeRegistrationsResponse{} }
func (m *GetNodeRegistrationsResponse) String() string { return proto.CompactTextString(m) }
func (*GetNodeRegistrationsResponse) ProtoMessage()    {}
func (*GetNodeRegistrationsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_df1af0ec89e31788, []int{2}
}

func (m *GetNodeRegistrationsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetNodeRegistrationsResponse.Unmarshal(m, b)
}
func (m *GetNodeRegistrationsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetNodeRegistrationsResponse.Marshal(b, m, deterministic)
}
func (m *GetNodeRegistrationsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetNodeRegistrationsResponse.Merge(m, src)
}
func (m *GetNodeRegistrationsResponse) XXX_Size() int {
	return xxx_messageInfo_GetNodeRegistrationsResponse.Size(m)
}
func (m *GetNodeRegistrationsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetNodeRegistrationsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetNodeRegistrationsResponse proto.InternalMessageInfo

func (m *GetNodeRegistrationsResponse) GetTotal() uint64 {
	if m != nil {
		return m.Total
	}
	return 0
}

func (m *GetNodeRegistrationsResponse) GetNodeRegistrations() []*NodeRegistration {
	if m != nil {
		return m.NodeRegistrations
	}
	return nil
}

// GetNodeRegistrationRequest create request for single node register
type GetNodeRegistrationRequest struct {
	NodePublicKey        []byte   `protobuf:"bytes,1,opt,name=NodePublicKey,proto3" json:"NodePublicKey,omitempty"`
	AccountAddressType   uint32   `protobuf:"varint,2,opt,name=AccountAddressType,proto3" json:"AccountAddressType,omitempty"`
	AccountAddress       []byte   `protobuf:"bytes,3,opt,name=AccountAddress,proto3" json:"AccountAddress,omitempty"`
	RegistrationHeight   uint32   `protobuf:"varint,4,opt,name=RegistrationHeight,proto3" json:"RegistrationHeight,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetNodeRegistrationRequest) Reset()         { *m = GetNodeRegistrationRequest{} }
func (m *GetNodeRegistrationRequest) String() string { return proto.CompactTextString(m) }
func (*GetNodeRegistrationRequest) ProtoMessage()    {}
func (*GetNodeRegistrationRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_df1af0ec89e31788, []int{3}
}

func (m *GetNodeRegistrationRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetNodeRegistrationRequest.Unmarshal(m, b)
}
func (m *GetNodeRegistrationRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetNodeRegistrationRequest.Marshal(b, m, deterministic)
}
func (m *GetNodeRegistrationRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetNodeRegistrationRequest.Merge(m, src)
}
func (m *GetNodeRegistrationRequest) XXX_Size() int {
	return xxx_messageInfo_GetNodeRegistrationRequest.Size(m)
}
func (m *GetNodeRegistrationRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetNodeRegistrationRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetNodeRegistrationRequest proto.InternalMessageInfo

func (m *GetNodeRegistrationRequest) GetNodePublicKey() []byte {
	if m != nil {
		return m.NodePublicKey
	}
	return nil
}

func (m *GetNodeRegistrationRequest) GetAccountAddressType() uint32 {
	if m != nil {
		return m.AccountAddressType
	}
	return 0
}

func (m *GetNodeRegistrationRequest) GetAccountAddress() []byte {
	if m != nil {
		return m.AccountAddress
	}
	return nil
}

func (m *GetNodeRegistrationRequest) GetRegistrationHeight() uint32 {
	if m != nil {
		return m.RegistrationHeight
	}
	return 0
}

type GetNodeRegistrationResponse struct {
	NodeRegistration     *NodeRegistration `protobuf:"bytes,1,opt,name=NodeRegistration,proto3" json:"NodeRegistration,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *GetNodeRegistrationResponse) Reset()         { *m = GetNodeRegistrationResponse{} }
func (m *GetNodeRegistrationResponse) String() string { return proto.CompactTextString(m) }
func (*GetNodeRegistrationResponse) ProtoMessage()    {}
func (*GetNodeRegistrationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_df1af0ec89e31788, []int{4}
}

func (m *GetNodeRegistrationResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetNodeRegistrationResponse.Unmarshal(m, b)
}
func (m *GetNodeRegistrationResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetNodeRegistrationResponse.Marshal(b, m, deterministic)
}
func (m *GetNodeRegistrationResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetNodeRegistrationResponse.Merge(m, src)
}
func (m *GetNodeRegistrationResponse) XXX_Size() int {
	return xxx_messageInfo_GetNodeRegistrationResponse.Size(m)
}
func (m *GetNodeRegistrationResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetNodeRegistrationResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetNodeRegistrationResponse proto.InternalMessageInfo

func (m *GetNodeRegistrationResponse) GetNodeRegistration() *NodeRegistration {
	if m != nil {
		return m.NodeRegistration
	}
	return nil
}

// NodeAdmissionTimestamp represent the timestamp for next node admission
type NodeAdmissionTimestamp struct {
	Timestamp            int64    `protobuf:"varint,1,opt,name=Timestamp,proto3" json:"Timestamp,omitempty"`
	BlockHeight          uint32   `protobuf:"varint,2,opt,name=BlockHeight,proto3" json:"BlockHeight,omitempty"`
	Latest               bool     `protobuf:"varint,3,opt,name=Latest,proto3" json:"Latest,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NodeAdmissionTimestamp) Reset()         { *m = NodeAdmissionTimestamp{} }
func (m *NodeAdmissionTimestamp) String() string { return proto.CompactTextString(m) }
func (*NodeAdmissionTimestamp) ProtoMessage()    {}
func (*NodeAdmissionTimestamp) Descriptor() ([]byte, []int) {
	return fileDescriptor_df1af0ec89e31788, []int{5}
}

func (m *NodeAdmissionTimestamp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NodeAdmissionTimestamp.Unmarshal(m, b)
}
func (m *NodeAdmissionTimestamp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NodeAdmissionTimestamp.Marshal(b, m, deterministic)
}
func (m *NodeAdmissionTimestamp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NodeAdmissionTimestamp.Merge(m, src)
}
func (m *NodeAdmissionTimestamp) XXX_Size() int {
	return xxx_messageInfo_NodeAdmissionTimestamp.Size(m)
}
func (m *NodeAdmissionTimestamp) XXX_DiscardUnknown() {
	xxx_messageInfo_NodeAdmissionTimestamp.DiscardUnknown(m)
}

var xxx_messageInfo_NodeAdmissionTimestamp proto.InternalMessageInfo

func (m *NodeAdmissionTimestamp) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *NodeAdmissionTimestamp) GetBlockHeight() uint32 {
	if m != nil {
		return m.BlockHeight
	}
	return 0
}

func (m *NodeAdmissionTimestamp) GetLatest() bool {
	if m != nil {
		return m.Latest
	}
	return false
}

// GetNodeRegistrationsByNodePublicKeys create request to get a list of node registry by a list of NodePublicKey
type GetNodeRegistrationsByNodePublicKeysRequest struct {
	NodePublicKeys [][]byte `protobuf:"bytes,1,rep,name=NodePublicKeys,proto3" json:"NodePublicKeys,omitempty"`
	// Fetch Node Registries based on Pagination field
	Pagination           *Pagination `protobuf:"bytes,2,opt,name=Pagination,proto3" json:"Pagination,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *GetNodeRegistrationsByNodePublicKeysRequest) Reset() {
	*m = GetNodeRegistrationsByNodePublicKeysRequest{}
}
func (m *GetNodeRegistrationsByNodePublicKeysRequest) String() string {
	return proto.CompactTextString(m)
}
func (*GetNodeRegistrationsByNodePublicKeysRequest) ProtoMessage() {}
func (*GetNodeRegistrationsByNodePublicKeysRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_df1af0ec89e31788, []int{6}
}

func (m *GetNodeRegistrationsByNodePublicKeysRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetNodeRegistrationsByNodePublicKeysRequest.Unmarshal(m, b)
}
func (m *GetNodeRegistrationsByNodePublicKeysRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetNodeRegistrationsByNodePublicKeysRequest.Marshal(b, m, deterministic)
}
func (m *GetNodeRegistrationsByNodePublicKeysRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetNodeRegistrationsByNodePublicKeysRequest.Merge(m, src)
}
func (m *GetNodeRegistrationsByNodePublicKeysRequest) XXX_Size() int {
	return xxx_messageInfo_GetNodeRegistrationsByNodePublicKeysRequest.Size(m)
}
func (m *GetNodeRegistrationsByNodePublicKeysRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetNodeRegistrationsByNodePublicKeysRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetNodeRegistrationsByNodePublicKeysRequest proto.InternalMessageInfo

func (m *GetNodeRegistrationsByNodePublicKeysRequest) GetNodePublicKeys() [][]byte {
	if m != nil {
		return m.NodePublicKeys
	}
	return nil
}

func (m *GetNodeRegistrationsByNodePublicKeysRequest) GetPagination() *Pagination {
	if m != nil {
		return m.Pagination
	}
	return nil
}

type GetNodeRegistrationsByNodePublicKeysResponse struct {
	// Number of node registry in total
	Total uint64 `protobuf:"varint,1,opt,name=Total,proto3" json:"Total,omitempty"`
	// NodeRegistrations list of NodeRegistration
	NodeRegistrations    []*NodeRegistration `protobuf:"bytes,2,rep,name=NodeRegistrations,proto3" json:"NodeRegistrations,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *GetNodeRegistrationsByNodePublicKeysResponse) Reset() {
	*m = GetNodeRegistrationsByNodePublicKeysResponse{}
}
func (m *GetNodeRegistrationsByNodePublicKeysResponse) String() string {
	return proto.CompactTextString(m)
}
func (*GetNodeRegistrationsByNodePublicKeysResponse) ProtoMessage() {}
func (*GetNodeRegistrationsByNodePublicKeysResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_df1af0ec89e31788, []int{7}
}

func (m *GetNodeRegistrationsByNodePublicKeysResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetNodeRegistrationsByNodePublicKeysResponse.Unmarshal(m, b)
}
func (m *GetNodeRegistrationsByNodePublicKeysResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetNodeRegistrationsByNodePublicKeysResponse.Marshal(b, m, deterministic)
}
func (m *GetNodeRegistrationsByNodePublicKeysResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetNodeRegistrationsByNodePublicKeysResponse.Merge(m, src)
}
func (m *GetNodeRegistrationsByNodePublicKeysResponse) XXX_Size() int {
	return xxx_messageInfo_GetNodeRegistrationsByNodePublicKeysResponse.Size(m)
}
func (m *GetNodeRegistrationsByNodePublicKeysResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetNodeRegistrationsByNodePublicKeysResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetNodeRegistrationsByNodePublicKeysResponse proto.InternalMessageInfo

func (m *GetNodeRegistrationsByNodePublicKeysResponse) GetTotal() uint64 {
	if m != nil {
		return m.Total
	}
	return 0
}

func (m *GetNodeRegistrationsByNodePublicKeysResponse) GetNodeRegistrations() []*NodeRegistration {
	if m != nil {
		return m.NodeRegistrations
	}
	return nil
}

type GetPendingNodeRegistrationsRequest struct {
	Limit                uint32   `protobuf:"varint,1,opt,name=Limit,proto3" json:"Limit,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetPendingNodeRegistrationsRequest) Reset()         { *m = GetPendingNodeRegistrationsRequest{} }
func (m *GetPendingNodeRegistrationsRequest) String() string { return proto.CompactTextString(m) }
func (*GetPendingNodeRegistrationsRequest) ProtoMessage()    {}
func (*GetPendingNodeRegistrationsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_df1af0ec89e31788, []int{8}
}

func (m *GetPendingNodeRegistrationsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetPendingNodeRegistrationsRequest.Unmarshal(m, b)
}
func (m *GetPendingNodeRegistrationsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetPendingNodeRegistrationsRequest.Marshal(b, m, deterministic)
}
func (m *GetPendingNodeRegistrationsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetPendingNodeRegistrationsRequest.Merge(m, src)
}
func (m *GetPendingNodeRegistrationsRequest) XXX_Size() int {
	return xxx_messageInfo_GetPendingNodeRegistrationsRequest.Size(m)
}
func (m *GetPendingNodeRegistrationsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetPendingNodeRegistrationsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetPendingNodeRegistrationsRequest proto.InternalMessageInfo

func (m *GetPendingNodeRegistrationsRequest) GetLimit() uint32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

type GetPendingNodeRegistrationsResponse struct {
	NodeRegistrations    []*NodeRegistration `protobuf:"bytes,1,rep,name=NodeRegistrations,proto3" json:"NodeRegistrations,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *GetPendingNodeRegistrationsResponse) Reset()         { *m = GetPendingNodeRegistrationsResponse{} }
func (m *GetPendingNodeRegistrationsResponse) String() string { return proto.CompactTextString(m) }
func (*GetPendingNodeRegistrationsResponse) ProtoMessage()    {}
func (*GetPendingNodeRegistrationsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_df1af0ec89e31788, []int{9}
}

func (m *GetPendingNodeRegistrationsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetPendingNodeRegistrationsResponse.Unmarshal(m, b)
}
func (m *GetPendingNodeRegistrationsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetPendingNodeRegistrationsResponse.Marshal(b, m, deterministic)
}
func (m *GetPendingNodeRegistrationsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetPendingNodeRegistrationsResponse.Merge(m, src)
}
func (m *GetPendingNodeRegistrationsResponse) XXX_Size() int {
	return xxx_messageInfo_GetPendingNodeRegistrationsResponse.Size(m)
}
func (m *GetPendingNodeRegistrationsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetPendingNodeRegistrationsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetPendingNodeRegistrationsResponse proto.InternalMessageInfo

func (m *GetPendingNodeRegistrationsResponse) GetNodeRegistrations() []*NodeRegistration {
	if m != nil {
		return m.NodeRegistrations
	}
	return nil
}

type GetMyNodePublicKeyResponse struct {
	NodePublicKey        []byte   `protobuf:"bytes,1,opt,name=NodePublicKey,proto3" json:"NodePublicKey,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetMyNodePublicKeyResponse) Reset()         { *m = GetMyNodePublicKeyResponse{} }
func (m *GetMyNodePublicKeyResponse) String() string { return proto.CompactTextString(m) }
func (*GetMyNodePublicKeyResponse) ProtoMessage()    {}
func (*GetMyNodePublicKeyResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_df1af0ec89e31788, []int{10}
}

func (m *GetMyNodePublicKeyResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetMyNodePublicKeyResponse.Unmarshal(m, b)
}
func (m *GetMyNodePublicKeyResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetMyNodePublicKeyResponse.Marshal(b, m, deterministic)
}
func (m *GetMyNodePublicKeyResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetMyNodePublicKeyResponse.Merge(m, src)
}
func (m *GetMyNodePublicKeyResponse) XXX_Size() int {
	return xxx_messageInfo_GetMyNodePublicKeyResponse.Size(m)
}
func (m *GetMyNodePublicKeyResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetMyNodePublicKeyResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetMyNodePublicKeyResponse proto.InternalMessageInfo

func (m *GetMyNodePublicKeyResponse) GetNodePublicKey() []byte {
	if m != nil {
		return m.NodePublicKey
	}
	return nil
}

func init() {
	proto.RegisterEnum("model.NodeRegistrationState", NodeRegistrationState_name, NodeRegistrationState_value)
	proto.RegisterType((*NodeRegistration)(nil), "model.NodeRegistration")
	proto.RegisterType((*GetNodeRegistrationsRequest)(nil), "model.GetNodeRegistrationsRequest")
	proto.RegisterType((*GetNodeRegistrationsResponse)(nil), "model.GetNodeRegistrationsResponse")
	proto.RegisterType((*GetNodeRegistrationRequest)(nil), "model.GetNodeRegistrationRequest")
	proto.RegisterType((*GetNodeRegistrationResponse)(nil), "model.GetNodeRegistrationResponse")
	proto.RegisterType((*NodeAdmissionTimestamp)(nil), "model.NodeAdmissionTimestamp")
	proto.RegisterType((*GetNodeRegistrationsByNodePublicKeysRequest)(nil), "model.GetNodeRegistrationsByNodePublicKeysRequest")
	proto.RegisterType((*GetNodeRegistrationsByNodePublicKeysResponse)(nil), "model.GetNodeRegistrationsByNodePublicKeysResponse")
	proto.RegisterType((*GetPendingNodeRegistrationsRequest)(nil), "model.GetPendingNodeRegistrationsRequest")
	proto.RegisterType((*GetPendingNodeRegistrationsResponse)(nil), "model.GetPendingNodeRegistrationsResponse")
	proto.RegisterType((*GetMyNodePublicKeyResponse)(nil), "model.GetMyNodePublicKeyResponse")
}

func init() {
	proto.RegisterFile("model/nodeRegistration.proto", fileDescriptor_df1af0ec89e31788)
}

var fileDescriptor_df1af0ec89e31788 = []byte{
	// 651 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x55, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0x66, 0xed, 0x26, 0xb4, 0x93, 0xa6, 0x4d, 0x57, 0x6d, 0xb0, 0xd2, 0x1e, 0x2c, 0x83, 0x2a,
	0xab, 0x40, 0x02, 0x81, 0x13, 0x27, 0x1a, 0x8a, 0x4a, 0x45, 0x5a, 0x15, 0xd3, 0x13, 0x37, 0xc7,
	0x1e, 0xd2, 0x55, 0x6d, 0x6f, 0x88, 0xd7, 0x12, 0xe5, 0xc2, 0x95, 0x1b, 0xaf, 0xc1, 0xb3, 0xf0,
	0x34, 0x3c, 0x02, 0xf2, 0xda, 0xb8, 0xfe, 0x6b, 0x95, 0x5c, 0xb8, 0x44, 0x99, 0xef, 0xdb, 0x99,
	0x9d, 0xbf, 0x6f, 0x0d, 0x7b, 0x3e, 0x77, 0xd1, 0x1b, 0x04, 0xdc, 0x45, 0x0b, 0xa7, 0x2c, 0x14,
	0x73, 0x5b, 0x30, 0x1e, 0xf4, 0x67, 0x73, 0x2e, 0x38, 0x6d, 0x48, 0xb6, 0xd7, 0x4d, 0x0e, 0xcd,
	0xec, 0x29, 0x0b, 0x72, 0x74, 0x6f, 0xf7, 0xc6, 0xf9, 0xd0, 0x75, 0xe7, 0x18, 0x86, 0x27, 0xc1,
	0x67, 0x9e, 0x90, 0xc6, 0x2f, 0x15, 0x3a, 0x67, 0xa5, 0xb0, 0xb4, 0x07, 0xcd, 0x18, 0x3b, 0x39,
	0xd2, 0x88, 0x4e, 0x4c, 0x75, 0xa4, 0x3c, 0x23, 0x56, 0x8a, 0xd0, 0x47, 0xd0, 0x8e, 0xff, 0x9d,
	0x47, 0x13, 0x8f, 0x39, 0xef, 0xf1, 0x5a, 0x53, 0x74, 0x62, 0xae, 0x5b, 0x45, 0x90, 0xf6, 0x81,
	0x1e, 0x3a, 0x0e, 0x8f, 0x02, 0x91, 0x5e, 0x79, 0x71, 0x3d, 0x43, 0x4d, 0xd5, 0x89, 0xd9, 0xb6,
	0x6a, 0x18, 0xba, 0x0f, 0x1b, 0x45, 0x54, 0x5b, 0x91, 0x61, 0x4b, 0x68, 0x1c, 0x37, 0x9f, 0xe9,
	0x3b, 0x64, 0xd3, 0x4b, 0xa1, 0x35, 0x92, 0xb8, 0x55, 0x86, 0x9a, 0xd0, 0x1e, 0x73, 0xe7, 0x0a,
	0xdd, 0x91, 0xed, 0xd9, 0x81, 0x83, 0x5a, 0x33, 0x2b, 0xa8, 0x48, 0x94, 0x23, 0x7f, 0x14, 0xb6,
	0x88, 0x42, 0xed, 0x7e, 0x35, 0x72, 0xc2, 0xd0, 0x2e, 0x34, 0xc7, 0xb6, 0xc0, 0x50, 0x68, 0xab,
	0x3a, 0x31, 0x57, 0xad, 0xd4, 0x8a, 0xf1, 0x34, 0xab, 0x35, 0xe9, 0x9b, 0x5a, 0xf4, 0x35, 0x6c,
	0x9e, 0x15, 0x27, 0xa0, 0x81, 0x4e, 0xcc, 0xd6, 0xb0, 0xdb, 0x97, 0xf3, 0xe9, 0x97, 0x58, 0xab,
	0x7c, 0xdc, 0xf8, 0x43, 0x60, 0xf7, 0x18, 0x45, 0x79, 0x5a, 0xa1, 0x85, 0x5f, 0xa2, 0xf8, 0xe6,
	0x21, 0x6c, 0x57, 0xf3, 0xc4, 0x50, 0x23, 0xba, 0x6a, 0xb6, 0xad, 0x5a, 0x8e, 0xbe, 0x84, 0x9d,
	0x53, 0x16, 0xd4, 0xb4, 0x54, 0x91, 0xc9, 0xd7, 0x93, 0xd2, 0xcb, 0xfe, 0x5a, 0xe3, 0xa5, 0xa6,
	0x5e, 0x75, 0x24, 0x7d, 0x0e, 0x70, 0x9e, 0xed, 0xa6, 0x9c, 0x6f, 0x6b, 0xb8, 0x95, 0x16, 0x7f,
	0x43, 0x58, 0xb9, 0x43, 0xc6, 0x77, 0xd8, 0xab, 0xaf, 0x38, 0x9c, 0xf1, 0x20, 0x44, 0xaa, 0x41,
	0xe3, 0x82, 0x0b, 0xdb, 0x93, 0x7b, 0xba, 0x22, 0xc7, 0x9a, 0x00, 0xf4, 0x2d, 0x6c, 0x55, 0xdc,
	0x34, 0x45, 0x57, 0xcd, 0xd6, 0xf0, 0x41, 0xae, 0xe1, 0x79, 0xde, 0xaa, 0x7a, 0x18, 0xbf, 0x09,
	0xf4, 0x6a, 0x32, 0xf8, 0xd7, 0xf2, 0x8a, 0x18, 0xc8, 0xe2, 0x62, 0x50, 0x96, 0x10, 0x83, 0xba,
	0x84, 0x18, 0x56, 0x6e, 0x13, 0x83, 0x31, 0xa9, 0xdd, 0x9f, 0xac, 0x99, 0x6f, 0xaa, 0x2f, 0x81,
	0xac, 0xe7, 0x8e, 0x8e, 0x55, 0x1c, 0x0c, 0x01, 0xdd, 0x64, 0x6f, 0x7d, 0x16, 0x86, 0x8c, 0x07,
	0x17, 0xcc, 0xc7, 0x50, 0xd8, 0xfe, 0x8c, 0xea, 0xb0, 0x96, 0x19, 0xb9, 0x77, 0x65, 0x2d, 0x7f,
	0xa2, 0x35, 0xf2, 0xb8, 0x73, 0x55, 0x58, 0xc1, 0x3c, 0x94, 0x13, 0x9d, 0x9a, 0x17, 0x9d, 0xf1,
	0x83, 0xc0, 0xe3, 0xba, 0x45, 0x19, 0x5d, 0x17, 0x26, 0x91, 0x49, 0x65, 0x1f, 0x36, 0x8a, 0x84,
	0x14, 0xc9, 0xba, 0x55, 0x42, 0x4b, 0x2b, 0xab, 0x2c, 0xb2, 0xb2, 0x3f, 0x09, 0x3c, 0x59, 0x2c,
	0x95, 0xff, 0xb5, 0xc3, 0xaf, 0xc0, 0x38, 0x46, 0x71, 0x8e, 0x81, 0xcb, 0x82, 0xe9, 0xad, 0xaf,
	0xc7, 0x36, 0x34, 0xc6, 0xcc, 0x67, 0x42, 0xa6, 0xd1, 0xb6, 0x12, 0xc3, 0xf0, 0xe0, 0xe1, 0x9d,
	0xbe, 0x69, 0x0d, 0xb5, 0x99, 0x92, 0xa5, 0x33, 0x1d, 0x49, 0xb1, 0x9d, 0x16, 0x3b, 0x95, 0x5d,
	0xb2, 0x90, 0xd8, 0x0e, 0xc6, 0xb0, 0x53, 0x0e, 0x1c, 0xbf, 0x76, 0x48, 0x69, 0x32, 0xf3, 0x84,
	0xc0, 0x39, 0xba, 0x9d, 0x7b, 0x74, 0x03, 0x20, 0xc6, 0x3e, 0x44, 0x18, 0xa1, 0xdb, 0x21, 0x74,
	0x13, 0x5a, 0xb1, 0x7d, 0x84, 0x1e, 0x0a, 0x74, 0x3b, 0xca, 0xe8, 0xe0, 0x93, 0x39, 0x65, 0xe2,
	0x32, 0x9a, 0xf4, 0x1d, 0xee, 0x0f, 0xbe, 0x71, 0x3e, 0x71, 0x92, 0xdf, 0xa7, 0x0e, 0x9f, 0xe3,
	0xc0, 0xe1, 0xbe, 0xcf, 0x83, 0x81, 0xac, 0x70, 0xd2, 0x94, 0x5f, 0xd4, 0x17, 0x7f, 0x03, 0x00,
	0x00, 0xff, 0xff, 0x15, 0x97, 0x78, 0x63, 0xad, 0x07, 0x00, 0x00,
}
