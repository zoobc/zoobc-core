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
	AccountAddress       string           `protobuf:"bytes,3,opt,name=AccountAddress,proto3" json:"AccountAddress,omitempty"`
	RegistrationHeight   uint32           `protobuf:"varint,4,opt,name=RegistrationHeight,proto3" json:"RegistrationHeight,omitempty"`
	LockedBalance        int64            `protobuf:"varint,5,opt,name=LockedBalance,proto3" json:"LockedBalance,omitempty"`
	RegistrationStatus   uint32           `protobuf:"varint,6,opt,name=RegistrationStatus,proto3" json:"RegistrationStatus,omitempty"`
	Latest               bool             `protobuf:"varint,7,opt,name=Latest,proto3" json:"Latest,omitempty"`
	Height               uint32           `protobuf:"varint,8,opt,name=Height,proto3" json:"Height,omitempty"`
	NodeAddressInfo      *NodeAddressInfo `protobuf:"bytes,9,opt,name=NodeAddressInfo,proto3" json:"NodeAddressInfo,omitempty"`
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

func (m *NodeRegistration) GetAccountAddress() string {
	if m != nil {
		return m.AccountAddress
	}
	return ""
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
	AccountAddress       string   `protobuf:"bytes,2,opt,name=AccountAddress,proto3" json:"AccountAddress,omitempty"`
	RegistrationHeight   uint32   `protobuf:"varint,3,opt,name=RegistrationHeight,proto3" json:"RegistrationHeight,omitempty"`
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

func (m *GetNodeRegistrationRequest) GetAccountAddress() string {
	if m != nil {
		return m.AccountAddress
	}
	return ""
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
	// 635 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x55, 0xcf, 0x53, 0xd3, 0x40,
	0x14, 0x76, 0x53, 0x5a, 0xe9, 0x2b, 0x85, 0xb2, 0x03, 0x35, 0x53, 0x38, 0x64, 0xa2, 0xc3, 0x64,
	0x50, 0x5b, 0xad, 0x9e, 0x3c, 0x49, 0xc5, 0x41, 0xc6, 0xc2, 0xe0, 0xca, 0xc9, 0x5b, 0x9a, 0x3c,
	0xcb, 0x0e, 0x49, 0xb6, 0x36, 0x9b, 0x19, 0xf1, 0xe2, 0xd5, 0x9b, 0x77, 0xff, 0x4a, 0x8f, 0x1e,
	0x9d, 0x6c, 0x62, 0xc9, 0x2f, 0x98, 0x72, 0xf1, 0xc2, 0xf0, 0xde, 0xb7, 0xdf, 0xee, 0xb7, 0xdf,
	0x7e, 0x2f, 0x85, 0x5d, 0x5f, 0xb8, 0xe8, 0x0d, 0x02, 0xe1, 0x22, 0xc3, 0x29, 0x0f, 0xe5, 0xdc,
	0x96, 0x5c, 0x04, 0xfd, 0xd9, 0x5c, 0x48, 0x41, 0xeb, 0x0a, 0xed, 0x75, 0x93, 0x45, 0x33, 0x7b,
	0xca, 0x83, 0x0c, 0xdc, 0xdb, 0xb9, 0x26, 0x1f, 0xb8, 0xee, 0x1c, 0xc3, 0xf0, 0x38, 0xf8, 0x2c,
	0x12, 0xd0, 0xfc, 0xa3, 0x41, 0xe7, 0xb4, 0xb0, 0x2d, 0xed, 0x41, 0x23, 0xee, 0x1d, 0x1f, 0xea,
	0xc4, 0x20, 0x56, 0x6d, 0xa4, 0x3d, 0x23, 0x2c, 0xed, 0xd0, 0x47, 0xd0, 0x8e, 0xff, 0x3b, 0x8b,
	0x26, 0x1e, 0x77, 0xde, 0xe3, 0x95, 0xae, 0x19, 0xc4, 0x5a, 0x63, 0xf9, 0x26, 0xdd, 0x83, 0xf5,
	0x03, 0xc7, 0x11, 0x51, 0x20, 0xd3, 0x23, 0xf5, 0x9a, 0x41, 0xac, 0x26, 0x2b, 0x74, 0x69, 0x1f,
	0x68, 0xf6, 0xe4, 0x77, 0xc8, 0xa7, 0x17, 0x52, 0x5f, 0x31, 0x88, 0xd5, 0x66, 0x15, 0x08, 0xb5,
	0xa0, 0x3d, 0x16, 0xce, 0x25, 0xba, 0x23, 0xdb, 0xb3, 0x03, 0x07, 0xf5, 0xfa, 0x42, 0x60, 0x1e,
	0x28, 0xee, 0xfc, 0x51, 0xda, 0x32, 0x0a, 0xf5, 0x46, 0x79, 0xe7, 0x04, 0xa1, 0x5d, 0x68, 0x8c,
	0x6d, 0x89, 0xa1, 0xd4, 0xef, 0x1b, 0xc4, 0x5a, 0x65, 0x69, 0x15, 0xf7, 0x53, 0x55, 0xab, 0x8a,
	0x9b, 0x56, 0xf4, 0x35, 0x6c, 0x9c, 0xe6, 0x1d, 0xd5, 0x9b, 0x06, 0xb1, 0x5a, 0xc3, 0x6e, 0x5f,
	0xf9, 0xdd, 0x2f, 0xa0, 0xac, 0xb8, 0xdc, 0xfc, 0x4d, 0x60, 0xe7, 0x08, 0x65, 0xd1, 0xfd, 0x90,
	0xe1, 0x97, 0x28, 0x3e, 0x79, 0x08, 0x5b, 0x65, 0x9d, 0x18, 0xea, 0xc4, 0xa8, 0x59, 0x6d, 0x56,
	0x89, 0xd1, 0x97, 0xb0, 0x7d, 0xc2, 0x83, 0x0a, 0x4b, 0x35, 0x25, 0xbe, 0x1a, 0x54, 0x2c, 0xfb,
	0x6b, 0x05, 0xab, 0x96, 0xb2, 0xaa, 0x40, 0xfa, 0x1c, 0xe0, 0x6c, 0x91, 0x35, 0xf5, 0x66, 0xad,
	0xe1, 0x66, 0x7a, 0xf9, 0x6b, 0x80, 0x65, 0x16, 0x99, 0xdf, 0x61, 0xb7, 0xfa, 0xc6, 0xe1, 0x4c,
	0x04, 0x21, 0x52, 0x1d, 0xea, 0xe7, 0x42, 0xda, 0x9e, 0xca, 0xdd, 0x8a, 0x7a, 0xd6, 0xa4, 0x41,
	0xdf, 0xc2, 0x66, 0x89, 0xa6, 0x6b, 0x46, 0xcd, 0x6a, 0x0d, 0x1f, 0x64, 0x0c, 0xcf, 0xe2, 0xac,
	0xcc, 0x30, 0x7f, 0x11, 0xe8, 0x55, 0x28, 0xf8, 0x67, 0x79, 0x29, 0xdc, 0x64, 0xb9, 0x70, 0x6b,
	0x77, 0x08, 0x77, 0xed, 0xa6, 0x70, 0x9b, 0x93, 0xca, 0x3c, 0x2c, 0xcc, 0x79, 0x53, 0x9e, 0x54,
	0xa5, 0xef, 0x16, 0x07, 0x4a, 0x04, 0x53, 0x42, 0x37, 0xc9, 0xa1, 0xcf, 0xc3, 0x90, 0x8b, 0xe0,
	0x9c, 0xfb, 0x18, 0x4a, 0xdb, 0x9f, 0x51, 0x03, 0x9a, 0x8b, 0x22, 0x33, 0xf7, 0xcd, 0xec, 0x8a,
	0xd6, 0xc8, 0x13, 0xce, 0x65, 0x2e, 0x52, 0xd9, 0x56, 0x66, 0x88, 0x6a, 0xd9, 0x21, 0x32, 0x7f,
	0x10, 0x78, 0x5c, 0xf5, 0xf0, 0xa3, 0xab, 0x9c, 0xb3, 0x8b, 0xe8, 0xef, 0xc1, 0x7a, 0x1e, 0x50,
	0xa1, 0x5f, 0x63, 0x85, 0x6e, 0x21, 0x82, 0xda, 0x32, 0x11, 0xfc, 0x49, 0xe0, 0xc9, 0x72, 0x52,
	0xfe, 0x57, 0x26, 0x5f, 0x81, 0x79, 0x84, 0xf2, 0x0c, 0x03, 0x97, 0x07, 0xd3, 0x1b, 0xbf, 0x06,
	0x5b, 0x50, 0x1f, 0x73, 0x9f, 0x4b, 0x25, 0xa3, 0xcd, 0x92, 0xc2, 0xf4, 0xe0, 0xe1, 0xad, 0xdc,
	0xf4, 0x0e, 0x95, 0x4a, 0xc9, 0x9d, 0x95, 0x8e, 0xd4, 0xf0, 0x9c, 0xe4, 0x9d, 0x5a, 0x1c, 0xb2,
	0xd4, 0xf0, 0xec, 0x8f, 0x61, 0xbb, 0xb8, 0x71, 0xfc, 0xf5, 0x42, 0x4a, 0x93, 0x37, 0x4f, 0x00,
	0x9c, 0xa3, 0xdb, 0xb9, 0x47, 0xd7, 0x01, 0xe2, 0xde, 0x87, 0x08, 0x23, 0x74, 0x3b, 0x84, 0x6e,
	0x40, 0x2b, 0xae, 0x0f, 0xd1, 0x43, 0x89, 0x6e, 0x47, 0x1b, 0xed, 0x7f, 0xb2, 0xa6, 0x5c, 0x5e,
	0x44, 0x93, 0xbe, 0x23, 0xfc, 0xc1, 0x37, 0x21, 0x26, 0x4e, 0xf2, 0xf7, 0xa9, 0x23, 0xe6, 0x38,
	0x70, 0x84, 0xef, 0x8b, 0x60, 0xa0, 0x6e, 0x38, 0x69, 0xa8, 0x5f, 0xbc, 0x17, 0x7f, 0x03, 0x00,
	0x00, 0xff, 0xff, 0xd3, 0xde, 0x49, 0xfa, 0x4d, 0x07, 0x00, 0x00,
}
