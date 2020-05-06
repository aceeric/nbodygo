package grpcserver

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"nbodygo/cmd/body"
	"nbodygo/cmd/nbodygrpc"
)

//
// These functions are the implementation of the gRPC interface. They simply delegate everything to the
// callbacks in the passed nbodyServiceServer struct, and do some assembly/disassembly of data structures
// to mediate between the simulation and gRPC
//

func (s *nbodyServiceServer) GetCurrentConfig(_ context.Context, in *empty.Empty) (*nbodygrpc.CurrentConfig, error) {
	_ = in
	return &nbodygrpc.CurrentConfig{
		Bodies:                 int64(s.callbacks.BodyCount()),
		ResultQueueSize:        int64(s.callbacks.ResultQueueSize()),
		ComputationThreads:     int64(s.callbacks.ComputationWorkers()),
		SmoothingFactor:        s.callbacks.Smoothing(),
		RestitutionCoefficient: s.callbacks.RestitutionCoefficient(),
	}, nil
}

func (s *nbodyServiceServer) SetComputationThreads(_ context.Context, in *nbodygrpc.ItemCount) (*nbodygrpc.ResultCode, error) {
	s.callbacks.SetComputationWorkers(int(in.ItemCount))
	return &nbodygrpc.ResultCode{
		ResultCode: nbodygrpc.ResultCode_OK,
		Message:    "OK",
	}, nil
}

func (s *nbodyServiceServer) SetResultQueueSize(_ context.Context, in *nbodygrpc.ItemCount) (*nbodygrpc.ResultCode, error) {
	s.callbacks.SetResultQueueSize(int(in.ItemCount)) // maybe it couldn't be resized but we don't care
	return &nbodygrpc.ResultCode{
		ResultCode: nbodygrpc.ResultCode_OK,
		Message:    "OK",
	}, nil
}

func (s *nbodyServiceServer) SetSmoothing(_ context.Context, in *nbodygrpc.Factor) (*nbodygrpc.ResultCode, error) {
	s.callbacks.SetSmoothing(in.Factor)
	return &nbodygrpc.ResultCode{
		ResultCode: nbodygrpc.ResultCode_OK,
		Message:    "OK",
	}, nil
}

func (s *nbodyServiceServer) SetRestitutionCoefficient(_ context.Context, in *nbodygrpc.RestitutionCoefficient) (*nbodygrpc.ResultCode, error) {
	s.callbacks.SetRestitutionCoefficient(in.RestitutionCoefficient)
	return &nbodygrpc.ResultCode{
		ResultCode: nbodygrpc.ResultCode_OK,
		Message:    "OK",
	}, nil
}

func (s *nbodyServiceServer) RemoveBodies(_ context.Context, in *nbodygrpc.ItemCount) (*nbodygrpc.ResultCode, error) {
	s.callbacks.RemoveBodies(int(in.ItemCount))
	return &nbodygrpc.ResultCode{
		ResultCode: nbodygrpc.ResultCode_OK,
		Message:    "OK",
	}, nil
}

func (s *nbodyServiceServer) AddBody(_ context.Context, in *nbodygrpc.BodyDescription) (*nbodygrpc.ResultCode, error) {
	mass := in.Mass
	x := in.X
	y := in.Y
	z := in.Z
	vx := in.Vx
	vy := in.Vy
	vz := in.Vz
	radius := in.Radius
	isSun := in.IsSun
	intensity := in.Intensity
	fragFactor := in.FragFactor
	fragStep := in.FragStep
	withTelemetry := in.WithTelemetry
	name := in.Name
	class := in.Class
	pinned := in.Pinned
	behavior := nbodygrpc.GrpcCbToSimCb(in.CollisionBehavior)
	bodyColor := nbodygrpc.GrpcColorToSimColor(in.BodyColor)
	id := s.callbacks.AddBody(mass, x, y, z, vx, vy, vz, radius, isSun, intensity, behavior, bodyColor,
		fragFactor, fragStep, withTelemetry, name, class, pinned)
	return &nbodygrpc.ResultCode{
		ResultCode: nbodygrpc.ResultCode_OK,
		Message:    fmt.Sprintf("Added body ID: %d", id),
	}, nil
}

func (s *nbodyServiceServer) ModBody(_ context.Context, in *nbodygrpc.ModBodyMessage) (*nbodygrpc.ResultCode, error) {
	result := s.callbacks.ModBody(int(in.Id), in.Name, in.Class, in.P)
	return &nbodygrpc.ResultCode{
		ResultCode: nbodygrpc.ResultCode_OK,
		Message:    result.String(),
	}, nil
}

func (s *nbodyServiceServer) GetBody(_ context.Context, in *nbodygrpc.ModBodyMessage) (*nbodygrpc.BodyDescription, error) {
	id := int(in.Id)
	name := in.Name
	b := s.callbacks.GetBody(id, name).(*body.Body)
	if b.Id == -1 {
		return nil, status.Errorf(codes.Unimplemented, "No such body ID: %v", id)
	}
	bd := nbodygrpc.BodyDescription{
		Id:                int64(b.Id),
		X:                 b.X,
		Y:                 b.Y,
		Z:                 b.Z,
		Vx:                b.Vx,
		Vy:                b.Vy,
		Vz:                b.Vz,
		Mass:              b.Mass,
		Radius:            b.Radius,
		IsSun:             b.IsSun,
		CollisionBehavior: nbodygrpc.SimCbToGrpcCb(b.CollisionBehavior),
		BodyColor:         nbodygrpc.SimColorToGrpcColor(b.BodyColor),
		FragFactor:        b.FragFactor,
		FragStep:          b.FragStep,
		WithTelemetry:     b.WithTelemetry,
		Name:              b.Name,
		Class:             b.Class,
		Pinned:            b.Pinned,
	}
	return &bd, nil
}

