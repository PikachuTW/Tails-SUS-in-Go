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
	enverr := godotenv.Load()
	var token = os.Getenv("DISCORD_TOKEN")
	if enverr != nil || token == "" {
		log.Fatal("failed to load env\n", enverr)
	}

	discord, discorderr := discordgo.New("Bot " + token)
	if discorderr != nil {
		log.Fatal("error create discord client\n", discorderr)
	}

	discord.Identify.Intents = discordgo.IntentsAll
	discord.State.TrackMembers = true
	discord.State.TrackChannels = true
	discord.State.TrackRoles = true

	discord.AddHandler(messageCreate)
	discord.AddHandler(OnReady)

	discordsocketerr := discord.Open()
	if discordsocketerr != nil {
		log.Fatal("error opening socket connection\n", discordsocketerr)
	}

	log.Println("bot running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	discord.Close()
}

func OnReady(bot *discordgo.Session, ready *discordgo.Ready) {
	println("[READY] ", bot.State.User.Username)
}

func messageCreate(bot *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.Bot || message.WebhookID != "" {
		return
	}

	if message.Content == "" || strings.ToLower(message.Content[:4]) != "sgo?" {
		return
	}

	var err = bot.ChannelMessageDelete(message.ChannelID, message.ID)
	if err != nil {
		log.Println("Error Deleteing message: ", err)
	}

	args := strings.Fields(message.Content)
	command := strings.ToLower(args[0])

	if command == "sgo?ping" {
		var replyContent string = "機器人延遲: " + strconv.FormatInt(bot.HeartbeatLatency().Milliseconds(), 10) + "ms"
		bot.ChannelMessageSend(message.ChannelID, replyContent)
	} else if command == "sgo?s" || command == "sgo?say" {
		if len(args) <= 2 {
			return
		}
		var target = message.Mentions[0]
		var content = strings.Join(args[2:], " ")
		webhooks, _ := bot.ChannelWebhooks(message.ChannelID)
		var foundwebhook bool = false
		var webhook *discordgo.Webhook
		for i := 0; i < len(webhooks); i++ {
			if webhooks[i].User.ID == bot.State.User.ID {
				foundwebhook = true
				webhook = webhooks[i]
				break
			}
		}
		if !foundwebhook {
			if len(webhooks) >= 10 {
				bot.ChannelMessageSend(message.ChannelID, "此頻道的Webhooks已經滿了，無法再創建，請刪除一些")
				return
			}
			newwebhook, _ := bot.WebhookCreate(message.ChannelID, "Tails SUS Golang", "")
			webhook = newwebhook
		}
		bot.WebhookExecute(webhook.ID, webhook.Token, true, &discordgo.WebhookParams{
			Content:   content,
			Username:  target.Username,
			AvatarURL: target.AvatarURL("4096"),
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeUsers},
			},
		})
	}
}
