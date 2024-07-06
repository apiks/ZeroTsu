package commands

import (
	"fmt"
	"strings"

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// setDailyScheduleCommand sets a channel ID as the autopost anime schedule target channel
func setDailyScheduleCommand(targetChannel *discordgo.Channel, enabled bool, guildID string) string {
	dailySchedule := db.GetGuildAutopost(guildID, "dailyschedule")

	if targetChannel == nil && dailySchedule == (entities.Cha{}) {
		return "Error: Daily anime schedule autopost is currently not set."
	} else if targetChannel == nil && dailySchedule != (entities.Cha{}) && enabled {
		return fmt.Sprintf("Current daily anime schedule autopost channel is: `%s - %s`", dailySchedule.GetName(), dailySchedule.GetID())
	}

	if dailySchedule == (entities.Cha{}) {
		dailySchedule = entities.NewCha("", "", "")
	}

	// Parse and save the target channel
	if !enabled {
		dailySchedule = entities.Cha{}
	} else {
		dailySchedule = entities.NewCha(targetChannel.Name, targetChannel.ID, "")
	}

	// Write
	db.SetGuildAutopost(guildID, "dailyschedule", dailySchedule)

	if dailySchedule == (entities.Cha{}) {
		return "Success: Daily anime schedule autopost has been disabled!"
	} else {
		return fmt.Sprintf("Success: New daily anime schedule autopost channel is: `%s - %s`", dailySchedule.GetName(), dailySchedule.GetID())
	}
}

// setDailyScheduleCommandHandler sets a channel ID as the autopost anime schedule target channel
func setDailyScheduleCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)
	dailySchedule := db.GetGuildAutopost(m.GuildID, "dailyschedule")
	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Displays current dailyschedule channel
	if len(commandStrings) == 1 {
		if dailySchedule == (entities.Cha{}) {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: Autopost Daily Anime Schedule channel is currently not set. Please use `%sdailyschedule [channel]`", guildSettings.GetPrefix()))
			if err != nil {
				common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
				return
			}
			return
		}
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current Autopost Daily Anime Schedule channel is: `%s - %s` \n\nTo change it please use `%sdailyschedule [channel]`\nTo disable it please use `%sdailyschedule disable`", dailySchedule.GetName(), dailySchedule.GetID(), guildSettings.GetPrefix(), guildSettings.GetPrefix()))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}
	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sdailyschedule [channel]`\nTo disable it please use `%sdailyschedule disable`", guildSettings.GetPrefix(), guildSettings.GetPrefix()))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	if dailySchedule == (entities.Cha{}) {
		dailySchedule = entities.NewCha("", "", "")
	}

	// Parse and save the target channel
	if commandStrings[1] == "disable" || commandStrings[1] == "0" || commandStrings[1] == "false" {
		dailySchedule = entities.Cha{}
	} else {
		channelID, channelName := common.ChannelParser(s, commandStrings[1], m.GuildID)
		dailySchedule = entities.NewCha(channelName, channelID, "")
	}

	// Write
	db.SetGuildAutopost(m.GuildID, "dailyschedule", dailySchedule)

	if dailySchedule == (entities.Cha{}) {
		_, err := s.ChannelMessageSend(m.ChannelID, "Success! Autopost Daily Anime Schedule has been disabled! If this was not intentional please verify the channel ID.")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New Autopost Daily Anime Schedule channel is: `%s - %s`", dailySchedule.GetName(), dailySchedule.GetID()))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// setNewEpisodesCommand sets a channel ID as the autopost new airing anime episodes target channel
func setNewEpisodesCommand(targetChannel *discordgo.Channel, enabled bool, role *discordgo.Role, guildID string) string {
	var newEpisodes = db.GetGuildAutopost(guildID, "newepisodes")

	if targetChannel == nil && newEpisodes == (entities.Cha{}) {
		return "Error: New anime episodes autopost is currently not set."
	} else if targetChannel == nil && newEpisodes != (entities.Cha{}) && enabled {
		if newEpisodes.GetRoleID() != "" {
			return fmt.Sprintf("Current New anime episodes autopost channel is: `%s - Channel ID: %s- Role ID: %s`", newEpisodes.GetName(), newEpisodes.GetID(), newEpisodes.GetRoleID())
		} else {
			return fmt.Sprintf("Current New anime episodes autopost channel is: `%s - Channel ID: %s`", newEpisodes.GetName(), newEpisodes.GetID())
		}
	}

	if newEpisodes == (entities.Cha{}) {
		newEpisodes = entities.NewCha("", "", "")
	}

	// Parse and save the target channel
	newEpisodes = entities.Cha{}
	if enabled {
		newEpisodes = entities.NewCha(targetChannel.Name, targetChannel.ID, "")
	}

	if role != nil {
		newEpisodes = newEpisodes.SetRoleID(role.ID)
	}

	// Write
	db.SetGuildAutopost(guildID, "newepisodes", newEpisodes)
	entities.SetupGuildSub(guildID)
	err := entities.AnimeSubsWrite(entities.SharedInfo.GetAnimeSubsMap())
	if err != nil {
		return "Error: " + err.Error()
	}

	if newEpisodes == (entities.Cha{}) {
		return "Success: New anime episodes autopost has been disabled!"
	} else {
		if newEpisodes.GetRoleID() != "" {
			return fmt.Sprintf("Success: New anime episodes autopost channel is: `%s - Channel ID: %s- Role ID: %s`", newEpisodes.GetName(), newEpisodes.GetID(), newEpisodes.GetRoleID())
		} else {
			return fmt.Sprintf("Success: New anime episodes autopost channel is: `%s - Channel ID: %s`", newEpisodes.GetName(), newEpisodes.GetID())
		}
	}
}

// setNewEpisodesCommandHandler sets a channel ID as the autopost new airing anime episodes target channel
func setNewEpisodesCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)
	newEpisodes := db.GetGuildAutopost(m.GuildID, "newepisodes")
	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Displays current new episodes channel
	if len(commandStrings) == 1 {
		if newEpisodes == (entities.Cha{}) {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: Autopost channel for new airing anime episodes is currently not set. Please use `%snewepisodes [channel]`", guildSettings.GetPrefix()))
			if err != nil {
				common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
				return
			}
			return
		}
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current Autopost channel for new airing anime episodes is: `%s - %s` \n\n To change it please use `%snewepisodes [channel]`\nTo disable it please use `%snewepisodes disable`", newEpisodes.GetName(), newEpisodes.GetID(), guildSettings.GetPrefix(), guildSettings.GetPrefix()))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}
	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%snewepisodes [channel]`\nTo disable it please use `%snewepisodes disable`", guildSettings.GetPrefix(), guildSettings.GetPrefix()))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	if newEpisodes == (entities.Cha{}) {
		newEpisodes = entities.NewCha("", "", "")
	}

	// Parse and save the target channel
	if commandStrings[1] == "disable" || commandStrings[1] == "0" || commandStrings[1] == "false" {
		newEpisodes = entities.Cha{}
	} else {
		channelID, channelName := common.ChannelParser(s, commandStrings[1], m.GuildID)
		newEpisodes = entities.NewCha(channelName, channelID, "")
	}

	// Write
	db.SetGuildAutopost(m.GuildID, "newepisodes", newEpisodes)

	if newEpisodes == (entities.Cha{}) {
		_, err := s.ChannelMessageSend(m.ChannelID, "Success! Autopost for new airing anime episodes has been disabled! If this was not intentional please verify the channel ID.")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	entities.SetupGuildSub(m.GuildID)
	err := entities.AnimeSubsWrite(entities.SharedInfo.GetAnimeSubsMap())
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New Autopost channel for new airing anime episodes is: `%s - %s`", newEpisodes.GetName(), newEpisodes.GetID()))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// DonghuaCommandHandler handles donghua disable or enable
func DonghuaCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	var (
		message string
		mode    bool
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	// Displays current mode setting
	if len(commandStrings) == 1 {
		if !guildSettings.GetDonghua() {
			message = fmt.Sprintf("Donghuas are disabled. Please use `%sdonghua true` to enable it.", guildSettings.GetPrefix())
		} else {
			message = fmt.Sprintf("Donghuas are enabled. Please use `%sdonghua false` to disable it.", guildSettings.GetPrefix())
		}

		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	} else if len(commandStrings) > 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sdonghua [true/false]`", guildSettings.GetPrefix()))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parses bool
	if commandStrings[1] == "true" ||
		commandStrings[1] == "1" ||
		commandStrings[1] == "enable" {
		mode = true
		message = "Success! Donghuas were enabled."
	} else if commandStrings[1] == "false" ||
		commandStrings[1] == "0" ||
		commandStrings[1] == "disable" {
		mode = false
		message = "Success! Donghuas were disabled."
	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: That is not a valid value. Please use `true` or `false`.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and writes mode bool to guild
	guildSettings = guildSettings.SetDonghua(mode)
	db.SetGuildSettings(m.GuildID, guildSettings)

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// toggleDonghuaCommand handles donghua disable or enable
func toggleDonghuaCommand(print bool, enabled bool, guildID string) string {
	guildSettings := db.GetGuildSettings(guildID)

	// Displays current mode setting
	if print {
		if guildSettings.GetDonghua() {
			return "Donghuas are enabled."
		} else {
			return "Donghuas are disabled."
		}
	}

	// Changes and writes enabled bool to guild
	guildSettings = guildSettings.SetDonghua(enabled)
	db.SetGuildSettings(guildID, guildSettings)

	if !enabled {
		return "Success! Donghuas were disabled."
	}

	return "Success! Donghuas were enabled."
}

func init() {
	Add(&Command{
		Execute:    setDailyScheduleCommandHandler,
		Name:       "daily-schedule",
		Aliases:    []string{"dailyschedul", "dayschedule", "dayschedul", "setdailyschedule", "setdailyschedul", "dailyschedule"},
		Desc:       "Sets the autopost channel for daily anime schedule",
		Permission: functionality.Mod,
		Module:     "autopost",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "The channel in which you want to set the daily schedule in.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "enabled",
				Description: "Whether the daily schedule autopost should be enabled or disabled.",
				Required:    false,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			err := VerifySlashCommand(s, "daily-schedule", i)
			if err != nil {
				errStr := err.Error()
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &errStr,
				})
				return
			}

			var targetChannel *discordgo.Channel
			enabled := true
			if i.ApplicationCommandData().Options != nil {
				for _, option := range i.ApplicationCommandData().Options {
					if option.Name == "channel" {
						targetChannel = option.ChannelValue(s)
					} else if option.Name == "enabled" {
						enabled = option.BoolValue()
					}
				}
			}

			respStr := setDailyScheduleCommand(targetChannel, enabled, i.GuildID)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &respStr,
			})
		},
	})
	Add(&Command{
		Execute:    setNewEpisodesCommandHandler,
		Name:       "new-episodes",
		Aliases:    []string{"newepisode", "newepisod", "episodes", "episode", "newepisodes"},
		Desc:       "Sets the autopost channel for new airing anime episodes",
		Permission: functionality.Mod,
		Module:     "autopost",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "The channel in which you want to set the new anime episodes in.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "enabled",
				Description: "Whether the new anime episodes autopost should be enabled or disabled.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        "role",
				Description: "Which role to ping every time an episode is released.",
				Required:    false,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			err := VerifySlashCommand(s, "new-episodes", i)
			if err != nil {
				errStr := err.Error()
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &errStr,
				})
				return
			}

			var (
				targetChannel *discordgo.Channel
				role          *discordgo.Role
				enabled       = true
			)

			if i.ApplicationCommandData().Options != nil {
				for _, option := range i.ApplicationCommandData().Options {
					if option.Name == "channel" {
						targetChannel = option.ChannelValue(s)
					} else if option.Name == "enabled" {
						enabled = option.BoolValue()
					} else if option.Name == "role" {
						role = option.RoleValue(s, i.GuildID)
					}
				}
			}

			respStr := setNewEpisodesCommand(targetChannel, enabled, role, i.GuildID)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &respStr,
			})
		},
	})
	Add(&Command{
		Execute:    DonghuaCommandHandler,
		Name:       "donghua",
		Aliases:    []string{"toggledonghua", "toggle-donghua"},
		Desc:       "Disable or enable whether donghua (Chinese Anime) should show in the anime autopost commands.",
		Permission: functionality.Mod,
		Module:     "autopost",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "enabled",
				Description: "Whether to enable or disable donghua (chinese anime) from displaying in the anime autopost commands.",
				Required:    false,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			err := VerifySlashCommand(s, "donghua", i)
			if err != nil {
				errStr := err.Error()
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &errStr,
				})
				return
			}

			enabled := true
			printModule := true
			if i.ApplicationCommandData().Options != nil {
				for _, option := range i.ApplicationCommandData().Options {
					if option.Name == "enabled" {
						enabled = option.BoolValue()
						printModule = false
					}
				}
			}

			respStr := toggleDonghuaCommand(printModule, enabled, i.GuildID)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &respStr,
			})
		},
	})
}
