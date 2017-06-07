package main

import (
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "github.com/bgonza2017/grpcdemo/grpcdemo"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) Join(ctx context.Context, in *pb.JoinRequest) (*pb.JoinReply, error) {
	return &pb.JoinReply{Message: "Welcome " + in.Name}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterLobbyServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
