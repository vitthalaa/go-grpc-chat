package interceptor

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/vitthalaa/go-grpc-chat/gen/go/chat/v1"
)

var (
	authorizationKey = "authorization"
)

type AuthInterceptor struct {
	userName string
}

func NewAuthClientInterceptor(username string) *AuthInterceptor {
	return &AuthInterceptor{
		userName: username,
	}
}

// AuthUnaryClientInterceptor adds authorization to outgoing context
func (i AuthInterceptor) AuthUnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}

	fmt.Println("setting authorization", i.userName)

	md.Set(authorizationKey, i.userName)
	ctx = metadata.NewOutgoingContext(ctx, md)

	return invoker(ctx, method, req, reply, cc, opts...)
}

func (i AuthInterceptor) AuthStreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	// connect method don't need authentication
	if method == pb.ChatService_Connect_FullMethodName {
		return streamer(ctx, desc, cc, method, opts...)
	}

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}

	fmt.Println("setting authorization", i.userName)

	md.Set(authorizationKey, i.userName)
	ctx = metadata.NewOutgoingContext(ctx, md)

	return streamer(ctx, desc, cc, method, opts...)
}
