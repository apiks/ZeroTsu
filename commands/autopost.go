package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Sets a channel ID as the autopost daily stats target channel
func setDailyStatsCommand(s *discordgo.Session, m *discordgo.Message) {

	var guildDailyStats *functionality.Cha

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	if dailyStats, ok := functionality.GuildMap[m.GuildID].Autoposts["dailystats"]; ok {
		guildDailyStats = dailyStats
	}
	functionality.MapMutex.Unlock()

	if guildDailyStats == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: Autopost Daily Stats channel is currently not set. Please use `%sdailystats [channel]`\nTo disable it please use `%sdailystats disable`", guildSettings.Prefix, guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Displays current dailystats channel
	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current Autopost Daily Stats channel is: `%s - %s` \n\nTo change it please use `%sdailystats [channel]`\nTo disable it please use `%sdailystats disable`", guildDailyStats.Name, guildDailyStats.ID, guildSettings.Prefix, guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sdailystats [channel]`", guildSettings.Prefix))
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parse and save the target channel
	if commandStrings[1] == "disable" {
		guildDailyStats = nil
	} else {
		channelID, channelName := functionality.ChannelParser(s, commandStrings[1], m.GuildID)
		guildDailyStats.ID = channelID
		guildDailyStats.Name = channelName
	}
	functionality.MapMutex.Lock()
	functionality.GuildMap[m.GuildID].Autoposts["dailystats"] = guildDailyStats
	_ = functionality.AutopostsWrite(functionality.GuildMap[m.GuildID].Autoposts, m.GuildID)
	functionality.MapMutex.Unlock()

	if guildDailyStats == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Success! Autopost Daily Stats has been disabled!")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New Autopost Daily Stats channel is: `%s - %s`", guildDailyStats.Name, guildDailyStats.ID))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Sets a channel ID as the autopost anime schedule target channel
func setDailyScheduleCommand(s *discordgo.Session, m *discordgo.Message) {

	var guildDailySchedule *functionality.Cha

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	if dailySchedule, ok := functionality.GuildMap[m.GuildID].Autoposts["dailyschedule"]; ok {
		guildDailySchedule = dailySchedule
	}
	functionality.MapMutex.Unlock()

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Displays current dailyschedule channel
	if len(commandStrings) == 1 {
		if guildDailySchedule == nil {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: Autopost Daily Anime Schedule channel is currently not set. Please use `%sdailyschedule [channel]`", guildSettings.Prefix))
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current Autopost Daily Anime Schedule channel is: `%s - %s` \n\nTo change it please use `%sdailyschedule [channel]`\nTo disable it please use `%sdailyschedule disable`", guildDailySchedule.Name, guildDailySchedule.ID, guildSettings.Prefix, guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sdailyschedule [channel]`\nTo disable it please use `%sdailyschedule disable`", guildSettings.Prefix, guildSettings.Prefix))
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	if guildDailySchedule == nil {
		guildDailySchedule = new(functionality.Cha)
	}

	// Parse and save the target channel
	if commandStrings[1] == "disable" {
		guildDailySchedule = nil
	} else {
		channelID, channelName := functionality.ChannelParser(s, commandStrings[1], m.GuildID)
		guildDailySchedule.ID = channelID
		guildDailySchedule.Name = channelName
	}
	functionality.MapMutex.Lock()
	functionality.GuildMap[m.GuildID].Autoposts["dailyschedule"] = guildDailySchedule
	_ = functionality.AutopostsWrite(functionality.GuildMap[m.GuildID].Autoposts, m.GuildID)
	functionality.MapMutex.Unlock()

	if guildDailySchedule == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Success! Autopost Daily Anime Schedule has been disabled!")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New Autopost Daily Anime Schedule channel is: `%s - %s`", guildDailySchedule.Name, guildDailySchedule.ID))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Sets a channel ID as the autopost new airing anime episodes target channel
func setNewEpisodesCommand(s *discordgo.Session, m *discordgo.Message) {

	var guildNewEpisodes *functionality.Cha

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	if newEpisodes, ok := functionality.GuildMap[m.GuildID].Autoposts["newepisodes"]; ok {
		guildNewEpisodes = newEpisodes
	}
	functionality.MapMutex.Unlock()

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Displays current new episodes channel
	if len(commandStrings) == 1 {
		if guildNewEpisodes == nil {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: Autopost channel for new airing anime episodes is currently not set. Please use `%snewepisodes [channel]`", guildSettings.Prefix))
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current Autopost channel for new airing anime episodes is: `%s - %s` \n\n To change it please use `%snewepisodes [channel]`\nTo disable it please use `%snewepisodes disable`", guildNewEpisodes.Name, guildNewEpisodes.ID, guildSettings.Prefix, guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%snewepisodes [channel]`\nTo disable it please use `%snewepisodes disable`", guildSettings.Prefix, guildSettings.Prefix))
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	if guildNewEpisodes == nil {
		guildNewEpisodes = new(functionality.Cha)
	}

	// Parse and save the target channel
	if commandStrings[1] == "disable" {
		guildNewEpisodes = nil
	} else {
		channelID, channelName := functionality.ChannelParser(s, commandStrings[1], m.GuildID)
		guildNewEpisodes.ID = channelID
		guildNewEpisodes.Name = channelName
	}
	functionality.MapMutex.Lock()
	functionality.GuildMap[m.GuildID].Autoposts["newepisodes"] = guildNewEpisodes
	_ = functionality.AutopostsWrite(functionality.GuildMap[m.GuildID].Autoposts, m.GuildID)

	if guildNewEpisodes == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Success! Autopost for new airing anime episodes has been disabled!")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}

	functionality.SetupGuildSub(m.GuildID)
	functionality.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New Autopost channel for new airing anime episodes is: `%s - %s`", guildNewEpisodes.Name, guildNewEpisodes.ID))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    setDailyStatsCommand,
		Trigger:    "dailystats",
		Aliases:    []string{"dailystat", "daystats", "daystat", "setdailystats", "setdailystat", "setdaystats", "setdaystat"},
		Desc:       "Sets the autopost channel for daily stats",
		Permission: functionality.Mod,
		Module:     "autopost",
	})
	functionality.Add(&functionality.Command{
		Execute:    setDailyScheduleCommand,
		Trigger:    "dailyschedule",
		Aliases:    []string{"dailyschedul", "dayschedule", "dayschedul", "setdailyschedule", "setdailyschedul"},
		Desc:       "Sets the autopost channel for daily anime schedule",
		Permission: functionality.Mod,
		Module:     "autopost",
	})
	functionality.Add(&functionality.Command{
		Execute:    setNewEpisodesCommand,
		Trigger:    "newepisodes",
		Aliases:    []string{"newepisode", "newepisod", "episodes", "episode"},
		Desc:       "Sets the autopost channel for new airing anime episodes",
		Permission: functionality.Mod,
		Module:     "autopost",
	})
}
