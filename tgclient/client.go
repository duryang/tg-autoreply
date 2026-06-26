package tgclient

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
)

type Message struct {
	Text           string
	SenderID       int64
	ChatID         int64
	SenderUsername string
	InputPeer      tg.InputPeerClass
}

type MessageHandler func(Message)

type Client struct {
	apiID       int
	apiHash     string
	phoneNumber string
	handler     MessageHandler
	internal    *telegram.Client
}

func NewClient(apiID int, apiHash string, phoneNumber string) *Client {
	return &Client{
		apiID:       apiID,
		apiHash:     apiHash,
		phoneNumber: phoneNumber,
	}
}

func (client *Client) OnMessage(handler MessageHandler) {
	client.handler = handler
}

func (client *Client) Start(ctx context.Context) error {
	dispatcher := tg.NewUpdateDispatcher()

	dispatcher.OnNewMessage(func(ctx context.Context, entities tg.Entities, update *tg.UpdateNewMessage) error {
		msg, ok := update.Message.(*tg.Message)
		if !ok || msg.Out {
			return nil
		}

		client.handleIncomingMessage(entities, msg)
		return nil
	})

	dispatcher.OnNewChannelMessage(func(ctx context.Context, entities tg.Entities, update *tg.UpdateNewChannelMessage) error {
		msg, ok := update.Message.(*tg.Message)

		// skip post messages in channels, we don't want to reply to them
		if !ok || msg.Out || msg.Post {
			return nil
		}

		client.handleIncomingMessage(entities, msg)
		return nil
	})

	// session storage so we don't re-authenticate every restart
	sessionStorage := &telegram.FileSessionStorage{
		Path: "session.json",
	}

	updatesManager := updates.New(updates.Config{Handler: dispatcher})

	client.internal = telegram.NewClient(client.apiID, client.apiHash, telegram.Options{
		SessionStorage: sessionStorage,
		UpdateHandler:  updatesManager,
	})

	// Auth flow — prompts for phone number and OTP code on first run
	flow := auth.NewFlow(
		auth.CodeOnly(client.phoneNumber, auth.CodeAuthenticatorFunc(func(ctx context.Context, _ *tg.AuthSentCode) (string, error) {
			fmt.Print("Enter OTP code: ")
			var code string
			fmt.Scan(&code)
			return code, nil
		})),
		auth.SendCodeOptions{},
	)

	return client.internal.Run(ctx, func(ctx context.Context) error {
		if err := client.internal.Auth().IfNecessary(ctx, flow); err != nil {
			return err
		}

		self, err := client.internal.Self(ctx)
		if err != nil {
			return err
		}

		fmt.Printf("Logged in as %s (@%s)\n", self.FirstName, self.Username)
		fmt.Println("Listening for messages...")

		return updatesManager.Run(ctx, client.internal.API(), self.ID, updates.AuthOptions{
			IsBot: self.Bot,
		})
	})
}

func (client *Client) handleIncomingMessage(entities tg.Entities, msg *tg.Message) {
	if client.handler != nil {
		client.handler(extractMsgDetails(msg, &entities))
	}
}

func extractMsgDetails(msg *tg.Message, entities *tg.Entities) Message {
	var senderID, chatID int64
	var senderUsername string
	var inputPeer tg.InputPeerClass

	switch peer := msg.GetPeerID().(type) {
	case *tg.PeerUser:
		fmt.Println("Direct message")

		// private chat - sender and chat IDs are the same
		senderID = peer.UserID
		chatID = peer.UserID
		inputPeer = &tg.InputPeerUser{UserID: chatID}
	case *tg.PeerChat:
		fmt.Println("Group messsage")

		// group chat - get chat ID from peer, sender from FromID
		chatID = peer.ChatID
		if fromPeer, ok := msg.FromID.(*tg.PeerUser); ok {
			senderID = fromPeer.UserID
		}
		inputPeer = &tg.InputPeerChat{ChatID: chatID}
	case *tg.PeerChannel:
		fmt.Println("Channel message")

		// channel - get chat ID from peer, sender from FromID
		chatID = peer.ChannelID
		if fromPeer, ok := msg.FromID.(*tg.PeerUser); ok {
			senderID = fromPeer.UserID
		}
		inputPeer = &tg.InputPeerChannel{
			ChannelID:  chatID,
			AccessHash: entities.Channels[chatID].AccessHash,
		}
	}

	// this is only working for direct message
	if user, ok := entities.Users[senderID]; ok {
		senderUsername = user.Username
	}

	return Message{
		Text:           msg.Message,
		SenderID:       senderID,
		ChatID:         chatID,
		SenderUsername: senderUsername,
		InputPeer:      inputPeer,
	}
}

// Sends a reply in the same chat that incoming message was from
func (client *Client) Reply(ctx context.Context, inputPeer tg.InputPeerClass, text string) error {
	sender := message.NewSender(client.internal.API())

	target := sender.To(inputPeer)

	_, err := target.Text(ctx, text)
	return err
}
