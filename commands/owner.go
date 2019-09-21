package commands

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Handles playing message view or change
func playingMsgCommand(s *discordgo.Session, m *discordgo.Message) {

	if m.Author.ID != config.OwnerID {
		return
	}

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	commandStrings := strings.SplitN(m.Content, " ", 2)

	// Displays current playing message if it's only that
	if len(commandStrings) == 1 {
		var playingMsgs string
		misc.MapMutex.Lock()
		for _, msg := range config.PlayingMsg {
			playingMsgs += fmt.Sprintf("\n`%v`,", msg)
		}
		misc.MapMutex.Unlock()
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current playing messages are: %v \n\nTo add more messages please use `%vplayingmsg [new message]`", playingMsgs, guildPrefix))
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Changes and writes new playing message to storage
	misc.MapMutex.Lock()
	config.PlayingMsg = append(config.PlayingMsg, commandStrings[1])
	err := config.WriteConfig()
	if err != nil {
		misc.MapMutex.Unlock()
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
	misc.MapMutex.Unlock()

	// Refreshes playing message
	err = s.UpdateStatus(0, commandStrings[1])
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New playing message added is: `%v`", commandStrings[1]))
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Handles removing a playing message
func removePlayingMsgCommand(s *discordgo.Session, m *discordgo.Message) {

	if m.Author.ID != config.OwnerID {
		return
	}

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	commandStrings := strings.SplitN(m.Content, " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vremoveplayingmsg [msg]`", guildPrefix))
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Changes and removes playing msg from storage
	misc.MapMutex.Lock()
	var index int
	for i, msg := range config.PlayingMsg {
		if msg == commandStrings[1] {
			index = i
			break
		}
	}
	config.PlayingMsg = append(config.PlayingMsg[:index], config.PlayingMsg[index+1:]...)
	err := config.WriteConfig()
	if err != nil {
		misc.MapMutex.Unlock()
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Refreshes playing message
	randMsg := rand.Intn(len(config.PlayingMsg)-1)
	err = s.UpdateStatus(0, config.PlayingMsg[randMsg])
	if err != nil {
		misc.MapMutex.Unlock()
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
	misc.MapMutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Removed playing message: `%v`", commandStrings[1]))
	if err != nil {
		_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
	}
}

// Prints in how many servers the BOT is
func serversCommand(s *discordgo.Session, m *discordgo.Message) {

	if m.Author.ID != config.OwnerID {
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("I am in %v servers.", len(s.State.Guilds)))
	if err != nil {

		misc.MapMutex.Lock()
		guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
		misc.MapMutex.Unlock()

		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Prints how many users the BOT observes
func usersCommand(s *discordgo.Session, m *discordgo.Message) {

	if m.Author.ID != config.OwnerID {
		return
	}

	misc.MapMutex.Lock()
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("I can see %v users.", len(misc.UserCounter)))
	if err != nil {
		guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
		_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
	}
	misc.MapMutex.Unlock()
}

func init() {
	add(&command{
		execute: playingMsgCommand,
		trigger: "playingmsg",
		desc:    "Prints or adds a BOT playing message.",
		elevated: true,
		admin: true,
	})
	add(&command{
		execute: removePlayingMsgCommand,
		trigger: "removeplayingmsg",
		aliases:  []string{"killplayingmsg"},
		desc:    "Removes a BOT playing message.",
		elevated: true,
		admin: true,
	})
	add(&command{
		execute: serversCommand,
		trigger: "servers",
		desc:    "Prints the number of servers the BOT is in.",
		elevated: true,
		admin: true,
	})
	//add(&command{
	//	execute: usersCommand,
	//	trigger: "users",
	//	desc:    "Prints the number of users the BOT can see.",
	//	elevated: true,
	//	admin: true,
	//})
}
