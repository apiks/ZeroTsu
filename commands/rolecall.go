package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Prints the amount of users in a guild have a specific role
func rolecallCommand(s *discordgo.Session, m *discordgo.Message) {
	var (
		role    entities.Role
		counter int
	)

	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: %srolecall [role]", guildSettings.GetPrefix()))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parse role for roleID
	roleID, roleName := common.RoleParser(s, commandStrings[1], m.GuildID)
	role = role.SetID(roleID)
	role = role.SetName(roleName)
	if role.GetID() == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role exists.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		guild, err = s.Guild(m.GuildID)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
	}
	for _, mem := range guild.Members {
		for _, memRole := range mem.Roles {
			if memRole == role.GetID() {
				counter++
				break
			}
		}
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%d users have the `%s` role.", counter, role.GetName()))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

func init() {
	Add(&Command{
		Execute:    rolecallCommand,
		Trigger:    "rolecall",
		Desc:       "Prints the amount of users that have a specific role",
		Permission: functionality.Mod,
		Module:     "misc",
	})
}
