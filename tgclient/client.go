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
	IsGroup        bool
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

		// skip outgoing messages
		if !ok || msg.Out {
			return nil
		}

		// // TODO: revert
		// if !ok {
		// 	return nil
		// }

		if client.handler != nil {
			client.handler(extractMsgDetails(msg, &entities))
		}
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

func extractMsgDetails(msg *tg.Message, entities *tg.Entities) Message {
	var senderID, chatID int64
	var senderUsername string
	var isGroup bool

	switch peer := msg.GetPeerID().(type) {
	case *tg.PeerUser:
		// private chat - sender and chat IDs are the same
		senderID = peer.UserID
		chatID = peer.UserID
	case *tg.PeerChat:
		// group chat - get chat ID from peer, sender from FromID
		chatID = peer.ChatID
		if fromPeer, ok := msg.FromID.(*tg.PeerUser); ok {
			senderID = fromPeer.UserID
		}
		isGroup = true
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
		IsGroup:        isGroup,
	}
}

// Sends a reply in the same chat that the incoming message was from
func (client *Client) Reply(ctx context.Context, msg Message, text string) error {
	sender := message.NewSender(client.internal.API())

	var peer tg.InputPeerClass
	if msg.IsGroup {
		peer = &tg.InputPeerChat{
			ChatID: msg.ChatID,
		}
	} else {
		peer = &tg.InputPeerUser{
        UserID: msg.ChatID,
		}
	}

	target := sender.To(peer)

	_, err := target.Text(ctx, text)
	return err
}
