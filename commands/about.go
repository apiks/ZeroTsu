package commands

import (
	"github.com/bwmarrin/discordgo"

	"ZeroTsu/functionality"
)

// Prints information about the BOT
func aboutCommand(s *discordgo.Session, m *discordgo.Message) {
	err := functionality.AboutEmbed(s, m)
	if err != nil && m.GuildID != "" {
		functionality.Mutex.RLock()
		guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
		functionality.Mutex.RUnlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute: aboutCommand,
		Trigger: "about",
		Desc:    "Display more information about me",
		Module:  "normal",
		DMAble:  true,
	})
}
