package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Returns a message on "about" for BOT information
func aboutCommand(s *discordgo.Session, m *discordgo.Message) {
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Hello, I'm %v and was made by Professor Apiks." +
		" I'm written in Go. Use `%vhelp` to list what commands are available to you.", s.State.User.Username, config.BotPrefix))
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

func init() {
	add(&command{
		execute: aboutCommand,
		trigger: "about",
		desc:    "Get info about me.",
		category:"normal",
	})
}