package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/embeds"
)

// Prints information about the BOT
func aboutCommand(s *discordgo.Session, m *discordgo.Message) {
	err := embeds.About(s, m)
	if err != nil && m.GuildID != "" {
		guildSettings := db.GetGuildSettings(m.GuildID)
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
	}
}

func init() {
	Add(&Command{
		Execute: aboutCommand,
		Name:    "about",
		Desc:    "Display more information about me.",
		Module:  "normal",
		DMAble:  true,
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						embeds.CreateAboutEmbed(s.State.User),
					},
				},
			})
		},
	})
}
