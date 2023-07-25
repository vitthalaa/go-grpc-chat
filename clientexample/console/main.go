package main

import (
	"context"
	"fmt"
	"log"

	"github.com/AlecAivazis/survey/v2"
	"github.com/vitthalaa/go-grpc-chat/clientexample/console/prompt"
	pb "github.com/vitthalaa/go-grpc-chat/gen/go/chat/v1"
	"google.golang.org/grpc"
)

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure())

	conn, err := grpc.Dial(":5400", opts...)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}

	ctx := context.Background()

	client := pb.NewChatServiceClient(conn)

	userName := ""
	err = survey.AskOne(&survey.Input{
		Message: "Username:",
	}, &userName)

	fmt.Println("Username:", userName)

	prompter := prompt.NewPrompter(client, userName)
	err = prompter.Run(ctx)
	if err != nil {
		log.Fatalf("Failed to run prompt: %v", err)
	}
}
