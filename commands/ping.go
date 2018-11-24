package commands

import (
	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Returns a message on "ping" to see if bot is alive
func pingCommand(s *discordgo.Session, m *discordgo.Message) {
	_, err := s.ChannelMessageSend(m.ChannelID, "Hmm? Do you want some honey, darling? Open wide~~")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	roles, err := s.GuildRoles(config.ServerID)
	if err != nil {
		return
	}
	for _, role := range roles {
		if role.Name == "Administratrator" {
			s.ChannelMessageSend(m.ChannelID, role.Name + " role ID is: " + role.ID)
		}
	}
}

func init() {
	add(&command{
		execute:  pingCommand,
		trigger:  "ping",
		aliases:  []string{"pingme"},
		desc:     "Am I alive?",
		elevated: true,
		category: "misc",
	})
}