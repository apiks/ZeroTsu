package commands

import (
	"strconv"
	"strings"

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/embeds"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// sayCommand sends a message from the bot to a channel
func sayCommand(s *discordgo.Session, message, targetChannelID string) string {
	_, err := s.ChannelMessageSend(targetChannelID, message)
	if err != nil {
		return err.Error()
	}

	return "Success! Message sent."
}

// sayCommandHandler sends a message from the bot to a channel
func sayCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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

// sayEmbedCommand sends a message embed from the bot to a channel
func sayEmbedCommand(s *discordgo.Session, message, targetChannelID string) string {
	err := embeds.Say(s, message, targetChannelID)
	if err != nil {
		return err.Error()
	}

	return "Success! Embed message sent."
}

// sayEmbedCommandHandler sends a message embed from the bot to a channel
func sayEmbedCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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

// editCommand edits a message sent by the bot with another message
func editCommand(s *discordgo.Session, targetChannelID, targetMessageID, message string) string {
	_, err := s.ChannelMessageEdit(targetChannelID, targetMessageID, message)
	if err != nil {
		return err.Error()
	}

	return "Success! Target message has been edited."
}

// editCommandHandler edits a message sent by the bot with another message
func editCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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

// editEmbedCommand edits an embed message sent by the bot with another embed message
func editEmbedCommand(s *discordgo.Session, targetChannelID, targetMessageID, message string) string {
	err := embeds.Edit(s, targetChannelID, targetMessageID, message)
	if err != nil {
		return err.Error()
	}

	return "Success! Target embed message has been edited."
}

// editEmbedCommandHandler edits an embed message sent by the bot with another embed message
func editEmbedCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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
		Execute:    sayCommandHandler,
		Name:       "say",
		Desc:       "Sends a message from bot in the target channel.",
		Permission: functionality.Mod,
		Module:     "misc",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message",
				Description: "The message you want to send.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "The channel in which you want to send the message to.",
				Required:    false,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "say", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			if i.ApplicationCommandData().Options == nil {
				return
			}

			message := ""
			targetChannelID := i.ChannelID
			for _, option := range i.ApplicationCommandData().Options {
				if option.Name == "message" {
					message = option.StringValue()
				} else if option.Name == "channel" {
					targetChannelID = option.ChannelValue(s).ID
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: sayCommand(s, message, targetChannelID),
				},
			})
		},
	})
	Add(&Command{
		Execute:    sayEmbedCommandHandler,
		Name:       "embed",
		Aliases:    []string{"esay", "sayembed"},
		Desc:       "Sends an embed message from bot in the command channel.",
		Permission: functionality.Mod,
		Module:     "misc",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message",
				Description: "The message you want to send.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "The channel in which you want to send the message to.",
				Required:    false,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "embed", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			if i.ApplicationCommandData().Options == nil {
				return
			}

			message := ""
			targetChannelID := i.ChannelID
			for _, option := range i.ApplicationCommandData().Options {
				if option.Name == "message" {
					message = option.StringValue()
				} else if option.Name == "channel" {
					targetChannelID = option.ChannelValue(s).ID
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: sayEmbedCommand(s, message, targetChannelID),
				},
			})
		},
	})
	Add(&Command{
		Execute:    editCommandHandler,
		Name:       "edit",
		Desc:       "Edits a message sent by the bot.",
		Permission: functionality.Mod,
		Module:     "misc",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "The channel in which the message is.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message-id",
				Description: "The ID of the message itself.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message",
				Description: "The new message with which to replace the old one.",
				Required:    true,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "edit", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			if i.ApplicationCommandData().Options == nil {
				return
			}

			targetChannelID := ""
			messageID := ""
			message := ""
			for _, option := range i.ApplicationCommandData().Options {
				if option.Name == "channel" {
					targetChannelID = option.ChannelValue(s).ID
				} else if option.Name == "message-id" {
					messageID = option.ChannelValue(s).ID
				} else if option.Name == "message" {
					message = option.StringValue()
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: editCommand(s, targetChannelID, messageID, message),
				},
			})
		},
	})
	Add(&Command{
		Execute:    editEmbedCommandHandler,
		Name:       "editembed",
		Desc:       "Edits a message embed sent by the bot.",
		Permission: functionality.Mod,
		Module:     "misc",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "The channel in which the message is.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message-id",
				Description: "The ID of the message itself.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message",
				Description: "The new message with which to replace the old one.",
				Required:    true,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "editembed", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			if i.ApplicationCommandData().Options == nil {
				return
			}

			targetChannelID := ""
			messageID := ""
			message := ""
			for _, option := range i.ApplicationCommandData().Options {
				if option.Name == "channel" {
					targetChannelID = option.ChannelValue(s).ID
				} else if option.Name == "message-id" {
					messageID = option.ChannelValue(s).ID
				} else if option.Name == "message" {
					message = option.StringValue()
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: editEmbedCommand(s, targetChannelID, messageID, message),
				},
			})
		},
	})
}
