package commands

import (
	"github.com/bwmarrin/discordgo"

	"../functionality"
)

// Prints information about the BOT
func aboutCommand(s *discordgo.Session, m *discordgo.Message) {
	err := functionality.AboutEmbed(s, m)
	if err != nil && m.GuildID != "" {
		functionality.MapMutex.Lock()
		guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
		functionality.MapMutex.Unlock()
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
