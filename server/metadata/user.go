package metadata

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	userNameKey      = "username"
	authorizationKey = "authorization"
)

func GetMetaDataWithUser(ctx context.Context) (metadata.MD, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no metadata")
	}

	username := md.Get(authorizationKey)
	if len(username) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no authorization")
	}

	fmt.Println("metadata username: ", username[0])

	md.Set(userNameKey, username[0])

	return md, nil
}

func GetUserName(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	username := md.Get(userNameKey)
	if len(username) == 0 {
		return ""
	}

	return username[0]
}
