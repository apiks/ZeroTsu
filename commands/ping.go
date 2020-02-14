package commands

import (
	"github.com/bwmarrin/discordgo"

	"ZeroTsu/functionality"
)

// Prints a message to see if the BOT is alive
func pingCommand(s *discordgo.Session, m *discordgo.Message) {
	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	err := functionality.PingEmbed(s, m, guildSettings)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    pingCommand,
		Trigger:    "ping",
		Aliases:    []string{"pingme"},
		Desc:       "See if I respond",
		Permission: functionality.Mod,
		Module:     "misc",
	})
}
