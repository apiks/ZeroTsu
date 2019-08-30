package commands

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

// Rolls a number between 1 and a specified number (defaults to 100)
func rollCommand(s *discordgo.Session, m *discordgo.Message) {

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	misc.MapMutex.Unlock()

	commandStrings := strings.Split(m.Content, " ")

	// Rolls a number between 1 and 100 if only the command is used
	if len(commandStrings) == 1 {
		randomNum := rand.Intn(99)+1
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**Rolled:** %v", randomNum))
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
		return
	}

	// Prints error if too many parameters
	if len(commandStrings) > 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vroll [number]`", guildPrefix))
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
		return
	}

	// Checks if the specified number is actually a number
	num, err := strconv.Atoi(commandStrings[1])
	if err != nil {

		// Handles different error types
		if err.(*strconv.NumError).Err == strconv.ErrRange {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: That number is too large. Please try a smaller one.")
			if err != nil {
				misc.CommandErrorHandler(s, m, err, guildBotLog)
				return
			}
			return
		}
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: That is not a valid number.")
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
		return
	}

	// Makes sure it's not a number outside of boundaries
	if num < 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid number. Please use a positive number.")
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
		return
	}

	// Rolls a specified number
	randomNum := rand.Intn(num)+1
	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**Rolled:** %v", randomNum))
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
}

func init() {
	add(&command{
		execute:  rollCommand,
		trigger:  "roll",
		aliases:  []string{"rol", "r"},
		desc:     "Rolls a number from 1 to 100. Specify a number to change the range.",
		category: "normal",
	})
}
