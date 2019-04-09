package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Handles prefix view or change
func prefixCommand(s *discordgo.Session, m *discordgo.Message) {
	commandStrings := strings.SplitN(m.Content, " ", 2)

	// Displays current prefix if it's only that
	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current prefix is: `%v` \n\n To change prefix please use `%vprefix [new prefix]`", config.BotPrefix, config.BotPrefix))
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Changes and writes new prefix to storage
	config.BotPrefix = commandStrings[1]
	err := config.WriteConfig()
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New prefix is: `%v`", config.BotPrefix))
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
		execute:  prefixCommand,
		trigger:  "prefix",
		desc:     "Views or changes the current prefix.",
		elevated: true,
		category: "misc",
	})
}