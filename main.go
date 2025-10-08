package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	if env := os.Getenv("ENV"); env != "production" {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}
	var token = os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is not in env")
	}

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("error create discord client: %v", err)
	}

	discord.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuildWebhooks

	discord.AddHandler(messageCreate)
	discord.AddHandler(OnReady)

	if err := discord.Open(); err != nil {
		log.Fatalf("error opening socket connection: %v", err)
	}

	log.Println("bot running")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	err = discord.Close()
	if err != nil {
		log.Fatalf("error closing discord: %v", err)
	}
}

func OnReady(bot *discordgo.Session, ready *discordgo.Ready) {
	println("[READY] ", bot.State.User.Username)
}

func messageCreate(bot *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.Bot || message.WebhookID != "" {
		return
	}

	if message.Content == "" || strings.ToLower(message.Content[:4]) != "s?" {
		return
	}

	var err = bot.ChannelMessageDelete(message.ChannelID, message.ID)
	if err != nil {
		log.Println("Error Deleteing message: ", err)
	}

	args := strings.Fields(message.Content)
	command := strings.ToLower(args[0])

	if command == "s?ping" {
		replyContent := "機器人延遲: " + strconv.FormatInt(bot.HeartbeatLatency().Milliseconds(), 10) + "ms"
		_, err := bot.ChannelMessageSend(message.ChannelID, replyContent)
		if err != nil {
			log.Printf("Error sending message: %v", err)
			return
		}
	} else if command == "s?s" || command == "s?say" {
		if len(args) <= 2 {
			return
		}
		var target = message.Mentions[0]
		var content = strings.Join(args[2:], " ")
		webhooks, _ := bot.ChannelWebhooks(message.ChannelID)
		var webhook *discordgo.Webhook
		for i := 0; i < len(webhooks); i++ {
			if webhooks[i].User.ID == bot.State.User.ID {
				webhook = webhooks[i]
				break
			}
		}
		if webhook == nil {
			if len(webhooks) >= 10 {
				_, err := bot.ChannelMessageSend(message.ChannelID, "此頻道的Webhooks已經滿了，無法再創建，請刪除一些")
				if err != nil {
					log.Print("Error sending message: ", err)
				}
				return
			}
			newWebhook, _ := bot.WebhookCreate(message.ChannelID, "Tails SUS Golang", "")
			webhook = newWebhook
		}
		_, err := bot.WebhookExecute(webhook.ID, webhook.Token, true, &discordgo.WebhookParams{
			Content:   content,
			Username:  target.Username,
			AvatarURL: target.AvatarURL("4096"),
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeUsers},
			},
		})
		if err != nil {
			log.Printf("Error sending webhook: %v", err)
			return
		}
	}
}
