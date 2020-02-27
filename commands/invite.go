package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/embeds"
)

// Prints Public ZeroTsu's invite link
func inviteCommand(s *discordgo.Session, m *discordgo.Message) {
	err := embeds.Invite(s, m)
	if err != nil {
		if m.GuildID != "" {
			guildSettings := db.GetGuildSettings(m.GuildID)
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		}
	}
}

func init() {
	Add(&Command{
		Execute: inviteCommand,
		Trigger: "invite",
		Aliases: []string{"inv", "invit"},
		Desc:    "Display my invite link",
		Module:  "normal",
		DMAble:  true,
	})
}
