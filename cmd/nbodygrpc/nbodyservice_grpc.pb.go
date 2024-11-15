// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.12.4
// source: nbodyservice.proto

package nbodygrpc

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	NBodyService_SetComputationThreads_FullMethodName     = "/NBodyService/SetComputationThreads"
	NBodyService_SetResultQueueSize_FullMethodName        = "/NBodyService/SetResultQueueSize"
	NBodyService_SetSmoothing_FullMethodName              = "/NBodyService/SetSmoothing"
	NBodyService_SetRestitutionCoefficient_FullMethodName = "/NBodyService/SetRestitutionCoefficient"
	NBodyService_RemoveBodies_FullMethodName              = "/NBodyService/RemoveBodies"
	NBodyService_AddBody_FullMethodName                   = "/NBodyService/AddBody"
	NBodyService_ModBody_FullMethodName                   = "/NBodyService/ModBody"
	NBodyService_GetBody_FullMethodName                   = "/NBodyService/GetBody"
	NBodyService_GetCurrentConfig_FullMethodName          = "/NBodyService/GetCurrentConfig"
)

// NBodyServiceClient is the client API for NBodyService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// Defines a gRPC Service that enables entities external to the running sim to view / modify
// simulation configurables, thus changing the behavior of the simulation on the fly
type NBodyServiceClient interface {
	// Sets the number of threads allocated to computing the body positions
	// (The render engine threading model is not modifiable at this time)
	SetComputationThreads(ctx context.Context, in *ItemCount, opts ...grpc.CallOption) (*ResultCode, error)
	// Sets the number of compute-ahead results allowed, in cases where the computation
	// thread outruns the render thread
	SetResultQueueSize(ctx context.Context, in *ItemCount, opts ...grpc.CallOption) (*ResultCode, error)
	// Changes the smoothing factor. When the body force and position computation runs
	// during each compute cycle, the force and resulting motion of the bodies is
	// smoothed by a factor which can be changed using this RPC method. The result is
	// that the apparent motion of the simulation is faster or slower
	SetSmoothing(ctx context.Context, in *Factor, opts ...grpc.CallOption) (*ResultCode, error)
	// Sets the coefficient of restitution for elastic collisions
	SetRestitutionCoefficient(ctx context.Context, in *RestitutionCoefficient, opts ...grpc.CallOption) (*ResultCode, error)
	// Removes the specified number of bodies from the sim
	RemoveBodies(ctx context.Context, in *ItemCount, opts ...grpc.CallOption) (*ResultCode, error)
	// Adds a body into the simulation
	AddBody(ctx context.Context, in *BodyDescription, opts ...grpc.CallOption) (*ResultCode, error)
	// Modifies body properties
	ModBody(ctx context.Context, in *ModBodyMessage, opts ...grpc.CallOption) (*ResultCode, error)
	// Gets body properties (use ModBodyMessage and ignore what is not needed)
	GetBody(ctx context.Context, in *ModBodyMessage, opts ...grpc.CallOption) (*BodyDescription, error)
	// Gets the current values of sim configurables
	GetCurrentConfig(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*CurrentConfig, error)
}

type nBodyServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNBodyServiceClient(cc grpc.ClientConnInterface) NBodyServiceClient {
	return &nBodyServiceClient{cc}
}

func (c *nBodyServiceClient) SetComputationThreads(ctx context.Context, in *ItemCount, opts ...grpc.CallOption) (*ResultCode, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ResultCode)
	err := c.cc.Invoke(ctx, NBodyService_SetComputationThreads_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nBodyServiceClient) SetResultQueueSize(ctx context.Context, in *ItemCount, opts ...grpc.CallOption) (*ResultCode, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ResultCode)
	err := c.cc.Invoke(ctx, NBodyService_SetResultQueueSize_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nBodyServiceClient) SetSmoothing(ctx context.Context, in *Factor, opts ...grpc.CallOption) (*ResultCode, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ResultCode)
	err := c.cc.Invoke(ctx, NBodyService_SetSmoothing_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nBodyServiceClient) SetRestitutionCoefficient(ctx context.Context, in *RestitutionCoefficient, opts ...grpc.CallOption) (*ResultCode, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ResultCode)
	err := c.cc.Invoke(ctx, NBodyService_SetRestitutionCoefficient_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nBodyServiceClient) RemoveBodies(ctx context.Context, in *ItemCount, opts ...grpc.CallOption) (*ResultCode, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ResultCode)
	err := c.cc.Invoke(ctx, NBodyService_RemoveBodies_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nBodyServiceClient) AddBody(ctx context.Context, in *BodyDescription, opts ...grpc.CallOption) (*ResultCode, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ResultCode)
	err := c.cc.Invoke(ctx, NBodyService_AddBody_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nBodyServiceClient) ModBody(ctx context.Context, in *ModBodyMessage, opts ...grpc.CallOption) (*ResultCode, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ResultCode)
	err := c.cc.Invoke(ctx, NBodyService_ModBody_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nBodyServiceClient) GetBody(ctx context.Context, in *ModBodyMessage, opts ...grpc.CallOption) (*BodyDescription, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BodyDescription)
	err := c.cc.Invoke(ctx, NBodyService_GetBody_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nBodyServiceClient) GetCurrentConfig(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*CurrentConfig, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CurrentConfig)
	err := c.cc.Invoke(ctx, NBodyService_GetCurrentConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NBodyServiceServer is the server API for NBodyService service.
// All implementations must embed UnimplementedNBodyServiceServer
// for forward compatibility.
//
// Defines a gRPC Service that enables entities external to the running sim to view / modify
// simulation configurables, thus changing the behavior of the simulation on the fly
type NBodyServiceServer interface {
	// Sets the number of threads allocated to computing the body positions
	// (The render engine threading model is not modifiable at this time)
	SetComputationThreads(context.Context, *ItemCount) (*ResultCode, error)
	// Sets the number of compute-ahead results allowed, in cases where the computation
	// thread outruns the render thread
	SetResultQueueSize(context.Context, *ItemCount) (*ResultCode, error)
	// Changes the smoothing factor. When the body force and position computation runs
	// during each compute cycle, the force and resulting motion of the bodies is
	// smoothed by a factor which can be changed using this RPC method. The result is
	// that the apparent motion of the simulation is faster or slower
	SetSmoothing(context.Context, *Factor) (*ResultCode, error)
	// Sets the coefficient of restitution for elastic collisions
	SetRestitutionCoefficient(context.Context, *RestitutionCoefficient) (*ResultCode, error)
	// Removes the specified number of bodies from the sim
	RemoveBodies(context.Context, *ItemCount) (*ResultCode, error)
	// Adds a body into the simulation
	AddBody(context.Context, *BodyDescription) (*ResultCode, error)
	// Modifies body properties
	ModBody(context.Context, *ModBodyMessage) (*ResultCode, error)
	// Gets body properties (use ModBodyMessage and ignore what is not needed)
	GetBody(context.Context, *ModBodyMessage) (*BodyDescription, error)
	// Gets the current values of sim configurables
	GetCurrentConfig(context.Context, *empty.Empty) (*CurrentConfig, error)
	mustEmbedUnimplementedNBodyServiceServer()
}

// UnimplementedNBodyServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedNBodyServiceServer struct{}

func (UnimplementedNBodyServiceServer) SetComputationThreads(context.Context, *ItemCount) (*ResultCode, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetComputationThreads not implemented")
}
func (UnimplementedNBodyServiceServer) SetResultQueueSize(context.Context, *ItemCount) (*ResultCode, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetResultQueueSize not implemented")
}
func (UnimplementedNBodyServiceServer) SetSmoothing(context.Context, *Factor) (*ResultCode, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetSmoothing not implemented")
}
func (UnimplementedNBodyServiceServer) SetRestitutionCoefficient(context.Context, *RestitutionCoefficient) (*ResultCode, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetRestitutionCoefficient not implemented")
}
func (UnimplementedNBodyServiceServer) RemoveBodies(context.Context, *ItemCount) (*ResultCode, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveBodies not implemented")
}
func (UnimplementedNBodyServiceServer) AddBody(context.Context, *BodyDescription) (*ResultCode, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddBody not implemented")
}
func (UnimplementedNBodyServiceServer) ModBody(context.Context, *ModBodyMessage) (*ResultCode, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ModBody not implemented")
}
func (UnimplementedNBodyServiceServer) GetBody(context.Context, *ModBodyMessage) (*BodyDescription, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBody not implemented")
}
func (UnimplementedNBodyServiceServer) GetCurrentConfig(context.Context, *empty.Empty) (*CurrentConfig, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCurrentConfig not implemented")
}
func (UnimplementedNBodyServiceServer) mustEmbedUnimplementedNBodyServiceServer() {}
func (UnimplementedNBodyServiceServer) testEmbeddedByValue()                      {}

// UnsafeNBodyServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NBodyServiceServer will
// result in compilation errors.
type UnsafeNBodyServiceServer interface {
	mustEmbedUnimplementedNBodyServiceServer()
}

func RegisterNBodyServiceServer(s grpc.ServiceRegistrar, srv NBodyServiceServer) {
	// If the following call pancis, it indicates UnimplementedNBodyServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&NBodyService_ServiceDesc, srv)
}

func _NBodyService_SetComputationThreads_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ItemCount)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NBodyServiceServer).SetComputationThreads(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NBodyService_SetComputationThreads_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NBodyServiceServer).SetComputationThreads(ctx, req.(*ItemCount))
	}
	return interceptor(ctx, in, info, handler)
}

func _NBodyService_SetResultQueueSize_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ItemCount)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NBodyServiceServer).SetResultQueueSize(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NBodyService_SetResultQueueSize_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NBodyServiceServer).SetResultQueueSize(ctx, req.(*ItemCount))
	}
	return interceptor(ctx, in, info, handler)
}

func _NBodyService_SetSmoothing_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Factor)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NBodyServiceServer).SetSmoothing(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NBodyService_SetSmoothing_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NBodyServiceServer).SetSmoothing(ctx, req.(*Factor))
	}
	return interceptor(ctx, in, info, handler)
}

func _NBodyService_SetRestitutionCoefficient_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RestitutionCoefficient)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NBodyServiceServer).SetRestitutionCoefficient(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NBodyService_SetRestitutionCoefficient_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NBodyServiceServer).SetRestitutionCoefficient(ctx, req.(*RestitutionCoefficient))
	}
	return interceptor(ctx, in, info, handler)
}

func _NBodyService_RemoveBodies_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ItemCount)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NBodyServiceServer).RemoveBodies(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NBodyService_RemoveBodies_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NBodyServiceServer).RemoveBodies(ctx, req.(*ItemCount))
	}
	return interceptor(ctx, in, info, handler)
}

func _NBodyService_AddBody_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BodyDescription)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NBodyServiceServer).AddBody(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NBodyService_AddBody_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NBodyServiceServer).AddBody(ctx, req.(*BodyDescription))
	}
	return interceptor(ctx, in, info, handler)
}

func _NBodyService_ModBody_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ModBodyMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NBodyServiceServer).ModBody(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NBodyService_ModBody_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NBodyServiceServer).ModBody(ctx, req.(*ModBodyMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _NBodyService_GetBody_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ModBodyMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NBodyServiceServer).GetBody(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NBodyService_GetBody_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NBodyServiceServer).GetBody(ctx, req.(*ModBodyMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _NBodyService_GetCurrentConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NBodyServiceServer).GetCurrentConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NBodyService_GetCurrentConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NBodyServiceServer).GetCurrentConfig(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// NBodyService_ServiceDesc is the grpc.ServiceDesc for NBodyService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var NBodyService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "NBodyService",
	HandlerType: (*NBodyServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetComputationThreads",
			Handler:    _NBodyService_SetComputationThreads_Handler,
		},
		{
			MethodName: "SetResultQueueSize",
			Handler:    _NBodyService_SetResultQueueSize_Handler,
		},
		{
			MethodName: "SetSmoothing",
			Handler:    _NBodyService_SetSmoothing_Handler,
		},
		{
			MethodName: "SetRestitutionCoefficient",
			Handler:    _NBodyService_SetRestitutionCoefficient_Handler,
		},
		{
			MethodName: "RemoveBodies",
			Handler:    _NBodyService_RemoveBodies_Handler,
		},
		{
			MethodName: "AddBody",
			Handler:    _NBodyService_AddBody_Handler,
		},
		{
			MethodName: "ModBody",
			Handler:    _NBodyService_ModBody_Handler,
		},
		{
			MethodName: "GetBody",
			Handler:    _NBodyService_GetBody_Handler,
		},
		{
			MethodName: "GetCurrentConfig",
			Handler:    _NBodyService_GetCurrentConfig_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "nbodyservice.proto",
}
