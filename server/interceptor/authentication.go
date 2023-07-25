package interceptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/vitthalaa/go-grpc-chat/gen/go/chat/v1"
	metadata2 "github.com/vitthalaa/go-grpc-chat/server/metadata"
)

// AuthUnaryInterceptor is authentication interceptor for non-stream grpc server methods
func AuthUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	md, err := metadata2.GetMetaDataWithUser(ctx)
	if err != nil {
		return nil, err
	}

	ctx = metadata.NewIncomingContext(ctx, md)

	return handler(ctx, req)
}

// AuthStreamInterceptor is authentication interceptor for stream grpc server methods
func AuthStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// connect method don't need authentication
	if info.FullMethod == pb.ChatService_Connect_FullMethodName {
		return handler(srv, ss)
	}

	md, err := metadata2.GetMetaDataWithUser(ss.Context())
	if err != nil {
		return err
	}

	ctx := metadata.NewIncomingContext(ss.Context(), md)

	return handler(srv, newStreamWrapper(ss, ctx))
}
