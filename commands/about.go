package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

// Returns a message on "about" for BOT information
func aboutCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		guildPrefix = "."
		guildBotLog string
	)

	if m.GuildID != "" {
		misc.MapMutex.Lock()
		guildPrefix = misc.GuildMap[m.GuildID].GuildConfig.Prefix
		guildBotLog = misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
		misc.MapMutex.Unlock()
	}

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Hello, I'm %v and was made by Professor Apiks."+
		" I'm written in Go. Use `%vhelp` to list what commands are available to you", s.State.User.Username, guildPrefix))
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
}

func init() {
	add(&command{
		execute:  aboutCommand,
		trigger:  "about",
		desc:     "Get info about me.",
		category: "normal",
		DMAble: true,
	})
}
