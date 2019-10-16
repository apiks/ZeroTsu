package commands

import (
	"fmt"
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

	var guildSettings = &functionality.GuildSettings{
		Prefix: ".",
	}

	if m.GuildID != "" {
		functionality.MapMutex.Lock()
		*guildSettings = functionality.GuildMap[m.GuildID].GetGuildSettings()
		functionality.MapMutex.Unlock()
	}

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	// Displays current playing message if it's only that
	if len(commandStrings) == 1 {
		var playingMsgs string
		functionality.MapMutex.Lock()
		for _, msg := range config.PlayingMsg {
			playingMsgs += fmt.Sprintf("\n`%v`,", msg)
		}
		functionality.MapMutex.Unlock()
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current playing messages are: %v \n\nTo add more messages please use `%vplayingmsg [new message]`", playingMsgs, guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and writes new playing message to storage
	functionality.MapMutex.Lock()
	config.PlayingMsg = append(config.PlayingMsg, commandStrings[1])
	err := config.WriteConfig()
	if err != nil {
		functionality.MapMutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	functionality.MapMutex.Unlock()

	// Refreshes playing message
	err = s.UpdateStatus(0, commandStrings[1])
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New playing message added is: `%v`", commandStrings[1]))
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
		functionality.MapMutex.Lock()
		*guildSettings = functionality.GuildMap[m.GuildID].GetGuildSettings()
		functionality.MapMutex.Unlock()
	}

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vremoveplayingmsg [msg]`", guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and removes playing msg from storage
	functionality.MapMutex.Lock()
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
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such playing message.")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}
	config.PlayingMsg = append(config.PlayingMsg[:index], config.PlayingMsg[index+1:]...)
	err := config.WriteConfig()
	if err != nil {
		functionality.MapMutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Removed playing message: `%v`", commandStrings[1]))
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
		functionality.MapMutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	functionality.MapMutex.Unlock()
}

// Prints in how many servers the BOT is
func serversCommand(s *discordgo.Session, m *discordgo.Message) {

	if m.Author.ID != config.OwnerID {
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("I am in %v servers.", len(s.State.Guilds)))
	if err != nil {
		if m.GuildID != "" {
			functionality.MapMutex.Lock()
			guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
			functionality.MapMutex.Unlock()
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		}
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
			functionality.MapMutex.Lock()
			guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
			functionality.MapMutex.Unlock()
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		}
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
}
