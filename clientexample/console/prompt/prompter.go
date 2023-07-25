package prompt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/AlecAivazis/survey/v2"
	pb "github.com/vitthalaa/go-grpc-chat/gen/go/chat/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type option string

const (
	ListAllChannels option = "List all channels"
	CreateGroupChat option = "Create a group chat"
	JoinGroupChat   option = "Join a group chat"
	LeaveGroupChat  option = "Leave a group chat"
	SendMessage     option = "Send a message"
)

var selectHelp = "Press down/up arrow to move cursor. Press enter to select option"

func (o option) String() string {
	return string(o)
}

type optionExecutor func(ctx context.Context) error

type Prompter struct {
	userName       string
	client         pb.ChatServiceClient
	channelCache   []*pb.Channel
	cacheDuration  time.Duration
	cacheExpiresAt time.Time
}

func NewPrompter(client pb.ChatServiceClient, username string) *Prompter {
	return &Prompter{
		userName:      username,
		client:        client,
		cacheDuration: time.Second * 10,
	}
}

func (p *Prompter) Run(ctx context.Context) error {
	msgChan := make(chan *pb.Message)
	defer close(msgChan)

	go p.connect(ctx, msgChan)

	for {
		// initial options
		executor, err := p.askOptions(ctx)
		if err != nil {
			return err
		}

		err = executor(ctx)
		if err != nil {
			return err
		}

		p.checkNewMessage(ctx, msgChan)
	}
}

func (p *Prompter) checkNewMessage(ctx context.Context, msgChan <-chan *pb.Message) {
	select {
	case msg := <-msgChan:
		sender := "@" + msg.GetSender()
		chn := msg.GetChannel()
		senderName := msg.GetSender()
		if chn.GetType() == pb.ChannelType_GROUP {
			senderName = chn.GetName()
			sender = fmt.Sprintf("group %s (%s)", chn.GetName(), sender)
		}

		msgWithQuestion := fmt.Sprintf("New message from %s: --> %s\nDo you want to reply?",
			sender, msg.GetMessage())

		reply := false
		err := survey.AskOne(&survey.Confirm{
			Message: msgWithQuestion,
			Default: false,
		}, &reply)

		if err != nil || !reply {
			return
		}

		err = p.inputMessageForChannel(ctx, senderName)
		if err != nil {
			// TODO log error
		}

	default:
		return
	}
}

func (p *Prompter) connect(ctx context.Context, msgChan chan<- *pb.Message) {
	stream, err := p.client.Connect(ctx, &pb.ConnectRequest{
		Username: p.userName,
	})

	if err != nil {
		log.Printf("failed to connect: %v", err)
		return
	}

	waitC := make(chan struct{})
	defer close(waitC)

	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				close(waitC)
				return
			}
			if err != nil {
				log.Printf("failed to receive message: %v", err)
				return
			}

			msgChan <- msg
		}
	}()

	<-waitC
}

func (p *Prompter) listChannels(ctx context.Context) error {
	channels, err := p.getChannelCache(ctx)
	if err != nil {
		return err
	}

	for _, channel := range channels {
		fmt.Println(fmt.Sprintf("- %s(%s)", channel.GetName(), channel.GetType()))
	}

	return nil
}

func (p *Prompter) joinGroup(ctx context.Context) error {
	return errors.New("not implemented")
}

func (p *Prompter) sendMessage(ctx context.Context) error {
	channel, err := p.askChannelOptions(ctx)
	if err != nil {
		return err
	}

	return p.inputMessageForChannel(ctx, channel)
}

func (p *Prompter) inputMessageForChannel(ctx context.Context, channel string) error {
	msg := ""
	err := survey.AskOne(&survey.Input{
		Message: "Message:",
		Help:    "message for " + channel,
	}, &msg)

	if err != nil {
		return err
	}

	stream, err := p.client.SendMessage(ctx)
	if err != nil {
		return err
	}

	req := &pb.SendMessageRequest{
		Receiver: channel,
		Message:  msg,
	}

	err = stream.Send(req)
	if err != nil {
		return err
	}

	fmt.Println("\nMessage sent")

	return stream.CloseSend()
}

func (p *Prompter) askOptions(ctx context.Context) (optionExecutor, error) {
	selectedOption := ""
	err := survey.AskOne(getRootOptions(), &selectedOption)
	if err != nil {
		return nil, err
	}

	executor := p.getOptionExecutor(option(selectedOption))
	if executor == nil {
		return nil, fmt.Errorf("unknown option: %s", selectedOption)
	}

	return executor, nil
}

func (p *Prompter) getOptionExecutor(op option) optionExecutor {
	if op == "" {
		return nil
	}

	switch op {
	case ListAllChannels:
		return p.listChannels
	case JoinGroupChat:
		return p.joinGroup
	case SendMessage:
		return p.sendMessage
	default:
		return nil
	}
}

func (p *Prompter) askChannelOptions(ctx context.Context) (string, error) {
	channels, err := p.getChannelCache(ctx)
	if err != nil {
		return "", err
	}

	channelOptions := make([]string, 0, len(channels))
	for _, channel := range channels {
		channelOptions = append(channelOptions, channel.GetName())
	}

	channel := ""
	err = survey.AskOne(&survey.Select{
		Message: "Select channel",
		Options: channelOptions,
		Help:    selectHelp,
	}, &channel)

	if err != nil {
		return "", err
	}

	return channel, nil

}

func (p *Prompter) getChannelCache(ctx context.Context) ([]*pb.Channel, error) {
	if len(p.channelCache) != 0 && time.Now().Before(p.cacheExpiresAt) {
		return p.channelCache, nil
	}

	res, err := p.client.ListChannels(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	channels := res.GetChannels()
	p.channelCache = channels
	p.cacheExpiresAt = time.Now().Add(p.cacheDuration)

	return channels, nil
}

func getRootOptions() *survey.Select {
	return &survey.Select{
		Message: "Select an option",
		Options: []string{
			ListAllChannels.String(), CreateGroupChat.String(), JoinGroupChat.String(),
			LeaveGroupChat.String(), SendMessage.String(),
		},
		Default: ListAllChannels.String(),
		Help:    selectHelp,
	}
}
