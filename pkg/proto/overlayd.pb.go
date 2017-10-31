// Code generated by protoc-gen-go. DO NOT EDIT.
// source: overlayd.proto

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	overlayd.proto

It has these top-level messages:
	OverlaydCommand
	OverlaydCommandResponse
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto1.ProtoPackageIsVersion2 // please upgrade the proto package

// OverlaydCommand represents a command for i3x3, mirroring the properties relevant for
// i3x3overlayd, the command is used to move around the grid overlay appropriately.
type OverlaydCommand struct {
	Target uint32 `protobuf:"varint,1,opt,name=target" json:"target,omitempty"`
}

func (m *OverlaydCommand) Reset()                    { *m = OverlaydCommand{} }
func (m *OverlaydCommand) String() string            { return proto1.CompactTextString(m) }
func (*OverlaydCommand) ProtoMessage()               {}
func (*OverlaydCommand) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *OverlaydCommand) GetTarget() uint32 {
	if m != nil {
		return m.Target
	}
	return 0
}

// OverlaydCommandResponse represents the result of a command for i3x3overlayd.
type OverlaydCommandResponse struct {
	Success bool `protobuf:"varint,1,opt,name=success" json:"success,omitempty"`
}

func (m *OverlaydCommandResponse) Reset()                    { *m = OverlaydCommandResponse{} }
func (m *OverlaydCommandResponse) String() string            { return proto1.CompactTextString(m) }
func (*OverlaydCommandResponse) ProtoMessage()               {}
func (*OverlaydCommandResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *OverlaydCommandResponse) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func init() {
	proto1.RegisterType((*OverlaydCommand)(nil), "proto.OverlaydCommand")
	proto1.RegisterType((*OverlaydCommandResponse)(nil), "proto.OverlaydCommandResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for OverlaydServer service

type OverlaydServerClient interface {
	SendCommand(ctx context.Context, in *OverlaydCommand, opts ...grpc.CallOption) (*OverlaydCommandResponse, error)
}

type overlaydServerClient struct {
	cc *grpc.ClientConn
}

func NewOverlaydServerClient(cc *grpc.ClientConn) OverlaydServerClient {
	return &overlaydServerClient{cc}
}

func (c *overlaydServerClient) SendCommand(ctx context.Context, in *OverlaydCommand, opts ...grpc.CallOption) (*OverlaydCommandResponse, error) {
	out := new(OverlaydCommandResponse)
	err := grpc.Invoke(ctx, "/proto.OverlaydServer/SendCommand", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for OverlaydServer service

type OverlaydServerServer interface {
	SendCommand(context.Context, *OverlaydCommand) (*OverlaydCommandResponse, error)
}

func RegisterOverlaydServerServer(s *grpc.Server, srv OverlaydServerServer) {
	s.RegisterService(&_OverlaydServer_serviceDesc, srv)
}

func _OverlaydServer_SendCommand_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OverlaydCommand)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OverlaydServerServer).SendCommand(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.OverlaydServer/SendCommand",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OverlaydServerServer).SendCommand(ctx, req.(*OverlaydCommand))
	}
	return interceptor(ctx, in, info, handler)
}

var _OverlaydServer_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.OverlaydServer",
	HandlerType: (*OverlaydServerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendCommand",
			Handler:    _OverlaydServer_SendCommand_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "overlayd.proto",
}

func init() { proto1.RegisterFile("overlayd.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 145 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xcb, 0x2f, 0x4b, 0x2d,
	0xca, 0x49, 0xac, 0x4c, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x05, 0x53, 0x4a, 0x9a,
	0x5c, 0xfc, 0xfe, 0x50, 0x09, 0xe7, 0xfc, 0xdc, 0xdc, 0xc4, 0xbc, 0x14, 0x21, 0x31, 0x2e, 0xb6,
	0x92, 0xc4, 0xa2, 0xf4, 0xd4, 0x12, 0x09, 0x46, 0x05, 0x46, 0x0d, 0xde, 0x20, 0x28, 0x4f, 0xc9,
	0x98, 0x4b, 0x1c, 0x4d, 0x69, 0x50, 0x6a, 0x71, 0x41, 0x7e, 0x5e, 0x71, 0xaa, 0x90, 0x04, 0x17,
	0x7b, 0x71, 0x69, 0x72, 0x72, 0x6a, 0x71, 0x31, 0x58, 0x0f, 0x47, 0x10, 0x8c, 0x6b, 0x14, 0xce,
	0xc5, 0x07, 0xd3, 0x14, 0x9c, 0x5a, 0x54, 0x96, 0x5a, 0x24, 0xe4, 0xca, 0xc5, 0x1d, 0x9c, 0x9a,
	0x87, 0xb0, 0x0d, 0xe2, 0x1e, 0x3d, 0x34, 0xa3, 0xa5, 0xe4, 0xb0, 0x8b, 0xc3, 0xac, 0x4c, 0x62,
	0x03, 0x4b, 0x1b, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0x03, 0x1d, 0xbb, 0x92, 0xd8, 0x00, 0x00,
	0x00,
}
