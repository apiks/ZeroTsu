package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Returns user avatar in channel as message
func avatarCommand(s *discordgo.Session, m *discordgo.Message) {

	commandStrings := strings.Split(m.Content, " ")

	if len(commandStrings) > 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage `%avatar [user]`"))
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}
	if len(commandStrings) == 1 {
		// Fetches user
		mem, err := s.User(m.Author.ID)
		if err != nil {
			misc.CommandErrorHandler(s, m, err)
			return
		}
		// Sends user who used the command's avatar
		_, err = s.ChannelMessageSend(m.ChannelID, mem.AvatarURL("256"))
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings
	userID, err := misc.GetUserID(s, m, commandStrings)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Fetches user
	mem, err := s.User(userID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Sends avatar
	_, err = s.ChannelMessageSend(m.ChannelID, mem.AvatarURL("256"))
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
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
		desc:    "Show user avatar. Add [@mention] or [userID] to specify a user.",
		category:"normal",
	})
}