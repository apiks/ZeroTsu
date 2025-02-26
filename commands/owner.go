package commands

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"

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
	err = s.UpdateGameStatus(0, commandStrings[1])
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
		err = s.UpdateGameStatus(0, config.PlayingMsg[randInt])
	} else if len(config.PlayingMsg) == 1 {
		err = s.UpdateGameStatus(0, config.PlayingMsg[0])
	} else {
		err = s.UpdateGameStatus(0, "")
	}
	if err != nil {
		entities.Mutex.Unlock()
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	entities.Mutex.Unlock()
}

// serversCommand prints in how many servers the BOT is
func serversCommand(s *discordgo.Session, m *discordgo.Message) {
	if m.Author.ID != config.OwnerID {
		return
	}

	// Fetch guild count from MongoDB
	guildIds, err := entities.LoadAllGuildIDs()
	if err != nil {
		log.Printf("Error fetching guild IDs: %v", err)
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error fetching server count.")
		return
	}

	// Send message with the guild count
	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("I am in %d servers.", len(guildIds)))
	if err != nil {
		if m.GuildID != "" {
			guildSettings := db.GetGuildSettings(m.GuildID)
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		}
	}
}

// Prints in how many servers the BOT is from the sharding manager
func serversShardCommand(s *discordgo.Session, m *discordgo.Message) {
	if m.Author.ID != config.OwnerID {
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("I am in %d servers.", config.Mgr.GuildCount()))
	if err != nil {
		guildSettings := db.GetGuildSettings(m.GuildID)
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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
	if len(commandStrings) < 2 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Usage: `messageBotLogs [message]`")
		return
	}

	// Load only guild IDs and bot log channel IDs from MongoDB
	guildLogs, err := entities.LoadGuildBotLogs()
	if err != nil {
		log.Println("Error loading bot logs from MongoDB:", err)
		return
	}

	for _, botLog := range guildLogs {
		if botLog == "" {
			continue
		}

		// Send message to the bot log channel
		_, _ = s.ChannelMessageSend(botLog, commandStrings[1])
		time.Sleep(1 * time.Second) // Prevent rate-limiting issues
	}
}

func init() {
	Add(&Command{
		Execute:    playingMsgCommand,
		Name:       "playingmsg",
		Desc:       "Prints or sets a BOT playing message",
		DMAble:     true,
		Permission: functionality.Owner,
	})
	Add(&Command{
		Execute:    removePlayingMsgCommand,
		Name:       "removeplayingmsg",
		Aliases:    []string{"killplayingmsg"},
		Desc:       "Removes a BOT playing message",
		DMAble:     true,
		Permission: functionality.Owner,
	})
	Add(&Command{
		Execute:    serversCommand,
		Name:       "servers",
		Desc:       "Prints the number of servers the BOT is in",
		DMAble:     true,
		Permission: functionality.Owner,
	})
	Add(&Command{
		Execute:    serversShardCommand,
		Name:       "serversshards",
		Desc:       "Prints the number of servers the BOT is in using the sharding manager",
		DMAble:     true,
		Permission: functionality.Owner,
	})
	Add(&Command{
		Execute:    uptimeCommand,
		Name:       "uptime",
		Desc:       "Print how long I've been on for",
		DMAble:     true,
		Permission: functionality.Owner,
	})
	Add(&Command{
		Execute:    messageBotLogsCommand,
		Name:       "messagelogs",
		Desc:       "Messages all BOT logs with a message",
		DMAble:     true,
		Permission: functionality.Owner,
	})
}
