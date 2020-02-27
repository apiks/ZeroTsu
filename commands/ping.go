package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/embeds"
	"github.com/r-anime/ZeroTsu/functionality"
)

// Prints a message to see if the BOT is alive
func pingCommand(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)
	err := embeds.Ping(s, m, guildSettings)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
	}
}

func init() {
	Add(&Command{
		Execute:    pingCommand,
		Trigger:    "ping",
		Aliases:    []string{"pingme"},
		Desc:       "See if I respond",
		Permission: functionality.Mod,
		Module:     "misc",
	})
}
