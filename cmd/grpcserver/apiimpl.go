package grpcserver

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"nbodygo/cmd/body"
	"nbodygo/cmd/globals"
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
	mass := float64(in.Mass) // todo remove redundant type conv
	x := float64(in.X)
	y := float64(in.Y)
	z := float64(in.Z)
	vx := float64(in.Vx)
	vy := float64(in.Vy)
	vz := float64(in.Vz)
	radius := float64(in.Radius)
	isSun := in.IsSun
	intensity := in.Intensity
	fragFactor := float64(in.FragFactor)
	fragStep := float64(in.FragStep)
	withTelemetry := in.WithTelemetry
	name := in.Name
	class := in.Class
	pinned := in.Pinned
	behavior := GrpcCbToSimCb(in.CollisionBehavior)
	bodyColor := GrpcColorToSimColor(in.BodyColor)
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
		CollisionBehavior: SimCbToGrpcCb(b.CollisionBehavior),
		BodyColor:         SimColorToGrpcColor(b.BodyColor),
		FragFactor:        b.FragFactor,
		FragStep:          b.FragStep,
		WithTelemetry:     b.WithTelemetry,
		Name:              b.Name,
		Class:             b.Class,
		Pinned:            b.Pinned,
	}
	return &bd, nil
}

func GrpcColorToSimColor(color nbodygrpc.BodyColorEnum) globals.BodyColor {
	switch color {
	case nbodygrpc.BodyColorEnum_RANDOM:
		return globals.Random
	case nbodygrpc.BodyColorEnum_BLACK:
		return globals.Black
	case nbodygrpc.BodyColorEnum_WHITE:
		return globals.White
	case nbodygrpc.BodyColorEnum_DARKGRAY:
		return globals.Darkgray
	case nbodygrpc.BodyColorEnum_GRAY:
		return globals.Gray
	case nbodygrpc.BodyColorEnum_LIGHTGRAY:
		return globals.Lightgray
	case nbodygrpc.BodyColorEnum_RED:
		return globals.Red
	case nbodygrpc.BodyColorEnum_GREEN:
		return globals.Green
	case nbodygrpc.BodyColorEnum_BLUE:
		return globals.Blue
	case nbodygrpc.BodyColorEnum_YELLOW:
		return globals.Yellow
	case nbodygrpc.BodyColorEnum_MAGENTA:
		return globals.Magenta
	case nbodygrpc.BodyColorEnum_CYAN:
		return globals.Cyan
	case nbodygrpc.BodyColorEnum_ORANGE:
		return globals.Orange
	case nbodygrpc.BodyColorEnum_BROWN:
		return globals.Brown
	case nbodygrpc.BodyColorEnum_PINK:
		return globals.Pink
	case nbodygrpc.BodyColorEnum_NOCOLOR:
		fallthrough
	default:
		return globals.Random
	}
}

func SimColorToGrpcColor(color globals.BodyColor) nbodygrpc.BodyColorEnum {
	switch color {
	case globals.Black:
		return nbodygrpc.BodyColorEnum_BLACK
	case globals.White:
		return nbodygrpc.BodyColorEnum_WHITE
	case globals.Darkgray:
		return nbodygrpc.BodyColorEnum_DARKGRAY
	case globals.Gray:
		return nbodygrpc.BodyColorEnum_GRAY
	case globals.Lightgray:
		return nbodygrpc.BodyColorEnum_LIGHTGRAY
	case globals.Red:
		return nbodygrpc.BodyColorEnum_RED
	case globals.Green:
		return nbodygrpc.BodyColorEnum_GREEN
	case globals.Blue:
		return nbodygrpc.BodyColorEnum_BLUE
	case globals.Yellow:
		return nbodygrpc.BodyColorEnum_YELLOW
	case globals.Magenta:
		return nbodygrpc.BodyColorEnum_MAGENTA
	case globals.Cyan:
		return nbodygrpc.BodyColorEnum_CYAN
	case globals.Orange:
		return nbodygrpc.BodyColorEnum_ORANGE
	case globals.Brown:
		return nbodygrpc.BodyColorEnum_BROWN
	case globals.Pink:
		return nbodygrpc.BodyColorEnum_PINK
	case globals.Random:
		fallthrough
	default:
		return nbodygrpc.BodyColorEnum_RANDOM
	}
}

func GrpcCbToSimCb(behavior nbodygrpc.CollisionBehaviorEnum) globals.CollisionBehavior {
	switch behavior {
	case nbodygrpc.CollisionBehaviorEnum_NONE:
		return globals.None
	case nbodygrpc.CollisionBehaviorEnum_SUBSUME:
		return globals.Subsume
	case nbodygrpc.CollisionBehaviorEnum_FRAGMENT:
		return globals.Fragment
	case nbodygrpc.CollisionBehaviorEnum_ELASTIC:
		fallthrough
	default:
		return globals.Elastic
	}
}

func SimCbToGrpcCb(behavior globals.CollisionBehavior) nbodygrpc.CollisionBehaviorEnum {
	switch behavior {
	case globals.None:
		return nbodygrpc.CollisionBehaviorEnum_NONE
	case globals.Subsume:
		return nbodygrpc.CollisionBehaviorEnum_SUBSUME
	case globals.Fragment:
		return nbodygrpc.CollisionBehaviorEnum_FRAGMENT
	default:
		return nbodygrpc.CollisionBehaviorEnum_ELASTIC
	}
}
