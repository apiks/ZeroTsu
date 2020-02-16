package commands

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Rolls a number between 1 and a specified number (defaults to 100)
func rollCommand(s *discordgo.Session, m *discordgo.Message) {

	var guildSettings = &functionality.GuildSettings{
		Prefix: ".",
	}

	if m.GuildID != "" {
		functionality.Mutex.RLock()
		guildSettings = functionality.GuildMap[m.GuildID].GetGuildSettings()
		functionality.Mutex.RUnlock()
	}

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	// Rolls a number between 1 and 100 if only the command is used
	if len(commandStrings) == 1 {
		randomNum := rand.Intn(99) + 1
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**Rolled:** %d", randomNum))
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Prints error if too many parameters
	if len(commandStrings) > 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sroll [number]`", guildSettings.Prefix))
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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
				functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
				return
			}
			return
		}
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: That is not a valid number.")
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Makes sure it's not a number outside of boundaries
	if num < 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid number. Please use a positive number.")
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Rolls a specified number
	randomNum := rand.Intn(num) + 1
	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**Rolled:** %v", randomNum))
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute: rollCommand,
		Trigger: "roll",
		Aliases: []string{"rol", "r"},
		Desc:    "Rolls a number from 1 to 100. Specify a positive number to change the range",
		Module:  "normal",
		DMAble:  true,
	})
}
