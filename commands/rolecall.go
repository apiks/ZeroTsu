package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"ZeroTsu/functionality"
)

// Prints the amount of users in a guild have a specific role
func rolecallCommand(s *discordgo.Session, m *discordgo.Message) {
	var (
		role    functionality.Role
		counter int
	)

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: %srolecall [role]", guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parse role for roleID
	role.ID, role.Name = functionality.RoleParser(s, commandStrings[1], m.GuildID)
	if role.ID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role exists.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	guild, err := s.Guild(m.GuildID)
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
	for _, mem := range guild.Members {
		for _, memRole := range mem.Roles {
			if memRole == role.ID {
				counter++
				break
			}
		}
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%d users have the `%s` role.", counter, role.Name))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    rolecallCommand,
		Trigger:    "rolecall",
		Desc:       "Prints the amount of users that have a specific role",
		Permission: functionality.Mod,
		Module:     "misc",
	})
}
