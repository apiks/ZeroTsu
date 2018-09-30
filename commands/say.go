package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
)

// Sends a message from the bot to the channel
func sayCommand(s *discordgo.Session, m *discordgo.Message) {

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "say [phrase]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
			if err != nil {
				return
			}
			return
		}
		return
	}

	sentence := strings.Replace(m.Content, config.BotPrefix+"say ", "", -1)

	// Sends the sentence to the channel the original message was in.
	_, err := s.ChannelMessageSend(m.ChannelID, sentence)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
		if err != nil {
			return
		}
		return
	}
}

func init() {
	add(&command{
		execute:  sayCommand,
		trigger:  "say",
		desc:     "Sends message from bot in command channel",
		elevated: true,
		deleteAfter: true,
	})
}