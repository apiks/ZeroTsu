package commands

import (
	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Prints Public ZeroTsu's invite link
func inviteCommand(s *discordgo.Session, m *discordgo.Message) {
	err := functionality.InviteEmbed(s, m)
	if err != nil {
		if m.GuildID != "" {
			functionality.Mutex.RLock()
			guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
			functionality.Mutex.RUnlock()
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		}
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute: inviteCommand,
		Trigger: "invite",
		Aliases: []string{"inv", "invit"},
		Desc:    "Display my invite link",
		Module:  "normal",
		DMAble:  true,
	})
}
