package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/duryang/tg-autoreply/config"
	"github.com/duryang/tg-autoreply/matching"
	"github.com/duryang/tg-autoreply/tgclient"
)

func main() {
	cfg, err := config.LoadConfig("config.toml")
	if err != nil {
		log.Fatal("failed to load config: ", err)
	}

	secrets, err := config.LoadSecrets("secrets.toml")
	if err != nil {
		log.Fatal("failed to load secrets: ", err)
	}

	client := tgclient.NewClient(secrets.APIID, secrets.APIHash, secrets.PhoneNumber)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client.OnMessage(func(msg tgclient.Message) {
		fmt.Printf("New incoming message: text=%q senderID=%d chatID=%d senderUsername=%q\n", msg.Text, msg.SenderID, msg.ChatID, msg.SenderUsername)
		if reply := matching.MatchRule(cfg, msg); reply != nil {
			scheduleReply(ctx, client, msg, *reply)
		}
	})

	if err := client.Start(ctx); err != nil {
		log.Fatal(err)
	}
}

func scheduleReply(ctx context.Context, client *tgclient.Client, msg tgclient.Message, reply config.Reply) {
	go func() {
		if reply.DelaySeconds > 0 {
			time.Sleep(time.Duration(reply.DelaySeconds) * time.Second)
		}
		if err := client.Reply(ctx, msg.InputPeer, reply.Text); err != nil {
			fmt.Println("failed to send reply:", err)
		}
	}()
}
