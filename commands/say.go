package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"fmt"
)

// Sends a message from the bot to the channel
func sayCommand(s *discordgo.Session, m *discordgo.Message) {

	if m.Content == config.BotPrefix+"say" {

		return
	}

	//Pulls the sentence from strings after "say "
	sentence := strings.Replace(m.Content, config.BotPrefix+"say ", "", -1)

	//Sends the sentence to the channel the original message was in.
	_, err := s.ChannelMessageSend(m.ChannelID, sentence)
	if err != nil {

		fmt.Println("Error:", err)
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