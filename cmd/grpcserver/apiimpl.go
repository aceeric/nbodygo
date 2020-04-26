package grpcserver

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"nbodygo/cmd/globals"
	"nbodygo/cmd/nbodygrpc"
)

//
// These functions are the implementation of the gRPC interface. They simply delegate everything to the
// callbacks in the passed nbodyServiceServer struct, and do some assembly/disassembly of data structures
//

func (s *nbodyServiceServer) GetCurrentConfig(_ context.Context, in *empty.Empty) (*nbodygrpc.CurrentConfig, error) {
	return &nbodygrpc.CurrentConfig{
		Bodies:                 int64(s.gpcSim.BodyCount()),
		ResultQueueSize:        int64(s.gpcSim.ResultQueueSize()),
		ComputationThreads:     int64(s.gpcSim.ComputationWorkers()),
		SmoothingFactor:        float32(s.gpcSim.Smoothing()),
		RestitutionCoefficient: float32(s.gpcSim.RestitutionCoefficient()),
	}, nil
}

// needs to remain unimplemented until the Worker Pool can support resize
func (s *nbodyServiceServer) SetComputationThreads(_ context.Context, in *nbodygrpc.ItemCount) (*nbodygrpc.ResultCode, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetComputationThreads not implemented")
}

// needs to remain unimplemented until the ResultQueueHolder can support resize in a performant manner
func (s *nbodyServiceServer) SetResultQueueSize(_ context.Context, in *nbodygrpc.ItemCount) (*nbodygrpc.ResultCode, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetResultQueueSize not implemented")
}

func (s *nbodyServiceServer) SetSmoothing(_ context.Context, in *nbodygrpc.Factor) (*nbodygrpc.ResultCode, error) {
	s.gpcSim.SetSmoothing(float64(in.GetFactor()))
	return &nbodygrpc.ResultCode{
		ResultCode: nbodygrpc.ResultCode_OK,
		Message:    "OK",
	}, nil
}

func (s *nbodyServiceServer) SetRestitutionCoefficient(_ context.Context, in *nbodygrpc.RestitutionCoefficient) (*nbodygrpc.ResultCode, error) {
	s.gpcSim.SetRestitutionCoefficient(float64(in.GetRestitutionCoefficient()))
	return &nbodygrpc.ResultCode{
		ResultCode: nbodygrpc.ResultCode_OK,
		Message:    "OK",
	}, nil
}

func (s *nbodyServiceServer) RemoveBodies(_ context.Context, in *nbodygrpc.ItemCount) (*nbodygrpc.ResultCode, error) {
	s.gpcSim.RemoveBodies(int(in.GetItemCount()))
	return &nbodygrpc.ResultCode{
		ResultCode: nbodygrpc.ResultCode_OK,
		Message:    "OK",
	}, nil
}

func (s *nbodyServiceServer) AddBody(_ context.Context, in *nbodygrpc.BodyDescription) (*nbodygrpc.ResultCode, error) {
	mass := float64(in.GetMass())
	x := float64(in.GetX())
	y := float64(in.GetY())
	z := float64(in.GetZ())
	vx := float64(in.GetVx())
	vy := float64(in.GetVy())
	vz := float64(in.GetVz())
	radius := float64(in.GetRadius())
	isSun := in.GetIsSun()
	fragFactor := float64(in.GetFragFactor())
	fragStep := float64(in.GetFragStep())
	withTelemetry := in.GetWithTelemetry()
	name := in.GetName()
	class := in.GetClass()
	pinned := in.GetPinned()
	behavior := grpcCbToSimCb(in.GetCollisionBehavior())
	bodyColor := grpcColorToSimColor(in.GetBodyColor())
	id := s.gpcSim.AddBody(mass, x, y, z, vx, vy, vz, radius, isSun, behavior, bodyColor,
		fragFactor, fragStep, withTelemetry, name, class, pinned)
	return &nbodygrpc.ResultCode{
		ResultCode: nbodygrpc.ResultCode_OK,
		Message:    fmt.Sprintf("Added body ID: %d", id),
	}, nil
}

func (s *nbodyServiceServer) ModBody(_ context.Context, in *nbodygrpc.ModBodyMessage) (*nbodygrpc.ResultCode, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ModBody not implemented")
}

func (s *nbodyServiceServer) GetBody(_ context.Context, in *nbodygrpc.ModBodyMessage) (*nbodygrpc.BodyDescription, error) {
	id := int(in.GetId())
	name := in.GetName()
	rb := s.gpcSim.GetBody(id, name)
	if rb.Id == -1 {
		return nil, status.Errorf(codes.Unimplemented, "No such body ID: %v", id)
	}
	bd := nbodygrpc.BodyDescription{
		Id:                rb.Id,
		X:                 rb.X,
		Y:                 rb.Y,
		Z:                 rb.Z,
		Vx:                rb.Vx,
		Vy:                rb.Vy,
		Vz:                rb.Vz,
		Mass:              rb.Mass,
		Radius:            rb.Radius,
		IsSun:             rb.IsSun,
		CollisionBehavior: simCbToGrpcCb(rb.CollisionBehavior),
		BodyColor:         SimColorToGrpcColor(rb.BodyColor),
		FragFactor:        rb.FragFactor,
		FragStep:          rb.FragStep,
		WithTelemetry:     rb.WithTelemetry,
		Name:              rb.Name,
		Class:             rb.Class,
		Pinned:            rb.Pinned,
	}
	return &bd, nil
}

func grpcColorToSimColor(color nbodygrpc.BodyColorEnum) globals.BodyColor {
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
	case globals.Black: return nbodygrpc.BodyColorEnum_BLACK
	case globals.White: return nbodygrpc.BodyColorEnum_WHITE
	case globals.Darkgray: return nbodygrpc.BodyColorEnum_DARKGRAY
	case globals.Gray: return nbodygrpc.BodyColorEnum_GRAY
	case globals.Lightgray: return nbodygrpc.BodyColorEnum_LIGHTGRAY
	case globals.Red: return nbodygrpc.BodyColorEnum_RED
	case globals.Green: return nbodygrpc.BodyColorEnum_GREEN
	case globals.Blue: return nbodygrpc.BodyColorEnum_BLUE
	case globals.Yellow: return nbodygrpc.BodyColorEnum_YELLOW
	case globals.Magenta: return nbodygrpc.BodyColorEnum_MAGENTA
	case globals.Cyan: return nbodygrpc.BodyColorEnum_CYAN
	case globals.Orange: return nbodygrpc.BodyColorEnum_ORANGE
	case globals.Brown: return nbodygrpc.BodyColorEnum_BROWN
	case globals.Pink: return nbodygrpc.BodyColorEnum_PINK
	case globals.Random:
		fallthrough
	default:
		return nbodygrpc.BodyColorEnum_RANDOM
	}
}

func grpcCbToSimCb(behavior nbodygrpc.CollisionBehaviorEnum) globals.CollisionBehavior {
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

func simCbToGrpcCb(behavior globals.CollisionBehavior) nbodygrpc.CollisionBehaviorEnum {
	switch behavior {
	case globals.None: return nbodygrpc.CollisionBehaviorEnum_NONE
	case globals.Subsume: return nbodygrpc.CollisionBehaviorEnum_SUBSUME
	case globals.Fragment: return nbodygrpc.CollisionBehaviorEnum_FRAGMENT
	default:
		return nbodygrpc.CollisionBehaviorEnum_ELASTIC
	}
}

