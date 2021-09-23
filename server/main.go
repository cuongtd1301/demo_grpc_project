package main

import (
	"log"
	"net"

	pb "demo-grpc/proto"
	"demo-grpc/server/infrastructure"
	"demo-grpc/server/service"

	"google.golang.org/grpc"
)

// var contentsTag = cascadia.MustCompile("p, h1, h2, h3, h4, h5, h6")

func main() {
	lis, err := net.Listen("tcp", infrastructure.GetGrpcAddress())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterMediaServiceServer(s, service.GetServerGrpcStruct())
	log.Printf("grpc server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to grpc serve: %v", err)
	}
}
