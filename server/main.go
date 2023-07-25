package main

import (
	"log"
	"net"

	pb "github.com/vitthalaa/go-grpc-chat/gen/go/chat/v1"
	"github.com/vitthalaa/go-grpc-chat/server/service"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", "localhost:5400")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	chatSvc := service.NewChatService()
	pb.RegisterChatServiceServer(grpcServer, chatSvc)

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
