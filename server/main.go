package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/vitthalaa/go-grpc-chat/gen/go/chat/v1"
	"github.com/vitthalaa/go-grpc-chat/server/interceptor"
	"github.com/vitthalaa/go-grpc-chat/server/service"
)

func main() {
	lis, err := net.Listen("tcp", "localhost:5400")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.AuthUnaryInterceptor),
		grpc.StreamInterceptor(interceptor.AuthStreamInterceptor),
	}

	grpcServer := grpc.NewServer(opts...)

	chatSvc := service.NewChatService()
	pb.RegisterChatServiceServer(grpcServer, chatSvc)

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
