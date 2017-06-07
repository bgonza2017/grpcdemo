package main

import (
	"log"
	"os"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "github.com/bgonza2017/grpcdemo/grpcdemo"
)

const (
	address     = "localhost:50051"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewLobbyClient(conn)

	// Contact the server and print out its response.
	name := "anon"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	r, err := c.Join(context.Background(), &pb.JoinRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Message)
}
