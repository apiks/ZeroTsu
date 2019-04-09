package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Handles playing message view or change
func playingMsgCommand(s *discordgo.Session, m *discordgo.Message) {
	commandStrings := strings.SplitN(m.Content, " ", 2)

	// Displays current playing message if it's only that
	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current playing message is: `%v` \n\n To change the message please use `%vplayingmsg [new message]`", config.PlayingMsg, config.BotPrefix))
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Changes and writes new playing message to storage
	config.PlayingMsg = commandStrings[1]
	err := config.WriteConfig()
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Refreshes playing message
	err = s.UpdateStatus(0, config.PlayingMsg)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New playing message is: `%v`", config.PlayingMsg))
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
		execute:  playingMsgCommand,
		trigger:  "playingmsg",
		desc:     "Views or changes the current BOT playing message.",
		elevated: true,
		category: "misc",
	})
}