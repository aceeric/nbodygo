package grpcserver

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"nbodygo/cmd/nbodygrpc"
)

//
// These functions are the implementation of the gRPC interface
//

func (s *nbodyServiceServer) GetCurrentConfig(ctx context.Context, in *empty.Empty) (*nbodygrpc.CurrentConfig, error) {
	return &nbodygrpc.CurrentConfig{
		Bodies:                 int64(s.gpcSim.BodyCount()),
		ResultQueueSize:        int64(s.gpcSim.ResultQueueSize()),
		ComputationThreads:     int64(s.gpcSim.ComputationWorkers()),
		SmoothingFactor:        float32(s.gpcSim.Smoothing()),
		RestitutionCoefficient: float32(s.gpcSim.RestitutionCoefficient()),
	}, nil
}

func (s *nbodyServiceServer) SetComputationThreads(ctx context.Context, in *nbodygrpc.ItemCount) (*nbodygrpc.ResultCode, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetComputationThreads not implemented")
}
func (s *nbodyServiceServer) SetResultQueueSize(ctx context.Context, in *nbodygrpc.ItemCount) (*nbodygrpc.ResultCode, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetResultQueueSize not implemented")
}
func (s *nbodyServiceServer) SetSmoothing(ctx context.Context, in *nbodygrpc.Factor) (*nbodygrpc.ResultCode, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetSmoothing not implemented")
}
func (s *nbodyServiceServer) SetRestitutionCoefficient(ctx context.Context, in *nbodygrpc.RestitutionCoefficient) (*nbodygrpc.ResultCode, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetRestitutionCoefficient not implemented")
}
func (s *nbodyServiceServer) RemoveBodies(ctx context.Context, in *nbodygrpc.ItemCount) (*nbodygrpc.ResultCode, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveBodies not implemented")
}
func (s *nbodyServiceServer) AddBody(ctx context.Context, in *nbodygrpc.BodyDescription) (*nbodygrpc.ResultCode, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddBody not implemented")
}
func (s *nbodyServiceServer) ModBody(ctx context.Context, in *nbodygrpc.ModBodyMessage) (*nbodygrpc.ResultCode, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ModBody not implemented")
}
func (s *nbodyServiceServer) GetBody(ctx context.Context, in *nbodygrpc.ModBodyMessage) (*nbodygrpc.BodyDescription, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBody not implemented")
}
