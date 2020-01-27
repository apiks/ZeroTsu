package commands

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Sends a message from the bot to a channel
func sayCommand(s *discordgo.Session, m *discordgo.Message) {

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 3)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"say OPTIONAL[channelID] [message]`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if the optional channel is present
	channelID, _ := functionality.ChannelParser(s, commandStrings[1], m.GuildID)

	// Sends the message to the channel the original message was in. Else continues to custom channel ID
	if channelID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, commandStrings[1])
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		err = s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if len(commandStrings) < 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"say OPTIONAL[channelID] [message]`\n\nError: Missing non-channel text.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(channelID, commandStrings[2])
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! Message sent.")
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Sends a message embed from the bot to a channel
func sayEmbedCommand(s *discordgo.Session, m *discordgo.Message) {

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 3)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"embed OPTIONAL[channelID] [message]`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if the optional channel is present
	channelID, _ := functionality.ChannelParser(s, commandStrings[1], m.GuildID)

	// Sends the message embed to the channel the original message was in. Else continues to custom channel ID
	if channelID == "" {
		err := functionality.SayEmbed(s, commandStrings[1], m.ChannelID)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		err = s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if len(commandStrings) < 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"embed OPTIONAL[channelID] [message]`\n\nError: Missing non-channel text.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	err := functionality.SayEmbed(s, commandStrings[2], m.ChannelID)
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! Message embed sent.")
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Edits a message sent by the bot with another message
func editCommand(s *discordgo.Session, m *discordgo.Message) {

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 4)

	if len(commandStrings) < 4 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"edit [channelID] [messageID] [message]`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if the channel is present and valid
	channelID, _ := functionality.ChannelParser(s, commandStrings[1], m.GuildID)
	if channelID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid channel.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if it's a valid message ID
	_, err := strconv.ParseInt(commandStrings[2], 10, 64)
	if len(commandStrings[2]) < 17 || err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid message.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Edits the target message
	_, err = s.ChannelMessageEdit(commandStrings[1], commandStrings[2], commandStrings[3])
	if err != nil {
		_, err = s.ChannelMessageSend(m.ChannelID, err.Error()+"\n"+functionality.ErrorLocation(err))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! Selected message edited.")
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Edits an embed message sent by the bot with another embed message
func editEmbedCommand(s *discordgo.Session, m *discordgo.Message) {

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 4)

	if len(commandStrings) < 4 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"edit [channelID] [messageID] [message]`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if the channel is present and valid
	channelID, _ := functionality.ChannelParser(s, commandStrings[1], m.GuildID)
	if channelID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid channel.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if it's a valid message ID
	_, err := strconv.ParseInt(commandStrings[2], 10, 64)
	if len(commandStrings[2]) < 17 || err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid message.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Edits the target message
	err = functionality.EditEmbed(s, commandStrings[1], commandStrings[2], commandStrings[3])
	if err != nil {
		_, err = s.ChannelMessageSend(m.ChannelID, err.Error()+"\n"+functionality.ErrorLocation(err))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! Selected message embed edited.")
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    sayCommand,
		Trigger:    "say",
		Desc:       "Sends a message from bot in the command channel",
		Permission: functionality.Mod,
		Module:     "misc",
	})
	functionality.Add(&functionality.Command{
		Execute:    sayEmbedCommand,
		Trigger:    "embed",
		Desc:       "Sends an embed message from bot in the command channel",
		Permission: functionality.Mod,
		Module:     "misc",
	})
	functionality.Add(&functionality.Command{
		Execute:    editCommand,
		Trigger:    "edit",
		Desc:       "Edits a message sent by the bot",
		Permission: functionality.Mod,
		Module:     "misc",
	})
	functionality.Add(&functionality.Command{
		Execute:    editEmbedCommand,
		Trigger:    "editembed",
		Desc:       "Edits a message embed sent by the bot",
		Permission: functionality.Mod,
		Module:     "misc",
	})
}
