package vb

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"vb-bot/internal/config"
	"vb-bot/pkg/logging"
)

type Bot struct {
	Session *discordgo.Session
	ID      string
	Logger  logging.Logger
}

func NewBotSession(config config.BotConfig, logger logging.Logger) (Bot, error) {
	bot := Bot{}
	session, err := discordgo.New(fmt.Sprintf("Bot %s", config.Token))
	if err != nil {
		logger.Warnf("Warning in creating bot session %s", err)
	}
	user, err := session.User("@me")
	if err != nil {
		logger.Warnf("Warning in getting user @me data %s", err)
	}

	bot.Session = session
	bot.ID = user.ID
	bot.Logger = logger

	return bot, err
}

// AddHandler wrap session handler
func (bot *Bot) AddHandler(handler interface{}) func() {
	return bot.Session.AddHandler(handler)
}
