package commands

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
)

// Sends a message from the bot to a channel
func sayCommand(s *discordgo.Session, m *discordgo.Message) {

	var channelID string

	command := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(command, " ", 3)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "say OPTIONAL[channelID] [message]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Checks if the optional command is present
	_, err := strconv.ParseInt(commandStrings[1], 10, 64)
	if len(commandStrings[1]) > 17 && err == nil {
		// Set variable to non-null value for below check
		channelID = "1"
	}

	// Sends the message to the channel the original message was in. Else continues to custom channel ID
	if channelID == "" {
		message := strings.TrimPrefix(m.Content, config.BotPrefix + "say ")
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		err = s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Pulls server channels and checks if it's a valid channel
	channels, err := s.GuildChannels(config.ServerID)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
	for i := 0; i < len(channels); i++ {
		if channels[i].ID == commandStrings[1] {
			channelID = channels[i].ID
			break
		}
	}
	if channelID == "1" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid channel.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	message := strings.TrimPrefix(m.Content, config.BotPrefix + "say " + channelID)
	_, err = s.ChannelMessageSend(channelID, message)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! Message sent.")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Edits a message sent by the bot with another message
func editCommand(s *discordgo.Session, m *discordgo.Message) {

	commandStrings := strings.SplitN(m.Content, " ", 4)

	if len(commandStrings) < 4 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "edit [channelID] [messageID] [message]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Checks if it's a valid channel ID
	_, err := strconv.ParseInt(commandStrings[1], 10, 64)
	if len(commandStrings[1]) < 17 || err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid channel.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}
	// Checks if it's a valid message ID
	_, err = strconv.ParseInt(commandStrings[2], 10, 64)
	if len(commandStrings[2]) < 17 || err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid message.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Edits the target message
	_, err = s.ChannelMessageEdit(commandStrings[1], commandStrings[2], commandStrings[3])
	if err != nil {
		_, err = s.ChannelMessageSend(m.ChannelID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! Selected message edited.")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

func init() {
	add(&command{
		execute:  sayCommand,
		trigger:  "say",
		desc:     "Sends message from bot in command channel",
		elevated: true,
		deleteAfter: false,
		category: "misc",
	})
	add(&command{
		execute:  editCommand,
		trigger:  "edit",
		desc:     "Edits a message sent by the bot with another message",
		elevated: true,
		deleteAfter: false,
		category: "misc",
	})
}