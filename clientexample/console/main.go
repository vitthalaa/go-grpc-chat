package main

import (
	"context"
	"fmt"
	"log"

	"github.com/AlecAivazis/survey/v2"
	"github.com/vitthalaa/go-grpc-chat/clientexample/console/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/vitthalaa/go-grpc-chat/clientexample/console/prompt"
	pb "github.com/vitthalaa/go-grpc-chat/gen/go/chat/v1"
)

func main() {
	userName := ""
	err := survey.AskOne(&survey.Input{
		Message: "Username:",
	}, &userName)

	if err != nil {
		log.Fatalf("failed to read username: %v", err)
	}

	fmt.Println("Username:", userName)

	ctx := context.Background()

	authInc := interceptor.NewAuthClientInterceptor(userName)

	var opts []grpc.DialOption
	opts = append(opts,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(authInc.AuthUnaryClientInterceptor),
		grpc.WithStreamInterceptor(authInc.AuthStreamClientInterceptor),
	)

	conn, err := grpc.Dial(":5400", opts...)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}

	client := pb.NewChatServiceClient(conn)

	prompter := prompt.NewPrompter(client, userName)
	err = prompter.Run(ctx)
	if err != nil {
		log.Fatalf("failed to run prompt: %v", err)
	}
}
