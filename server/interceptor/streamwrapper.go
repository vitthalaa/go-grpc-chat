package interceptor

import (
	"context"

	"google.golang.org/grpc"
)

// streamWrapper is a wrapper of grpc.ServerStream for providing authenticated user metadata in context
type streamWrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func newStreamWrapper(s grpc.ServerStream, ctx context.Context) *streamWrapper {
	return &streamWrapper{s, ctx}
}

// Context wrapper for Context method of grpc.ServerStream
func (w *streamWrapper) Context() context.Context {
	return w.ctx
}
