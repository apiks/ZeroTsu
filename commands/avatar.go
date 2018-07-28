package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
	"github.com/r-anime/ZeroTsu/config"
)

// Returns a user avatar in channel
func avatarCommand(s *discordgo.Session, m *discordgo.Message) {

	commandStrings := strings.Split(m.Content, " ")

	if len(commandStrings) != 2 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix + "avatar [@user or userID]`")
		if err != nil {

			_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
			if err != nil {

				return
			}
			return
		}
	}

	// Pulls userID from 2nd parameter of commandStrings
	userID := misc.GetUserID(s, m, commandStrings)
	if userID == "" {
		return
	}

	// Fetches user
	mem, err := s.User(userID)
	if err != nil {

		misc.CommandErrorHandler(s, m, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, mem.AvatarURL("256"))
	if err != nil {

		_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
		if err != nil {

			return
		}
		return
	}
}

func init() {
	add(&command{
		execute: avatarCommand,
		trigger: "avatar",
		desc:    "Show user avatar.",
	})
}