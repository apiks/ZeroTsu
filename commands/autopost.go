package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/misc"
	"strings"
)

// Sets a channel ID as the autopost daily stats target channel
func setDailyStatsCommand(s *discordgo.Session, m *discordgo.Message) {

	var guildDailyStats *misc.Cha

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	if dailyStats, ok := misc.GuildMap[m.GuildID].Autoposts["dailystats"]; ok {
		guildDailyStats = dailyStats
	}
	misc.MapMutex.Unlock()

	if guildDailyStats == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: Autopost Daily Stats channel is currently not set. Please use `%vdailystats [channel]`\nTo disable it please use `%vdailystats disable`", guildPrefix, guildPrefix))
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error())
			if err != nil {
				return
			}
			return
		}
		return
	}

	commandStrings := strings.Split(strings.ToLower(m.Content), " ")

	// Displays current dailystats channel
	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current Autopost Daily Stats channel is: `%v - %v` \n\nTo change it please use `%vdailystats [channel]`\nTo disable it please use `%vdailystats disable`", guildDailyStats.Name, guildDailyStats.ID, guildPrefix, guildPrefix))
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error())
			if err != nil {
				return
			}
			return
		}
		return
	}
	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vdailystats [channel]`", guildPrefix))
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
		return
	}

	// Parse and save the target channel
	if commandStrings[1] == "disable" {
		guildDailyStats = nil
	} else {
		channelID, channelName := misc.ChannelParser(s, commandStrings[1], m.GuildID)
		guildDailyStats.ID = channelID
		guildDailyStats.Name = channelName
	}
	misc.MapMutex.Lock()
	misc.GuildMap[m.GuildID].Autoposts["dailystats"] = guildDailyStats
	_ = misc.AutopostsWrite(misc.GuildMap[m.GuildID].Autoposts, m.GuildID)
	misc.MapMutex.Unlock()

	if guildDailyStats == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Success! Autopost Daily Stats has been disabled!")
		if err != nil {
			_, _ = s.ChannelMessageSend(guildBotLog, err.Error())
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New Autopost Daily Stats channel is: `%v - %v`", guildDailyStats.Name, guildDailyStats.ID))
	if err != nil {
		_, _ = s.ChannelMessageSend(guildBotLog, err.Error())
	}
}

// Sets a channel ID as the autopost anime schedule target channel
func setDailyScheduleCommand(s *discordgo.Session, m *discordgo.Message) {

	var guildDailySchedule *misc.Cha

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	if dailySchedule, ok := misc.GuildMap[m.GuildID].Autoposts["dailyschedule"]; ok {
		guildDailySchedule = dailySchedule
	}
	misc.MapMutex.Unlock()

	commandStrings := strings.Split(strings.ToLower(m.Content), " ")

	// Displays current dailyschedule channel
	if len(commandStrings) == 1 {
		if guildDailySchedule == nil {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: Autopost Daily Anime Schedule channel is currently not set. Please use `%vdailyschedule [channel]`", guildPrefix))
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error())
				if err != nil {
					return
				}
				return
			}
			return
		}
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current Autopost Daily Anime Schedule channel is: `%v - %v` \n\nTo change it please use `%vdailyschedule [channel]`\nTo disable it please use `%vdailyschedule disable`", guildDailySchedule.Name, guildDailySchedule.ID, guildPrefix, guildPrefix))
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error())
			if err != nil {
				return
			}
			return
		}
		return
	}
	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vdailyschedule [channel]`\nTo disable it please use `%vdailyschedule disable`", guildPrefix, guildPrefix))
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
		return
	}

	if guildDailySchedule == nil {
		guildDailySchedule = new(misc.Cha)
	}

	// Parse and save the target channel
	if commandStrings[1] == "disable" {
		guildDailySchedule = nil
	} else {
		channelID, channelName := misc.ChannelParser(s, commandStrings[1], m.GuildID)
		guildDailySchedule.ID = channelID
		guildDailySchedule.Name = channelName
	}
	misc.MapMutex.Lock()
	misc.GuildMap[m.GuildID].Autoposts["dailyschedule"] = guildDailySchedule
	_ = misc.AutopostsWrite(misc.GuildMap[m.GuildID].Autoposts, m.GuildID)
	misc.MapMutex.Unlock()

	if guildDailySchedule == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Success! Autopost Daily Anime Schedule has been disabled!")
		if err != nil {
			_, _ = s.ChannelMessageSend(guildBotLog, err.Error())
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New Autopost Daily Anime Schedule channel is: `%v - %v`", guildDailySchedule.Name, guildDailySchedule.ID))
	if err != nil {
		_, _ = s.ChannelMessageSend(guildBotLog, err.Error())
	}
}

// Sets a channel ID as the autopost new airing anime episodes target channel
func setNewEpisodesCommand(s *discordgo.Session, m *discordgo.Message) {

	var guildNewEpisodes *misc.Cha

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	if newEpisodes, ok := misc.GuildMap[m.GuildID].Autoposts["newepisodes"]; ok {
		guildNewEpisodes = newEpisodes
	}
	misc.MapMutex.Unlock()

	commandStrings := strings.Split(strings.ToLower(m.Content), " ")

	// Displays current new episodes channel
	if len(commandStrings) == 1 {
		if guildNewEpisodes == nil {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: Autopost channel for new airing anime episodes is currently not set. Please use `%vnewepisodes [channel]`", guildPrefix))
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error())
				if err != nil {
					return
				}
				return
			}
			return
		}
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current Autopost channel for new airing anime episodes is: `%v - %v` \n\n To change it please use `%vnewepisodes [channel]`\nTo disable it please use `%vnewepisodes disable`", guildNewEpisodes.Name, guildNewEpisodes.ID, guildPrefix, guildPrefix))
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error())
			if err != nil {
				return
			}
			return
		}
		return
	}
	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vnewepisodes [channel]`\nTo disable it please use `%vnewepisodes disable`", guildPrefix, guildPrefix))
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
		return
	}

	if guildNewEpisodes == nil {
		guildNewEpisodes = new(misc.Cha)
	}

	// Parse and save the target channel
	if commandStrings[1] == "disable" {
		guildNewEpisodes = nil
	} else {
		channelID, channelName := misc.ChannelParser(s, commandStrings[1], m.GuildID)
		guildNewEpisodes.ID = channelID
		guildNewEpisodes.Name = channelName
	}
	misc.MapMutex.Lock()
	misc.GuildMap[m.GuildID].Autoposts["newepisodes"] = guildNewEpisodes
	_ = misc.AutopostsWrite(misc.GuildMap[m.GuildID].Autoposts, m.GuildID)
	misc.MapMutex.Unlock()

	if guildNewEpisodes == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Success! Autopost for new airing anime episodes has been disabled!")
		if err != nil {
			_, _ = s.ChannelMessageSend(guildBotLog, err.Error())
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New Autopost channel for new airing anime episodes is: `%v - %v`", guildNewEpisodes.Name, guildNewEpisodes.ID))
	if err != nil {
		_, _ = s.ChannelMessageSend(guildBotLog, err.Error())
	}
}

func init() {
	add(&command{
		execute:  setDailyStatsCommand,
		trigger:  "dailystats",
		aliases:  []string{"dailystat", "daystats", "daystat", "setdailystats", "setdailystat", "setdaystats", "setdaystat"},
		desc:     "Sets the autopost channel for daily stats",
		elevated: true,
		category: "autopost",
	})
	add(&command{
		execute:  setDailyScheduleCommand,
		trigger:  "dailyschedule",
		aliases:  []string{"dailyschedul", "dayschedule", "dayschedul", "setdailyschedule", "setdailyschedul"},
		desc:     "Sets the autopost channel for daily anime schedule",
		elevated: true,
		category: "autopost",
	})
	add(&command{
		execute:  setNewEpisodesCommand,
		trigger:  "newepisodes",
		aliases:  []string{"newepisode", "newepisod", "episodes", "episode"},
		desc:     "Sets the autopost channel for new airing anime episodes",
		elevated: true,
		category: "autopost",
	})
}
