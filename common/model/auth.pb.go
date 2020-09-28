// Code generated by protoc-gen-go. DO NOT EDIT.
// source: model/auth.proto

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

// RequestType used to sign a node administration request
type RequestType int32

const (
	RequestType_GetNodeHardware                   RequestType = 0
	RequestType_GetProofOfOwnership               RequestType = 1
	RequestType_GeneratetNodeKey                  RequestType = 2
	RequestType_GetPendingNodeRegistrationsStream RequestType = 3
)

var RequestType_name = map[int32]string{
	0: "GetNodeHardware",
	1: "GetProofOfOwnership",
	2: "GeneratetNodeKey",
	3: "GetPendingNodeRegistrationsStream",
}

var RequestType_value = map[string]int32{
	"GetNodeHardware":                   0,
	"GetProofOfOwnership":               1,
	"GeneratetNodeKey":                  2,
	"GetPendingNodeRegistrationsStream": 3,
}

func (x RequestType) String() string {
	return proto.EnumName(RequestType_name, int32(x))
}

func (RequestType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_50b87fa4d1f0d790, []int{0}
}

func init() {
	proto.RegisterEnum("model.RequestType", RequestType_name, RequestType_value)
}

func init() { proto.RegisterFile("model/auth.proto", fileDescriptor_50b87fa4d1f0d790) }

var fileDescriptor_50b87fa4d1f0d790 = []byte{
	// 194 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x24, 0x8e, 0xbf, 0x4b, 0xc5, 0x30,
	0x14, 0x85, 0xfd, 0x81, 0x0e, 0x71, 0x30, 0xe4, 0x09, 0xce, 0x0e, 0x82, 0x3c, 0xf0, 0x65, 0xf0,
	0x3f, 0x70, 0xa9, 0x20, 0xf8, 0xa4, 0x3a, 0xb9, 0xa5, 0xc9, 0x69, 0x1b, 0x30, 0xb9, 0xf5, 0xe6,
	0x96, 0x5a, 0xff, 0x7a, 0x69, 0xbb, 0x9c, 0xe1, 0xe3, 0x7c, 0xf0, 0x29, 0x9d, 0x28, 0xe0, 0xdb,
	0xba, 0x51, 0xfa, 0xc3, 0xc0, 0x24, 0x64, 0x2e, 0x56, 0xb2, 0xff, 0x55, 0x57, 0x35, 0x7e, 0x46,
	0x14, 0xf9, 0x9c, 0x07, 0x98, 0x9d, 0xba, 0xae, 0x20, 0x6f, 0x14, 0xf0, 0xe2, 0x38, 0x4c, 0x8e,
	0xa1, 0x4f, 0xcc, 0xad, 0xda, 0x55, 0x90, 0x77, 0x26, 0x6a, 0x8f, 0xed, 0x71, 0xca, 0xe0, 0xd2,
	0xc7, 0x41, 0x9f, 0x9a, 0x1b, 0xa5, 0x2b, 0x64, 0xb0, 0x93, 0xcd, 0x79, 0xc5, 0xac, 0xcf, 0xcc,
	0xbd, 0xba, 0x5b, 0xee, 0xc8, 0x21, 0xe6, 0x6e, 0xc1, 0x35, 0xba, 0x58, 0x84, 0x9d, 0x44, 0xca,
	0xe5, 0x43, 0x18, 0x2e, 0xe9, 0xf3, 0xe7, 0xfd, 0xd7, 0x43, 0x17, 0xa5, 0x1f, 0x9b, 0x83, 0xa7,
	0x64, 0xff, 0x88, 0x1a, 0xbf, 0xed, 0xa3, 0x27, 0x86, 0xf5, 0x94, 0x12, 0x65, 0xbb, 0x56, 0x36,
	0x97, 0x6b, 0xf3, 0xd3, 0x7f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x55, 0x12, 0x2e, 0x13, 0xc7, 0x00,
	0x00, 0x00,
}
