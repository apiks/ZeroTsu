package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/embeds"
	"github.com/r-anime/ZeroTsu/functionality"
)

// Prints a message to see if the BOT is alive
func pingCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)
	err := embeds.Ping(s, m, guildSettings)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
	}
}

func init() {
	Add(&Command{
		Execute:    pingCommandHandler,
		Name:       "ping",
		Aliases:    []string{"pingme"},
		Desc:       "See if I respond",
		Permission: functionality.Mod,
		Module:     "misc",
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			emptyContent := ""
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			err := VerifySlashCommand(s, "ping", i)
			if err != nil {
				errStr := err.Error()
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &errStr,
				})
				return
			}

			embed := embeds.CreatePingEmbed(s.State.User, db.GetGuildSettings(i.GuildID))
			then := time.Now()
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &emptyContent,
				Embeds: &[]*discordgo.MessageEmbed{
					embed,
				},
			})
			embed.Title += fmt.Sprintf(" %s", time.Since(then).Truncate(time.Millisecond).String())
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &emptyContent,
				Embeds: &[]*discordgo.MessageEmbed{
					embed,
				},
			})
		},
	})
}
