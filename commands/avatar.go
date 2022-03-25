package commands

import (
	"fmt"
	"strings"

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"

	"github.com/bwmarrin/discordgo"
)

// avatarCommand returns the target user avatar
func avatarCommand(targetUser *discordgo.User) string {
	return targetUser.AvatarURL("256")
}

// avatarCommandHandler returns the target user avatar
func avatarCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	var (
		err           error
		guildSettings = entities.GuildSettings{Prefix: "."}
	)

	if m.GuildID != "" {
		guildSettings = db.GetGuildSettings(m.GuildID)
	}

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) > 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage `%vavatar [user]`", guildSettings.GetPrefix()))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}
	if len(commandStrings) == 1 {
		// Fetches user
		mem, err := s.User(m.Author.ID)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		// Sends user who used the command's avatar
		_, err = s.ChannelMessageSend(m.ChannelID, mem.AvatarURL("256"))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings
	userID, err := common.GetUserID(m, commandStrings)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Fetches user
	mem, err := s.User(userID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Sends avatar
	_, err = s.ChannelMessageSend(m.ChannelID, mem.AvatarURL("256"))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

func init() {
	Add(&Command{
		Execute: avatarCommandHandler,
		Name:    "avatar",
		Desc:    "Show a user's avatar.",
		Module:  "normal",
		DMAble:  true,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "The user of whom you want to see the avatar of.",
				Required:    false,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			user := i.Member.User
			if i.ApplicationCommandData().Options != nil {
				for _, option := range i.ApplicationCommandData().Options {
					if option.Name == "user" {
						user = option.UserValue(s)
					}
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: avatarCommand(user),
				},
			})
		},
	})
}
