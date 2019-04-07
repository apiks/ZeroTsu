package commands

import (
	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Returns a message for bot information
func aboutCommand(s *discordgo.Session, m *discordgo.Message) {
	_, err := s.ChannelMessageSend(m.ChannelID, "Hello, darling. I'm " + s.State.User.Username + " and was made by Professor Apiks for /r/anime. I'm written in Go. "+
		"He says the original me is from Darling in the Franxx but that's just a bunch of nonsense to me. Use `" + config.BotPrefix +
		"help` to list what commands are available to you. I hope you brought sweets.")
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