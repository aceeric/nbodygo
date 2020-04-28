package grpcserver

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"nbodygo/cmd/grpcsimcb"
	"nbodygo/cmd/nbodygrpc"
	"net"
)

const (
	port int = 50051
)

var grpcServer *grpc.Server

type nbodyServiceServer struct {
	nbodygrpc.UnimplementedNBodyServiceServer
	callbacks grpcsimcb.GrpcSimCallbacks
}

func newServer(callbacks grpcsimcb.GrpcSimCallbacks) nbodygrpc.NBodyServiceServer {
	return &nbodyServiceServer{callbacks: callbacks}
}

func Start(callbacks grpcsimcb.GrpcSimCallbacks) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer = grpc.NewServer()
	reflection.Register(grpcServer)
	nbodygrpc.RegisterNBodyServiceServer(grpcServer, newServer(callbacks))
	go grpcServer.Serve(lis)
}

func Stop() {
	grpcServer.Stop()
}
