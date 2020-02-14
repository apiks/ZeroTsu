package commands

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"ZeroTsu/config"
	"ZeroTsu/functionality"
)

// Handles playing message view or change
func playingMsgCommand(s *discordgo.Session, m *discordgo.Message) {

	if m.Author.ID != config.OwnerID {
		return
	}

	var guildSettings = &functionality.GuildSettings{
		Prefix: ".",
	}

	if m.GuildID != "" {
		functionality.Mutex.RLock()
		guildSettings = functionality.GuildMap[m.GuildID].GetGuildSettings()
		functionality.Mutex.RUnlock()
	}

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	// Displays current playing message if it's only that
	if len(commandStrings) == 1 {
		var playingMsgs string
		functionality.Mutex.RLock()
		for _, msg := range config.PlayingMsg {
			playingMsgs += fmt.Sprintf("\n`%v`,", msg)
		}
		functionality.Mutex.RUnlock()
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current playing messages are: %s \n\nTo add more messages please use `%splayingmsg [new message]`", playingMsgs, guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and writes new playing message to storage
	functionality.Mutex.Lock()
	config.PlayingMsg = append(config.PlayingMsg, commandStrings[1])
	err := config.WriteConfig()
	if err != nil {
		functionality.Mutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	functionality.Mutex.Unlock()

	// Refreshes playing message
	err = s.UpdateStatus(0, commandStrings[1])
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New playing message added is: `%s`", commandStrings[1]))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Handles removing a playing message
func removePlayingMsgCommand(s *discordgo.Session, m *discordgo.Message) {

	if m.Author.ID != config.OwnerID {
		return
	}

	var guildSettings = &functionality.GuildSettings{
		Prefix: ".",
	}

	if m.GuildID != "" {
		functionality.Mutex.RLock()
		guildSettings = functionality.GuildMap[m.GuildID].GetGuildSettings()
		functionality.Mutex.RUnlock()
	}

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sremoveplayingmsg [msg]`", guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and removes playing msg from storage
	functionality.Mutex.Lock()
	var index int
	var foundIndex bool
	for i, msg := range config.PlayingMsg {
		if msg == commandStrings[1] {
			index = i
			foundIndex = true
			break
		}
	}
	if !foundIndex {
		functionality.Mutex.Unlock()
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such playing message.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	config.PlayingMsg = append(config.PlayingMsg[:index], config.PlayingMsg[index+1:]...)
	err := config.WriteConfig()
	if err != nil {
		functionality.Mutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Removed playing message: `%s`", commandStrings[1]))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
	}

	// Updates playing status
	if len(config.PlayingMsg) > 1 {
		rand.Seed(time.Now().UnixNano())
		randInt := rand.Intn(len(config.PlayingMsg))
		err = s.UpdateStatus(0, config.PlayingMsg[randInt])
	} else if len(config.PlayingMsg) == 1 {
		err = s.UpdateStatus(0, config.PlayingMsg[0])
	} else {
		err = s.UpdateStatus(0, "")
	}
	if err != nil {
		functionality.Mutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	functionality.Mutex.Unlock()
}

// Prints in how many servers the BOT is
func serversCommand(s *discordgo.Session, m *discordgo.Message) {

	if m.Author.ID != config.OwnerID {
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("I am in %d servers.", len(s.State.Guilds)))
	if err != nil {
		if m.GuildID != "" {
			functionality.Mutex.RLock()
			guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
			functionality.Mutex.RUnlock()
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}
}

// Prints BOT uptime
func uptimeCommand(s *discordgo.Session, m *discordgo.Message) {
	if m.Author.ID != config.OwnerID {
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("I've been online for %s.", functionality.Uptime().Truncate(time.Second).String()))
	if err != nil {
		if m.GuildID != "" {
			functionality.Mutex.RLock()
			guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
			functionality.Mutex.RUnlock()
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}
}

func flushCommand(s *discordgo.Session, m *discordgo.Message) {
	if m.Author.ID != config.OwnerID {
		return
	}

	guilds := s.State.Guilds
	functionality.Mutex.Lock()
	for _, guild := range guilds {
		_ = functionality.WriteMemberInfo(functionality.GuildMap[guild.ID].MemberInfoMap, guild.ID)
		_ = functionality.EmojiStatsWrite(functionality.GuildMap[guild.ID].EmojiStats, guild.ID)
		_, _ = functionality.ChannelStatsWrite(functionality.GuildMap[guild.ID].ChannelStats, guild.ID)
		_, _ = functionality.UserChangeStatsWrite(functionality.GuildMap[guild.ID].UserChangeStats, guild.ID)
		_ = functionality.VerifiedStatsWrite(functionality.GuildMap[guild.ID].VerifiedStats, guild.ID)
		_ = functionality.VoteInfoWrite(functionality.GuildMap[guild.ID].VoteInfoMap, guild.ID)
		_ = functionality.TempChaWrite(functionality.GuildMap[guild.ID].TempChaMap, guild.ID)
		_ = functionality.ReactJoinWrite(functionality.GuildMap[guild.ID].ReactJoinMap, guild.ID)
		_ = functionality.RafflesWrite(functionality.GuildMap[guild.ID].Raffles, guild.ID)
		_ = functionality.WaifusWrite(functionality.GuildMap[guild.ID].Waifus, guild.ID)
		_ = functionality.WaifuTradesWrite(functionality.GuildMap[guild.ID].WaifuTrades, guild.ID)
		_ = functionality.AutopostsWrite(functionality.GuildMap[guild.ID].Autoposts, guild.ID)
		_ = functionality.PunishedUsersWrite(functionality.GuildMap[guild.ID].PunishedUsers, guild.ID)
		_ = functionality.GuildSettingsWrite(functionality.GuildMap[guild.ID].GuildConfig, guild.ID)
	}
	functionality.Mutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, "Flushed to storage successfuly!")
	if err != nil {
		if m.GuildID != "" {
			functionality.Mutex.RLock()
			guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
			functionality.Mutex.RUnlock()
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    playingMsgCommand,
		Trigger:    "playingmsg",
		Desc:       "Prints or adds a BOT playing message",
		DMAble:     true,
		Permission: functionality.Owner,
	})
	functionality.Add(&functionality.Command{
		Execute:    removePlayingMsgCommand,
		Trigger:    "removeplayingmsg",
		Aliases:    []string{"killplayingmsg"},
		Desc:       "Removes a BOT playing message",
		DMAble:     true,
		Permission: functionality.Owner,
	})
	functionality.Add(&functionality.Command{
		Execute:    serversCommand,
		Trigger:    "servers",
		Desc:       "Prints the number of servers the BOT is in",
		DMAble:     true,
		Permission: functionality.Owner,
	})
	functionality.Add(&functionality.Command{
		Execute:    uptimeCommand,
		Trigger:    "uptime",
		Desc:       "Print how long I've been on for",
		DMAble:     true,
		Permission: functionality.Owner,
	})
	functionality.Add(&functionality.Command{
		Execute:    flushCommand,
		Trigger:    "flush",
		Desc:       "Write everything in memory to disk",
		DMAble:     true,
		Permission: functionality.Owner,
	})
}
