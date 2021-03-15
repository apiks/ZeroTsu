package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Sets a channel ID as the autopost anime schedule target channel
func setDailyScheduleCommand(s *discordgo.Session, m *discordgo.Message) {
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
		dailySchedule = entities.NewCha("", "")
	}

	// Parse and save the target channel
	if commandStrings[1] == "disable" || commandStrings[1] == "0" || commandStrings[1] == "false" {
		dailySchedule = entities.Cha{}
	} else {
		channelID, channelName := common.ChannelParser(s, commandStrings[1], m.GuildID)
		dailySchedule = entities.NewCha(channelName, channelID)
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

// Sets a channel ID as the autopost new airing anime episodes target channel
func setNewEpisodesCommand(s *discordgo.Session, m *discordgo.Message) {
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
		newEpisodes = entities.NewCha("", "")
	}

	// Parse and save the target channel
	if commandStrings[1] == "disable" || commandStrings[1] == "0" || commandStrings[1] == "false" {
		newEpisodes = entities.Cha{}
	} else {
		channelID, channelName := common.ChannelParser(s, commandStrings[1], m.GuildID)
		newEpisodes = entities.NewCha(channelName, channelID)
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

	entities.Mutex.Lock()
	entities.SharedInfo.Lock()
	entities.AnimeSchedule.RLock()
	entities.SetupGuildSub(m.GuildID)
	entities.AnimeSchedule.RUnlock()
	entities.SharedInfo.Unlock()
	entities.Mutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New Autopost channel for new airing anime episodes is: `%s - %s`", newEpisodes.GetName(), newEpisodes.GetID()))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

func init() {
	Add(&Command{
		Execute:    setDailyScheduleCommand,
		Trigger:    "dailyschedule",
		Aliases:    []string{"dailyschedul", "dayschedule", "dayschedul", "setdailyschedule", "setdailyschedul"},
		Desc:       "Sets the autopost channel for daily anime schedule",
		Permission: functionality.Mod,
		Module:     "autopost",
	})
	Add(&Command{
		Execute:    setNewEpisodesCommand,
		Trigger:    "newepisodes",
		Aliases:    []string{"newepisode", "newepisod", "episodes", "episode"},
		Desc:       "Sets the autopost channel for new airing anime episodes",
		Permission: functionality.Mod,
		Module:     "autopost",
	})
}
