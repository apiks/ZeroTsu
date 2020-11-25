package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/events"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/functionality"
)

// Handles playing message view or change
func playingMsgCommand(s *discordgo.Session, m *discordgo.Message) {
	if m.Author.ID != config.OwnerID {
		return
	}

	var (
		err           error
		guildSettings = entities.GuildSettings{Prefix: "."}
	)

	if m.GuildID != "" {
		guildSettings = db.GetGuildSettings(m.GuildID)
	}

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	// Displays current playing message if it's only that
	if len(commandStrings) == 1 {
		var playingMsgs string
		entities.Mutex.RLock()
		for _, msg := range config.PlayingMsg {
			playingMsgs += fmt.Sprintf("\n`%v`,", msg)
		}
		entities.Mutex.RUnlock()
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current playing messages are: %s \n\nTo add more messages please use `%splayingmsg [new message]`", playingMsgs, guildSettings.GetPrefix()))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and writes new playing message to storage
	entities.Mutex.Lock()
	config.PlayingMsg = append(config.PlayingMsg, commandStrings[1])
	err = config.WriteConfig()
	if err != nil {
		entities.Mutex.Unlock()
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	entities.Mutex.Unlock()

	// Refreshes playing message
	err = s.UpdateStatus(0, commandStrings[1])
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New playing message added is: `%s`", commandStrings[1]))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Handles removing a playing message
func removePlayingMsgCommand(s *discordgo.Session, m *discordgo.Message) {
	if m.Author.ID != config.OwnerID {
		return
	}

	var (
		err           error
		guildSettings = entities.GuildSettings{Prefix: "."}
	)

	if m.GuildID != "" {
		guildSettings = db.GetGuildSettings(m.GuildID)
	}

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sremoveplayingmsg [msg]`", guildSettings.GetPrefix()))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and removes playing msg from storage
	entities.Mutex.Lock()
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
		entities.Mutex.Unlock()
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such playing message.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	config.PlayingMsg = append(config.PlayingMsg[:index], config.PlayingMsg[index+1:]...)
	err = config.WriteConfig()
	if err != nil {
		entities.Mutex.Unlock()
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Removed playing message: `%s`", commandStrings[1]))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
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
		entities.Mutex.Unlock()
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	entities.Mutex.Unlock()
}

// Prints in how many servers the BOT is
func serversCommand(s *discordgo.Session, m *discordgo.Message) {
	if m.Author.ID != config.OwnerID {
		return
	}

	events.GuildIds.RLock()
	defer events.GuildIds.RUnlock()
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("I am in %d servers.", len(events.GuildIds.Ids)))
	if err != nil {
		if m.GuildID != "" {
			guildSettings := db.GetGuildSettings(m.GuildID)
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("I've been online for %s.", common.Uptime().Truncate(time.Second).String()))
	if err != nil {
		if m.GuildID != "" {
			guildSettings := db.GetGuildSettings(m.GuildID)
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		}
	}
}

// Messages all BOT logs with a message
func messageBotLogsCommand(s *discordgo.Session, m *discordgo.Message) {
	if m.Author.ID != config.OwnerID {
		return
	}
	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	for guildID, guild := range entities.Guilds.DB {
		guild.RLock()
		if db.GetGuildSettings(guildID).GetBotLog() == (entities.Cha{}) ||
			db.GetGuildSettings(guildID).GetBotLog().GetName() == "" ||
			db.GetGuildSettings(guildID).GetBotLog().GetID() == "" {
			continue
		}
		_, _ = s.ChannelMessageSend(db.GetGuildSettings(guildID).GetBotLog().GetID(), commandStrings[1])
		guild.RUnlock()
		time.Sleep(1 * time.Second)
	}
}

func init() {
	Add(&Command{
		Execute:    playingMsgCommand,
		Trigger:    "playingmsg",
		Desc:       "Prints or adds a BOT playing message",
		DMAble:     true,
		Permission: functionality.Owner,
	})
	Add(&Command{
		Execute:    removePlayingMsgCommand,
		Trigger:    "removeplayingmsg",
		Aliases:    []string{"killplayingmsg"},
		Desc:       "Removes a BOT playing message",
		DMAble:     true,
		Permission: functionality.Owner,
	})
	Add(&Command{
		Execute:    serversCommand,
		Trigger:    "servers",
		Desc:       "Prints the number of servers the BOT is in",
		DMAble:     true,
		Permission: functionality.Owner,
	})
	Add(&Command{
		Execute:    uptimeCommand,
		Trigger:    "uptime",
		Desc:       "Print how long I've been on for",
		DMAble:     true,
		Permission: functionality.Owner,
	})
	Add(&Command{
		Execute:    messageBotLogsCommand,
		Trigger:    "messagelogs",
		Desc:       "Messages all BOT logs with a message",
		DMAble:     true,
		Permission: functionality.Owner,
	})
}
