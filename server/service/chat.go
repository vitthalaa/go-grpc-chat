package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/vitthalaa/go-grpc-chat/gen/go/chat/v1"
	"github.com/vitthalaa/go-grpc-chat/server/metadata"
)

type Channel struct {
	Type  pb.ChannelType
	Name  string
	Users []string
}

type ChatService struct {
	pb.UnimplementedChatServiceServer
	userChannels map[string]chan *pb.Message
	channels     map[string]Channel
}

func NewChatService() *ChatService {
	return &ChatService{
		userChannels: make(map[string]chan *pb.Message),
		channels:     make(map[string]Channel),
	}
}

var errUnimplemented = errors.New("not implemented")

func (s *ChatService) Connect(req *pb.ConnectRequest, stream pb.ChatService_ConnectServer) error {
	userName := req.GetUsername()
	if _, ok := s.userChannels[userName]; ok {
		return errors.New("already present")
	}

	channel := Channel{
		Type: pb.ChannelType_USER,
		Name: userName,
	}

	s.channels[userName] = channel

	msgChannel := make(chan *pb.Message)
	s.userChannels[userName] = msgChannel

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case msg := <-msgChannel:
			fmt.Printf("GO ROUTINE (got message): %v \n", msg)
			err := stream.Send(msg)
			if err != nil {
				log.Default().Println(err)
			}
		}
	}
}

func (s *ChatService) CreateGroupChat(ctx context.Context, req *pb.CreateGroupChatRequest) (*emptypb.Empty, error) {
	_, err := s.getAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	return nil, errUnimplemented
}

func (s *ChatService) JoinGroupChat(ctx context.Context, req *pb.JoinGroupChatRequest) (*emptypb.Empty, error) {
	_, err := s.getAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	return nil, errUnimplemented
}

func (s *ChatService) LeaveGroupChat(ctx context.Context, req *pb.LeaveGroupChatRequest) (*emptypb.Empty, error) {
	_, err := s.getAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	return nil, errUnimplemented
}

func (s *ChatService) SendMessage(msgStream pb.ChatService_SendMessageServer) error {
	sender, err := s.getAuthUser(msgStream.Context())
	if err != nil {
		return err
	}

	req, err := msgStream.Recv()
	if err == io.EOF {
		return nil
	}

	if err != nil {
		return err
	}

	channel, ok := s.channels[req.GetReceiver()]
	if !ok {
		return errors.New("invalid receiver")
	}

	err = msgStream.SendAndClose(&emptypb.Empty{})
	if err != nil {
		return err
	}

	pbChannel := &pb.Channel{
		Type: channel.Type,
		Name: channel.Name,
	}

	if channel.Type == pb.ChannelType_USER {
		err = s.sendUserMessage(channel.Name, sender, req.GetMessage(), pbChannel)
		if err != nil {
			return err
		}

		return nil
	}

	for _, user := range channel.Users {
		err = s.sendUserMessage(user, sender, req.GetMessage(), pbChannel)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

func (s *ChatService) ListChannels(ctx context.Context, req *emptypb.Empty) (*pb.ListChannelsResponse, error) {
	user, err := s.getAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	resChan := make([]*pb.Channel, 0, len(s.channels)-1)
	for _, c := range s.channels {
		// not including user who requested list
		if c.Name == user {
			continue
		}

		resChan = append(resChan, &pb.Channel{
			Type: c.Type,
			Name: c.Name,
		})
	}

	return &pb.ListChannelsResponse{
		Channels: resChan,
	}, nil
}

func (s *ChatService) sendUserMessage(user, sender, msg string, channel *pb.Channel) error {
	msgChannel, ok := s.userChannels[user]
	if !ok {
		return errors.New("invalid receiver")
	}

	msgChannel <- &pb.Message{
		Channel: channel,
		Message: msg,
		Sender:  sender,
		Time:    timestamppb.New(time.Now()),
	}

	return nil
}

func (s *ChatService) getAuthUser(ctx context.Context) (string, error) {
	username := metadata.GetUserName(ctx)
	if username == "" {
		return "", status.Error(codes.Unauthenticated, "unauthenticated")
	}

	if _, ok := s.userChannels[username]; !ok {
		return "", status.Error(codes.Unauthenticated, "unauthenticated")
	}

	return username, nil
}
