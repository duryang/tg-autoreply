package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

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
		fmt.Printf("New message from %d (@%s): %s\n", msg.SenderID, msg.SenderUsername, msg.Text)
		if reply := matching.MatchRule(cfg, msg); reply != nil {
			if err := client.Reply(ctx, msg, reply.Text); err != nil {
				fmt.Println("failed to send reply:", err)
			}
		}
	})

	if err := client.Start(ctx); err != nil {
		log.Fatal(err)
	}
}
