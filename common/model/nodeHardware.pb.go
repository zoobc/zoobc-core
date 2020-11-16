// Code generated by protoc-gen-go. DO NOT EDIT.
// source: model/nodeHardware.proto

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

type GetNodeHardwareResponse struct {
	NodeHardware         *NodeHardware `protobuf:"bytes,1,opt,name=NodeHardware,proto3" json:"NodeHardware,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *GetNodeHardwareResponse) Reset()         { *m = GetNodeHardwareResponse{} }
func (m *GetNodeHardwareResponse) String() string { return proto.CompactTextString(m) }
func (*GetNodeHardwareResponse) ProtoMessage()    {}
func (*GetNodeHardwareResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ce967b26533b171c, []int{0}
}

func (m *GetNodeHardwareResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetNodeHardwareResponse.Unmarshal(m, b)
}
func (m *GetNodeHardwareResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetNodeHardwareResponse.Marshal(b, m, deterministic)
}
func (m *GetNodeHardwareResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetNodeHardwareResponse.Merge(m, src)
}
func (m *GetNodeHardwareResponse) XXX_Size() int {
	return xxx_messageInfo_GetNodeHardwareResponse.Size(m)
}
func (m *GetNodeHardwareResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetNodeHardwareResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetNodeHardwareResponse proto.InternalMessageInfo

func (m *GetNodeHardwareResponse) GetNodeHardware() *NodeHardware {
	if m != nil {
		return m.NodeHardware
	}
	return nil
}

type GetNodeHardwareRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetNodeHardwareRequest) Reset()         { *m = GetNodeHardwareRequest{} }
func (m *GetNodeHardwareRequest) String() string { return proto.CompactTextString(m) }
func (*GetNodeHardwareRequest) ProtoMessage()    {}
func (*GetNodeHardwareRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ce967b26533b171c, []int{1}
}

func (m *GetNodeHardwareRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetNodeHardwareRequest.Unmarshal(m, b)
}
func (m *GetNodeHardwareRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetNodeHardwareRequest.Marshal(b, m, deterministic)
}
func (m *GetNodeHardwareRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetNodeHardwareRequest.Merge(m, src)
}
func (m *GetNodeHardwareRequest) XXX_Size() int {
	return xxx_messageInfo_GetNodeHardwareRequest.Size(m)
}
func (m *GetNodeHardwareRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetNodeHardwareRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetNodeHardwareRequest proto.InternalMessageInfo

// Get Node Current Time Based on UTC
type GetNodeTimeResponse struct {
	NodeTime             int64    `protobuf:"varint,1,opt,name=NodeTime,proto3" json:"NodeTime,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetNodeTimeResponse) Reset()         { *m = GetNodeTimeResponse{} }
func (m *GetNodeTimeResponse) String() string { return proto.CompactTextString(m) }
func (*GetNodeTimeResponse) ProtoMessage()    {}
func (*GetNodeTimeResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ce967b26533b171c, []int{2}
}

func (m *GetNodeTimeResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetNodeTimeResponse.Unmarshal(m, b)
}
func (m *GetNodeTimeResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetNodeTimeResponse.Marshal(b, m, deterministic)
}
func (m *GetNodeTimeResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetNodeTimeResponse.Merge(m, src)
}
func (m *GetNodeTimeResponse) XXX_Size() int {
	return xxx_messageInfo_GetNodeTimeResponse.Size(m)
}
func (m *GetNodeTimeResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetNodeTimeResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetNodeTimeResponse proto.InternalMessageInfo

func (m *GetNodeTimeResponse) GetNodeTime() int64 {
	if m != nil {
		return m.NodeTime
	}
	return 0
}

type NodeHardware struct {
	CPUInformation       []*CPUInformation   `protobuf:"bytes,1,rep,name=CPUInformation,proto3" json:"CPUInformation,omitempty"`
	MemoryInformation    *MemoryInformation  `protobuf:"bytes,2,opt,name=MemoryInformation,proto3" json:"MemoryInformation,omitempty"`
	StorageInformation   *StorageInformation `protobuf:"bytes,3,opt,name=StorageInformation,proto3" json:"StorageInformation,omitempty"`
	HostInformation      *HostInformation    `protobuf:"bytes,4,opt,name=HostInformation,proto3" json:"HostInformation,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *NodeHardware) Reset()         { *m = NodeHardware{} }
func (m *NodeHardware) String() string { return proto.CompactTextString(m) }
func (*NodeHardware) ProtoMessage()    {}
func (*NodeHardware) Descriptor() ([]byte, []int) {
	return fileDescriptor_ce967b26533b171c, []int{3}
}

func (m *NodeHardware) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NodeHardware.Unmarshal(m, b)
}
func (m *NodeHardware) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NodeHardware.Marshal(b, m, deterministic)
}
func (m *NodeHardware) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NodeHardware.Merge(m, src)
}
func (m *NodeHardware) XXX_Size() int {
	return xxx_messageInfo_NodeHardware.Size(m)
}
func (m *NodeHardware) XXX_DiscardUnknown() {
	xxx_messageInfo_NodeHardware.DiscardUnknown(m)
}

var xxx_messageInfo_NodeHardware proto.InternalMessageInfo

func (m *NodeHardware) GetCPUInformation() []*CPUInformation {
	if m != nil {
		return m.CPUInformation
	}
	return nil
}

func (m *NodeHardware) GetMemoryInformation() *MemoryInformation {
	if m != nil {
		return m.MemoryInformation
	}
	return nil
}

func (m *NodeHardware) GetStorageInformation() *StorageInformation {
	if m != nil {
		return m.StorageInformation
	}
	return nil
}

func (m *NodeHardware) GetHostInformation() *HostInformation {
	if m != nil {
		return m.HostInformation
	}
	return nil
}

type CPUInformation struct {
	Family               string   `protobuf:"bytes,1,opt,name=Family,proto3" json:"Family,omitempty"`
	CPUIndex             int32    `protobuf:"varint,2,opt,name=CPUIndex,proto3" json:"CPUIndex,omitempty"`
	Model                string   `protobuf:"bytes,3,opt,name=Model,proto3" json:"Model,omitempty"`
	ModelName            string   `protobuf:"bytes,4,opt,name=ModelName,proto3" json:"ModelName,omitempty"`
	VendorId             string   `protobuf:"bytes,5,opt,name=VendorId,proto3" json:"VendorId,omitempty"`
	Mhz                  float64  `protobuf:"fixed64,6,opt,name=Mhz,proto3" json:"Mhz,omitempty"`
	CacheSize            int32    `protobuf:"varint,7,opt,name=CacheSize,proto3" json:"CacheSize,omitempty"`
	UsedPercent          float64  `protobuf:"fixed64,8,opt,name=UsedPercent,proto3" json:"UsedPercent,omitempty"`
	CoreID               string   `protobuf:"bytes,9,opt,name=CoreID,proto3" json:"CoreID,omitempty"`
	Cores                int32    `protobuf:"varint,10,opt,name=Cores,proto3" json:"Cores,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CPUInformation) Reset()         { *m = CPUInformation{} }
func (m *CPUInformation) String() string { return proto.CompactTextString(m) }
func (*CPUInformation) ProtoMessage()    {}
func (*CPUInformation) Descriptor() ([]byte, []int) {
	return fileDescriptor_ce967b26533b171c, []int{4}
}

func (m *CPUInformation) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CPUInformation.Unmarshal(m, b)
}
func (m *CPUInformation) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CPUInformation.Marshal(b, m, deterministic)
}
func (m *CPUInformation) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CPUInformation.Merge(m, src)
}
func (m *CPUInformation) XXX_Size() int {
	return xxx_messageInfo_CPUInformation.Size(m)
}
func (m *CPUInformation) XXX_DiscardUnknown() {
	xxx_messageInfo_CPUInformation.DiscardUnknown(m)
}

var xxx_messageInfo_CPUInformation proto.InternalMessageInfo

func (m *CPUInformation) GetFamily() string {
	if m != nil {
		return m.Family
	}
	return ""
}

func (m *CPUInformation) GetCPUIndex() int32 {
	if m != nil {
		return m.CPUIndex
	}
	return 0
}

func (m *CPUInformation) GetModel() string {
	if m != nil {
		return m.Model
	}
	return ""
}

func (m *CPUInformation) GetModelName() string {
	if m != nil {
		return m.ModelName
	}
	return ""
}

func (m *CPUInformation) GetVendorId() string {
	if m != nil {
		return m.VendorId
	}
	return ""
}

func (m *CPUInformation) GetMhz() float64 {
	if m != nil {
		return m.Mhz
	}
	return 0
}

func (m *CPUInformation) GetCacheSize() int32 {
	if m != nil {
		return m.CacheSize
	}
	return 0
}

func (m *CPUInformation) GetUsedPercent() float64 {
	if m != nil {
		return m.UsedPercent
	}
	return 0
}

func (m *CPUInformation) GetCoreID() string {
	if m != nil {
		return m.CoreID
	}
	return ""
}

func (m *CPUInformation) GetCores() int32 {
	if m != nil {
		return m.Cores
	}
	return 0
}

type HostInformation struct {
	Uptime                 uint64   `protobuf:"varint,1,opt,name=Uptime,proto3" json:"Uptime,omitempty"`
	OS                     string   `protobuf:"bytes,2,opt,name=OS,proto3" json:"OS,omitempty"`
	Platform               string   `protobuf:"bytes,3,opt,name=Platform,proto3" json:"Platform,omitempty"`
	PlatformFamily         string   `protobuf:"bytes,4,opt,name=PlatformFamily,proto3" json:"PlatformFamily,omitempty"`
	PlatformVersion        string   `protobuf:"bytes,5,opt,name=PlatformVersion,proto3" json:"PlatformVersion,omitempty"`
	NumberOfRunningProcess uint64   `protobuf:"varint,6,opt,name=NumberOfRunningProcess,proto3" json:"NumberOfRunningProcess,omitempty"`
	HostID                 string   `protobuf:"bytes,7,opt,name=HostID,proto3" json:"HostID,omitempty"`
	HostName               string   `protobuf:"bytes,8,opt,name=HostName,proto3" json:"HostName,omitempty"`
	XXX_NoUnkeyedLiteral   struct{} `json:"-"`
	XXX_unrecognized       []byte   `json:"-"`
	XXX_sizecache          int32    `json:"-"`
}

func (m *HostInformation) Reset()         { *m = HostInformation{} }
func (m *HostInformation) String() string { return proto.CompactTextString(m) }
func (*HostInformation) ProtoMessage()    {}
func (*HostInformation) Descriptor() ([]byte, []int) {
	return fileDescriptor_ce967b26533b171c, []int{5}
}

func (m *HostInformation) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HostInformation.Unmarshal(m, b)
}
func (m *HostInformation) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HostInformation.Marshal(b, m, deterministic)
}
func (m *HostInformation) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HostInformation.Merge(m, src)
}
func (m *HostInformation) XXX_Size() int {
	return xxx_messageInfo_HostInformation.Size(m)
}
func (m *HostInformation) XXX_DiscardUnknown() {
	xxx_messageInfo_HostInformation.DiscardUnknown(m)
}

var xxx_messageInfo_HostInformation proto.InternalMessageInfo

func (m *HostInformation) GetUptime() uint64 {
	if m != nil {
		return m.Uptime
	}
	return 0
}

func (m *HostInformation) GetOS() string {
	if m != nil {
		return m.OS
	}
	return ""
}

func (m *HostInformation) GetPlatform() string {
	if m != nil {
		return m.Platform
	}
	return ""
}

func (m *HostInformation) GetPlatformFamily() string {
	if m != nil {
		return m.PlatformFamily
	}
	return ""
}

func (m *HostInformation) GetPlatformVersion() string {
	if m != nil {
		return m.PlatformVersion
	}
	return ""
}

func (m *HostInformation) GetNumberOfRunningProcess() uint64 {
	if m != nil {
		return m.NumberOfRunningProcess
	}
	return 0
}

func (m *HostInformation) GetHostID() string {
	if m != nil {
		return m.HostID
	}
	return ""
}

func (m *HostInformation) GetHostName() string {
	if m != nil {
		return m.HostName
	}
	return ""
}

type MemoryInformation struct {
	Total uint64 `protobuf:"varint,1,opt,name=Total,proto3" json:"Total,omitempty"`
	// This is the kernel's notion of free memory; RAM chips whose bits nobody
	// cares about the value of right now. For a human consumable number,
	// Available is what you really want.
	Free uint64 `protobuf:"varint,2,opt,name=Free,proto3" json:"Free,omitempty"`
	// RAM available for programs to allocate
	Available            uint64   `protobuf:"varint,3,opt,name=Available,proto3" json:"Available,omitempty"`
	Used                 uint64   `protobuf:"varint,4,opt,name=Used,proto3" json:"Used,omitempty"`
	UsedPercent          float64  `protobuf:"fixed64,5,opt,name=UsedPercent,proto3" json:"UsedPercent,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MemoryInformation) Reset()         { *m = MemoryInformation{} }
func (m *MemoryInformation) String() string { return proto.CompactTextString(m) }
func (*MemoryInformation) ProtoMessage()    {}
func (*MemoryInformation) Descriptor() ([]byte, []int) {
	return fileDescriptor_ce967b26533b171c, []int{6}
}

func (m *MemoryInformation) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MemoryInformation.Unmarshal(m, b)
}
func (m *MemoryInformation) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MemoryInformation.Marshal(b, m, deterministic)
}
func (m *MemoryInformation) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MemoryInformation.Merge(m, src)
}
func (m *MemoryInformation) XXX_Size() int {
	return xxx_messageInfo_MemoryInformation.Size(m)
}
func (m *MemoryInformation) XXX_DiscardUnknown() {
	xxx_messageInfo_MemoryInformation.DiscardUnknown(m)
}

var xxx_messageInfo_MemoryInformation proto.InternalMessageInfo

func (m *MemoryInformation) GetTotal() uint64 {
	if m != nil {
		return m.Total
	}
	return 0
}

func (m *MemoryInformation) GetFree() uint64 {
	if m != nil {
		return m.Free
	}
	return 0
}

func (m *MemoryInformation) GetAvailable() uint64 {
	if m != nil {
		return m.Available
	}
	return 0
}

func (m *MemoryInformation) GetUsed() uint64 {
	if m != nil {
		return m.Used
	}
	return 0
}

func (m *MemoryInformation) GetUsedPercent() float64 {
	if m != nil {
		return m.UsedPercent
	}
	return 0
}

type StorageInformation struct {
	FsType               string   `protobuf:"bytes,1,opt,name=FsType,proto3" json:"FsType,omitempty"`
	Total                uint64   `protobuf:"varint,2,opt,name=Total,proto3" json:"Total,omitempty"`
	Free                 uint64   `protobuf:"varint,3,opt,name=Free,proto3" json:"Free,omitempty"`
	Used                 uint64   `protobuf:"varint,4,opt,name=Used,proto3" json:"Used,omitempty"`
	UsedPercent          float64  `protobuf:"fixed64,5,opt,name=UsedPercent,proto3" json:"UsedPercent,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StorageInformation) Reset()         { *m = StorageInformation{} }
func (m *StorageInformation) String() string { return proto.CompactTextString(m) }
func (*StorageInformation) ProtoMessage()    {}
func (*StorageInformation) Descriptor() ([]byte, []int) {
	return fileDescriptor_ce967b26533b171c, []int{7}
}

func (m *StorageInformation) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StorageInformation.Unmarshal(m, b)
}
func (m *StorageInformation) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StorageInformation.Marshal(b, m, deterministic)
}
func (m *StorageInformation) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StorageInformation.Merge(m, src)
}
func (m *StorageInformation) XXX_Size() int {
	return xxx_messageInfo_StorageInformation.Size(m)
}
func (m *StorageInformation) XXX_DiscardUnknown() {
	xxx_messageInfo_StorageInformation.DiscardUnknown(m)
}

var xxx_messageInfo_StorageInformation proto.InternalMessageInfo

func (m *StorageInformation) GetFsType() string {
	if m != nil {
		return m.FsType
	}
	return ""
}

func (m *StorageInformation) GetTotal() uint64 {
	if m != nil {
		return m.Total
	}
	return 0
}

func (m *StorageInformation) GetFree() uint64 {
	if m != nil {
		return m.Free
	}
	return 0
}

func (m *StorageInformation) GetUsed() uint64 {
	if m != nil {
		return m.Used
	}
	return 0
}

func (m *StorageInformation) GetUsedPercent() float64 {
	if m != nil {
		return m.UsedPercent
	}
	return 0
}

func init() {
	proto.RegisterType((*GetNodeHardwareResponse)(nil), "model.GetNodeHardwareResponse")
	proto.RegisterType((*GetNodeHardwareRequest)(nil), "model.GetNodeHardwareRequest")
	proto.RegisterType((*GetNodeTimeResponse)(nil), "model.GetNodeTimeResponse")
	proto.RegisterType((*NodeHardware)(nil), "model.NodeHardware")
	proto.RegisterType((*CPUInformation)(nil), "model.CPUInformation")
	proto.RegisterType((*HostInformation)(nil), "model.HostInformation")
	proto.RegisterType((*MemoryInformation)(nil), "model.MemoryInformation")
	proto.RegisterType((*StorageInformation)(nil), "model.StorageInformation")
}

func init() {
	proto.RegisterFile("model/nodeHardware.proto", fileDescriptor_ce967b26533b171c)
}

var fileDescriptor_ce967b26533b171c = []byte{
	// 607 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x54, 0xcf, 0x8f, 0xd2, 0x40,
	0x14, 0x4e, 0x4b, 0x41, 0x3a, 0x18, 0x56, 0x67, 0x15, 0xab, 0x31, 0x86, 0xf4, 0x60, 0x88, 0x89,
	0x60, 0xd6, 0xa8, 0x27, 0x13, 0x5d, 0x36, 0xb8, 0x1c, 0xf8, 0x91, 0x01, 0xf6, 0xe0, 0xad, 0xb4,
	0x6f, 0xa1, 0x49, 0xdb, 0xc1, 0x69, 0x51, 0xe1, 0x6f, 0xf0, 0x62, 0xf4, 0x2f, 0xf0, 0x2f, 0x35,
	0xf3, 0x3a, 0x40, 0x69, 0xf1, 0xe2, 0x85, 0xbc, 0xef, 0xbd, 0x79, 0xdf, 0x7c, 0xef, 0x7b, 0x1d,
	0x88, 0x15, 0x72, 0x0f, 0x82, 0x4e, 0xc4, 0x3d, 0xb8, 0x76, 0x84, 0xf7, 0xcd, 0x11, 0xd0, 0x5e,
	0x09, 0x9e, 0x70, 0x5a, 0xc6, 0x8a, 0xcd, 0xc8, 0xa3, 0x4f, 0x90, 0x0c, 0x33, 0x75, 0x06, 0xf1,
	0x8a, 0x47, 0x31, 0xd0, 0x77, 0xe4, 0x6e, 0x36, 0x6f, 0x69, 0x4d, 0xad, 0x55, 0xbb, 0x38, 0x6f,
	0x63, 0x63, 0xfb, 0xa8, 0xe5, 0xe8, 0xa0, 0x6d, 0x91, 0x46, 0x81, 0xf3, 0xcb, 0x1a, 0xe2, 0xc4,
	0x7e, 0x43, 0xce, 0x55, 0x65, 0xea, 0x87, 0x87, 0x9b, 0x9e, 0x91, 0xea, 0x2e, 0x87, 0xb7, 0x94,
	0x2e, 0xf5, 0x57, 0x1a, 0xdb, 0xe7, 0xec, 0x3f, 0xfa, 0xb1, 0x14, 0xfa, 0x9e, 0xd4, 0xbb, 0xe3,
	0x59, 0x3f, 0xba, 0xe5, 0x22, 0x74, 0x12, 0x9f, 0x47, 0x96, 0xd6, 0x2c, 0xb5, 0x6a, 0x17, 0x0f,
	0x95, 0xb8, 0xe3, 0x22, 0xcb, 0x1d, 0xa6, 0x3d, 0x72, 0x7f, 0x00, 0x21, 0x17, 0x9b, 0x2c, 0x83,
	0x8e, 0xe3, 0x59, 0x8a, 0xa1, 0x50, 0x67, 0xc5, 0x16, 0xda, 0x27, 0x74, 0x92, 0x70, 0xe1, 0x2c,
	0x20, 0x4b, 0x54, 0x42, 0xa2, 0xc7, 0x8a, 0xa8, 0x78, 0x80, 0x9d, 0x68, 0xa2, 0x1f, 0xc8, 0xd9,
	0x35, 0x8f, 0x93, 0x2c, 0x8f, 0x81, 0x3c, 0x0d, 0xc5, 0x93, 0xab, 0xb2, 0xfc, 0x71, 0xfb, 0xb7,
	0x9e, 0x37, 0x85, 0x36, 0x48, 0xa5, 0xe7, 0x84, 0x7e, 0xb0, 0x41, 0x57, 0x4d, 0xa6, 0x10, 0x7d,
	0x42, 0xaa, 0x78, 0xd2, 0x83, 0xef, 0x38, 0x76, 0x99, 0xed, 0x31, 0x7d, 0x40, 0xca, 0x03, 0x79,
	0x21, 0x8e, 0x61, 0xb2, 0x14, 0xd0, 0xa7, 0xc4, 0xc4, 0x60, 0xe8, 0x84, 0x80, 0xc2, 0x4c, 0x76,
	0x48, 0x48, 0xbe, 0x1b, 0x88, 0x3c, 0x2e, 0xfa, 0x9e, 0x55, 0xc6, 0xe2, 0x1e, 0xd3, 0x7b, 0xa4,
	0x34, 0x58, 0x6e, 0xad, 0x4a, 0x53, 0x6b, 0x69, 0x4c, 0x86, 0x92, 0xab, 0xeb, 0xb8, 0x4b, 0x98,
	0xf8, 0x5b, 0xb0, 0xee, 0xe0, 0xf5, 0x87, 0x04, 0x6d, 0x92, 0xda, 0x2c, 0x06, 0x6f, 0x0c, 0xc2,
	0x85, 0x28, 0xb1, 0xaa, 0xd8, 0x97, 0x4d, 0xc9, 0xa9, 0xba, 0x5c, 0x40, 0xff, 0xca, 0x32, 0xd3,
	0xa9, 0x52, 0x24, 0x95, 0xcb, 0x28, 0xb6, 0x08, 0x72, 0xa6, 0xc0, 0xfe, 0xa5, 0x17, 0x9c, 0x95,
	0x0c, 0xb3, 0x55, 0xb2, 0xfb, 0xda, 0x0c, 0xa6, 0x10, 0xad, 0x13, 0x7d, 0x34, 0x41, 0x47, 0x4c,
	0xa6, 0x8f, 0x26, 0x72, 0xae, 0x71, 0xe0, 0x24, 0xb2, 0x51, 0xd9, 0xb1, 0xc7, 0xf4, 0x39, 0xa9,
	0xef, 0x62, 0xe5, 0x71, 0x6a, 0x4b, 0x2e, 0x4b, 0x5b, 0xe4, 0x6c, 0x97, 0xb9, 0x01, 0x11, 0xcb,
	0xc5, 0xa6, 0x16, 0xe5, 0xd3, 0xf4, 0x2d, 0x69, 0x0c, 0xd7, 0xe1, 0x1c, 0xc4, 0xe8, 0x96, 0xad,
	0xa3, 0xc8, 0x8f, 0x16, 0x63, 0xc1, 0x5d, 0x88, 0x63, 0x34, 0xcf, 0x60, 0xff, 0xa8, 0xca, 0x69,
	0x70, 0xc0, 0x2b, 0x34, 0xd3, 0x64, 0x0a, 0x49, 0xf5, 0x32, 0xc2, 0x95, 0x55, 0x53, 0xf5, 0x3b,
	0x6c, 0xff, 0xd4, 0x4e, 0x3c, 0x01, 0xe9, 0xe0, 0x94, 0x27, 0x4e, 0xa0, 0x6c, 0x49, 0x01, 0xa5,
	0xc4, 0xe8, 0x09, 0x00, 0xf4, 0xc5, 0x60, 0x18, 0xcb, 0x1d, 0x7e, 0xfc, 0xea, 0xf8, 0x81, 0x33,
	0x0f, 0x00, 0xad, 0x31, 0xd8, 0x21, 0x21, 0x3b, 0xe4, 0xc2, 0xd0, 0x11, 0x83, 0x61, 0x9c, 0xdf,
	0x6b, 0xb9, 0xb0, 0x57, 0xfb, 0x87, 0x76, 0xea, 0x39, 0xe1, 0x47, 0x1c, 0x4f, 0x37, 0x2b, 0xd8,
	0x7f, 0xc4, 0x88, 0x0e, 0x62, 0xf5, 0x53, 0x62, 0x4b, 0x19, 0xb1, 0xff, 0x25, 0xe7, 0xf2, 0xc5,
	0xe7, 0xd6, 0xc2, 0x4f, 0x96, 0xeb, 0x79, 0xdb, 0xe5, 0x61, 0x67, 0xcb, 0xf9, 0xdc, 0x4d, 0x7f,
	0x5f, 0xba, 0x5c, 0x40, 0xc7, 0xe5, 0x61, 0xc8, 0xa3, 0x0e, 0x3e, 0xce, 0x79, 0x05, 0xff, 0x53,
	0x5f, 0xff, 0x0d, 0x00, 0x00, 0xff, 0xff, 0x63, 0xf2, 0x0c, 0x2d, 0x6f, 0x05, 0x00, 0x00,
}
