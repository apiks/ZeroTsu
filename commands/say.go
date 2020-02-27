package commands

import (
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/embeds"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Sends a message from the bot to a channel
func sayCommand(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)
	cmdStrs := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 3)

	if len(cmdStrs) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"say OPTIONAL[channelID] [message]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if the optional channel is present
	channelID, _ := common.ChannelParser(s, cmdStrs[1], m.GuildID)

	// Sends the message to the channel the original message was in. Else continues to custom channel ID
	if channelID == "" {
		var message strings.Builder

		message.WriteString(cmdStrs[1])
		if len(cmdStrs) == 3 {
			message.WriteString(" ")
			message.WriteString(cmdStrs[2])
		}

		_, err := s.ChannelMessageSend(m.ChannelID, message.String())
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		err = s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if len(cmdStrs) < 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"say OPTIONAL[channelID] [message]`\n\nError: Missing non-channel text. Maybe the first word was the name of a channel?")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(channelID, cmdStrs[2])
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! Message sent.")
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Sends a message embed from the bot to a channel
func sayEmbedCommand(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)
	cmdStrs := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 3)

	if len(cmdStrs) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"embed OPTIONAL[channelID] [message]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if the optional channel is present
	channelID, _ := common.ChannelParser(s, cmdStrs[1], m.GuildID)

	// Sends the message embed to the channel the original message was in. Else continues to custom channel ID
	if channelID == "" {
		var message strings.Builder

		message.WriteString(cmdStrs[1])
		if len(cmdStrs) == 3 {
			message.WriteString(" ")
			message.WriteString(cmdStrs[2])
		}

		err := embeds.Say(s, message.String(), m.ChannelID)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		err = s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if len(cmdStrs) < 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"embed OPTIONAL[channelID] [message]`\n\nError: Missing non-channel text.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	err := embeds.Say(s, cmdStrs[2], m.ChannelID)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! Message embed sent.")
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Edits a message sent by the bot with another message
func editCommand(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)
	cmdStrs := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 4)

	if len(cmdStrs) < 4 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"edit [channelID] [messageID] [message]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if the channel is present and valid
	channelID, _ := common.ChannelParser(s, cmdStrs[1], m.GuildID)
	if channelID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid channel.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if it's a valid message ID
	_, err := strconv.ParseInt(cmdStrs[2], 10, 64)
	if len(cmdStrs[2]) < 17 || err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid message ID.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Edits the target message
	_, err = s.ChannelMessageEdit(cmdStrs[1], cmdStrs[2], cmdStrs[3])
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! Selected message edited.")
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Edits an embed message sent by the bot with another embed message
func editEmbedCommand(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)
	cmdStrs := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 4)

	if len(cmdStrs) < 4 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"edit [channelID] [messageID] [message]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if the channel is present and valid
	channelID, _ := common.ChannelParser(s, cmdStrs[1], m.GuildID)
	if channelID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid channel.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if it's a valid message ID
	_, err := strconv.ParseInt(cmdStrs[2], 10, 64)
	if len(cmdStrs[2]) < 17 || err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid message.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Edits the target message
	err = embeds.Edit(s, cmdStrs[1], cmdStrs[2], cmdStrs[3])
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! Selected message embed edited.")
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

func init() {
	Add(&Command{
		Execute:    sayCommand,
		Trigger:    "say",
		Desc:       "Sends a message from bot in the command channel",
		Permission: functionality.Mod,
		Module:     "misc",
	})
	Add(&Command{
		Execute:    sayEmbedCommand,
		Trigger:    "embed",
		Aliases:    []string{"esay"},
		Desc:       "Sends an embed message from bot in the command channel",
		Permission: functionality.Mod,
		Module:     "misc",
	})
	Add(&Command{
		Execute:    editCommand,
		Trigger:    "edit",
		Desc:       "Edits a message sent by the bot",
		Permission: functionality.Mod,
		Module:     "misc",
	})
	Add(&Command{
		Execute:    editEmbedCommand,
		Trigger:    "editembed",
		Desc:       "Edits a message embed sent by the bot",
		Permission: functionality.Mod,
		Module:     "misc",
	})
}
