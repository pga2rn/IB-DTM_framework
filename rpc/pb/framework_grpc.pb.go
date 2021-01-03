// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// FrameworkStatisticsQueryClient is the client API for FrameworkStatisticsQuery service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FrameworkStatisticsQueryClient interface {
	GetLatestData(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*StatisticsBundle, error)
	GetDataForEpoch(ctx context.Context, in *QueryEpoch, opts ...grpc.CallOption) (*StatisticsBundle, error)
	EchoEpoch(ctx context.Context, in *QueryEpoch, opts ...grpc.CallOption) (*QueryEpoch, error)
}

type frameworkStatisticsQueryClient struct {
	cc grpc.ClientConnInterface
}

func NewFrameworkStatisticsQueryClient(cc grpc.ClientConnInterface) FrameworkStatisticsQueryClient {
	return &frameworkStatisticsQueryClient{cc}
}

func (c *frameworkStatisticsQueryClient) GetLatestData(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*StatisticsBundle, error) {
	out := new(StatisticsBundle)
	err := c.cc.Invoke(ctx, "/rpc.pb.FrameworkStatisticsQuery/GetLatestData", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *frameworkStatisticsQueryClient) GetDataForEpoch(ctx context.Context, in *QueryEpoch, opts ...grpc.CallOption) (*StatisticsBundle, error) {
	out := new(StatisticsBundle)
	err := c.cc.Invoke(ctx, "/rpc.pb.FrameworkStatisticsQuery/GetDataForEpoch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *frameworkStatisticsQueryClient) EchoEpoch(ctx context.Context, in *QueryEpoch, opts ...grpc.CallOption) (*QueryEpoch, error) {
	out := new(QueryEpoch)
	err := c.cc.Invoke(ctx, "/rpc.pb.FrameworkStatisticsQuery/EchoEpoch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FrameworkStatisticsQueryServer is the server API for FrameworkStatisticsQuery service.
// All implementations must embed UnimplementedFrameworkStatisticsQueryServer
// for forward compatibility
type FrameworkStatisticsQueryServer interface {
	GetLatestData(context.Context, *emptypb.Empty) (*StatisticsBundle, error)
	GetDataForEpoch(context.Context, *QueryEpoch) (*StatisticsBundle, error)
	EchoEpoch(context.Context, *QueryEpoch) (*QueryEpoch, error)
	mustEmbedUnimplementedFrameworkStatisticsQueryServer()
}

// UnimplementedFrameworkStatisticsQueryServer must be embedded to have forward compatible implementations.
type UnimplementedFrameworkStatisticsQueryServer struct {
}

func (UnimplementedFrameworkStatisticsQueryServer) GetLatestData(context.Context, *emptypb.Empty) (*StatisticsBundle, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLatestData not implemented")
}
func (UnimplementedFrameworkStatisticsQueryServer) GetDataForEpoch(context.Context, *QueryEpoch) (*StatisticsBundle, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDataForEpoch not implemented")
}
func (UnimplementedFrameworkStatisticsQueryServer) EchoEpoch(context.Context, *QueryEpoch) (*QueryEpoch, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EchoEpoch not implemented")
}
func (UnimplementedFrameworkStatisticsQueryServer) mustEmbedUnimplementedFrameworkStatisticsQueryServer() {
}

// UnsafeFrameworkStatisticsQueryServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FrameworkStatisticsQueryServer will
// result in compilation errors.
type UnsafeFrameworkStatisticsQueryServer interface {
	mustEmbedUnimplementedFrameworkStatisticsQueryServer()
}

func RegisterFrameworkStatisticsQueryServer(s grpc.ServiceRegistrar, srv FrameworkStatisticsQueryServer) {
	s.RegisterService(&_FrameworkStatisticsQuery_serviceDesc, srv)
}

func _FrameworkStatisticsQuery_GetLatestData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FrameworkStatisticsQueryServer).GetLatestData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.pb.FrameworkStatisticsQuery/GetLatestData",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FrameworkStatisticsQueryServer).GetLatestData(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _FrameworkStatisticsQuery_GetDataForEpoch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryEpoch)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FrameworkStatisticsQueryServer).GetDataForEpoch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.pb.FrameworkStatisticsQuery/GetDataForEpoch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FrameworkStatisticsQueryServer).GetDataForEpoch(ctx, req.(*QueryEpoch))
	}
	return interceptor(ctx, in, info, handler)
}

func _FrameworkStatisticsQuery_EchoEpoch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryEpoch)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FrameworkStatisticsQueryServer).EchoEpoch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.pb.FrameworkStatisticsQuery/EchoEpoch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FrameworkStatisticsQueryServer).EchoEpoch(ctx, req.(*QueryEpoch))
	}
	return interceptor(ctx, in, info, handler)
}

var _FrameworkStatisticsQuery_serviceDesc = grpc.ServiceDesc{
	ServiceName: "rpc.pb.FrameworkStatisticsQuery",
	HandlerType: (*FrameworkStatisticsQueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetLatestData",
			Handler:    _FrameworkStatisticsQuery_GetLatestData_Handler,
		},
		{
			MethodName: "GetDataForEpoch",
			Handler:    _FrameworkStatisticsQuery_GetDataForEpoch_Handler,
		},
		{
			MethodName: "EchoEpoch",
			Handler:    _FrameworkStatisticsQuery_EchoEpoch_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pb/framework.proto",
}
