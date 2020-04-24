package grpcserver

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"nbodygo/cmd/grpcsimcb"
	"nbodygo/cmd/nbodygrpc"
	"net"
	"strconv"
)

const (
	port int = 50051
)

var grpcServer *grpc.Server

type nbodyServiceServer struct {
	nbodygrpc.UnimplementedNBodyServiceServer
	gpcSim grpcsimcb.GrpcSimCallbacks
}

func newServer(gpcSim grpcsimcb.GrpcSimCallbacks) nbodygrpc.NBodyServiceServer {
	return &nbodyServiceServer{gpcSim: gpcSim}
}

func Start(gpcSim grpcsimcb.GrpcSimCallbacks) {
	grpcServer = grpc.NewServer()
	reflection.Register(grpcServer)
	nbodygrpc.RegisterNBodyServiceServer(grpcServer, newServer(gpcSim))
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		panic("Failed to open port " + strconv.Itoa(port))
	}
	go grpcServer.Serve(lis) // todo can you go a func and handle the error?
}

func Stop() {
	grpcServer.Stop()
}

