// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: proto/metrics.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	MetricsService_SendMetrics_FullMethodName = "/metrics.MetricsService/SendMetrics"
	MetricsService_GetMetrics_FullMethodName  = "/metrics.MetricsService/GetMetrics"
)

// MetricsServiceClient is the client API for MetricsService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MetricsServiceClient interface {
	// Отправка метрик
	SendMetrics(ctx context.Context, in *MetricsRequest, opts ...grpc.CallOption) (*SendMetricsResponse, error)
	// Запрос метрик
	GetMetrics(ctx context.Context, in *GetMetricsRequest, opts ...grpc.CallOption) (*GetMetricsResponse, error)
}

type metricsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMetricsServiceClient(cc grpc.ClientConnInterface) MetricsServiceClient {
	return &metricsServiceClient{cc}
}

func (c *metricsServiceClient) SendMetrics(ctx context.Context, in *MetricsRequest, opts ...grpc.CallOption) (*SendMetricsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SendMetricsResponse)
	err := c.cc.Invoke(ctx, MetricsService_SendMetrics_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricsServiceClient) GetMetrics(ctx context.Context, in *GetMetricsRequest, opts ...grpc.CallOption) (*GetMetricsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetMetricsResponse)
	err := c.cc.Invoke(ctx, MetricsService_GetMetrics_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MetricsServiceServer is the server API for MetricsService service.
// All implementations must embed UnimplementedMetricsServiceServer
// for forward compatibility.
type MetricsServiceServer interface {
	// Отправка метрик
	SendMetrics(context.Context, *MetricsRequest) (*SendMetricsResponse, error)
	// Запрос метрик
	GetMetrics(context.Context, *GetMetricsRequest) (*GetMetricsResponse, error)
	mustEmbedUnimplementedMetricsServiceServer()
}

// UnimplementedMetricsServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedMetricsServiceServer struct{}

func (UnimplementedMetricsServiceServer) SendMetrics(context.Context, *MetricsRequest) (*SendMetricsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMetrics not implemented")
}
func (UnimplementedMetricsServiceServer) GetMetrics(context.Context, *GetMetricsRequest) (*GetMetricsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMetrics not implemented")
}
func (UnimplementedMetricsServiceServer) mustEmbedUnimplementedMetricsServiceServer() {}
func (UnimplementedMetricsServiceServer) testEmbeddedByValue()                        {}

// UnsafeMetricsServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MetricsServiceServer will
// result in compilation errors.
type UnsafeMetricsServiceServer interface {
	mustEmbedUnimplementedMetricsServiceServer()
}

func RegisterMetricsServiceServer(s grpc.ServiceRegistrar, srv MetricsServiceServer) {
	// If the following call pancis, it indicates UnimplementedMetricsServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&MetricsService_ServiceDesc, srv)
}

func _MetricsService_SendMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServiceServer).SendMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetricsService_SendMetrics_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServiceServer).SendMetrics(ctx, req.(*MetricsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricsService_GetMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMetricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServiceServer).GetMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetricsService_GetMetrics_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServiceServer).GetMetrics(ctx, req.(*GetMetricsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MetricsService_ServiceDesc is the grpc.ServiceDesc for MetricsService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MetricsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "metrics.MetricsService",
	HandlerType: (*MetricsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendMetrics",
			Handler:    _MetricsService_SendMetrics_Handler,
		},
		{
			MethodName: "GetMetrics",
			Handler:    _MetricsService_GetMetrics_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/metrics.proto",
}
